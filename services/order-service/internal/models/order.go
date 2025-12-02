package models

import "time"

type Order struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	ProductID   uint      `gorm:"not null" json:"product_id"`
	Quantity    int       `gorm:"not null" json:"quantity"`
	TotalAmount float64   `gorm:"not null" json:"total_amount"`
	Status      string    `gorm:"not null;default:'PENDING'" json:"status"`
	PaymentID   *string   `json:"payment_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateOrderRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}
