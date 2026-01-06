package audit

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID                uuid.UUID       `db:"id" json:"id"`
	ActorAccountID    uuid.UUID       `db:"actor_account_id" json:"actor_account_id"`
	ActorRole         string          `db:"actor_role" json:"actor_role"`
	ActorIP           *string         `db:"actor_ip" json:"actor_ip,omitempty"`
	TargetAccountID   *uuid.UUID      `db:"target_account_id" json:"target_account_id,omitempty"`
	TargetCharacterID *uuid.UUID      `db:"target_character_id" json:"target_character_id,omitempty"`
	Action            string          `db:"action" json:"action"`
	OldValue          json.RawMessage `db:"old_value" json:"old_value,omitempty"`
	NewValue          json.RawMessage `db:"new_value" json:"new_value,omitempty"`
	Reason            *string         `db:"reason" json:"reason,omitempty"`
	Metadata          json.RawMessage `db:"metadata" json:"metadata"`
	CreatedAt         time.Time       `db:"created_at" json:"created_at"`
}

// AuditLogRequest represents a request to create an audit log
type AuditLogRequest struct {
	ActorAccountID    uuid.UUID
	ActorRole         string
	ActorIP           *string
	TargetAccountID   *uuid.UUID
	TargetCharacterID *uuid.UUID
	Action            string
	OldValue          json.RawMessage
	NewValue          json.RawMessage
	Reason            *string
	Metadata          json.RawMessage
}
