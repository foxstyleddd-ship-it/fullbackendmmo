package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	pkgredis "github.com/hp-mmo/backend/internal/pkg/redis"
)

// Service handles authentication business logic
type Service struct {
	repo       *Repository
	jwt        *JWTService
	redis      *pkgredis.Client
	sessionTTL time.Duration
}

// NewService creates a new auth service
func NewService(repo *Repository, jwt *JWTService, redis *pkgredis.Client, sessionTTL time.Duration) *Service {
	return &Service{
		repo:       repo,
		jwt:        jwt,
		redis:      redis,
		sessionTTL: sessionTTL,
	}
}

// Register creates a new account
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*Account, error) {
	// Validate email uniqueness
	emailExists, err := s.repo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if emailExists {
		return nil, fmt.Errorf("email already registered")
	}

	// Validate username uniqueness
	usernameExists, err := s.repo.UsernameExists(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if usernameExists {
		return nil, fmt.Errorf("username already taken")
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create account
	account := &Account{
		ID:           uuid.New(),
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(passwordHash),
		Role:         "player",
	}

	if req.DisplayName != "" {
		account.DisplayName.String = req.DisplayName
		account.DisplayName.Valid = true
	}

	if err := s.repo.CreateAccount(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return account, nil
}

// Login authenticates a user and returns tokens
func (s *Service) Login(ctx context.Context, req LoginRequest, ip string) (*LoginResponse, error) {
	// Get account
	account, err := s.repo.GetAccountByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if account is locked
	if account.LockedUntil.Valid && account.LockedUntil.Time.After(time.Now()) {
		return nil, fmt.Errorf("account locked until %s", account.LockedUntil.Time.Format(time.RFC3339))
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.Password)); err != nil {
		// Increment failed attempts
		_ = s.repo.IncrementFailedLoginAttempts(ctx, account.ID)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last login
	if err := s.repo.UpdateLastLogin(ctx, account.ID, ip); err != nil {
		return nil, fmt.Errorf("failed to update login info: %w", err)
	}

	// Create session
	sessionID := uuid.New().String()
	session := &Session{
		ID:        sessionID,
		AccountID: account.ID,
		ClientIP:  toNullString(ip),
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Store session in Redis
	sessionKey := fmt.Sprintf("session:%s", sessionID)
	if err := s.redis.SetWithExpiry(ctx, sessionKey, account.ID.String(), s.sessionTTL); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	// Generate tokens
	accessToken, err := s.jwt.GenerateAccessToken(account.ID, account.Role, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(account.ID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Build response
	response := &LoginResponse{
		AccountID:   account.ID,
		Username:    account.Username,
		Role:        account.Role,
		Characters:  []CharacterInfo{}, // Will be populated by handler
		Tokens: TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    int64(24 * time.Hour.Seconds()),
			TokenType:    "Bearer",
		},
	}

	if account.DisplayName.Valid {
		response.DisplayName = account.DisplayName.String
	}
	if account.LastLoginAt.Valid {
		lastLogin := account.LastLoginAt.Time
		response.LastLoginAt = &lastLogin
	}

	return response, nil
}

// RefreshToken refreshes an access token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	// Validate refresh token
	claims, err := s.jwt.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if refresh scope
	hasRefreshScope := false
	for _, scope := range claims.Scopes {
		if scope == "refresh" {
			hasRefreshScope = true
			break
		}
	}
	if !hasRefreshScope {
		return nil, fmt.Errorf("token is not a refresh token")
	}

	// Check session exists in Redis
	sessionKey := fmt.Sprintf("session:%s", claims.SessionID)
	exists, err := s.redis.Exists(ctx, sessionKey)
	if err != nil || !exists {
		return nil, fmt.Errorf("session expired or invalid")
	}

	// Parse account ID
	accountID, err := uuid.Parse(claims.AccountID)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID: %w", err)
	}

	// Generate new access token
	accessToken, err := s.jwt.GenerateAccessToken(accountID, claims.Role, claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate new refresh token
	newRefreshToken, err := s.jwt.GenerateRefreshToken(accountID, claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(24 * time.Hour.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// Logout invalidates a session
func (s *Service) Logout(ctx context.Context, sessionID string) error {
	// Remove from Redis
	sessionKey := fmt.Sprintf("session:%s", sessionID)
	if err := s.redis.Del(ctx, sessionKey).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Mark as disconnected in DB
	if err := s.repo.EndSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to end session: %w", err)
	}

	return nil
}

// ValidateSession checks if a session is valid
func (s *Service) ValidateSession(ctx context.Context, sessionID string) (bool, error) {
	sessionKey := fmt.Sprintf("session:%s", sessionID)
	exists, err := s.redis.Exists(ctx, sessionKey)
	if err != nil {
		return false, err
	}

	if exists {
		// Update activity
		_ = s.repo.UpdateSessionActivity(ctx, sessionID)
	}

	return exists, nil
}

// Helper functions
func toNullString(s string) NullString {
	if s == "" {
		return NullString{}
	}
	return NullString{String: s, Valid: true}
}

type NullString struct {
	String string
	Valid  bool
}
