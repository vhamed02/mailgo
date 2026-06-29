package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"

	"github.com/mailgo/backend/internal/domain"
)

const (
	TypeMailboxProvision   = "mailbox:provision"
	TypeMailboxDelete      = "mailbox:delete"
	TypeDomainSetup        = "domain:setup"
	TypeDomainVerification = "domain:verify"
	TypeEmailSend          = "email:send"
	TypeAuditLogProcess    = "audit:process"
)

type Handlers struct {
	mailboxRepo   domain.MailboxRepository
	domainRepo    domain.DomainRepository
	mailServer    domain.MailServerAdapter
	emailSender   domain.EmailSenderAdapter
	emailProvider domain.EmailProviderAdapter
	cache         domain.CacheAdapter
}

func NewHandlers(
	mailboxRepo domain.MailboxRepository,
	domainRepo domain.DomainRepository,
	mailServer domain.MailServerAdapter,
	emailSender domain.EmailSenderAdapter,
	emailProvider domain.EmailProviderAdapter,
	cache domain.CacheAdapter,
) *Handlers {
	return &Handlers{
		mailboxRepo:   mailboxRepo,
		domainRepo:    domainRepo,
		mailServer:    mailServer,
		emailSender:   emailSender,
		emailProvider: emailProvider,
		cache:         cache,
	}
}

type MailboxProvisionPayload struct {
	MailboxID   uuid.UUID `json:"mailbox_id"`
	Email       string    `json:"email"`
	PasswordKey string    `json:"password_key"`
	Password    string    `json:"password,omitempty"`
}

func (h *Handlers) HandleMailboxProvision(ctx context.Context, task *asynq.Task) error {
	var payload MailboxProvisionPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Info().Str("mailbox_id", payload.MailboxID.String()).Str("email", payload.Email).Msg("Provisioning mailbox")

	mailbox, err := h.mailboxRepo.GetByID(ctx, payload.MailboxID)
	if err != nil {
		return fmt.Errorf("failed to get mailbox: %w", err)
	}

	password := payload.Password
	if payload.PasswordKey != "" {
		password, err = h.cache.Get(ctx, payload.PasswordKey)
		if err != nil {
			return fmt.Errorf("failed to get mailbox password from cache: %w", err)
		}
	}
	if password == "" {
		return fmt.Errorf("mailbox password is empty")
	}

	req := domain.CreateMailboxRequest{
		Email:       mailbox.Email,
		Password:    password,
		DisplayName: mailbox.DisplayName,
		QuotaBytes:  mailbox.QuotaBytes,
	}

	if err := h.mailServer.CreateMailbox(ctx, req); err != nil {
		log.Error().Err(err).Str("email", mailbox.Email).Msg("Failed to provision mailbox")
		return fmt.Errorf("failed to provision mailbox: %w", err)
	}

	if payload.PasswordKey != "" {
		if err := h.cache.Delete(ctx, payload.PasswordKey); err != nil {
			log.Warn().Err(err).Str("email", mailbox.Email).Msg("Failed to delete cached mailbox password")
		}
	}

	log.Info().Str("email", mailbox.Email).Msg("Mailbox provisioned")
	return nil
}

type MailboxDeletePayload struct {
	Email string `json:"email"`
}

func (h *Handlers) HandleMailboxDelete(ctx context.Context, task *asynq.Task) error {
	var payload MailboxDeletePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if err := h.mailServer.DeleteMailbox(ctx, payload.Email); err != nil {
		log.Error().Err(err).Str("email", payload.Email).Msg("Failed to delete mailbox")
		return fmt.Errorf("failed to delete mailbox: %w", err)
	}

	log.Info().Str("email", payload.Email).Msg("Mailbox deleted")
	return nil
}

type DomainSetupPayload struct {
	DomainID uuid.UUID `json:"domain_id"`
}

