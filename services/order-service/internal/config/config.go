package config

import (
	"os"
)

type Config struct {
	Port              string
	DatabaseURL       string
	RedisURL          string
	UserServiceURL    string
	ProductServiceURL string
	PaymentServiceURL string
	JWTSecret         string
}

func LoadConfig() *Config {
	return &Config{
		Port:              getEnv("PORT", "8003"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/order_db?sslmode=disable"),
		RedisURL:          getEnv("REDIS_URL", "redis://localhost:6379/2"),
		UserServiceURL:    getEnv("USER_SERVICE_URL", "http://localhost:8001"),
		ProductServiceURL: getEnv("PRODUCT_SERVICE_URL", "http://localhost:8000"),
		PaymentServiceURL: getEnv("PAYMENT_SERVICE_URL", "http://localhost:8085"),
		JWTSecret:         getEnv("JWT_SECRET", "your-secret-key"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
