package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"order-service/internal/cache"
	"order-service/internal/config"
	"order-service/internal/database"
	"order-service/internal/handler"
	"order-service/internal/middleware"
	"order-service/internal/models"
	"order-service/internal/repository"
	"order-service/internal/service"
	"order-service/internal/telemetry"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize OpenTelemetry
	shutdown, err := telemetry.InitTracer(cfg.ServiceName, cfg.OTELEndpoint, cfg.Environment)
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			log.Printf("Failed to shutdown tracer: %v", err)
		}
	}()

	// Connect to database
	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate database schema
	if err := db.AutoMigrate(&models.Order{}, &models.OrderItem{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Connect to Redis
	redisCache, err := cache.NewRedisCache(cfg.RedisURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Println("Continuing without cache...")
	}

	// Initialize layers
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, redisCache, cfg.ProductServiceURL, cfg.InventoryServiceURL, cfg.PaymentServiceURL)
	orderHandler := handler.NewOrderHandler(orderService)
	middleware.InitUserClient(cfg.UserServiceURL)

	// Setup Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())
	router.Use(otelgin.Middleware(cfg.ServiceName))

	// Health check endpoints
	router.GET("/health", healthCheck)
	router.GET("/ready", readyCheck(db))

	// API routes
	api := router.Group("/api")
	{
		orders := api.Group("/orders")
		orders.Use(middleware.ExtractToken()) // Protect all order endpoints
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("", orderHandler.GetAllOrders)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.PATCH("/:id/status", orderHandler.UpdateOrderStatus)
			orders.POST("/:id/cancel", orderHandler.CancelOrder)
		}

		users := api.Group("/users")
		users.Use(middleware.ExtractToken()) // Protect user endpoints
		{
			users.GET("/:user_id/orders", orderHandler.GetUserOrders)
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("User Service URL: %s", cfg.UserServiceURL)
		log.Printf("Inventory Service URL: %s", cfg.InventoryServiceURL)
		log.Printf("Payment Service URL: %s", cfg.PaymentServiceURL)
		log.Printf("Starting Order Service on port %s", cfg.ServerPort)
		log.Printf("Environment: %s", cfg.Environment)
		log.Printf("Product Service URL: %s", cfg.ProductServiceURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Order service is healthy",
		"data": gin.H{
			"service": "order-service",
			"version": "1.0.0",
		},
	})
}

func readyCheck(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"message": "Database connection failed",
			})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"message": "Database connection failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Order service is ready",
		})
	}
}
