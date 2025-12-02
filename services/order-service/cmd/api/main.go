package main

import (
	"fmt"
	"log"

	"order-service/internal/config"
	"order-service/internal/handler"
	"order-service/internal/models"
	"order-service/internal/repository"
	"order-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load config
	cfg := config.LoadConfig()

	// Database connection
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(&models.Order{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize layers
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, cfg)
	orderHandler := handler.NewOrderHandler(orderService)

	// Setup Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Routes
	api := router.Group("/api")
	{
		api.GET("/health", orderHandler.Health)

		orders := api.Group("/orders")
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("", orderHandler.GetAllOrders)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.GET("/user/:user_id", orderHandler.GetUserOrders)
			orders.PUT("/:id/cancel", orderHandler.CancelOrder)
		}
	}

	// Start server
	port := cfg.Port
	fmt.Printf("🚀 Order Service starting on port %s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
