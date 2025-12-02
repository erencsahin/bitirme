package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"order-service/internal/cache"
	"order-service/internal/models"
	"order-service/internal/repository"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

var tracer = otel.Tracer("order-service")

type ProductResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp,omitempty"`
	Data      struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Price       string `json:"price"`
		Stock       int    `json:"stock"`
		InStock     bool   `json:"in_stock"`
		SKU         string `json:"sku,omitempty"`
	} `json:"data"`
}

type OrderService struct {
	repo              *repository.OrderRepository
	cache             *cache.RedisCache
	httpClient        *http.Client
	productServiceURL string
	inventoryClient   *InventoryClient
	paymentClient     *PaymentClient
}

func NewOrderService(repo *repository.OrderRepository, cache *cache.RedisCache, productServiceURL, inventoryServiceURL, paymentServiceURL string) *OrderService {
	return &OrderService{
		repo:              repo,
		cache:             cache,
		productServiceURL: productServiceURL,
		inventoryClient:   NewInventoryClient(inventoryServiceURL),
		paymentClient:     NewPaymentClient(paymentServiceURL),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, order *models.Order) error {
	ctx, span := tracer.Start(ctx, "CreateOrder")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", order.UserID),
		attribute.Int("items_count", len(order.OrderItems)),
	)

	// Validate and calculate order
	totalAmount := 0.0
	for i := range order.OrderItems {
		item := &order.OrderItems[i]

		// Get product info from Product Service
		product, err := s.getProductWithRetry(ctx, item.ProductID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed to get product info")
			return fmt.Errorf("failed to get product %s: %w", item.ProductID, err)
		}

		// Check stock
		if product.Data.Stock < item.Quantity {
			err := fmt.Errorf("insufficient stock for product %s", item.ProductID)
			span.RecordError(err)
			return err
		}

		// ✅ String price'ı float'a çevir
		price, err := strconv.ParseFloat(product.Data.Price, 64)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("invalid price format for product %s: %w", item.ProductID, err)
		}

		// Calculate prices
		item.UnitPrice = price
		item.Subtotal = price * float64(item.Quantity)
		totalAmount += item.Subtotal
	}

	// Check stock availability with Inventory Service
	token, ok := ctx.Value("token").(string)
	if !ok || token == "" {
		return fmt.Errorf("authentication token not found")
	}

	for _, item := range order.OrderItems {
		productID, err := strconv.Atoi(item.ProductID)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("invalid product ID format: %w", err)
		}

		available, err := s.inventoryClient.CheckStock(ctx, productID, item.Quantity, token)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to check inventory: %w", err)
		}

		if !available {
			err := fmt.Errorf("insufficient stock in inventory for product %s", item.ProductID)
			span.RecordError(err)
			return err
		}
	}

	// Reserve stock
	for i, item := range order.OrderItems {
		productID, err := strconv.Atoi(item.ProductID)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("invalid product ID format: %w", err)
		}

		if err := s.inventoryClient.ReserveStock(ctx, productID, item.Quantity, token); err != nil {
			span.RecordError(err)
			// Release already reserved items
			for j := 0; j < i; j++ {
				prevProductID, _ := strconv.Atoi(order.OrderItems[j].ProductID)
				s.inventoryClient.ReleaseStock(ctx, prevProductID, order.OrderItems[j].Quantity, token)
			}
			return fmt.Errorf("failed to reserve stock: %w", err)
		}
	}

	order.TotalAmount = totalAmount
	order.Status = models.OrderStatusPending

	// Process payment
	paymentReq := CreatePaymentRequest{
		OrderID:       order.ID,
		UserID:        order.UserID,
		Amount:        totalAmount,
		Currency:      "USD",
		PaymentMethod: "CREDIT_CARD", // Default
	}

	paymentID, err := s.paymentClient.ProcessPayment(ctx, paymentReq, token)
	if err != nil {
		span.RecordError(err)
		// Release reserved stock if payment fails
		for _, item := range order.OrderItems {
			productID, _ := strconv.Atoi(item.ProductID)
			s.inventoryClient.ReleaseStock(ctx, productID, item.Quantity, token)
		}
		return fmt.Errorf("payment failed: %w", err)
	}

	span.SetAttributes(attribute.String("payment_id", paymentID))
	order.Status = models.OrderStatusConfirmed // Payment successful

	// Create order in database
	if err := s.repo.Create(ctx, order); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create order")
		return err
	}

	// Invalidate user's orders cache
	cacheKey := fmt.Sprintf("user_orders:%s", order.UserID)
	s.cache.DeletePattern(ctx, cacheKey+"*")

	span.SetStatus(codes.Ok, "Order created successfully")
	return nil
}

