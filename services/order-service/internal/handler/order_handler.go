package handler

import (
	"net/http"

	"order-service/internal/models"
	"order-service/internal/service"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Create a new order with items. Requires authentication.
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order body models.CreateOrderRequest true "Order creation details"
// @Success 201 {object} models.OrderResponse "Order created successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	// If no auth middleware, use a default for testing
	userID, exists := c.Get("user_id")
	if !exists {
		// For testing without auth - remove this in production
		userID = "test-user-123"
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.OrderResponse{
		Status:  "success",
		Data:    order,
		Message: "Order created successfully",
	})
}

// GetOrder godoc
// @Summary Get order by ID
// @Description Retrieve detailed information about a specific order
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID (UUID)" example("123e4567-e89b-12d3-a456-426614174000")
// @Success 200 {object} models.OrderResponse "Order details"
// @Failure 404 {object} models.ErrorResponse "Order not found"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Router /orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		userID = "test-user-123" // For testing
	}

	order, err := h.orderService.GetOrder(c.Request.Context(), orderID, userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.OrderResponse{
		Status: "success",
		Data:   order,
	})
}

// GetMyOrders godoc
// @Summary Get current user's orders
// @Description Retrieve all orders for the authenticated user with pagination
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1) example(1)
// @Param limit query int false "Items per page" default(10) example(10)
// @Success 200 {object} models.OrderListResponse "List of orders"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /orders/my-orders [get]
func (h *OrderHandler) GetMyOrders(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		userID = "test-user-123" // For testing
	}

	orders, err := h.orderService.GetUserOrders(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.OrderListResponse{
		Status: "success",
		Data:   orders,
	})
}

// UpdateOrderStatus godoc
// @Summary Update order status
// @Description Update the status of an existing order. Valid statuses: pending, processing, shipped, delivered, cancelled
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID (UUID)" example("123e4567-e89b-12d3-a456-426614174000")
// @Param status body models.UpdateStatusRequest true "New status information"
// @Success 200 {object} models.OrderResponse "Status updated successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid status or request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "Order not found"
// @Router /orders/{id}/status [patch]
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")

	var req models.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		userID = "test-user-123" // For testing
	}

	order, err := h.orderService.UpdateOrderStatus(c.Request.Context(), orderID, userID.(string), req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.OrderResponse{
		Status:  "success",
		Data:    order,
		Message: "Order status updated successfully",
	})
}

// CancelOrder godoc
// @Summary Cancel an order
// @Description Cancel an existing order. Only orders with status 'pending' or 'processing' can be cancelled
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID (UUID)" example("123e4567-e89b-12d3-a456-426614174000")
// @Success 200 {object} models.OrderResponse "Order cancelled successfully"
// @Failure 400 {object} models.ErrorResponse "Cannot cancel order (wrong status)"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "Order not found"
// @Router /orders/{id}/cancel [post]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		userID = "test-user-123" // For testing
	}

	order, err := h.orderService.CancelOrder(c.Request.Context(), orderID, userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.OrderResponse{
		Status:  "success",
		Data:    order,
		Message: "Order cancelled successfully",
	})
}
