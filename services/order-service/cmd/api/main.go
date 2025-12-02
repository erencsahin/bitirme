package main

import (
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
	godotenv.Load()

	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&models.Order{})

	repo := repository.New(db)
	svc := service.New(repo, cfg)
	h := handler.New(svc)

	r := gin.Default()

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

	log.Printf("🚀 Order Service on port %s\n", cfg.Port)
	r.Run(":" + cfg.Port)
}