func (s *OrderService) getProductWithRetry(ctx context.Context, productID string) (*ProductResponse, error) {
	ctx, span := tracer.Start(ctx, "getProductWithRetry")
	defer span.End()

	var lastErr error
	maxRetries := 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		product, err := s.getProduct(ctx, productID)
		if err == nil {
			span.SetAttributes(attribute.Int("retry_attempt", attempt))
			return product, nil
		}

		lastErr = err
		if attempt < maxRetries-1 {
			backoff := time.Duration(attempt+1) * time.Second
			time.Sleep(backoff)
		}
	}

	span.RecordError(lastErr)
	span.SetStatus(codes.Error, "All retry attempts failed")
	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

func (s *OrderService) getProduct(ctx context.Context, productID string) (*ProductResponse, error) {
	ctx, span := tracer.Start(ctx, "getProduct")
	defer span.End()

	url := fmt.Sprintf("%s/api/products/%s/", s.productServiceURL, productID) // ✅ Sondaki slash

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Inject tracing headers for distributed tracing
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("product service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("product service returned %d: %s", resp.StatusCode, string(body))
		span.RecordError(err)
		return nil, err
	}

	var product ProductResponse
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to decode product response: %w", err)
	}

	return &product, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, id string) (*models.Order, error) {
	ctx, span := tracer.Start(ctx, "GetOrderByID")
	defer span.End()

	span.SetAttributes(attribute.String("order_id", id))

	// Try cache first
	cacheKey := fmt.Sprintf("order:%s", id)
	var order models.Order
	if err := s.cache.Get(ctx, cacheKey, &order); err == nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		return &order, nil
	}

	span.SetAttributes(attribute.Bool("cache_hit", false))

	// Get from database
	orderPtr, err := s.repo.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Order not found")
		return nil, err
	}

	// Cache for 5 minutes
	s.cache.Set(ctx, cacheKey, orderPtr, 5*time.Minute)

	return orderPtr, nil
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID string, page, pageSize int) ([]models.Order, int64, error) {
	ctx, span := tracer.Start(ctx, "GetUserOrders")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID),
		attribute.Int("page", page),
		attribute.Int("page_size", pageSize),
	)

	offset := (page - 1) * pageSize

	// Try cache
	cacheKey := fmt.Sprintf("user_orders:%s:page:%d:size:%d", userID, page, pageSize)
	var cachedResult struct {
		Orders []models.Order `json:"orders"`
		Total  int64          `json:"total"`
	}

	if err := s.cache.Get(ctx, cacheKey, &cachedResult); err == nil {
		span.SetAttributes(attribute.Bool("cache_hit", true))
		return cachedResult.Orders, cachedResult.Total, nil
	}

	span.SetAttributes(attribute.Bool("cache_hit", false))

	orders, total, err := s.repo.GetByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		span.RecordError(err)
		return nil, 0, err
	}

	// Cache for 2 minutes
	cachedResult.Orders = orders
	cachedResult.Total = total
	s.cache.Set(ctx, cacheKey, cachedResult, 2*time.Minute)

	return orders, total, nil
}

func (s *OrderService) GetAllOrders(ctx context.Context, page, pageSize int) ([]models.Order, int64, error) {
	ctx, span := tracer.Start(ctx, "GetAllOrders")
	defer span.End()

	offset := (page - 1) * pageSize
	return s.repo.GetAll(ctx, pageSize, offset)
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, id string, status models.OrderStatus) error {
	ctx, span := tracer.Start(ctx, "UpdateOrderStatus")
	defer span.End()

	span.SetAttributes(
		attribute.String("order_id", id),
		attribute.String("new_status", string(status)),
	)

	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		span.RecordError(err)
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("order:%s", id)
	s.cache.Delete(ctx, cacheKey)

	return nil
}

func (s *OrderService) CancelOrder(ctx context.Context, id string) error {
	return s.UpdateOrderStatus(ctx, id, models.OrderStatusCancelled)
}
