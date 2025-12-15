package main

import (
	"log"
	"os"

	"order-service/internal/config"
	"order-service/internal/handler"
	"order-service/internal/middleware"
	"order-service/internal/models"
	"order-service/internal/repository"
	"order-service/internal/service"
	"order-service/internal/tracing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "order-service/docs"
)

// @title Order Service API
// @version 1.0
// @description Microservices Order Management API - Handles order creation, management, and status tracking

// @contact.name Eren
// @contact.email eren@example.com

// @host localhost:8003
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your JWT token in the format: Bearer <token>

// @tag.name Health
// @tag.description Health check and service status endpoints

// @tag.name Orders
// @tag.description Order management operations

func main() {
	// Initialize OpenTelemetry tracing
	cleanup := tracing.InitTracer()
	defer cleanup()

	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(&models.Order{}, &models.OrderItem{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize repositories
	orderRepo := repository.New(db)

	// Initialize services
	orderService := service.New(orderRepo, cfg)

	// Initialize handlers
	orderHandler := handler.NewOrderHandler(orderService)
	healthHandler := handler.NewHealthHandler(db)

	// Setup Gin router
	router := gin.Default()

	// Add tracing middleware
	router.Use(middleware.TracingMiddleware())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Public routes (no auth required)
	router.GET("/health", healthHandler.Health)
	router.GET("/api/health", healthHandler.Health)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes (authentication required)
	api := router.Group("/api")
	// Note: Add auth middleware here if you have one
	// api.Use(middleware.AuthMiddleware(cfg.UserServiceURL))
	{
		// Order routes
		api.POST("/orders", orderHandler.CreateOrder)
		api.GET("/orders/:id", orderHandler.GetOrder)
		api.GET("/orders/my-orders", orderHandler.GetMyOrders)
		api.PATCH("/orders/:id/status", orderHandler.UpdateOrderStatus)
		api.POST("/orders/:id/cancel", orderHandler.CancelOrder)
	}

	// Start server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8003"
	}

	log.Printf("🚀 Order Service starting on port %s", port)
	log.Printf("📚 Swagger UI: http://localhost:%s/swagger/index.html", port)
	log.Printf("💚 Health Check: http://localhost:%s/health", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
