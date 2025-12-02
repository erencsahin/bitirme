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
	"order-service/internal/metrics"
	"order-service/internal/models"
	"order-service/internal/repository"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type OrderService struct {
	repo              *repository.OrderRepository
	userServiceURL    string
	productServiceURL string
	paymentServiceURL string
	client            *http.Client
	tracer            trace.Tracer
}

func New(repo *repository.OrderRepository, cfg *config.Config) *OrderService {
	return &OrderService{
		repo:              repo,
		userServiceURL:    cfg.UserServiceURL,
		productServiceURL: cfg.ProductServiceURL,
		paymentServiceURL: cfg.PaymentServiceURL,
		client:            &http.Client{Timeout: 10 * time.Second},
		tracer:            otel.Tracer("order-service"),
	}
}

func (s *OrderService) ValidateToken(ctx context.Context, token string) (uint, error) {
	ctx, span := s.tracer.Start(ctx, "ValidateToken")
	defer span.End()

	req, _ := http.NewRequestWithContext(ctx, "GET", s.userServiceURL+"/api/auth/validate", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "token validation failed")
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		span.SetStatus(codes.Error, "invalid token")
		return 0, errors.New("invalid token")
	}

	var result struct {
		Data struct {
			UserID uint `json:"user_id"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	span.SetAttributes(attribute.Int("user_id", int(result.Data.UserID)))
	return result.Data.UserID, nil
}

func (s *OrderService) CheckStock(ctx context.Context, productID uint, quantity int) (bool, error) {
	ctx, span := s.tracer.Start(ctx, "CheckStock")
	defer span.End()

	span.SetAttributes(
		attribute.Int("product_id", int(productID)),
		attribute.Int("quantity", quantity),
	)

	start := time.Now()
	url := fmt.Sprintf("%s/api/products/%d/check_stock/?quantity=%d", s.productServiceURL, productID, quantity)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := s.client.Do(req)

	metrics.StockCheckDuration.Observe(time.Since(start).Seconds())

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "stock check failed")
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		span.SetStatus(codes.Error, "stock check failed")
		return false, errors.New("stock check failed")
	}

	var result struct {
		Data struct {
			IsAvailable bool `json:"is_available"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	span.SetAttributes(attribute.Bool("is_available", result.Data.IsAvailable))
	return result.Data.IsAvailable, nil
}

func (s *OrderService) UpdateStock(ctx context.Context, productID uint, quantity int, token string) error {
	ctx, span := s.tracer.Start(ctx, "UpdateStock")
	defer span.End()

	span.SetAttributes(
		attribute.Int("product_id", int(productID)),
		attribute.Int("quantity", -quantity),
	)

	url := fmt.Sprintf("%s/api/products/%d/update_stock/", s.productServiceURL, productID)
	payload := map[string]int{"quantity": -quantity}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "stock update failed")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		span.SetStatus(codes.Error, "stock update failed")
		return errors.New("stock update failed")
	}

	return nil
}

func (s *OrderService) ProcessPayment(ctx context.Context, orderID uint, amount float64) (string, error) {
	ctx, span := s.tracer.Start(ctx, "ProcessPayment")
	defer span.End()

	span.SetAttributes(
		attribute.Int("order_id", int(orderID)),
		attribute.Float64("amount", amount),
	)

	start := time.Now()
	url := s.paymentServiceURL + "/api/payments"
	payload := map[string]interface{}{"order_id": orderID, "amount": amount}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)

	metrics.PaymentProcessDuration.Observe(time.Since(start).Seconds())

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "payment failed")
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		span.SetStatus(codes.Error, "payment failed")
		return "", errors.New("payment failed")
	}

	var result struct {
		Data struct {
			PaymentID string `json:"payment_id"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	span.SetAttributes(attribute.String("payment_id", result.Data.PaymentID))
	return result.Data.PaymentID, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, userID, productID uint, quantity int, token string) (*models.Order, error) {
	ctx, span := s.tracer.Start(ctx, "CreateOrder")
	defer span.End()

	span.SetAttributes(
		attribute.Int("user_id", int(userID)),
		attribute.Int("product_id", int(productID)),
		attribute.Int("quantity", quantity),
	)

	start := time.Now()
	defer func() {
		metrics.OrderDuration.Observe(time.Since(start).Seconds())
	}()

	// 1. Check stock
	available, err := s.CheckStock(ctx, productID, quantity)
	if err != nil || !available {
		metrics.OrdersFailed.Inc()
		span.RecordError(errors.New("insufficient stock"))
		span.SetStatus(codes.Error, "insufficient stock")
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
		metrics.OrdersFailed.Inc()
		span.RecordError(err)
		span.SetStatus(codes.Error, "order creation failed")
		return nil, err
	}

	span.SetAttributes(attribute.Int("order_id", int(order.ID)))

	// 3. Process payment
	paymentID, err := s.ProcessPayment(ctx, order.ID, order.TotalAmount)
	if err != nil {
		order.Status = "CANCELLED"
		s.repo.Update(ctx, order)
		metrics.OrdersFailed.Inc()
		span.SetStatus(codes.Error, "payment failed")
		return nil, err
	}

	// 4. Update stock
	if err := s.UpdateStock(ctx, productID, quantity, token); err != nil {
		order.Status = "CANCELLED"
		s.repo.Update(ctx, order)
		metrics.OrdersFailed.Inc()
		span.SetStatus(codes.Error, "stock update failed")
		return nil, err
	}

	// 5. Complete order
	order.Status = "COMPLETED"
	order.PaymentID = &paymentID
	s.repo.Update(ctx, order)

	metrics.OrdersCreated.Inc()
	span.SetStatus(codes.Ok, "order completed")
	return order, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, id uint) (*models.Order, error) {
	ctx, span := s.tracer.Start(ctx, "GetOrderByID")
	defer span.End()
	span.SetAttributes(attribute.Int("order_id", int(id)))

	return s.repo.GetByID(ctx, id)
}

func (s *OrderService) GetOrdersByUser(ctx context.Context, userID uint) ([]*models.Order, error) {
	ctx, span := s.tracer.Start(ctx, "GetOrdersByUser")
	defer span.End()
	span.SetAttributes(attribute.Int("user_id", int(userID)))

	return s.repo.GetByUserID(ctx, userID)
}

func (s *OrderService) GetAllOrders(ctx context.Context) ([]*models.Order, error) {
	ctx, span := s.tracer.Start(ctx, "GetAllOrders")
	defer span.End()

	return s.repo.GetAll(ctx)
}

func (s *OrderService) CancelOrder(ctx context.Context, id uint) error {
	ctx, span := s.tracer.Start(ctx, "CancelOrder")
	defer span.End()
	span.SetAttributes(attribute.Int("order_id", int(id)))

	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	order.Status = "CANCELLED"

	metrics.OrdersCancelled.Inc()
	return s.repo.Update(ctx, order)
}
