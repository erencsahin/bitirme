package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort          string
	DatabaseURL         string
	RedisURL            string
	OTELEndpoint        string
	ServiceName         string
	Environment         string
	ProductServiceURL   string
	InventoryServiceURL string
	PaymentServiceURL   string
	UserServiceURL      string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	return &Config{
		ServerPort:          getEnv("SERVER_PORT", "8003"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/orders?sslmode=disable"),
		RedisURL:            getEnv("REDIS_URL", "localhost:6379"),
		OTELEndpoint:        getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317"),
		ServiceName:         getEnv("SERVICE_NAME", "order-service"),
		Environment:         getEnv("ENVIRONMENT", "development"),
		ProductServiceURL:   getEnv("PRODUCT_SERVICE_URL", "http://localhost:8002"),
		UserServiceURL:      getEnv("USER_SERVICE_URL", "http://localhost:8083"),
		InventoryServiceURL: getEnv("INVENTORY_SERVICE_URL", "http://localhost:8084"),
		PaymentServiceURL:   getEnv("PAYMENT_SERVICE_URL", "http://localhost:8085"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
