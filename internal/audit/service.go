package audit

import (
	"context"
	"encoding/json"
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
func (s *Service) LogGradeChange(ctx context.Context, actorID, characterID uuid.UUID, actorRole string, oldGrade, newGrade int, ip, userAgent string) error {
	oldVal, _ := json.Marshal(map[string]interface{}{"grade": oldGrade})
	newVal, _ := json.Marshal(map[string]interface{}{"grade": newGrade})

	metadata := map[string]interface{}{
		"user_agent": userAgent,
		"success":    true,
	}
	metadataJSON, _ := json.Marshal(metadata)

	reason := fmt.Sprintf("Grade changed from %d to %d", oldGrade, newGrade)

	req := AuditLogRequest{
		ActorAccountID:    actorID,
		ActorRole:         actorRole,
		ActorIP:           &ip,
		TargetCharacterID: &characterID,
		Action:            "update_grade",
		OldValue:          oldVal,
		NewValue:          newVal,
		Reason:            &reason,
		Metadata:          metadataJSON,
	}

	return s.repo.CreateAuditLog(ctx, req)
}

// LogTeleport logs a teleport event
func (s *Service) LogTeleport(ctx context.Context, actorID, characterID uuid.UUID, actorRole string, fromZone, toZone string, ip, userAgent string) error {
	oldVal, _ := json.Marshal(map[string]interface{}{"zone": fromZone})
	newVal, _ := json.Marshal(map[string]interface{}{"zone": toZone})

	metadata := map[string]interface{}{
		"user_agent": userAgent,
		"success":    true,
	}
	metadataJSON, _ := json.Marshal(metadata)

	reason := fmt.Sprintf("Teleported from %s to %s", fromZone, toZone)

	req := AuditLogRequest{
		ActorAccountID:    actorID,
		ActorRole:         actorRole,
		ActorIP:           &ip,
		TargetCharacterID: &characterID,
		Action:            "teleport",
		OldValue:          oldVal,
		NewValue:          newVal,
		Reason:            &reason,
		Metadata:          metadataJSON,
	}

	return s.repo.CreateAuditLog(ctx, req)
}

// LogItemGrant logs an item grant event
func (s *Service) LogItemGrant(ctx context.Context, actorID, characterID, itemID uuid.UUID, actorRole string, quantity int, ip, userAgent string) error {
	newVal, _ := json.Marshal(map[string]interface{}{
		"item_id":  itemID.String(),
		"quantity": quantity,
	})

	metadata := map[string]interface{}{
		"user_agent": userAgent,
		"success":    true,
	}
	metadataJSON, _ := json.Marshal(metadata)

	reason := fmt.Sprintf("Granted %d of item %s", quantity, itemID.String())

	req := AuditLogRequest{
		ActorAccountID:    actorID,
		ActorRole:         actorRole,
		ActorIP:           &ip,
		TargetCharacterID: &characterID,
		Action:            "grant_item",
		NewValue:          newVal,
		Reason:            &reason,
		Metadata:          metadataJSON,
	}

	return s.repo.CreateAuditLog(ctx, req)
}

// LogAdminAction logs a generic admin action
func (s *Service) LogAdminAction(ctx context.Context, actorID uuid.UUID, actorRole string, targetID *uuid.UUID, action, details string, success bool, errorMsg *string, ip, userAgent string) error {
	metadata := map[string]interface{}{
		"user_agent": userAgent,
		"success":    success,
	}
	if errorMsg != nil {
		metadata["error_message"] = *errorMsg
	}
	metadataJSON, _ := json.Marshal(metadata)

	req := AuditLogRequest{
		ActorAccountID:    actorID,
		ActorRole:         actorRole,
		ActorIP:           &ip,
		TargetCharacterID: targetID,
		Action:            action,
		Reason:            &details,
		Metadata:          metadataJSON,
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
