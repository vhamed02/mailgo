package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mailgo/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type MailboxService struct {
	mailboxRepo   domain.MailboxRepository
	domainRepo    domain.DomainRepository
	quotaRepo     domain.QuotaRepository
	auditRepo     domain.AuditLogRepository
	mailServer    domain.MailServerAdapter
	queue         domain.QueueAdapter
}

func NewMailboxService(
	mailboxRepo domain.MailboxRepository,
	domainRepo domain.DomainRepository,
	quotaRepo domain.QuotaRepository,
	auditRepo domain.AuditLogRepository,
	mailServer domain.MailServerAdapter,
	queue domain.QueueAdapter,
) *MailboxService {
	return &MailboxService{
		mailboxRepo: mailboxRepo,
		domainRepo:  domainRepo,
		quotaRepo:   quotaRepo,
		auditRepo:   auditRepo,
		mailServer:  mailServer,
		queue:       queue,
	}
}

type CreateMailboxRequest struct {
	OrganizationID uuid.UUID
	DomainID       uuid.UUID
	LocalPart      string
	Password       string
	DisplayName    string
	QuotaBytes     int64
	UserID         uuid.UUID // For audit
}

type UpdateMailboxRequest struct {
	ID          uuid.UUID
	DisplayName *string
	Password    *string
	QuotaBytes  *int64
	UserID      uuid.UUID // For audit
}

