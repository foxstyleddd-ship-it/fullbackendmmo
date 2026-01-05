package audit

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service handles audit logging business logic
type Service struct {
	repo *Repository
}

// NewService creates a new audit service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// LogGradeChange logs a grade change event
func (s *Service) LogGradeChange(ctx context.Context, actorID, characterID uuid.UUID, oldGrade, newGrade int, ip, userAgent string) error {
	oldVal := fmt.Sprintf("%d", oldGrade)
	newVal := fmt.Sprintf("%d", newGrade)

	req := AuditLogRequest{
		EventType:  "grade_change",
		ActorID:    actorID,
		ActorType:  "account",
		TargetID:   characterID,
		TargetType: "character",
		Action:     "update_grade",
		Details:    fmt.Sprintf("Grade changed from %d to %d", oldGrade, newGrade),
		OldValue:   &oldVal,
		NewValue:   &newVal,
		IPAddress:  &ip,
		UserAgent:  &userAgent,
		Success:    true,
	}

	return s.repo.CreateAuditLog(ctx, req)
}

// LogTeleport logs a teleport event
func (s *Service) LogTeleport(ctx context.Context, actorID, characterID uuid.UUID, fromZone, toZone string, ip, userAgent string) error {
	req := AuditLogRequest{
		EventType:  "teleport",
		ActorID:    actorID,
		ActorType:  "account",
		TargetID:   characterID,
		TargetType: "character",
		Action:     "teleport",
		Details:    fmt.Sprintf("Teleported from %s to %s", fromZone, toZone),
		OldValue:   &fromZone,
		NewValue:   &toZone,
		IPAddress:  &ip,
		UserAgent:  &userAgent,
		Success:    true,
	}

	return s.repo.CreateAuditLog(ctx, req)
}

// LogItemGrant logs an item grant event
func (s *Service) LogItemGrant(ctx context.Context, actorID, characterID, itemID uuid.UUID, quantity int, ip, userAgent string) error {
	quantityStr := fmt.Sprintf("%d", quantity)

	req := AuditLogRequest{
		EventType:  "item_grant",
		ActorID:    actorID,
		ActorType:  "account",
		TargetID:   characterID,
		TargetType: "character",
		Action:     "grant_item",
		Details:    fmt.Sprintf("Granted %d of item %s", quantity, itemID.String()),
		NewValue:   &quantityStr,
		IPAddress:  &ip,
		UserAgent:  &userAgent,
		Success:    true,
	}

	return s.repo.CreateAuditLog(ctx, req)
}

// LogAdminAction logs a generic admin action
func (s *Service) LogAdminAction(ctx context.Context, actorID, targetID uuid.UUID, action, details string, success bool, errorMsg *string, ip, userAgent string) error {
	req := AuditLogRequest{
		EventType:    "admin_action",
		ActorID:      actorID,
		ActorType:    "account",
		TargetID:     targetID,
		TargetType:   "character",
		Action:       action,
		Details:      details,
		IPAddress:    &ip,
		UserAgent:    &userAgent,
		Success:      success,
		ErrorMessage: errorMsg,
	}

	return s.repo.CreateAuditLog(ctx, req)
}

// GetCharacterAuditLogs retrieves audit logs for a character
func (s *Service) GetCharacterAuditLogs(ctx context.Context, characterID uuid.UUID, limit int) ([]AuditLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repo.GetAuditLogsByTarget(ctx, characterID, limit)
}

// GetAdminAuditLogs retrieves audit logs for an admin user
func (s *Service) GetAdminAuditLogs(ctx context.Context, actorID uuid.UUID, limit int) ([]AuditLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repo.GetAuditLogsByActor(ctx, actorID, limit)
}
