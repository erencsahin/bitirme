package models

import (
	"time"
)

type Order struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	ProductID   uint      `gorm:"not null;index" json:"product_id"`
	Quantity    int       `gorm:"not null" json:"quantity"`
	TotalAmount float64   `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Status      string    `gorm:"type:varchar(50);not null;default:'PENDING'" json:"status"` // PENDING, COMPLETED, CANCELLED
	PaymentID   *string   `gorm:"type:varchar(255)" json:"payment_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// OrderStatus constants
const (
	OrderStatusPending   = "PENDING"
	OrderStatusCompleted = "COMPLETED"
	OrderStatusCancelled = "CANCELLED"
)

type CreateOrderRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=PENDING COMPLETED CANCELLED"`
}
