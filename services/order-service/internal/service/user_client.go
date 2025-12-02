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

type UserClient struct {
	baseURL    string
	httpClient *http.Client
	tracer     trace.Tracer
}

type ValidateTokenResponse struct {
	Status string `json:"status"`
	Data   struct {
		Valid  bool   `json:"valid"`
		UserID string `json:"userId"`
		Email  string `json:"email"`
	} `json:"data"`
}

func NewUserClient(baseURL string) *UserClient {
	return &UserClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		tracer: otel.Tracer("user-client"),
	}
}

func (c *UserClient) ValidateToken(ctx context.Context, token string) (bool, string, error) {
	ctx, span := c.tracer.Start(ctx, "user.validate_token")
	defer span.End()

	url := fmt.Sprintf("%s/api/auth/validate", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		span.RecordError(err)
		return false, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return false, "", fmt.Errorf("failed to validate token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		span.SetAttributes(attribute.Bool("valid", false))
		return false, "", fmt.Errorf("token validation failed: %s", string(body))
	}

	var result ValidateTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		span.RecordError(err)
		return false, "", fmt.Errorf("failed to decode response: %w", err)
	}

	span.SetAttributes(
		attribute.Bool("valid", result.Data.Valid),
		attribute.String("user_id", result.Data.UserID),
	)

	return result.Data.Valid, result.Data.UserID, nil
}
