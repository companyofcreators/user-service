package app

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// AuthClient is an HTTP client for the auth-service internal API.
type AuthClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAuthClient creates a new AuthClient with the given auth-service base URL.
func NewAuthClient(baseURL string) *AuthClient {
	return &AuthClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// AddRole adds a role to a user in auth-service (idempotent).
// POST /internal/users/{userID}/roles/{role}
func (c *AuthClient) AddRole(ctx context.Context, userID, role string) error {
	url := fmt.Sprintf("%s/internal/users/%s/roles/%s", c.baseURL, userID, role)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth-service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth-service returned status %d", resp.StatusCode)
	}

	return nil
}

// RemoveRole removes a role from a user in auth-service (idempotent).
// DELETE /internal/users/{userID}/roles/{role}
func (c *AuthClient) RemoveRole(ctx context.Context, userID, role string) error {
	url := fmt.Sprintf("%s/internal/users/%s/roles/%s", c.baseURL, userID, role)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth-service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth-service returned status %d", resp.StatusCode)
	}

	return nil
}