func (h *Handlers) HandleDomainSetup(ctx context.Context, task *asynq.Task) error {
	var payload DomainSetupPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Info().Str("domain_id", payload.DomainID.String()).Msg("Setting up domain with email provider")

	dom, err := h.domainRepo.GetByID(ctx, payload.DomainID)
	if err != nil {
		return fmt.Errorf("failed to get domain: %w", err)
	}

	config, err := h.emailProvider.RegisterDomain(ctx, dom.Name)
	if err != nil {
		log.Error().Err(err).Str("domain", dom.Name).Msg("Failed to register domain with email provider")
		dom.Status = domain.DomainStatusFailed
		dom.UpdatedAt = time.Now()
		_ = h.domainRepo.Update(ctx, dom)
		return fmt.Errorf("failed to register domain: %w", err)
	}

	dom.BrevoCodeValue = config.BrevoCodeValue
	dom.BrevoDkim1Host = config.BrevoDkim1Host
	dom.BrevoDkim1Value = config.BrevoDkim1Value
	dom.BrevoDkim2Host = config.BrevoDkim2Host
	dom.BrevoDkim2Value = config.BrevoDkim2Value
	dom.BrevoDmarcValue = config.BrevoDmarcValue
	dom.Status = domain.DomainStatusDNSPending
	dom.UpdatedAt = time.Now()

	if err := h.domainRepo.Update(ctx, dom); err != nil {
		return fmt.Errorf("failed to update domain: %w", err)
	}

	log.Info().Str("domain", dom.Name).Msg("Domain registered with email provider, DNS records stored")
	return nil
}

type DomainVerificationPayload struct {
	DomainID  uuid.UUID `json:"domain_id"`
	UserEmail string    `json:"user_email"`
}

func (h *Handlers) HandleDomainVerification(ctx context.Context, task *asynq.Task) error {
	var payload DomainVerificationPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Info().Str("domain_id", payload.DomainID.String()).Msg("Verifying domain")

	dom, err := h.domainRepo.GetByID(ctx, payload.DomainID)
	if err != nil {
		return fmt.Errorf("failed to get domain: %w", err)
	}

	spfOK := checkSPF(dom.Name)

	providerOK := false
	if err := h.emailProvider.AuthenticateDomain(ctx, dom.Name); err != nil {
		log.Warn().Err(err).Str("domain", dom.Name).Msg("Email provider domain authentication failed")
	} else {
		providerOK = true
	}

	if spfOK && providerOK {
		now := time.Now()
		dom.DNSVerified = true
		dom.BrevoVerified = true
		dom.BrevoAuthenticated = true
		dom.Status = domain.DomainStatusActive
		dom.VerifiedAt = &now
		dom.UpdatedAt = now

		if err := h.domainRepo.Update(ctx, dom); err != nil {
			return fmt.Errorf("failed to update domain: %w", err)
		}

		if payload.UserEmail != "" {
			_ = h.emailSender.SendTransactionalEmail(ctx, domain.SendEmailRequest{
				To:      []string{payload.UserEmail},
				Subject: fmt.Sprintf("Domain %s is now active", dom.Name),
				Body:    fmt.Sprintf(`<html><body><h2>Domain verified!</h2><p>Your domain <strong>%s</strong> is now active and ready for use.</p><p>— The MailGo Team</p></body></html>`, dom.Name),
				IsHTML:  true,
			})
		}

		log.Info().Str("domain", dom.Name).Msg("Domain fully verified and active")
	} else {
		log.Warn().
			Str("domain", dom.Name).
			Bool("spf_ok", spfOK).
			Bool("provider_ok", providerOK).
			Msg("Domain verification incomplete — DNS records not yet propagated")
	}

	return nil
}

func checkSPF(domainName string) bool {
	txts, err := net.LookupTXT(domainName)
	if err != nil {
		return false
	}
	for _, txt := range txts {
		if strings.HasPrefix(txt, "v=spf1") {
			return true
		}
	}
	return false
}

type EmailSendPayload struct {
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	IsHTML  bool     `json:"is_html"`
}

func (h *Handlers) HandleEmailSend(ctx context.Context, task *asynq.Task) error {
	var payload EmailSendPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if len(payload.To) == 0 {
		return nil
	}

	log.Info().Strs("to", payload.To).Str("subject", payload.Subject).Msg("Sending email")

	if err := h.emailSender.SendTransactionalEmail(ctx, domain.SendEmailRequest{
		To:      payload.To,
		Subject: payload.Subject,
		Body:    payload.Body,
		IsHTML:  payload.IsHTML,
	}); err != nil {
		log.Error().Err(err).Strs("to", payload.To).Msg("Failed to send email")
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Info().Strs("to", payload.To).Msg("Email sent")
	return nil
}

type AuditLogProcessPayload struct {
	LogID uuid.UUID `json:"log_id"`
}

func (h *Handlers) HandleAuditLogProcess(ctx context.Context, task *asynq.Task) error {
	var payload AuditLogProcessPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	log.Info().Str("log_id", payload.LogID.String()).Msg("Processing audit log")
	return nil
}
