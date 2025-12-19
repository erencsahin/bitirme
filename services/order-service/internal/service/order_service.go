package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"strings"

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
		client: &http.Client{
  Timeout: 10 * time.Second,
  CheckRedirect: func(req *http.Request, via []*http.Request) error {
    return nil // allow redirects
  },
},
		tracer:            otel.Tracer("order-service"),
	}
}

// --- Helper calls to other services ---

// CheckStock checks if the given quantity is available for the product.
func (s *OrderService) CheckStock(ctx context.Context, productID string, quantity int) (bool, error) {
	ctx, span := s.tracer.Start(ctx, "CheckStock")
	defer span.End()

	span.SetAttributes(
		attribute.String("product_id", productID),
		attribute.Int("quantity", quantity),
	)

	start := time.Now()
	base := strings.TrimRight(s.productServiceURL, "/")
	url := fmt.Sprintf("%s/api/products/%s/check_stock/?quantity=%d", base, strings.Trim(productID, "/"), quantity)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := s.client.Do(req)

	metrics.StockCheckDuration.Observe(time.Since(start).Seconds())

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "stock check failed")
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
  		b, _ := io.ReadAll(resp.Body)
		log.Printf("stock-service error: status=%d body=%s url=%s", resp.StatusCode, string(b), url)
		span.SetStatus(codes.Error, "stock check failed (non-200)")
		return false, errors.New("stock check failed")
	}

	var result struct {
		Data struct {
			IsAvailable bool `json:"is_available"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode failed")
		return false, err
	}

	span.SetAttributes(attribute.Bool("is_available", result.Data.IsAvailable))
	return result.Data.IsAvailable, nil
}

// UpdateStock decreases stock for a product (negative quantity).
func (s *OrderService) UpdateStock(ctx context.Context, productID string, quantity int) error {
	ctx, span := s.tracer.Start(ctx, "UpdateStock")
	defer span.End()

	span.SetAttributes(
		attribute.String("product_id", productID),
		attribute.Int("quantity_delta", -quantity),
	)

	base := strings.TrimRight(s.productServiceURL, "/")
	url := fmt.Sprintf("%s/api/products/%s/update_stock/", base, productID)
	payload := map[string]int{"quantity": -quantity}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "stock update failed")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		span.SetStatus(codes.Error, "stock update failed (non-200)")
		return errors.New("stock update failed")
	}

	return nil
}

// ProcessPayment sends a payment request to the payment service.
func (s *OrderService) ProcessPayment(
	ctx context.Context,
	orderID string,
	userID string,
	amount float64,
) (string, error) {
	ctx, span := s.tracer.Start(ctx, "ProcessPayment")
	defer span.End()

	span.SetAttributes(
		attribute.String("order_id", orderID),
		attribute.String("user_id", userID),
		attribute.Float64("amount", amount),
	)

	start := time.Now()
	url := s.paymentServiceURL + "/api/payments"

	payload := map[string]interface{}{
		"order_id":       orderID,
		"user_id":        userID,
		"amount":         amount,
		"currency":       "TRY",
		"payment_method": "CREDIT_CARD",
	}

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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		log.Printf("payment-service error: status=%d body=%s", resp.StatusCode, string(b))

		span.SetStatus(codes.Error, "payment failed")
		return "", errors.New("payment failed")
	}

	var result struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode payment response failed")
		return "", err
	}

	span.SetAttributes(attribute.String("payment_id", result.Data.ID))
	return result.Data.ID, nil
}

// --- Public API used by handlers ---

// CreateOrder creates a new order from a request.
func (s *OrderService) CreateOrder(ctx context.Context, userID string, req *models.CreateOrderRequest) (*models.Order, error) {
	ctx, span := s.tracer.Start(ctx, "CreateOrder")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID),
		attribute.Int("item_count", len(req.Items)),
	)

	start := time.Now()
	defer func() {
		metrics.OrderDuration.Observe(time.Since(start).Seconds())
	}()

	if len(req.Items) == 0 {
		err := errors.New("order must contain at least one item")
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation error")
		metrics.OrdersFailed.Inc()
		return nil, err
	}

	// 1) Stock check & total amount
	var total float64
	for _, item := range req.Items {
		ok, err := s.CheckStock(ctx, item.ProductID, item.Quantity)
		if err != nil {
			metrics.OrdersFailed.Inc()
			span.RecordError(err)
			span.SetStatus(codes.Error, "stock check failed")
			return nil, err
		}
		if !ok {
			err := fmt.Errorf("insufficient stock for product %s", item.ProductID)
			metrics.OrdersFailed.Inc()
			span.RecordError(err)
			span.SetStatus(codes.Error, "insufficient stock")
			return nil, err
		}
		total += float64(item.Quantity) * item.Price
	}

	// 2) Create order model
	order := &models.Order{
		UserID:          userID,
		Status:          "pending",
		TotalAmount:     total,
		ShippingAddress: req.ShippingAddress,
		BillingAddress:  req.BillingAddress,
	}

	for _, itemReq := range req.Items {
		order.Items = append(order.Items, models.OrderItem{
			ProductID: itemReq.ProductID,
			Quantity:  itemReq.Quantity,
			Price:     itemReq.Price,
		})
	}

	// 3) Save order in DB
	if err := s.repo.Create(ctx, order); err != nil {
		metrics.OrdersFailed.Inc()
		span.RecordError(err)
		span.SetStatus(codes.Error, "order creation failed")
		return nil, err
	}

	span.SetAttributes(
		attribute.String("order_id", order.ID.String()),
		attribute.Float64("total_amount", order.TotalAmount),
	)

	// 4) Process payment
	paymentID, err := s.ProcessPayment(
		ctx,
		order.ID.String(), // order ID uuid ise
		userID,            // handler’dan gelen user_id (string)
		order.TotalAmount, // hesapladığın toplam
	)
	if err != nil {
		order.Status = "cancelled"
		_ = s.repo.Update(ctx, order)
		metrics.OrdersFailed.Inc()
		span.RecordError(err)
		span.SetStatus(codes.Error, "payment failed")
		return nil, err
	}
	order.PaymentID = &paymentID

	// 5) Update stock for each item
	for _, item := range order.Items {
		if err := s.UpdateStock(ctx, item.ProductID, item.Quantity); err != nil {
			order.Status = "cancelled"
			_ = s.repo.Update(ctx, order)
			metrics.OrdersFailed.Inc()
			span.RecordError(err)
			span.SetStatus(codes.Error, "stock update failed")
			return nil, err
		}
	}

	// 6) Final status
	order.Status = "processing"
	if err := s.repo.Update(ctx, order); err != nil {
		metrics.OrdersFailed.Inc()
		span.RecordError(err)
		span.SetStatus(codes.Error, "order update failed")
		return nil, err
	}

	metrics.OrdersCreated.Inc()
	span.SetStatus(codes.Ok, "order created")
	return order, nil
}

// GetOrder returns a single order for a user.
func (s *OrderService) GetOrder(ctx context.Context, orderID, userID string) (*models.Order, error) {
	ctx, span := s.tracer.Start(ctx, "GetOrder")
	defer span.End()

	span.SetAttributes(
		attribute.String("order_id", orderID),
		attribute.String("user_id", userID),
	)

	order, err := s.repo.GetByIDAndUser(ctx, orderID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "order not found")
		return nil, err
	}

	span.SetStatus(codes.Ok, "order fetched")
	return order, nil
}

// GetUserOrders returns all orders for a given user.
func (s *OrderService) GetUserOrders(ctx context.Context, userID string) ([]*models.Order, error) {
	ctx, span := s.tracer.Start(ctx, "GetUserOrders")
	defer span.End()

	span.SetAttributes(attribute.String("user_id", userID))

	orders, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "fetch user orders failed")
		return nil, err
	}

	span.SetAttributes(attribute.Int("order_count", len(orders)))
	span.SetStatus(codes.Ok, "orders fetched")
	return orders, nil
}

// UpdateOrderStatus updates the status of an order for a user.
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID, userID, status string) (*models.Order, error) {
	ctx, span := s.tracer.Start(ctx, "UpdateOrderStatus")
	defer span.End()

	span.SetAttributes(
		attribute.String("order_id", orderID),
		attribute.String("user_id", userID),
		attribute.String("new_status", status),
	)

	order, err := s.repo.GetByIDAndUser(ctx, orderID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "order not found")
		return nil, err
	}

	order.Status = status
	if err := s.repo.Update(ctx, order); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "status update failed")
		return nil, err
	}

	span.SetStatus(codes.Ok, "status updated")
	return order, nil
}

// CancelOrder cancels an order if allowed.
func (s *OrderService) CancelOrder(ctx context.Context, orderID, userID string) (*models.Order, error) {
	ctx, span := s.tracer.Start(ctx, "CancelOrder")
	defer span.End()

	span.SetAttributes(
		attribute.String("order_id", orderID),
		attribute.String("user_id", userID),
	)

	order, err := s.repo.GetByIDAndUser(ctx, orderID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "order not found")
		return nil, err
	}

	if order.Status != "pending" && order.Status != "processing" {
		err := errors.New("only pending or processing orders can be cancelled")
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid status for cancellation")
		return nil, err
	}

	order.Status = "cancelled"
	if err := s.repo.Update(ctx, order); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cancel update failed")
		return nil, err
	}

	metrics.OrdersCancelled.Inc()
	span.SetStatus(codes.Ok, "order cancelled")
	return order, nil
}
