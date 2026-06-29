package mailcow

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mailgo/backend/internal/domain"
)

// Adapter implements domain.MailServerAdapter for Mailcow
// This adapter isolates all Mailcow-specific logic from the business layer
type Adapter struct {
	apiURL     string
	apiKey     string
	httpClient *http.Client
}

func NewAdapter(apiURL, apiKey string, timeout time.Duration) *Adapter {
	return &Adapter{
		apiURL: apiURL,
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

// CreateMailbox provisions a mailbox on Mailcow
func (a *Adapter) CreateMailbox(ctx context.Context, req domain.CreateMailboxRequest) error {
	// Mailcow API request structure
	payload := map[string]interface{}{
		"local_part": extractLocalPart(req.Email),
		"domain":     extractDomain(req.Email),
		"name":       req.DisplayName,
		"password":   req.Password,
		"password2":  req.Password,
		"quota":      req.QuotaBytes / (1024 * 1024), // Convert to MB
		"active":     1,
	}

	return a.makeRequest(ctx, "POST", "/mailbox", payload, nil)
}

// UpdateMailbox updates mailbox properties on Mailcow
func (a *Adapter) UpdateMailbox(ctx context.Context, req domain.UpdateMailboxRequest) error {
	payload := map[string]interface{}{
		"items": []string{req.Email},
		"attr":  map[string]interface{}{},
	}

	attrs := payload["attr"].(map[string]interface{})

	if req.DisplayName != nil {
		attrs["name"] = *req.DisplayName
	}
	if req.Password != nil {
		attrs["password"] = *req.Password
		attrs["password2"] = *req.Password
	}
	if req.QuotaBytes != nil {
		attrs["quota"] = *req.QuotaBytes / (1024 * 1024) // Convert to MB
	}

	return a.makeRequest(ctx, "POST", "/mailbox", payload, nil)
}

// DeleteMailbox removes a mailbox from Mailcow
func (a *Adapter) DeleteMailbox(ctx context.Context, email string) error {
	payload := map[string]interface{}{
		"items": []string{email},
	}

	return a.makeRequest(ctx, "DELETE", "/mailbox", payload, nil)
}

// SuspendMailbox disables a mailbox temporarily
func (a *Adapter) SuspendMailbox(ctx context.Context, email string) error {
	payload := map[string]interface{}{
		"items": []string{email},
		"attr": map[string]interface{}{
			"active": 0,
		},
	}

	return a.makeRequest(ctx, "POST", "/mailbox", payload, nil)
}

// UnsuspendMailbox re-enables a suspended mailbox
func (a *Adapter) UnsuspendMailbox(ctx context.Context, email string) error {
	payload := map[string]interface{}{
		"items": []string{email},
		"attr": map[string]interface{}{
			"active": 1,
		},
	}

	return a.makeRequest(ctx, "POST", "/mailbox", payload, nil)
}

// CreateAlias creates an email alias on Mailcow
func (a *Adapter) CreateAlias(ctx context.Context, source, destination string) error {
	payload := map[string]interface{}{
		"address": source,
		"goto":    destination,
		"active":  1,
	}

	return a.makeRequest(ctx, "POST", "/alias", payload, nil)
}

// DeleteAlias removes an email alias from Mailcow
func (a *Adapter) DeleteAlias(ctx context.Context, source string) error {
	payload := map[string]interface{}{
		"items": []string{source},
	}

	return a.makeRequest(ctx, "DELETE", "/alias", payload, nil)
}

// GetMailboxStats retrieves mailbox usage statistics
func (a *Adapter) GetMailboxStats(ctx context.Context, email string) (*domain.MailboxStats, error) {
	var result map[string]interface{}
	
	err := a.makeRequest(ctx, "GET", fmt.Sprintf("/mailbox/%s", email), nil, &result)
	if err != nil {
		return nil, err
	}

	// Parse Mailcow response
	// Note: Actual response structure depends on Mailcow API version
	stats := &domain.MailboxStats{
		UsedBytes:    int64(result["quota_used"].(float64)) * 1024 * 1024, // Convert from MB
		QuotaBytes:   int64(result["quota"].(float64)) * 1024 * 1024,
		MessageCount: int(result["messages"].(float64)),
	}

	return stats, nil
}

// makeRequest is a helper method for making HTTP requests to Mailcow API
func (a *Adapter) makeRequest(ctx context.Context, method, endpoint string, payload interface{}, result interface{}) error {
	var body io.Reader
	
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, a.apiURL+endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", a.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mailcow API error: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Helper functions to extract email parts
func extractLocalPart(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		return parts[0]
	}
	return email
}

func extractDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}
