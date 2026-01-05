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
			id, event_type, actor_id, actor_type, target_id, target_type,
			action, details, old_value, new_value, ip_address, user_agent,
			success, error_message
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		uuid.New(),
		req.EventType,
		req.ActorID,
		req.ActorType,
		req.TargetID,
		req.TargetType,
		req.Action,
		req.Details,
		req.OldValue,
		req.NewValue,
		req.IPAddress,
		req.UserAgent,
		req.Success,
		req.ErrorMessage,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetAuditLogsByTarget retrieves audit logs for a specific target
func (r *Repository) GetAuditLogsByTarget(ctx context.Context, targetID uuid.UUID, limit int) ([]AuditLog, error) {
	query := `
		SELECT id, event_type, actor_id, actor_type, target_id, target_type,
		       action, details, old_value, new_value, ip_address, user_agent,
		       success, error_message, created_at
		FROM audit_logs
		WHERE target_id = $1
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
		SELECT id, event_type, actor_id, actor_type, target_id, target_type,
		       action, details, old_value, new_value, ip_address, user_agent,
		       success, error_message, created_at
		FROM audit_logs
		WHERE actor_id = $1
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

// GetAuditLogsByEventType retrieves audit logs by event type
func (r *Repository) GetAuditLogsByEventType(ctx context.Context, eventType string, limit int) ([]AuditLog, error) {
	query := `
		SELECT id, event_type, actor_id, actor_type, target_id, target_type,
		       action, details, old_value, new_value, ip_address, user_agent,
		       success, error_message, created_at
		FROM audit_logs
		WHERE event_type = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	var logs []AuditLog
	err := r.db.SelectContext(ctx, &logs, query, eventType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return logs, nil
}
