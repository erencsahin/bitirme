package service

import (
	"bytes"
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

type PaymentClient struct {
	baseURL    string
	httpClient *http.Client
	tracer     trace.Tracer
}

type CreatePaymentRequest struct {
	OrderID       string  `json:"order_id"`
	UserID        string  `json:"user_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method"`
}

type PaymentResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func NewPaymentClient(baseURL string) *PaymentClient {
	return &PaymentClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		tracer: otel.Tracer("payment-client"),
	}
}

func (c *PaymentClient) ProcessPayment(ctx context.Context, req CreatePaymentRequest, token string) (string, error) {
	ctx, span := c.tracer.Start(ctx, "payment.process")
	defer span.End()

	span.SetAttributes(
		attribute.String("order_id", req.OrderID),
		attribute.Float64("amount", req.Amount),
		attribute.String("currency", req.Currency),
	)

	url := fmt.Sprintf("%s/api/payments", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to process payment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("payment processing failed: %s", string(body))
	}

	var result PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return "", fmt.Errorf("payment failed: %s", result.Message)
	}

	// Extract payment ID from response
	paymentData, ok := result.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid payment response format")
	}

	paymentID, ok := paymentData["id"].(string)
	if !ok {
		return "", fmt.Errorf("payment ID not found in response")
	}

	return paymentID, nil
}

func (c *PaymentClient) GetPaymentByOrder(ctx context.Context, orderID string, token string) (map[string]interface{}, error) {
	ctx, span := c.tracer.Start(ctx, "payment.get_by_order")
	defer span.End()

	span.SetAttributes(attribute.String("order_id", orderID))

	url := fmt.Sprintf("%s/api/payments/order/%s", c.baseURL, orderID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("payment not found for order: %s", orderID)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get payment: %s", string(body))
	}

	var result PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	paymentData, ok := result.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid payment response format")
	}

	return paymentData, nil
}
