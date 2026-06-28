package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mailgo/backend/internal/application"
	"github.com/mailgo/backend/internal/domain"
)

type AuthHandler struct {
	authService *application.AuthService
}

func NewAuthHandler(authService *application.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type RegisterRequestDTO struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	OrgName   string `json:"org_name" validate:"required"`
	OrgSlug   string `json:"org_slug" validate:"required"`
}

type LoginRequestDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Register handles user registration
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	appReq := application.RegisterRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		OrgName:   req.OrgName,
		OrgSlug:   req.OrgSlug,
	}

	resp, err := h.authService.Register(c.Request().Context(), appReq)
	if err != nil {
		if err == domain.ErrAlreadyExists {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "email already registered",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "registration failed",
		})
	}

	return c.JSON(http.StatusCreated, resp)
}

// Login handles user authentication
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	appReq := application.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := h.authService.Login(c.Request().Context(), appReq)
	if err != nil {
		if err == domain.ErrInvalidCredentials {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid email or password",
			})
		}
		if err == domain.ErrUserSuspended {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "account suspended",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "login failed",
		})
	}

	return c.JSON(http.StatusOK, resp)
}

// Logout handles user logout
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c echo.Context) error {
	// In a stateless JWT system, logout is handled client-side
	// Here we can implement token blacklisting if needed
	return c.JSON(http.StatusOK, map[string]string{
		"message": "logged out successfully",
	})
}

// Me returns current user information
// GET /api/v1/auth/me
func (h *AuthHandler) Me(c echo.Context) error {
	// Extract user from context (set by auth middleware)
	userID := c.Get("user_id")
	orgID := c.Get("organization_id")
	role := c.Get("role")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id":         userID,
		"organization_id": orgID,
		"role":            role,
	})
}
