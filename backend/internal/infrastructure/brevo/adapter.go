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

// Adapter implements domain.EmailSenderAdapter for Brevo
// This adapter isolates all Brevo-specific logic from the business layer
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
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// SendTransactionalEmail sends a single transactional email via Brevo
func (a *Adapter) SendTransactionalEmail(ctx context.Context, req domain.SendEmailRequest) error {
	// Build recipients
	to := make([]map[string]string, len(req.To))
	for i, email := range req.To {
		to[i] = map[string]string{"email": email}
	}

	// Build sender
	from := map[string]string{
		"email": a.fromEmail,
		"name":  a.fromName,
	}
	if req.From != nil {
		from["email"] = req.From.Email
		from["name"] = req.From.Name
	}

	// Build Brevo API payload
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

// SendWelcomeEmail sends a welcome email to new users
func (a *Adapter) SendWelcomeEmail(ctx context.Context, to, name string) error {
	subject := "Welcome to MailGo!"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h1>Welcome to MailGo, %s!</h1>
			<p>Thank you for joining our email hosting platform.</p>
			<p>You can now start managing your domains and mailboxes.</p>
			<p>Best regards,<br>The MailGo Team</p>
		</body>
		</html>
	`, name)

	return a.SendTransactionalEmail(ctx, domain.SendEmailRequest{
		To:      []string{to},
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	})
}

// SendPasswordResetEmail sends a password reset email
func (a *Adapter) SendPasswordResetEmail(ctx context.Context, to, resetLink string) error {
	subject := "Reset Your Password"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h1>Password Reset Request</h1>
			<p>You requested to reset your password.</p>
			<p>Click the link below to reset your password:</p>
			<p><a href="%s">Reset Password</a></p>
			<p>This link will expire in 1 hour.</p>
			<p>If you didn't request this, please ignore this email.</p>
			<p>Best regards,<br>The MailGo Team</p>
		</body>
		</html>
	`, resetLink)

	return a.SendTransactionalEmail(ctx, domain.SendEmailRequest{
		To:      []string{to},
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	})
}

// SendVerificationEmail sends an email verification email
func (a *Adapter) SendVerificationEmail(ctx context.Context, to, verifyLink string) error {
	subject := "Verify Your Email Address"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h1>Email Verification</h1>
			<p>Please verify your email address to complete your registration.</p>
			<p>Click the link below to verify:</p>
			<p><a href="%s">Verify Email</a></p>
			<p>This link will expire in 24 hours.</p>
			<p>Best regards,<br>The MailGo Team</p>
		</body>
		</html>
	`, verifyLink)

	return a.SendTransactionalEmail(ctx, domain.SendEmailRequest{
		To:      []string{to},
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	})
}

// makeRequest is a helper method for making HTTP requests to Brevo API
func (a *Adapter) makeRequest(ctx context.Context, method, endpoint string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, a.apiURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("api-key", a.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("brevo API error: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
