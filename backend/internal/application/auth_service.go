package application

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mailgo/backend/internal/domain"
	"github.com/mailgo/backend/internal/infrastructure/postgres"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo    domain.UserRepository
	orgRepo     domain.OrganizationRepository
	orgUserRepo domain.OrganizationUserRepository
	quotaRepo   domain.QuotaRepository
	auditRepo   domain.AuditLogRepository
	txManager   *postgres.TxManager
	queue       domain.QueueAdapter
	jwtSecret   string
	jwtExpiry   time.Duration
}

func NewAuthService(
	userRepo domain.UserRepository,
	orgRepo domain.OrganizationRepository,
	orgUserRepo domain.OrganizationUserRepository,
	quotaRepo domain.QuotaRepository,
	auditRepo domain.AuditLogRepository,
	txManager *postgres.TxManager,
	queue domain.QueueAdapter,
	jwtSecret string,
	jwtExpiry time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		orgRepo:     orgRepo,
		orgUserRepo: orgUserRepo,
		quotaRepo:   quotaRepo,
		auditRepo:   auditRepo,
		txManager:   txManager,
		queue:       queue,
		jwtSecret:   jwtSecret,
		jwtExpiry:   jwtExpiry,
	}
}

type RegisterRequest struct {
	Email        string
	Password     string
	FirstName    string
	LastName     string
	OrgName      string
	OrgSlug      string
}

type LoginRequest struct {
	Email    string
	Password string
}

type AuthResponse struct {
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	User         *domain.User   `json:"user"`
	Organization *domain.Organization `json:"organization,omitempty"`
	ExpiresAt    time.Time      `json:"expires_at"`
}

// Register creates a new user and organization atomically.
// All operations run inside a single database transaction — if any step fails
// the entire registration is rolled back, preventing orphaned records.
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Check if user already exists (outside transaction — read-only, safe to do early)
	existing, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, domain.ErrAlreadyExists
	}

	// Hash password before the transaction to avoid holding the tx open during CPU work
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var (
		user    *domain.User
		org     *domain.Organization
		orgUser *domain.OrganizationUser
	)

	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		user = &domain.User{
			ID:           uuid.New(),
			Email:        req.Email,
			PasswordHash: string(passwordHash),
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			Status:       domain.UserStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := s.userRepo.Create(txCtx, user); err != nil {
			return err
		}

		org = &domain.Organization{
			ID:        uuid.New(),
			Name:      req.OrgName,
			Slug:      req.OrgSlug,
			Status:    domain.OrgStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.orgRepo.Create(txCtx, org); err != nil {
			return err
		}

		orgUser = &domain.OrganizationUser{
			ID:             uuid.New(),
			OrganizationID: org.ID,
			UserID:         user.ID,
			Role:           domain.RoleOwner,
			JoinedAt:       time.Now(),
		}
		if err := s.orgUserRepo.Create(txCtx, orgUser); err != nil {
			return err
		}

		if err := s.quotaRepo.CreateDefault(txCtx, org.ID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Audit log — best effort, outside transaction so it never blocks registration
	_ = s.auditRepo.Create(ctx, &domain.AuditLog{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		UserID:         &user.ID,
		Action:         domain.ActionCreate,
		EntityType:     "user",
		EntityID:       &user.ID,
		CreatedAt:      time.Now(),
	})

	_ = s.queue.EnqueueEmailSend(ctx, domain.SendEmailRequest{
		To:      []string{user.Email},
		Subject: "Welcome to MailGo!",
		Body: fmt.Sprintf(`<html><body>
<h2>Welcome, %s!</h2>
<p>Your account and organisation <strong>%s</strong> are ready.</p>
<p>Start by adding a domain, then create mailboxes for your team.</p>
<p>— The MailGo Team</p>
</body></html>`, user.FirstName, org.Name),
		IsHTML: true,
	})

	// Generate tokens
	accessToken, expiresAt, err := s.generateToken(user.ID, org.ID, orgUser.Role, user.Email)
	if err != nil {
		return nil, err
	}
	refreshToken, _, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		Organization: org,
		ExpiresAt:    expiresAt,
	}, nil
}

// Login authenticates a user
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	// Get user
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Log the actual error for debugging
		println("ERROR: GetByEmail failed:", err.Error())
		return nil, domain.ErrInvalidCredentials
	}

	println("DEBUG: Found user:", user.Email, "with hash length:", len(user.PasswordHash))

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		println("ERROR: Password comparison failed:", err.Error())
		return nil, domain.ErrInvalidCredentials
	}

	println("DEBUG: Password verified successfully")

	// Check user status
	if user.Status == domain.UserStatusSuspended {
		return nil, domain.ErrUserSuspended
	}
	if user.Status == domain.UserStatusDeleted {
		return nil, domain.ErrUserDeleted
	}

	// Get user's primary organization
	orgUsers, err := s.orgUserRepo.ListByUser(ctx, user.ID)
	if err != nil || len(orgUsers) == 0 {
		println("ERROR: ListByUser failed or empty, err:", err, "count:", len(orgUsers))
		return nil, domain.ErrNotOrgMember
	}

	orgUser := orgUsers[0] // Use first organization
	org, err := s.orgRepo.GetByID(ctx, orgUser.OrganizationID)
	if err != nil {
		println("ERROR: GetByID for org failed:", err.Error())
		return nil, err
	}

	// Check organization status
	if org.Status == domain.OrgStatusSuspended {
		return nil, domain.ErrOrgSuspended
	}
	if org.Status == domain.OrgStatusDeleted {
		return nil, domain.ErrOrgDeleted
	}

	// Update last login
	now := time.Now()
	s.userRepo.UpdateLastLogin(ctx, user.ID, now)

	// Audit log
	s.auditRepo.Create(ctx, &domain.AuditLog{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		UserID:         &user.ID,
		Action:         domain.ActionLogin,
		EntityType:     "user",
		EntityID:       &user.ID,
		CreatedAt:      time.Now(),
	})

	// Generate tokens
	accessToken, expiresAt, err := s.generateToken(user.ID, org.ID, orgUser.Role, user.Email)
	if err != nil {
		println("ERROR: generateToken failed:", err.Error())
		return nil, err
	}

	refreshToken, _, err := s.generateRefreshToken(user.ID)
	if err != nil {
		println("ERROR: generateRefreshToken failed:", err.Error())
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		Organization: org,
		ExpiresAt:    expiresAt,
	}, nil
}

// JWT Claims
type Claims struct {
	UserID         uuid.UUID   `json:"user_id"`
	OrganizationID uuid.UUID   `json:"organization_id"`
	Role           domain.Role `json:"role"`
	Email          string      `json:"email"`
	jwt.RegisteredClaims
}

func (s *AuthService) generateToken(userID, orgID uuid.UUID, role domain.Role, email string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.jwtExpiry)

	claims := &Claims{
		UserID:         userID,
		OrganizationID: orgID,
		Role:           role,
		Email:          email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "mailgo",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func (s *AuthService) generateRefreshToken(userID uuid.UUID) (string, time.Time, error) {
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
	
	claims := &jwt.RegisteredClaims{
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "mailgo",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateToken validates and parses a JWT token
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, domain.ErrUnauthorized
}
