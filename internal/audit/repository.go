package audit

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository handles audit log data access
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new audit repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateAuditLog creates a new audit log entry
func (r *Repository) CreateAuditLog(ctx context.Context, req AuditLogRequest) error {
	query := `
		INSERT INTO audit_logs (
			id, actor_account_id, actor_role, actor_ip,
			target_account_id, target_character_id,
			action, old_value, new_value, reason, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		uuid.New(),
		req.ActorAccountID,
		req.ActorRole,
		req.ActorIP,
		req.TargetAccountID,
		req.TargetCharacterID,
		req.Action,
		req.OldValue,
		req.NewValue,
		req.Reason,
		req.Metadata,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetAuditLogsByTarget retrieves audit logs for a specific target
func (r *Repository) GetAuditLogsByTarget(ctx context.Context, targetID uuid.UUID, limit int) ([]AuditLog, error) {
	query := `
		SELECT id, actor_account_id, actor_role, actor_ip,
		       target_account_id, target_character_id,
		       action, old_value, new_value, reason, metadata, created_at
		FROM audit_logs
		WHERE target_character_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	var logs []AuditLog
	err := r.db.SelectContext(ctx, &logs, query, targetID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return logs, nil
}

// GetAuditLogsByActor retrieves audit logs for a specific actor
func (r *Repository) GetAuditLogsByActor(ctx context.Context, actorID uuid.UUID, limit int) ([]AuditLog, error) {
	query := `
		SELECT id, actor_account_id, actor_role, actor_ip,
		       target_account_id, target_character_id,
		       action, old_value, new_value, reason, metadata, created_at
		FROM audit_logs
		WHERE actor_account_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	var logs []AuditLog
	err := r.db.SelectContext(ctx, &logs, query, actorID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return logs, nil
}

// GetAuditLogsByAction retrieves audit logs by action type
func (r *Repository) GetAuditLogsByAction(ctx context.Context, action string, limit int) ([]AuditLog, error) {
	query := `
		SELECT id, actor_account_id, actor_role, actor_ip,
		       target_account_id, target_character_id,
		       action, old_value, new_value, reason, metadata, created_at
		FROM audit_logs
		WHERE action = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	var logs []AuditLog
	err := r.db.SelectContext(ctx, &logs, query, action, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return logs, nil
}
