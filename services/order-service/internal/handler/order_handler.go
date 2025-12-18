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

// getAuthenticatedUserID: Tüm handlerlar için ortak ID yönetim fonksiyonu
func getAuthenticatedUserID(c *gin.Context) string {
	// 1. Eğer Auth Middleware (JWT) devredeyse context'ten al
	userIDVal, exists := c.Get("user_id")
	if exists {
		return userIDVal.(string)
	}
	// 2. Middleware yoksa (Test aşaması), pgAdmin'de bakiye verdiğin gerçek ID'yi kullan
	return "eadf7070-8f2b-4323-b660-83c7938df731"
}

// CreateOrder godoc
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

	userID := getAuthenticatedUserID(c)

	order, err := h.orderService.CreateOrder(c.Request.Context(), userID, &req)
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
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")
	userID := getAuthenticatedUserID(c)

	order, err := h.orderService.GetOrder(c.Request.Context(), orderID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Status:  "error",
			Message: "Sipariş bulunamadı veya yetkiniz yok",
		})
		return
	}

	c.JSON(http.StatusOK, models.OrderResponse{
		Status: "success",
		Data:   order,
	})
}

// GetMyOrders godoc
func (h *OrderHandler) GetMyOrders(c *gin.Context) {
	userID := getAuthenticatedUserID(c)

	orders, err := h.orderService.GetUserOrders(c.Request.Context(), userID)
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
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	var req models.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
		return
	}

	userID := getAuthenticatedUserID(c)

	order, err := h.orderService.UpdateOrderStatus(c.Request.Context(), orderID, userID, req.Status)
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
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")
	userID := getAuthenticatedUserID(c)

	order, err := h.orderService.CancelOrder(c.Request.Context(), orderID, userID)
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