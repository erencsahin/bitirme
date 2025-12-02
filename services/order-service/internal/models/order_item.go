package models

import (
	"time"
)

type OrderItem struct {
	ID        string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrderID   string    `gorm:"type:uuid;not null" json:"order_id"`
	ProductID string    `gorm:"type:varchar(255);not null" json:"product_id"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	UnitPrice float64   `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	Subtotal  float64   `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (OrderItem) TableName() string {
	return "order_items"
}
