package config

import "os"

type Config struct {
	Port              string
	DatabaseURL       string
	UserServiceURL    string
	ProductServiceURL string
	PaymentServiceURL string
	OTELEndpoint      string
	ServiceName       string
}

func Load() *Config {
	return &Config{
		Port:              getEnv("SERVER_PORT", "8003"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/orders?sslmode=disable"),
		UserServiceURL:    getEnv("USER_SERVICE_URL", "http://localhost:8001"),
		ProductServiceURL: getEnv("PRODUCT_SERVICE_URL", "http://localhost:8000"),
		PaymentServiceURL: getEnv("PAYMENT_SERVICE_URL", "http://localhost:8085"),
		OTELEndpoint:      getEnv("OTEL_ENDPOINT", "localhost:4317"),
		ServiceName:       getEnv("SERVICE_NAME", "order-service"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
