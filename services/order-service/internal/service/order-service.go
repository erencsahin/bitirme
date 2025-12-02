package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	httpClient        *http.Client
}

func NewOrderService(repo *repository.OrderRepository, cfg *config.Config) *OrderService {
	return &OrderService{
		repo:              repo,
		userServiceURL:    cfg.UserServiceURL,
		productServiceURL: cfg.ProductServiceURL,
		paymentServiceURL: cfg.PaymentServiceURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ValidateToken - User Service'den token doğrulama
func (s *OrderService) ValidateToken(token string) (uint, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/auth/validate", s.userServiceURL), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New("invalid token")
	}

	var result struct {
		Status string `json:"status"`
		Data   struct {
			UserID uint `json:"user_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result.Data.UserID, nil
}

// CheckStock - Product Service'den stok kontrolü
func (s *OrderService) CheckStock(productID uint, quantity int, token string) (bool, error) {
	url := fmt.Sprintf("%s/api/products/%d/check_stock/?quantity=%d",
		s.productServiceURL, productID, quantity)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to check stock: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("stock check failed: %s", string(body))
	}

	var result struct {
		Status string `json:"status"`
		Data   struct {
			IsAvailable bool `json:"is_available"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return result.Data.IsAvailable, nil
}

// UpdateStock - Product Service'de stok güncelleme (azaltma)
func (s *OrderService) UpdateStock(productID uint, quantity int, token string) error {
	url := fmt.Sprintf("%s/api/products/%d/update_stock/", s.productServiceURL, productID)

	payload := map[string]int{
		"quantity": -quantity, // Negative = decrease stock
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("stock update failed: %s", string(body))
	}

	return nil
}

// ProcessPayment - Payment Service'e ödeme isteği
func (s *OrderService) ProcessPayment(orderID uint, amount float64, token string) (string, error) {
	url := fmt.Sprintf("%s/api/payments", s.paymentServiceURL)

	payload := map[string]interface{}{
		"order_id": orderID,
		"amount":   amount,
		"currency": "USD",
		"method":   "credit_card",
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to process payment: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			PaymentID string `json:"payment_id"`
			Status    string `json:"status"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("payment failed: %s", result.Message)
	}

	return result.Data.PaymentID, nil
}

// CreateOrder - Yeni order oluşturma
func (s *OrderService) CreateOrder(ctx context.Context, userID uint, productID uint, quantity int, token string) (*models.Order, error) {
	// 1. Stock kontrolü
	available, err := s.CheckStock(productID, quantity, token)
	if err != nil {
		return nil, fmt.Errorf("stock check failed: %w", err)
	}
	if !available {
		return nil, errors.New("insufficient stock")
	}

	// 2. Order oluştur (status: PENDING)
	order := &models.Order{
		UserID:      userID,
		ProductID:   productID,
		Quantity:    quantity,
		Status:      models.OrderStatusPending,
		TotalAmount: 99.99, // TODO: Get real price from Product Service
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// 3. Payment işlemi
	paymentID, err := s.ProcessPayment(order.ID, order.TotalAmount, token)
	if err != nil {
		// Payment başarısız - order'ı iptal et
		order.Status = models.OrderStatusCancelled
		s.repo.Update(ctx, order)
		return nil, fmt.Errorf("payment failed: %w", err)
	}

	// 4. Payment başarılı - Stok azalt
	if err := s.UpdateStock(productID, quantity, token); err != nil {
		// Stok azaltma başarısız - order'ı iptal et
		order.Status = models.OrderStatusCancelled
		s.repo.Update(ctx, order)
		return nil, fmt.Errorf("stock update failed: %w", err)
	}

	// 5. Order'ı COMPLETED olarak işaretle
	order.Status = models.OrderStatusCompleted
	order.PaymentID = &paymentID
	if err := s.repo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	return order, nil
}

// GetOrderByID - Order'ı ID ile getir
func (s *OrderService) GetOrderByID(ctx context.Context, id uint) (*models.Order, error) {
	return s.repo.GetByID(ctx, id)
}

// GetOrdersByUser - Kullanıcının tüm order'larını getir
func (s *OrderService) GetOrdersByUser(ctx context.Context, userID uint) ([]*models.Order, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// GetAllOrders - Tüm order'ları getir
func (s *OrderService) GetAllOrders(ctx context.Context) ([]*models.Order, error) {
	return s.repo.GetAll(ctx)
}

// UpdateOrderStatus - Order status güncelleme
func (s *OrderService) UpdateOrderStatus(ctx context.Context, id uint, status string) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	order.Status = status
	return s.repo.Update(ctx, order)
}

// CancelOrder - Order iptali
func (s *OrderService) CancelOrder(ctx context.Context, id uint) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	order.Status = models.OrderStatusCancelled
	return s.repo.Update(ctx, order)
}
