package handler

import (
	"net/http"
	"strconv"
	"strings"

	"order-service/internal/models"
	"order-service/internal/service"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	svc *service.OrderService
}

func New(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) extractToken(c *gin.Context) (string, error) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return "", http.ErrNotSupported
	}
	parts := strings.Split(auth, " ")
	if len(parts) != 2 {
		return "", http.ErrNotSupported
	}
	return parts[1], nil
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	token, err := h.extractToken(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	userID, err := h.svc.ValidateToken(c, token)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid token"})
		return
	}

	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	order, err := h.svc.CreateOrder(c, userID, req.ProductID, req.Quantity, token)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"status": "success", "data": order})
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	order, err := h.svc.GetOrderByID(c, uint(id))
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.JSON(200, gin.H{"status": "success", "data": order})
}

func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("user_id"), 10, 32)
	orders, _ := h.svc.GetOrdersByUser(c, uint(userID))
	c.JSON(200, gin.H{"status": "success", "data": orders})
}

func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	orders, _ := h.svc.GetAllOrders(c)
	c.JSON(200, gin.H{"status": "success", "data": orders})
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.svc.CancelOrder(c, uint(id)); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "success", "message": "order cancelled"})
}

func (h *OrderHandler) Health(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
