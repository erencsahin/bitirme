package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type InventoryClient struct {
	baseURL    string
	httpClient *http.Client
	tracer     trace.Tracer
}

type CheckStockRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type ReserveStockRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type InventoryResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func NewInventoryClient(baseURL string) *InventoryClient {
	return &InventoryClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		tracer: otel.Tracer("inventory-client"),
	}
}

func (c *InventoryClient) CheckStock(ctx context.Context, productID int, quantity int, token string) (bool, error) {
	ctx, span := c.tracer.Start(ctx, "inventory.check_stock")
	defer span.End()

	span.SetAttributes(
		attribute.Int("product_id", productID),
		attribute.Int("quantity", quantity),
	)

	url := fmt.Sprintf("%s/api/inventory/product/%d/check?quantity=%d", c.baseURL, productID, quantity)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("failed to check stock: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("stock check failed: %s", string(body))
	}

	var result InventoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	// Data is boolean
	available, ok := result.Data.(bool)
	if !ok {
		return false, fmt.Errorf("invalid response format")
	}

	return available, nil
}

func (c *InventoryClient) ReserveStock(ctx context.Context, productID int, quantity int, token string) error {
	ctx, span := c.tracer.Start(ctx, "inventory.reserve_stock")
	defer span.End()

	span.SetAttributes(
		attribute.Int("product_id", productID),
		attribute.Int("quantity", quantity),
	)

	url := fmt.Sprintf("%s/api/inventory/product/%d/reserve?quantity=%d", c.baseURL, productID, quantity)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to reserve stock: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("stock reservation failed: %s", string(body))
	}

	return nil
}

func (c *InventoryClient) ReleaseStock(ctx context.Context, productID int, quantity int, token string) error {
	ctx, span := c.tracer.Start(ctx, "inventory.release_stock")
	defer span.End()

	span.SetAttributes(
		attribute.Int("product_id", productID),
		attribute.Int("quantity", quantity),
	)

	url := fmt.Sprintf("%s/api/inventory/product/%d/release?quantity=%d", c.baseURL, productID, quantity)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to release stock: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("stock release failed: %s", string(body))
	}

	return nil
}
