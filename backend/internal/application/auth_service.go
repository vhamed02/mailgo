package application

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mailgo/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo    domain.UserRepository
	orgRepo     domain.OrganizationRepository
	orgUserRepo domain.OrganizationUserRepository
	quotaRepo   domain.QuotaRepository
	auditRepo   domain.AuditLogRepository
	jwtSecret   string
	jwtExpiry   time.Duration
}

func NewAuthService(
	userRepo domain.UserRepository,
	orgRepo domain.OrganizationRepository,
	orgUserRepo domain.OrganizationUserRepository,
	quotaRepo domain.QuotaRepository,
	auditRepo domain.AuditLogRepository,
	jwtSecret string,
	jwtExpiry time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		orgRepo:     orgRepo,
		orgUserRepo: orgUserRepo,
		quotaRepo:   quotaRepo,
		auditRepo:   auditRepo,
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

// Register creates a new user and organization
// NOTE: This should be wrapped in a database transaction in a future refactor
// For now, operations are executed sequentially with error handling
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Check if user already exists
	existing, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, domain.ErrAlreadyExists
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		println("ERROR: Password hashing failed:", err.Error())
		return nil, err
	}

	// Create user
	user := &domain.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Status:       domain.UserStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		println("ERROR: User creation failed:", err.Error())
		return nil, err
	}

	// Create organization
	org := &domain.Organization{
		ID:        uuid.New(),
		Name:      req.OrgName,
		Slug:      req.OrgSlug,
		Status:    domain.OrgStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.orgRepo.Create(ctx, org); err != nil {
		println("ERROR: Organization creation failed:", err.Error())
		// TODO: Rollback user creation in transaction
		return nil, err
	}

	// Create organization-user relationship (owner role)
	orgUser := &domain.OrganizationUser{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		UserID:         user.ID,
		Role:           domain.RoleOwner,
		JoinedAt:       time.Now(),
	}

	if err := s.orgUserRepo.Create(ctx, orgUser); err != nil {
		println("ERROR: Organization-user relationship creation failed:", err.Error())
		// TODO: Rollback user and org creation in transaction
		return nil, err
	}

	// Create default quota
	if err := s.quotaRepo.CreateDefault(ctx, org.ID); err != nil {
		println("ERROR: Quota creation failed:", err.Error())
		// TODO: Rollback all previous operations in transaction
		return nil, err
	}

	// Audit log (best effort - don't fail registration if this fails)
	if err := s.auditRepo.Create(ctx, &domain.AuditLog{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		UserID:         &user.ID,
		Action:         domain.ActionCreate,
		EntityType:     "user",
		EntityID:       &user.ID,
		CreatedAt:      time.Now(),
	}); err != nil {
		println("WARN: Audit log creation failed:", err.Error())
		// Continue - audit log failure should not block registration
	}

	println("INFO: Registration successful for", user.Email)

	// Generate tokens
	accessToken, expiresAt, err := s.generateToken(user.ID, org.ID, orgUser.Role)
	if err != nil {
		println("ERROR: Token generation failed:", err.Error())
		return nil, err
	}

	refreshToken, _, err := s.generateRefreshToken(user.ID)
	if err != nil {
		println("ERROR: Refresh token generation failed:", err.Error())
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
	accessToken, expiresAt, err := s.generateToken(user.ID, org.ID, orgUser.Role)
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
	jwt.RegisteredClaims
}

func (s *AuthService) generateToken(userID, orgID uuid.UUID, role domain.Role) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.jwtExpiry)
	
	claims := &Claims{
		UserID:         userID,
		OrganizationID: orgID,
		Role:           role,
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
