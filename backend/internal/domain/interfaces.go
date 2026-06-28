package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository Interfaces (Ports for Data Layer)

type OrganizationRepository interface {
	Create(ctx context.Context, org *Organization) error
	GetByID(ctx context.Context, id uuid.UUID) (*Organization, error)
	GetBySlug(ctx context.Context, slug string) (*Organization, error)
	List(ctx context.Context, userID uuid.UUID) ([]*Organization, error)
	Update(ctx context.Context, org *Organization) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLogin time.Time) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type OrganizationUserRepository interface {
	Create(ctx context.Context, orgUser *OrganizationUser) error
	GetByOrganizationAndUser(ctx context.Context, orgID, userID uuid.UUID) (*OrganizationUser, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*OrganizationUser, error)
	ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*OrganizationUser, error)
	Update(ctx context.Context, orgUser *OrganizationUser) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type DomainRepository interface {
	Create(ctx context.Context, domain *Domain) error
	GetByID(ctx context.Context, id uuid.UUID) (*Domain, error)
	GetByName(ctx context.Context, name string) (*Domain, error)
	ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*Domain, error)
	Update(ctx context.Context, domain *Domain) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByOrganization(ctx context.Context, orgID uuid.UUID) (int, error)
}

type MailboxRepository interface {
	Create(ctx context.Context, mailbox *Mailbox) error
	GetByID(ctx context.Context, id uuid.UUID) (*Mailbox, error)
	GetByEmail(ctx context.Context, email string) (*Mailbox, error)
	ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*Mailbox, error)
	ListByDomain(ctx context.Context, domainID uuid.UUID) ([]*Mailbox, error)
	Update(ctx context.Context, mailbox *Mailbox) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByOrganization(ctx context.Context, orgID uuid.UUID) (int, error)
}

type AliasRepository interface {
	Create(ctx context.Context, alias *Alias) error
	GetByID(ctx context.Context, id uuid.UUID) (*Alias, error)
	GetBySource(ctx context.Context, source string) (*Alias, error)
	ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*Alias, error)
	ListByDomain(ctx context.Context, domainID uuid.UUID) ([]*Alias, error)
	Delete(ctx context.Context, id uuid.UUID) error
	CountByOrganization(ctx context.Context, orgID uuid.UUID) (int, error)
}

type QuotaRepository interface {
	GetByOrganization(ctx context.Context, orgID uuid.UUID) (*Quota, error)
	CreateDefault(ctx context.Context, orgID uuid.UUID) error
	Update(ctx context.Context, quota *Quota) error
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	ListByOrganization(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*AuditLog, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*AuditLog, error)
}

// Infrastructure Adapter Interfaces (Ports for External Systems)

// MailServerAdapter abstracts mail server operations (e.g., Mailcow)
// This interface ensures business logic never depends on specific mail server implementation
type MailServerAdapter interface {
	// CreateMailbox provisions a mailbox on the mail server
	CreateMailbox(ctx context.Context, req CreateMailboxRequest) error
	
	// UpdateMailbox updates mailbox properties (quota, password)
	UpdateMailbox(ctx context.Context, req UpdateMailboxRequest) error
	
	// DeleteMailbox removes a mailbox from the mail server
	DeleteMailbox(ctx context.Context, email string) error
	
	// SuspendMailbox disables a mailbox temporarily
	SuspendMailbox(ctx context.Context, email string) error
	
	// UnsuspendMailbox re-enables a suspended mailbox
	UnsuspendMailbox(ctx context.Context, email string) error
	
	// CreateAlias creates an email alias
	CreateAlias(ctx context.Context, source, destination string) error
	
	// DeleteAlias removes an email alias
	DeleteAlias(ctx context.Context, source string) error
	
	// GetMailboxStats retrieves mailbox usage statistics
	GetMailboxStats(ctx context.Context, email string) (*MailboxStats, error)
}

type CreateMailboxRequest struct {
	Email       string
	Password    string
	DisplayName string
	QuotaBytes  int64
}

type UpdateMailboxRequest struct {
	Email       string
	Password    *string // Optional: only update if provided
	DisplayName *string // Optional: only update if provided
	QuotaBytes  *int64  // Optional: only update if provided
}

type MailboxStats struct {
	UsedBytes int64
	QuotaBytes int64
	MessageCount int
}

type EmailSenderAdapter interface {
	SendTransactionalEmail(ctx context.Context, req SendEmailRequest) error
	SendWelcomeEmail(ctx context.Context, to, name string) error
	SendPasswordResetEmail(ctx context.Context, to, resetLink string) error
	SendVerificationEmail(ctx context.Context, to, verifyLink string) error
}

type EmailProviderAdapter interface {
	RegisterDomain(ctx context.Context, domainName string) (*EmailProviderDomainConfig, error)
	AuthenticateDomain(ctx context.Context, domainName string) error
	GetDomainConfig(ctx context.Context, domainName string) (*EmailProviderDomainConfig, error)
}

type EmailProviderDomainConfig struct {
	SetupRecord     string
	SetupDKIMRecord string
	Verified        bool
}

type SendEmailRequest struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
	From    *EmailAddress // Optional: use default if not provided
}

type EmailAddress struct {
	Email string
	Name  string
}

// CacheAdapter abstracts caching operations (e.g., Redis)
type CacheAdapter interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// QueueAdapter abstracts async job queue operations (e.g., Asynq)
type QueueAdapter interface {
	EnqueueMailboxProvision(ctx context.Context, mailboxID uuid.UUID, email, password string) error
	EnqueueMailboxDelete(ctx context.Context, email string) error
	EnqueueDomainSetup(ctx context.Context, domainID uuid.UUID) error
	EnqueueDomainVerification(ctx context.Context, domainID uuid.UUID, userEmail string) error
	EnqueueEmailSend(ctx context.Context, req SendEmailRequest) error
	EnqueueAuditLogProcess(ctx context.Context, logID uuid.UUID) error
}
