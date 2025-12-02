package models

import (
	"time"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID              string      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID          string      `gorm:"type:varchar(255);not null" json:"user_id"`
	Status          OrderStatus `gorm:"type:varchar(50);not null;default:'pending'" json:"status"`
	TotalAmount     float64     `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Currency        string      `gorm:"type:varchar(3);not null;default:'USD'" json:"currency"`
	ShippingAddress string      `gorm:"type:text" json:"shipping_address"`
	BillingAddress  string      `gorm:"type:text" json:"billing_address"`
	Notes           string      `gorm:"type:text" json:"notes"`
	OrderItems      []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"order_items"`
	CreatedAt       time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time   `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Order) TableName() string {
	return "orders"
}
