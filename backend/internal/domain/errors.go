package domain

import (
	"errors"
	"fmt"
)

// Common domain errors
var (
	// Entity errors
	ErrNotFound           = errors.New("entity not found")
	ErrAlreadyExists      = errors.New("entity already exists")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrForbidden          = errors.New("forbidden operation")
	
	// User errors
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserSuspended      = errors.New("user account suspended")
	ErrUserDeleted        = errors.New("user account deleted")
	ErrEmailNotVerified   = errors.New("email not verified")
	
	// Organization errors
	ErrOrgSuspended       = errors.New("organization suspended")
	ErrOrgDeleted         = errors.New("organization deleted")
	ErrNotOrgMember       = errors.New("not a member of organization")
	ErrInsufficientRole   = errors.New("insufficient role permissions")
	
	// Quota errors
	ErrQuotaExceeded      = errors.New("quota exceeded")
	ErrDomainLimitReached = errors.New("domain limit reached")
	ErrMailboxLimitReached = errors.New("mailbox limit reached")
	ErrStorageLimitReached = errors.New("storage limit reached")
	ErrAliasLimitReached  = errors.New("alias limit reached")
	
	// Domain errors
	ErrDomainNotVerified  = errors.New("domain not verified")
	ErrInvalidDomainName  = errors.New("invalid domain name")
	ErrDomainInUse        = errors.New("domain already in use")
	
	// Mailbox errors
	ErrMailboxSuspended   = errors.New("mailbox suspended")
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrMailboxInUse       = errors.New("mailbox already exists")
	
	// Infrastructure errors
	ErrMailServerFailure  = errors.New("mail server operation failed")
	ErrEmailSendFailure   = errors.New("email send failed")
	ErrCacheFailure       = errors.New("cache operation failed")
	ErrQueueFailure       = errors.New("queue operation failed")
)

// DomainError wraps errors with additional context
type DomainError struct {
	Op      string // Operation that failed
	Err     error  // Underlying error
	Message string // User-friendly message
}

func (e *DomainError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s (%v)", e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError creates a new domain error
func NewDomainError(op string, err error, message string) *DomainError {
	return &DomainError{
		Op:      op,
		Err:     err,
		Message: message,
	}
}
