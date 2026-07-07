package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/repository"
)

type AuditService struct {
	repos repository.AuditRepository
}

func NewAuditService(repos repository.AuditRepository) *AuditService {
	return &AuditService{repos: repos}
}

func (s *AuditService) List(ctx context.Context, orgID uuid.UUID, limit int) ([]domain.StaffAuditLog, error) {
	return s.repos.List(ctx, orgID, limit)
}

func (s *AuditService) Log(
	ctx context.Context,
	orgID, staffID uuid.UUID,
	entityType domain.AuditEntityType,
	entityID uuid.UUID,
	action domain.AuditAction,
	metadata map[string]any,
) error {
	if metadata == nil {
		metadata = map[string]any{}
	}
	entry := &domain.StaffAuditLog{
		ID:         uuid.New(),
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		Metadata:   metadata,
		CreatedAt:  time.Now(),
	}
	var staffPtr *uuid.UUID
	if staffID != uuid.Nil {
		staffPtr = &staffID
	}
	return s.repos.Create(ctx, orgID, staffPtr, entry)
}
