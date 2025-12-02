package repository

import (
	"context"
	"order-service/internal/models"

	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *models.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *OrderRepository) GetByID(ctx context.Context, id uint) (*models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).First(&order, id).Error
	return &order, err
}

func (r *OrderRepository) GetByUserID(ctx context.Context, userID uint) ([]*models.Order, error) {
	var orders []*models.Order
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&orders).Error
	return orders, err
}

func (r *OrderRepository) GetAll(ctx context.Context) ([]*models.Order, error) {
	var orders []*models.Order
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&orders).Error
	return orders, err
}

func (r *OrderRepository) Update(ctx context.Context, order *models.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}
