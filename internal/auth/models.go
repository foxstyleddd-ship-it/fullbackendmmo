package auth

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Account represents a user account
type Account struct {
	ID                   uuid.UUID      `db:"id" json:"id"`
	Email                string         `db:"email" json:"email"`
	EmailVerified        bool           `db:"email_verified" json:"email_verified"`
	PasswordHash         string         `db:"password_hash" json:"-"`
	Username             string         `db:"username" json:"username"`
	DisplayName          sql.NullString `db:"display_name" json:"display_name,omitempty"`
	Role                 string         `db:"role" json:"role"`
	CreatedAt            time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time      `db:"updated_at" json:"updated_at"`
	LastLoginAt          sql.NullTime   `db:"last_login_at" json:"last_login_at,omitempty"`
	LastLoginIP          sql.NullString `db:"last_login_ip" json:"last_login_ip,omitempty"`
	FailedLoginAttempts  int            `db:"failed_login_attempts" json:"-"`
	LockedUntil          sql.NullTime   `db:"locked_until" json:"-"`
	TwoFactorEnabled     bool           `db:"two_factor_enabled" json:"two_factor_enabled"`
	TwoFactorSecret      sql.NullString `db:"two_factor_secret" json:"-"`
	DeletedAt            sql.NullTime   `db:"deleted_at" json:"-"`
}

// Session represents an active session
type Session struct {
	ID             string         `db:"id" json:"session_id"`
	AccountID      uuid.UUID      `db:"account_id" json:"account_id"`
	CharacterID    uuid.NullUUID  `db:"character_id" json:"character_id,omitempty"`
	ZoneID         sql.NullString `db:"zone_id" json:"zone_id,omitempty"`
	ZoneShard      sql.NullString `db:"zone_shard" json:"zone_shard,omitempty"`
	ConnectedAt    time.Time      `db:"connected_at" json:"connected_at"`
	LastActivityAt time.Time      `db:"last_activity_at" json:"last_activity_at"`
	DisconnectedAt sql.NullTime   `db:"disconnected_at" json:"disconnected_at,omitempty"`
	ClientIP       sql.NullString `db:"client_ip" json:"client_ip,omitempty"`
	ClientVersion  sql.NullString `db:"client_version" json:"client_version,omitempty"`
	Platform       sql.NullString `db:"platform" json:"platform,omitempty"`
	Metadata       []byte         `db:"metadata" json:"metadata,omitempty"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Username    string `json:"username" binding:"required,min=3,max=32"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name,omitempty"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// TokenResponse represents the authentication token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	AccountID   uuid.UUID       `json:"account_id"`
	Username    string          `json:"username"`
	DisplayName string          `json:"display_name,omitempty"`
	Role        string          `json:"role"`
	LastLoginAt *time.Time      `json:"last_login_at,omitempty"`
	Tokens      TokenResponse   `json:"tokens"`
	Characters  []CharacterInfo `json:"characters"`
}

// CharacterInfo represents minimal character info for login response
type CharacterInfo struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	House        string     `json:"house"`
	Grade        int        `json:"grade"`
	Level        int        `json:"level"`
	LastPlayedAt *time.Time `json:"last_played_at,omitempty"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	AccountID   string   `json:"account_id"`
	CharacterID string   `json:"character_id,omitempty"`
	Role        string   `json:"role"`
	House       string   `json:"house,omitempty"`
	Grade       int      `json:"grade,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	SessionID   string   `json:"session_id"`
	Scopes      []string `json:"scopes"`
}

// NullUUID for handling nullable UUIDs
type NullUUID struct {
	uuid.UUID
	Valid bool
}

func (nu *NullUUID) Scan(value interface{}) error {
	if value == nil {
		nu.Valid = false
		return nil
	}
	nu.Valid = true
	return nu.UUID.Scan(value)
}
