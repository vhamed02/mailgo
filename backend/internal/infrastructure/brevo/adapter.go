package brevo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mailgo/backend/internal/domain"
)

type Adapter struct {
	apiURL     string
	apiKey     string
	fromEmail  string
	fromName   string
	httpClient *http.Client
}

func NewAdapter(apiURL, apiKey, fromEmail, fromName string, timeout time.Duration) *Adapter {
	return &Adapter{
		apiURL:     apiURL,
		apiKey:     apiKey,
		fromEmail:  fromEmail,
		fromName:   fromName,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (a *Adapter) RegisterDomain(ctx context.Context, domainName string) (*domain.EmailProviderDomainConfig, error) {
	payload := map[string]string{"name": domainName}
	respBody, err := a.makeRequestWithResponse(ctx, "POST", "/senders/domains", payload)
	if err != nil {
		return nil, err
	}

	var result struct {
		DNSRecords struct {
			BrevoCode struct {
				Value string `json:"value"`
			} `json:"brevo_code"`
			DKIM struct {
				Value string `json:"value"`
			} `json:"dkim_record"`
		} `json:"dns_records"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse domain registration response: %w", err)
	}

	return &domain.EmailProviderDomainConfig{
		SetupRecord:     result.DNSRecords.BrevoCode.Value,
		SetupDKIMRecord: result.DNSRecords.DKIM.Value,
		Verified:        false,
	}, nil
}

func (a *Adapter) GetDomainConfig(ctx context.Context, domainName string) (*domain.EmailProviderDomainConfig, error) {
	respBody, err := a.makeRequestWithResponse(ctx, "GET", "/senders/domains/"+domainName, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Verified   bool `json:"verified"`
		DNSRecords struct {
			BrevoCode struct {
				Value string `json:"value"`
			} `json:"brevo_code"`
			DKIM struct {
				Value string `json:"value"`
			} `json:"dkim_record"`
		} `json:"dns_records"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse domain config response: %w", err)
	}

	return &domain.EmailProviderDomainConfig{
		SetupRecord:     result.DNSRecords.BrevoCode.Value,
		SetupDKIMRecord: result.DNSRecords.DKIM.Value,
		Verified:        result.Verified,
	}, nil
}

func (a *Adapter) AuthenticateDomain(ctx context.Context, domainName string) error {
	_, err := a.makeRequestWithResponse(ctx, "PUT", "/senders/domains/"+domainName+"/authenticate", nil)
	return err
}

func (a *Adapter) SendTransactionalEmail(ctx context.Context, req domain.SendEmailRequest) error {
	to := make([]map[string]string, len(req.To))
	for i, email := range req.To {
		to[i] = map[string]string{"email": email}
	}
	from := map[string]string{"email": a.fromEmail, "name": a.fromName}
	if req.From != nil {
		from["email"] = req.From.Email
		from["name"] = req.From.Name
	}
	payload := map[string]interface{}{
		"sender":  from,
		"to":      to,
		"subject": req.Subject,
	}
	if req.IsHTML {
		payload["htmlContent"] = req.Body
	} else {
		payload["textContent"] = req.Body
	}
	return a.makeRequest(ctx, "POST", "/smtp/email", payload)
}

func (a *Adapter) SendWelcomeEmail(ctx context.Context, to, name string) error {
	return a.SendTransactionalEmail(ctx, domain.SendEmailRequest{
		To:      []string{to},
		Subject: "Welcome to MailGo!",
		Body:    fmt.Sprintf(`<html><body><h1>Welcome, %s!</h1><p>Your account is ready. Start by adding a domain.</p><p>— The MailGo Team</p></body></html>`, name),
		IsHTML:  true,
	})
}

func (a *Adapter) SendPasswordResetEmail(ctx context.Context, to, resetLink string) error {
	return a.SendTransactionalEmail(ctx, domain.SendEmailRequest{
		To:      []string{to},
		Subject: "Reset Your Password",
		Body:    fmt.Sprintf(`<html><body><h1>Password Reset</h1><p><a href="%s">Reset your password</a></p><p>Expires in 1 hour.</p></body></html>`, resetLink),
		IsHTML:  true,
	})
}

func (a *Adapter) SendVerificationEmail(ctx context.Context, to, verifyLink string) error {
	return a.SendTransactionalEmail(ctx, domain.SendEmailRequest{
		To:      []string{to},
		Subject: "Verify Your Email",
		Body:    fmt.Sprintf(`<html><body><h1>Verify Email</h1><p><a href="%s">Verify your email</a></p><p>Expires in 24 hours.</p></body></html>`, verifyLink),
		IsHTML:  true,
	})
}

func (a *Adapter) makeRequest(ctx context.Context, method, endpoint string, payload interface{}) error {
	_, err := a.makeRequestWithResponse(ctx, method, endpoint, payload)
	return err
}

func (a *Adapter) makeRequestWithResponse(ctx context.Context, method, endpoint string, payload interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, a.apiURL+endpoint, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("api-key", a.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("api error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}
