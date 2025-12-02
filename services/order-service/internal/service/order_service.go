package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"order-service/internal/config"
	"order-service/internal/models"
	"order-service/internal/repository"
)

type OrderService struct {
	repo              *repository.OrderRepository
	userServiceURL    string
	productServiceURL string
	paymentServiceURL string
	client            *http.Client
}

func New(repo *repository.OrderRepository, cfg *config.Config) *OrderService {
	return &OrderService{
		repo:              repo,
		userServiceURL:    cfg.UserServiceURL,
		productServiceURL: cfg.ProductServiceURL,
		paymentServiceURL: cfg.PaymentServiceURL,
		client:            &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *OrderService) ValidateToken(token string) (uint, error) {
	req, _ := http.NewRequest("GET", s.userServiceURL+"/api/auth/validate", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, errors.New("invalid token")
	}

	var result struct {
		Data struct {
			UserID uint `json:"user_id"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Data.UserID, nil
}

func (s *OrderService) CheckStock(productID uint, quantity int) (bool, error) {
	url := fmt.Sprintf("%s/api/products/%d/check_stock/?quantity=%d", s.productServiceURL, productID, quantity)
	resp, err := s.client.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, errors.New("stock check failed")
	}

	var result struct {
		Data struct {
			IsAvailable bool `json:"is_available"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Data.IsAvailable, nil
}

func (s *OrderService) UpdateStock(productID uint, quantity int, token string) error {
	url := fmt.Sprintf("%s/api/products/%d/update_stock/", s.productServiceURL, productID)
	payload := map[string]int{"quantity": -quantity}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("stock update failed")
	}
	return nil
}

func (s *OrderService) ProcessPayment(orderID uint, amount float64) (string, error) {
	url := s.paymentServiceURL + "/api/payments"
	payload := map[string]interface{}{"order_id": orderID, "amount": amount}
	body, _ := json.Marshal(payload)

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return "", errors.New("payment failed")
	}

	var result struct {
		Data struct {
			PaymentID string `json:"payment_id"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Data.PaymentID, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, userID, productID uint, quantity int, token string) (*models.Order, error) {
	// 1. Check stock
	available, err := s.CheckStock(productID, quantity)
	if err != nil || !available {
		return nil, errors.New("insufficient stock")
	}

	// 2. Create order
	order := &models.Order{
		UserID:      userID,
		ProductID:   productID,
		Quantity:    quantity,
		Status:      "PENDING",
		TotalAmount: 99.99,
	}
	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	// 3. Process payment
	paymentID, err := s.ProcessPayment(order.ID, order.TotalAmount)
	if err != nil {
		order.Status = "CANCELLED"
		s.repo.Update(ctx, order)
		return nil, err
	}

	// 4. Update stock
	if err := s.UpdateStock(productID, quantity, token); err != nil {
		order.Status = "CANCELLED"
		s.repo.Update(ctx, order)
		return nil, err
	}

	// 5. Complete order
	order.Status = "COMPLETED"
	order.PaymentID = &paymentID
	s.repo.Update(ctx, order)

	return order, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, id uint) (*models.Order, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *OrderService) GetOrdersByUser(ctx context.Context, userID uint) ([]*models.Order, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *OrderService) GetAllOrders(ctx context.Context) ([]*models.Order, error) {
	return s.repo.GetAll(ctx)
}

func (s *OrderService) CancelOrder(ctx context.Context, id uint) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	order.Status = "CANCELLED"
	return s.repo.Update(ctx, order)
}
