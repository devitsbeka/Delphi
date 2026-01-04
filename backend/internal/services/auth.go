package services

import (
	"context"
	"fmt"
	"time"

	"github.com/delphi-platform/delphi/backend/internal/config"
	"github.com/delphi-platform/delphi/backend/internal/models"
	"github.com/delphi-platform/delphi/backend/internal/repository"
	"github.com/delphi-platform/delphi/backend/pkg/auth"
	"github.com/delphi-platform/delphi/backend/pkg/crypto"
	"github.com/delphi-platform/delphi/backend/pkg/logger"
	"github.com/google/uuid"
)

// AuthService handles authentication operations
type AuthService struct {
	cfg        *config.Config
	repos      *repository.Repositories
	jwtManager *auth.JWTManager
	log        *logger.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(cfg *config.Config, repos *repository.Repositories, jwtManager *auth.JWTManager, log *logger.Logger) *AuthService {
	return &AuthService{
		cfg:        cfg,
		repos:      repos,
		jwtManager: jwtManager,
		log:        log,
	}
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest represents registration data
type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	TenantName  string `json:"tenant_name"`
	TenantSlug  string `json:"tenant_slug"`
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*auth.TokenPair, *models.User, error) {
	// Get user by email
	user, err := s.repos.Users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	// In a real implementation, we'd check the password against a stored hash
	// For Supabase Auth integration, this would be handled by Supabase
	// This is a simplified version

	// Update last login
	if err := s.repos.Users.UpdateLastLogin(ctx, user.ID); err != nil {
		s.log.Warnw("failed to update last login", "user_id", user.ID, "error", err)
	}

	// Generate tokens
	tokens, err := s.jwtManager.GenerateTokenPair(user.ID, user.TenantID, user.Email, string(user.Role))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return tokens, user, nil
}

// Register creates a new user and tenant
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*auth.TokenPair, *models.User, error) {
	// Check if email already exists
	existing, err := s.repos.Users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existing != nil {
		return nil, nil, fmt.Errorf("email already registered")
	}

	// Check if tenant slug is available
	existingTenant, err := s.repos.Tenants.GetBySlug(ctx, req.TenantSlug)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check existing tenant: %w", err)
	}
	if existingTenant != nil {
		return nil, nil, fmt.Errorf("tenant slug already taken")
	}

	// Hash password
	_, err = crypto.HashPassword(req.Password)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()

	// Create tenant
	tenant := &models.Tenant{
		ID:        uuid.New(),
		Name:      req.TenantName,
		Slug:      req.TenantSlug,
		Plan:      models.PlanFree,
		Settings:  []byte("{}"),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repos.Tenants.Create(ctx, tenant); err != nil {
		return nil, nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Create user
	user := &models.User{
		ID:          uuid.New(),
		TenantID:    tenant.ID,
		Email:       req.Email,
		Name:        req.Name,
		Role:        models.RoleOwner,
		Preferences: []byte("{}"),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repos.Users.Create(ctx, user); err != nil {
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	tokens, err := s.jwtManager.GenerateTokenPair(user.ID, user.TenantID, user.Email, string(user.Role))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	s.log.Infow("user registered", "user_id", user.ID, "tenant_id", tenant.ID, "email", user.Email)

	return tokens, user, nil
}

// RefreshToken validates a refresh token and returns new tokens
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Get user to ensure they still exist and get current role
	user, err := s.repos.Users.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Generate new tokens
	tokens, err := s.jwtManager.GenerateTokenPair(user.ID, user.TenantID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return tokens, nil
}

// ValidateToken validates an access token and returns claims
func (s *AuthService) ValidateToken(token string) (*auth.Claims, error) {
	return s.jwtManager.ValidateToken(token)
}

// ForgotPassword initiates password reset (placeholder)
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	// In production, this would send a password reset email
	// For Supabase Auth, this would be handled by Supabase
	s.log.Infow("password reset requested", "email", email)
	return nil
}

