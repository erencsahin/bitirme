package main

import (
	"context"
	"log"
	"time"

	"order-service/internal/config"
	"order-service/internal/handler"
	"order-service/internal/models"
	"order-service/internal/repository"
	"order-service/internal/service"
	"order-service/internal/telemetry"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	godotenv.Load()

	cfg := config.Load()

	// Initialize OpenTelemetry
	shutdown, err := telemetry.InitTracer(cfg.ServiceName, cfg.OTELEndpoint)
	if err != nil {
		log.Printf("⚠️  OpenTelemetry init failed: %v (continuing without tracing)", err)
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := shutdown(ctx); err != nil {
				log.Printf("Error shutting down tracer: %v", err)
			}
		}()
		log.Println("✅ OpenTelemetry initialized")
	}

	// Database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&models.Order{})

	// Services
	repo := repository.New(db)
	svc := service.New(repo, cfg)
	h := handler.New(svc)

	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API routes
	api := r.Group("/api")
	{
		api.GET("/health", h.Health)
		orders := api.Group("/orders")
		{
			orders.POST("", h.CreateOrder)
			orders.GET("", h.GetAllOrders)
			orders.GET("/:id", h.GetOrder)
			orders.GET("/user/:user_id", h.GetUserOrders)
			orders.PUT("/:id/cancel", h.CancelOrder)
		}
	}

	log.Printf("🚀 Order Service on port %s", cfg.Port)
	log.Printf("📊 Metrics: http://localhost:%s/metrics", cfg.Port)
	log.Printf("🔍 Traces: %s", cfg.OTELEndpoint)

	r.Run(":" + cfg.Port)
}
