package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OrderClient is an HTTP client for the order-service internal API.
type OrderClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewOrderClient creates a new OrderClient with the given order-service base URL.
func NewOrderClient(baseURL string) *OrderClient {
	return &OrderClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// listOrdersResponse represents the order-service ListOrders JSON response.
type listOrdersResponse struct {
	Orders []interface{} `json:"orders"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

// HasActiveOrders checks if a user has any active orders (as customer or master).
// Active statuses: created, negotiation, assigned, in_progress.
// It calls the order-service internal endpoint: GET /internal/orders?user_id={id}&active=true
// Retries up to 3 times with exponential backoff (1s, 2s, 4s) on failure.
func (c *OrderClient) HasActiveOrders(ctx context.Context, userID string) (bool, error) {
	url := fmt.Sprintf("%s/internal/orders?user_id=%s&active=true&limit=1", c.baseURL, userID)

	var lastErr error
	backoff := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

	for attempt := 0; attempt <= len(backoff); attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return false, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoff[attempt-1]):
			}
		}

		result, err := c.doHasActiveOrders(ctx, url)
		if err == nil {
			return result, nil
		}

		lastErr = err
	}

	return false, fmt.Errorf("order-service request failed after %d retries: %w", len(backoff), lastErr)
}

// doHasActiveOrders performs a single HasActiveOrders HTTP call.
func (c *OrderClient) doHasActiveOrders(ctx context.Context, url string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("order-service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("order-service returned status %d", resp.StatusCode)
	}

	var result listOrdersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("decode response: %w", err)
	}

	return result.Total > 0, nil
}
