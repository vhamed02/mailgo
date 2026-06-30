package handlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/mailgo/backend/internal/application"
	"github.com/mailgo/backend/internal/domain"
)

type MailboxHandler struct {
	service *application.MailboxService
}

func NewMailboxHandler(service *application.MailboxService) *MailboxHandler {
	return &MailboxHandler{service: service}
}

// ── Request / Response DTOs ───────────────────────────────────────────────────

type CreateMailboxRequestDTO struct {
	DomainID    string `json:"domain_id"    validate:"required,uuid"`
	LocalPart   string `json:"local_part"   validate:"required"`
	Password    string `json:"password"     validate:"required,min=8"`
	DisplayName string `json:"display_name" validate:"required"`
	QuotaBytes  int64  `json:"quota_bytes"`
}

type UpdateMailboxRequestDTO struct {
	DisplayName *string `json:"display_name"`
	Password    *string `json:"password"     validate:"omitempty,min=8"`
	QuotaBytes  *int64  `json:"quota_bytes"`
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (h *MailboxHandler) orgID(c echo.Context) (uuid.UUID, error) {
	v := c.Get("organization_id")
	if v == nil {
		return uuid.Nil, errors.New("missing organization context")
	}
	return v.(uuid.UUID), nil
}

func (h *MailboxHandler) userID(c echo.Context) (uuid.UUID, error) {
	v := c.Get("user_id")
	if v == nil {
		return uuid.Nil, errors.New("missing user context")
	}
	return v.(uuid.UUID), nil
}

func mailboxError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{"error": "mailbox not found"})
	case errors.Is(err, domain.ErrForbidden):
		return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
	case errors.Is(err, domain.ErrMailboxInUse):
		return c.JSON(http.StatusConflict, map[string]string{"error": "email address already exists"})
	case errors.Is(err, domain.ErrDomainNotVerified):
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "domain is not verified yet"})
	case errors.Is(err, domain.ErrMailboxLimitReached):
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "mailbox limit reached for this organisation"})
	case errors.Is(err, domain.ErrStorageLimitReached):
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "storage limit reached for this organisation"})
	default:
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// List returns all mailboxes for the authenticated organisation.
// GET /api/v1/mailboxes
func (h *MailboxHandler) List(c echo.Context) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	mailboxes, err := h.service.ListMailboxes(c.Request().Context(), orgID)
	if err != nil {
		return mailboxError(c, err)
	}

	if mailboxes == nil {
		mailboxes = []*domain.Mailbox{}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  mailboxes,
		"total": len(mailboxes),
	})
}

// Create creates a new mailbox.
// POST /api/v1/mailboxes
func (h *MailboxHandler) Create(c echo.Context) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	userID, err := h.userID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	var req CreateMailboxRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	domainID, err := uuid.Parse(req.DomainID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid domain_id"})
	}

	// Default quota: 1 GB
	quotaBytes := req.QuotaBytes
	if quotaBytes <= 0 {
		quotaBytes = 1 * 1024 * 1024 * 1024
	}

	mailbox, err := h.service.CreateMailbox(c.Request().Context(), application.CreateMailboxRequest{
		OrganizationID: orgID,
		DomainID:       domainID,
		LocalPart:      req.LocalPart,
		Password:       req.Password,
		DisplayName:    req.DisplayName,
		QuotaBytes:     quotaBytes,
		UserID:         userID,
	})
	if err != nil {
		return mailboxError(c, err)
	}

	return c.JSON(http.StatusCreated, mailbox)
}

// Get returns a single mailbox by ID.
// GET /api/v1/mailboxes/:id
func (h *MailboxHandler) Get(c echo.Context) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid mailbox id"})
	}

	mailbox, err := h.service.GetMailbox(c.Request().Context(), id, orgID)
	if err != nil {
		return mailboxError(c, err)
	}

	return c.JSON(http.StatusOK, mailbox)
}

// WebmailPassword returns the decrypted IMAP password for a mailbox owned by
// the authenticated organization. The webmail client uses this to sign in
// automatically while the dashboard session is active.
// GET /api/v1/mailboxes/:id/webmail-password
func (h *MailboxHandler) WebmailPassword(c echo.Context) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid mailbox id"})
	}

	password, err := h.service.GetMailboxPassword(c.Request().Context(), id, orgID)
	if err != nil {
		return mailboxError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{"password": password})
}

// Update partially updates a mailbox (display name, password, quota).
// PATCH /api/v1/mailboxes/:id
func (h *MailboxHandler) Update(c echo.Context) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	userID, err := h.userID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid mailbox id"})
	}

	var req UpdateMailboxRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	mailbox, err := h.service.UpdateMailbox(c.Request().Context(), application.UpdateMailboxRequest{
		ID:          id,
		DisplayName: req.DisplayName,
		Password:    req.Password,
		QuotaBytes:  req.QuotaBytes,
		UserID:      userID,
	}, orgID)
	if err != nil {
		return mailboxError(c, err)
	}

	return c.JSON(http.StatusOK, mailbox)
}

// Suspend suspends a mailbox.
// POST /api/v1/mailboxes/:id/suspend
func (h *MailboxHandler) Suspend(c echo.Context) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	userID, err := h.userID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid mailbox id"})
	}

	if err := h.service.SuspendMailbox(c.Request().Context(), id, orgID, userID); err != nil {
		return mailboxError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "mailbox suspended"})
}

// Unsuspend reactivates a suspended mailbox.
// POST /api/v1/mailboxes/:id/unsuspend
func (h *MailboxHandler) Unsuspend(c echo.Context) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	userID, err := h.userID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid mailbox id"})
	}

	if err := h.service.UnsuspendMailbox(c.Request().Context(), id, orgID, userID); err != nil {
		return mailboxError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "mailbox unsuspended"})
}

// Delete soft-deletes a mailbox.
// DELETE /api/v1/mailboxes/:id
func (h *MailboxHandler) Delete(c echo.Context) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	userID, err := h.userID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid mailbox id"})
	}

	if err := h.service.DeleteMailbox(c.Request().Context(), id, orgID, userID); err != nil {
		return mailboxError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