// CreateMailbox creates a new mailbox (control plane) and provisions it (infrastructure)
func (s *MailboxService) CreateMailbox(ctx context.Context, req CreateMailboxRequest) (*domain.Mailbox, error) {
	// Get domain
	dom, err := s.domainRepo.GetByID(ctx, req.DomainID)
	if err != nil {
		return nil, domain.NewDomainError("CreateMailbox", err, "domain not found")
	}

	// Verify domain belongs to organization
	if dom.OrganizationID != req.OrganizationID {
		return nil, domain.ErrForbidden
	}

	// Check if domain is verified
	if !dom.DNSVerified {
		return nil, domain.ErrDomainNotVerified
	}

	// Check quota
	quota, err := s.quotaRepo.GetByOrganization(ctx, req.OrganizationID)
	if err != nil {
		if err == domain.ErrNotFound {
			if createErr := s.quotaRepo.CreateDefault(ctx, req.OrganizationID); createErr != nil {
				return nil, createErr
			}
			quota, err = s.quotaRepo.GetByOrganization(ctx, req.OrganizationID)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	currentCount, err := s.mailboxRepo.CountByOrganization(ctx, req.OrganizationID)
	if err != nil {
		return nil, err
	}

	if currentCount >= quota.MaxMailboxes {
		return nil, domain.ErrMailboxLimitReached
	}

	// Construct full email
	email := fmt.Sprintf("%s@%s", req.LocalPart, dom.Name)

	// Check if mailbox already exists
	existing, _ := s.mailboxRepo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, domain.ErrMailboxInUse
	}

	// Hash password for storage
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create mailbox entity
	mailbox := &domain.Mailbox{
		ID:             uuid.New(),
		OrganizationID: req.OrganizationID,
		DomainID:       req.DomainID,
		Email:          email,
		LocalPart:      req.LocalPart,
		DisplayName:    req.DisplayName,
		Status:         domain.MailboxStatusActive,
		QuotaBytes:     req.QuotaBytes,
		UsedBytes:      0,
		PasswordHash:   string(passwordHash),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Save to database (control plane)
	if err := s.mailboxRepo.Create(ctx, mailbox); err != nil {
		return nil, err
	}

	// Enqueue async provisioning on mail server (infrastructure)
	// This decouples control plane from infrastructure timing/failures
	if err := s.queue.EnqueueMailboxProvision(ctx, mailbox.ID, mailbox.Email, req.Password); err != nil {
		fmt.Printf("Failed to enqueue mailbox provision: %v\n", err)
	}

	_ = s.queue.EnqueueEmailSend(ctx, domain.SendEmailRequest{
		To:      []string{mailbox.Email},
		Subject: "Your new mailbox is ready",
		Body: fmt.Sprintf(`<html><body>
<h2>Mailbox created!</h2>
<p>Your mailbox <strong>%s</strong> has been created and is being provisioned.</p>
<p>It will be ready within a few minutes.</p>
<p>— The MailGo Team</p>
</body></html>`, mailbox.Email),
		IsHTML: true,
	})

	// Audit log
	s.auditRepo.Create(ctx, &domain.AuditLog{
		ID:             uuid.New(),
		OrganizationID: req.OrganizationID,
		UserID:         &req.UserID,
		Action:         domain.ActionCreate,
		EntityType:     "mailbox",
		EntityID:       &mailbox.ID,
		Details: map[string]interface{}{
			"email": email,
		},
		CreatedAt: time.Now(),
	})

	return mailbox, nil
}

// GetMailbox retrieves a mailbox by ID
func (s *MailboxService) GetMailbox(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.Mailbox, error) {
	mailbox, err := s.mailboxRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if mailbox.OrganizationID != orgID {
		return nil, domain.ErrForbidden
	}

	return mailbox, nil
}

// ListMailboxes lists all mailboxes for an organization
func (s *MailboxService) ListMailboxes(ctx context.Context, orgID uuid.UUID) ([]*domain.Mailbox, error) {
	return s.mailboxRepo.ListByOrganization(ctx, orgID)
}

// UpdateMailbox updates a mailbox
func (s *MailboxService) UpdateMailbox(ctx context.Context, req UpdateMailboxRequest, orgID uuid.UUID) (*domain.Mailbox, error) {
	mailbox, err := s.mailboxRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if mailbox.OrganizationID != orgID {
		return nil, domain.ErrForbidden
	}

	// Update fields
	if req.DisplayName != nil {
		mailbox.DisplayName = *req.DisplayName
	}
	if req.QuotaBytes != nil {
		mailbox.QuotaBytes = *req.QuotaBytes
	}
	if req.Password != nil {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		mailbox.PasswordHash = string(passwordHash)
	}

	mailbox.UpdatedAt = time.Now()

	// Update in database
	if err := s.mailboxRepo.Update(ctx, mailbox); err != nil {
		return nil, err
	}

	// Update on mail server (sync operation via adapter)
	// Note: In production, consider making this async for better resilience
	updateReq := domain.UpdateMailboxRequest{
		Email:       mailbox.Email,
		DisplayName: req.DisplayName,
		QuotaBytes:  req.QuotaBytes,
	}
	if req.Password != nil {
		updateReq.Password = req.Password
	}

	if err := s.mailServer.UpdateMailbox(ctx, updateReq); err != nil {
		// Log but don't fail - can be retried
		fmt.Printf("Failed to update mailbox on mail server: %v\n", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, &domain.AuditLog{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         &req.UserID,
		Action:         domain.ActionUpdate,
		EntityType:     "mailbox",
		EntityID:       &mailbox.ID,
		CreatedAt:      time.Now(),
	})

	return mailbox, nil
}

// SuspendMailbox suspends a mailbox
func (s *MailboxService) SuspendMailbox(ctx context.Context, id uuid.UUID, orgID uuid.UUID, userID uuid.UUID) error {
	mailbox, err := s.mailboxRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if mailbox.OrganizationID != orgID {
		return domain.ErrForbidden
	}

	mailbox.Status = domain.MailboxStatusSuspended
	now := time.Now()
	mailbox.SuspendedAt = &now
	mailbox.UpdatedAt = now

	if err := s.mailboxRepo.Update(ctx, mailbox); err != nil {
		return err
	}

	// Suspend on mail server
	if err := s.mailServer.SuspendMailbox(ctx, mailbox.Email); err != nil {
		fmt.Printf("Failed to suspend mailbox on mail server: %v\n", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, &domain.AuditLog{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         &userID,
		Action:         domain.ActionSuspend,
		EntityType:     "mailbox",
		EntityID:       &mailbox.ID,
		CreatedAt:      time.Now(),
	})

	return nil
}

// UnsuspendMailbox reactivates a suspended mailbox
func (s *MailboxService) UnsuspendMailbox(ctx context.Context, id uuid.UUID, orgID uuid.UUID, userID uuid.UUID) error {
	mailbox, err := s.mailboxRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if mailbox.OrganizationID != orgID {
		return domain.ErrForbidden
	}

	mailbox.Status = domain.MailboxStatusActive
	mailbox.SuspendedAt = nil
	mailbox.UpdatedAt = time.Now()

	if err := s.mailboxRepo.Update(ctx, mailbox); err != nil {
		return err
	}

	// Unsuspend on mail server
	if err := s.mailServer.UnsuspendMailbox(ctx, mailbox.Email); err != nil {
		fmt.Printf("Failed to unsuspend mailbox on mail server: %v\n", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, &domain.AuditLog{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         &userID,
		Action:         domain.ActionActivate,
		EntityType:     "mailbox",
		EntityID:       &mailbox.ID,
		CreatedAt:      time.Now(),
	})

	return nil
}

// DeleteMailbox deletes a mailbox
func (s *MailboxService) DeleteMailbox(ctx context.Context, id uuid.UUID, orgID uuid.UUID, userID uuid.UUID) error {
	mailbox, err := s.mailboxRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if mailbox.OrganizationID != orgID {
		return domain.ErrForbidden
	}

	// Soft delete in database
	mailbox.Status = domain.MailboxStatusDeleted
	mailbox.UpdatedAt = time.Now()

	if err := s.mailboxRepo.Update(ctx, mailbox); err != nil {
		return err
	}

	// Enqueue async deletion on mail server
	if err := s.queue.EnqueueMailboxDelete(ctx, mailbox.Email); err != nil {
		fmt.Printf("Failed to enqueue mailbox deletion: %v\n", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, &domain.AuditLog{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         &userID,
		Action:         domain.ActionDelete,
		EntityType:     "mailbox",
		EntityID:       &mailbox.ID,
		Details: map[string]interface{}{
			"email": mailbox.Email,
		},
		CreatedAt: time.Now(),
	})

	return nil
}

// RefreshMailboxStats fetches latest usage stats from mail server
func (s *MailboxService) RefreshMailboxStats(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.Mailbox, error) {
	mailbox, err := s.mailboxRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if mailbox.OrganizationID != orgID {
		return nil, domain.ErrForbidden
	}

	// Get stats from mail server
	stats, err := s.mailServer.GetMailboxStats(ctx, mailbox.Email)
	if err != nil {
		return nil, domain.NewDomainError("RefreshMailboxStats", err, "failed to fetch stats from mail server")
	}

	// Update mailbox
	mailbox.UsedBytes = stats.UsedBytes
	mailbox.UpdatedAt = time.Now()

	if err := s.mailboxRepo.Update(ctx, mailbox); err != nil {
		return nil, err
	}

	return mailbox, nil
}
