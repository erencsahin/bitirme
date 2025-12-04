package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Order represents an order in the system
type Order struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	UserID          string         `json:"user_id" gorm:"type:varchar(255);not null;index" example:"user-123e4567"`
	Status          string         `json:"status" gorm:"type:varchar(50);not null;default:'pending'" example:"pending"`
	TotalAmount     float64        `json:"total_amount" gorm:"type:decimal(10,2);not null" example:"3499.99"`
	ShippingAddress string         `json:"shipping_address" gorm:"type:text;not null" example:"123 Tech Street, Istanbul, Turkey"`
	BillingAddress  string         `json:"billing_address" gorm:"type:text;not null" example:"123 Tech Street, Istanbul, Turkey"`
	PaymentID       *string        `json:"payment_id,omitempty" gorm:"type:varchar(255)" example:"pay_1234567890"`
	Items           []OrderItem    `json:"items" gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
	CreatedAt       time.Time      `json:"created_at" example:"2025-12-04T15:30:00Z"`
	UpdatedAt       time.Time      `json:"updated_at" example:"2025-12-04T15:30:00Z"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()" example:"item-123e4567"`
	OrderID   uuid.UUID      `json:"order_id" gorm:"type:uuid;not null;index" example:"123e4567-e89b-12d3-a456-426614174000"`
	ProductID string         `json:"product_id" gorm:"type:varchar(255);not null" example:"prod-123e4567-e89b-12d3-a456-426614174000"`
	Quantity  int            `json:"quantity" gorm:"not null" example:"2"`
	Price     float64        `json:"price" gorm:"type:decimal(10,2);not null" example:"1749.99"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// CreateOrderRequest represents the request body for creating an order
type CreateOrderRequest struct {
	Items           []OrderItemRequest `json:"items" binding:"required,min=1" example:"[{\"product_id\":\"prod-123\",\"quantity\":2,\"price\":1749.99}]"`
	ShippingAddress string             `json:"shipping_address" binding:"required" example:"123 Tech Street, Kadikoy, Istanbul, Turkey"`
	BillingAddress  string             `json:"billing_address" binding:"required" example:"123 Tech Street, Kadikoy, Istanbul, Turkey"`
}

// OrderItemRequest represents an item in the order creation request
type OrderItemRequest struct {
	ProductID string  `json:"product_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Quantity  int     `json:"quantity" binding:"required,min=1" example:"2"`
	Price     float64 `json:"price" binding:"required,gt=0" example:"1749.99"`
}

// UpdateStatusRequest represents the request to update order status
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending processing shipped delivered cancelled" example:"processing" enums:"pending,processing,shipped,delivered,cancelled"`
}

// OrderResponse represents the standard response for order operations
type OrderResponse struct {
	Status  string `json:"status" example:"success"`
	Data    *Order `json:"data"`
	Message string `json:"message,omitempty" example:"Order created successfully"`
}

// OrderListResponse represents a list of orders
type OrderListResponse struct {
	Status string   `json:"status" example:"success"`
	Data   []*Order `json:"data"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Invalid request"`
	Error   string `json:"error,omitempty" example:"validation error: field is required"`
}

// TableName specifies the table name for Order model
func (Order) TableName() string {
	return "orders"
}

// TableName specifies the table name for OrderItem model
func (OrderItem) TableName() string {
	return "order_items"
}
