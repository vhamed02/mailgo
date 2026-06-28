package handlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/mailgo/backend/internal/application"
	"github.com/mailgo/backend/internal/domain"
)

type DomainHandler struct {
	service *application.DomainService
}

func NewDomainHandler(service *application.DomainService) *DomainHandler {
	return &DomainHandler{service: service}
}

// ── Request / Response DTOs ───────────────────────────────────────────────────

type AddDomainRequestDTO struct {
	Name string `json:"name" validate:"required"`
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (h *DomainHandler) orgID(c echo.Context) (uuid.UUID, error) {
	v := c.Get("organization_id")
	if v == nil {
		return uuid.Nil, errors.New("missing organization context")
	}
	return v.(uuid.UUID), nil
}

func (h *DomainHandler) userID(c echo.Context) (uuid.UUID, error) {
	v := c.Get("user_id")
	if v == nil {
		return uuid.Nil, errors.New("missing user context")
	}
	return v.(uuid.UUID), nil
}

func domainError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrDomainInUse):
		return c.JSON(http.StatusConflict, map[string]string{"error": "domain already in use"})
	case errors.Is(err, domain.ErrInvalidDomainName):
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid domain name"})
	case errors.Is(err, domain.ErrDomainLimitReached):
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "domain limit reached for this organisation"})
	case errors.Is(err, domain.ErrForbidden):
		return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
	case errors.Is(err, domain.ErrNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{"error": "domain not found"})
	default:
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// List returns all domains for the authenticated organisation.
// GET /api/v1/domains
func (h *DomainHandler) List(c echo.Context) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	domains, err := h.service.ListDomains(c.Request().Context(), orgID)
	if err != nil {
		return domainError(c, err)
	}

	if domains == nil {
		domains = []*domain.Domain{}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  domains,
		"total": len(domains),
	})
}

// Create adds a new domain to the organisation.
// POST /api/v1/domains
func (h *DomainHandler) Create(c echo.Context) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	userID, err := h.userID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	var req AddDomainRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	d, err := h.service.AddDomain(c.Request().Context(), application.AddDomainRequest{
		OrganizationID: orgID,
		Name:           req.Name,
		UserID:         userID,
	})
	if err != nil {
		return domainError(c, err)
	}

	return c.JSON(http.StatusCreated, d)
}

// Get returns a single domain by ID.
// GET /api/v1/domains/:id
func (h *DomainHandler) Get(c echo.Context) error {
	orgID, err := h.orgID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid domain id"})
	}

	d, err := h.service.GetDomain(c.Request().Context(), id, orgID)
	if err != nil {
		return domainError(c, err)
	}

	return c.JSON(http.StatusOK, d)
}

// Verify triggers DNS verification for a domain.
// POST /api/v1/domains/:id/verify
func (h *DomainHandler) Verify(c echo.Context) error {
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
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid domain id"})
	}

	d, err := h.service.VerifyDomain(c.Request().Context(), id, orgID, userID)
	if err != nil {
		return domainError(c, err)
	}

	return c.JSON(http.StatusOK, d)
}

// Delete removes a domain.
// DELETE /api/v1/domains/:id
func (h *DomainHandler) Delete(c echo.Context) error {
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
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid domain id"})
	}

	if err := h.service.DeleteDomain(c.Request().Context(), id, orgID, userID); err != nil {
		return domainError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
