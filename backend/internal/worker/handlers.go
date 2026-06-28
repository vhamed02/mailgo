package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/mailgo/backend/internal/domain"
)

// Task type constants
const (
	TypeMailboxProvision   = "mailbox:provision"
	TypeMailboxDelete      = "mailbox:delete"
	TypeDomainVerification = "domain:verify"
	TypeEmailSend          = "email:send"
	TypeAuditLogProcess    = "audit:process"
)

type Handlers struct {
	mailboxRepo   domain.MailboxRepository
	domainRepo    domain.DomainRepository
	mailServer    domain.MailServerAdapter
	emailSender   domain.EmailSenderAdapter
}

func NewHandlers(
	mailboxRepo domain.MailboxRepository,
	domainRepo domain.DomainRepository,
	mailServer domain.MailServerAdapter,
	emailSender domain.EmailSenderAdapter,
) *Handlers {
	return &Handlers{
		mailboxRepo:   mailboxRepo,
		domainRepo:    domainRepo,
		mailServer:    mailServer,
		emailSender:   emailSender,
	}
}

// Mailbox Provision Payload
type MailboxProvisionPayload struct {
	MailboxID uuid.UUID `json:"mailbox_id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
}

// HandleMailboxProvision provisions a mailbox on the mail server
func (h *Handlers) HandleMailboxProvision(ctx context.Context, task *asynq.Task) error {
	var payload MailboxProvisionPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Info().
		Str("mailbox_id", payload.MailboxID.String()).
		Str("email", payload.Email).
		Msg("Provisioning mailbox")

	// Get mailbox from database for display name and quota
	mailbox, err := h.mailboxRepo.GetByID(ctx, payload.MailboxID)
	if err != nil {
		return fmt.Errorf("failed to get mailbox: %w", err)
	}

	// Provision on mail server using the plaintext password from the job payload
	req := domain.CreateMailboxRequest{
		Email:       mailbox.Email,
		Password:    payload.Password,
		DisplayName: mailbox.DisplayName,
		QuotaBytes:  mailbox.QuotaBytes,
	}

	if err := h.mailServer.CreateMailbox(ctx, req); err != nil {
		log.Error().
			Err(err).
			Str("email", mailbox.Email).
			Msg("Failed to provision mailbox on mail server")
		return fmt.Errorf("failed to provision mailbox: %w", err)
	}

	log.Info().
		Str("email", mailbox.Email).
		Msg("Mailbox provisioned successfully")

	return nil
}

// Mailbox Delete Payload
type MailboxDeletePayload struct {
	Email string `json:"email"`
}

// HandleMailboxDelete removes a mailbox from the mail server
func (h *Handlers) HandleMailboxDelete(ctx context.Context, task *asynq.Task) error {
	var payload MailboxDeletePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Info().
		Str("email", payload.Email).
		Msg("Deleting mailbox")

	// Delete from mail server
	if err := h.mailServer.DeleteMailbox(ctx, payload.Email); err != nil {
		log.Error().
			Err(err).
			Str("email", payload.Email).
			Msg("Failed to delete mailbox from mail server")
		return fmt.Errorf("failed to delete mailbox: %w", err)
	}

	log.Info().
		Str("email", payload.Email).
		Msg("Mailbox deleted successfully")

	return nil
}

// Domain Verification Payload
type DomainVerificationPayload struct {
	DomainID uuid.UUID `json:"domain_id"`
}

// HandleDomainVerification verifies DNS records for a domain
func (h *Handlers) HandleDomainVerification(ctx context.Context, task *asynq.Task) error {
	var payload DomainVerificationPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Info().
		Str("domain_id", payload.DomainID.String()).
		Msg("Verifying domain DNS records")

	// Get domain from database
	domain, err := h.domainRepo.GetByID(ctx, payload.DomainID)
	if err != nil {
		return fmt.Errorf("failed to get domain: %w", err)
	}

	// TODO: Implement actual DNS verification
	// For now, just log
	log.Info().
		Str("domain", domain.Name).
		Msg("DNS verification not yet implemented")

	return nil
}

// Email Send Payload
type EmailSendPayload struct {
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	IsHTML  bool     `json:"is_html"`
}

// HandleEmailSend sends an email via the email sender adapter
func (h *Handlers) HandleEmailSend(ctx context.Context, task *asynq.Task) error {
	var payload EmailSendPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Info().
		Strs("to", payload.To).
		Str("subject", payload.Subject).
		Msg("Sending email")

	req := domain.SendEmailRequest{
		To:      payload.To,
		Subject: payload.Subject,
		Body:    payload.Body,
		IsHTML:  payload.IsHTML,
	}

	if err := h.emailSender.SendTransactionalEmail(ctx, req); err != nil {
		log.Error().
			Err(err).
			Strs("to", payload.To).
			Msg("Failed to send email")
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Info().
		Strs("to", payload.To).
		Msg("Email sent successfully")

	return nil
}

// Audit Log Process Payload
type AuditLogProcessPayload struct {
	LogID uuid.UUID `json:"log_id"`
}

// HandleAuditLogProcess processes audit log entries (e.g., for analytics)
func (h *Handlers) HandleAuditLogProcess(ctx context.Context, task *asynq.Task) error {
	var payload AuditLogProcessPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Info().
		Str("log_id", payload.LogID.String()).
		Msg("Processing audit log")

	// TODO: Implement audit log processing
	// Could involve: aggregation, alerting, external logging service, etc.

	return nil
}
