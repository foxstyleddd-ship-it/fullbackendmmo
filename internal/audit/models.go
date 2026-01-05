package audit

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID            uuid.UUID `db:"id" json:"id"`
	EventType     string    `db:"event_type" json:"event_type"`
	ActorID       uuid.UUID `db:"actor_id" json:"actor_id"`
	ActorType     string    `db:"actor_type" json:"actor_type"`
	TargetID      uuid.UUID `db:"target_id" json:"target_id"`
	TargetType    string    `db:"target_type" json:"target_type"`
	Action        string    `db:"action" json:"action"`
	Details       string    `db:"details" json:"details"`
	OldValue      *string   `db:"old_value" json:"old_value,omitempty"`
	NewValue      *string   `db:"new_value" json:"new_value,omitempty"`
	IPAddress     *string   `db:"ip_address" json:"ip_address,omitempty"`
	UserAgent     *string   `db:"user_agent" json:"user_agent,omitempty"`
	Success       bool      `db:"success" json:"success"`
	ErrorMessage  *string   `db:"error_message" json:"error_message,omitempty"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

// AuditLogRequest represents a request to create an audit log
type AuditLogRequest struct {
	EventType    string
	ActorID      uuid.UUID
	ActorType    string
	TargetID     uuid.UUID
	TargetType   string
	Action       string
	Details      string
	OldValue     *string
	NewValue     *string
	IPAddress    *string
	UserAgent    *string
	Success      bool
	ErrorMessage *string
}
