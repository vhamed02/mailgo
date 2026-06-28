package domain

import (
	"time"

	"github.com/google/uuid"
)

// Organization represents a tenant in the multi-tenant system
type Organization struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	Status    OrgStatus  `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type OrgStatus string

const (
	OrgStatusActive    OrgStatus = "active"
	OrgStatusSuspended OrgStatus = "suspended"
	OrgStatusDeleted   OrgStatus = "deleted"
)

// User represents an authenticated user
type User struct {
	ID             uuid.UUID  `json:"id"`
	Email          string     `json:"email"`
	PasswordHash   string     `json:"-"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	EmailVerified  bool       `json:"email_verified"`
	Status         UserStatus `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty"`
}

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDeleted   UserStatus = "deleted"
)

// OrganizationUser represents the many-to-many relationship with roles
type OrganizationUser struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	UserID         uuid.UUID `json:"user_id"`
	Role           Role      `json:"role"`
	JoinedAt       time.Time `json:"joined_at"`
}

type Role string

const (
	RoleOwner Role = "owner"
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

// Domain represents an email domain
type Domain struct {
	ID                  uuid.UUID    `json:"id"`
	OrganizationID      uuid.UUID    `json:"organization_id"`
	Name                string       `json:"name"`
	Status              DomainStatus `json:"status"`
	DNSVerified         bool         `json:"dns_verified"`
	SPFRecord           string       `json:"spf_record"`
	DKIMRecord          string       `json:"dkim_record"`
	DMARCRecord         string       `json:"dmarc_record"`
	BrevoCodeValue      string       `json:"brevo_code_value"`
	BrevoDkim1Host      string       `json:"brevo_dkim1_host"`
	BrevoDkim1Value     string       `json:"brevo_dkim1_value"`
	BrevoDkim2Host      string       `json:"brevo_dkim2_host"`
	BrevoDkim2Value     string       `json:"brevo_dkim2_value"`
	BrevoDmarcValue     string       `json:"brevo_dmarc_value"`
	BrevoVerified       bool         `json:"brevo_verified"`
	BrevoAuthenticated  bool         `json:"brevo_authenticated"`
	VerifiedAt          *time.Time   `json:"verified_at,omitempty"`
	CreatedAt           time.Time    `json:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at"`
}

type DomainStatus string

const (
	DomainStatusPending    DomainStatus = "pending"
	DomainStatusDNSPending DomainStatus = "dns_pending"
	DomainStatusActive     DomainStatus = "active"
	DomainStatusInactive   DomainStatus = "inactive"
	DomainStatusFailed     DomainStatus = "failed"
)

// Mailbox represents an email account
type Mailbox struct {
	ID             uuid.UUID      `json:"id"`
	OrganizationID uuid.UUID      `json:"organization_id"`
	DomainID       uuid.UUID      `json:"domain_id"`
	Email          string         `json:"email"`
	LocalPart      string         `json:"local_part"`
	DisplayName    string         `json:"display_name"`
	Status         MailboxStatus  `json:"status"`
	QuotaBytes     int64          `json:"quota_bytes"`
	UsedBytes      int64          `json:"used_bytes"`
	PasswordHash   string         `json:"-"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	SuspendedAt    *time.Time     `json:"suspended_at,omitempty"`
}

type MailboxStatus string

const (
	MailboxStatusActive    MailboxStatus = "active"
	MailboxStatusSuspended MailboxStatus = "suspended"
	MailboxStatusDeleted   MailboxStatus = "deleted"
)

// Alias represents an email alias or forwarding rule
type Alias struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	DomainID       uuid.UUID `json:"domain_id"`
	Source         string    `json:"source"`
	Destination    string    `json:"destination"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
}

// Quota represents resource limits for an organization
type Quota struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	MaxDomains     int       `json:"max_domains"`
	MaxMailboxes   int       `json:"max_mailboxes"`
	MaxStorageGB   int       `json:"max_storage_gb"`
	MaxAliases     int       `json:"max_aliases"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID             uuid.UUID              `json:"id"`
	OrganizationID uuid.UUID              `json:"organization_id"`
	UserID         *uuid.UUID             `json:"user_id,omitempty"`
	Action         string                 `json:"action"`
	EntityType     string                 `json:"entity_type"`
	EntityID       *uuid.UUID             `json:"entity_id,omitempty"`
	Details        map[string]interface{} `json:"details"`
	IPAddress      string                 `json:"ip_address"`
	UserAgent      string                 `json:"user_agent"`
	CreatedAt      time.Time              `json:"created_at"`
}

// Common audit actions
const (
	ActionCreate   = "create"
	ActionUpdate   = "update"
	ActionDelete   = "delete"
	ActionSuspend  = "suspend"
	ActionActivate = "activate"
	ActionLogin    = "login"
	ActionLogout   = "logout"
)
