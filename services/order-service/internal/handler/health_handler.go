package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status" example:"ok" description:"Service status"`
	Service   string    `json:"service,omitempty" example:"order-service" description:"Service name"`
	Timestamp time.Time `json:"timestamp,omitempty" example:"2025-12-04T15:30:00Z" description:"Current timestamp"`
	Database  string    `json:"database,omitempty" example:"connected" description:"Database connection status"`
}

// Health godoc
// @Summary Health check endpoint
// @Description Returns the health status of the Order Service including database connectivity
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse "Service is healthy"
// @Failure 503 {object} HealthResponse "Service is degraded or unavailable"
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	// Check database connection
	sqlDB, err := h.db.DB()
	dbStatus := "connected"

	if err != nil {
		dbStatus = "error"
	} else {
		if err := sqlDB.Ping(); err != nil {
			dbStatus = "disconnected"
		}
	}

	response := HealthResponse{
		Status:    "ok",
		Service:   "order-service",
		Timestamp: time.Now(),
		Database:  dbStatus,
	}

	// Set status code based on health
	statusCode := http.StatusOK
	if dbStatus != "connected" {
		statusCode = http.StatusServiceUnavailable
		response.Status = "degraded"
	}

	c.JSON(statusCode, response)
}
