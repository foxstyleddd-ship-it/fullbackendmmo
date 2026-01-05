package auth

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository handles database operations for authentication
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new auth repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateAccount creates a new account
func (r *Repository) CreateAccount(ctx context.Context, account *Account) error {
	query := `
		INSERT INTO accounts (id, email, password_hash, username, display_name, role)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`

	return r.db.QueryRowContext(
		ctx, query,
		account.ID,
		account.Email,
		account.PasswordHash,
		account.Username,
		toNullString(account.DisplayName.String),
		account.Role,
	).Scan(&account.CreatedAt, &account.UpdatedAt)
}

// GetAccountByEmail retrieves an account by email
func (r *Repository) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	var account Account
	query := `
		SELECT id, email, email_verified, password_hash, username, display_name, role,
		       created_at, updated_at, last_login_at, last_login_ip, failed_login_attempts,
		       locked_until, two_factor_enabled, two_factor_secret, deleted_at
		FROM accounts
		WHERE email = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &account, query, email)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("account not found")
	}
	return &account, err
}

// GetAccountByID retrieves an account by ID
func (r *Repository) GetAccountByID(ctx context.Context, id uuid.UUID) (*Account, error) {
	var account Account
	query := `
		SELECT id, email, email_verified, password_hash, username, display_name, role,
		       created_at, updated_at, last_login_at, last_login_ip, failed_login_attempts,
		       locked_until, two_factor_enabled, two_factor_secret, deleted_at
		FROM accounts
		WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &account, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("account not found")
	}
	return &account, err
}

// EmailExists checks if an email is already registered
func (r *Repository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM accounts WHERE email = $1 AND deleted_at IS NULL)`
	err := r.db.GetContext(ctx, &exists, query, email)
	return exists, err
}

// UsernameExists checks if a username is already taken
func (r *Repository) UsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM accounts WHERE username = $1 AND deleted_at IS NULL)`
	err := r.db.GetContext(ctx, &exists, query, username)
	return exists, err
}

// UpdateLastLogin updates the last login timestamp and IP
func (r *Repository) UpdateLastLogin(ctx context.Context, accountID uuid.UUID, ip string) error {
	query := `
		UPDATE accounts
		SET last_login_at = NOW(),
		    last_login_ip = $2,
		    failed_login_attempts = 0,
		    updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, accountID, ip)
	return err
}

// IncrementFailedLoginAttempts increments failed login counter
func (r *Repository) IncrementFailedLoginAttempts(ctx context.Context, accountID uuid.UUID) error {
	query := `
		UPDATE accounts
		SET failed_login_attempts = failed_login_attempts + 1,
		    locked_until = CASE
		        WHEN failed_login_attempts >= 4 THEN NOW() + INTERVAL '15 minutes'
		        ELSE locked_until
		    END,
		    updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, accountID)
	return err
}

// CreateSession creates a new session record
func (r *Repository) CreateSession(ctx context.Context, session *Session) error {
	query := `
		INSERT INTO sessions (id, account_id, character_id, client_ip, client_version, platform, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING connected_at, last_activity_at`

	return r.db.QueryRowContext(
		ctx, query,
		session.ID,
		session.AccountID,
		toNullUUID(session.CharacterID),
		session.ClientIP,
		session.ClientVersion,
		session.Platform,
		session.Metadata,
	).Scan(&session.ConnectedAt, &session.LastActivityAt)
}

// GetSession retrieves a session by ID
func (r *Repository) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	var session Session
	query := `
		SELECT id, account_id, character_id, zone_id, zone_shard,
		       connected_at, last_activity_at, disconnected_at,
		       client_ip, client_version, platform, metadata
		FROM sessions
		WHERE id = $1 AND disconnected_at IS NULL`

	err := r.db.GetContext(ctx, &session, query, sessionID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	return &session, err
}

// UpdateSessionActivity updates the last activity timestamp
func (r *Repository) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	query := `UPDATE sessions SET last_activity_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, sessionID)
	return err
}

// EndSession marks a session as disconnected
func (r *Repository) EndSession(ctx context.Context, sessionID string) error {
	query := `UPDATE sessions SET disconnected_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, sessionID)
	return err
}

// Helper functions
func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func toNullUUID(nu uuid.NullUUID) interface{} {
	if !nu.Valid {
		return nil
	}
	return nu.UUID
}
