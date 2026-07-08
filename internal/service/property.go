package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/repository"
	"github.com/nivas/server/pkg/apperror"
)

type PropertyService struct {
	repos repository.PropertyRepository
}

func NewPropertyService(repos repository.PropertyRepository) *PropertyService {
	return &PropertyService{repos: repos}
}

func (s *PropertyService) List(ctx context.Context, orgID uuid.UUID) ([]domain.Property, error) {
	return s.repos.List(ctx, orgID)
}

func (s *PropertyService) Create(ctx context.Context, orgID uuid.UUID, name string, address *string) (*domain.Property, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, apperror.BadRequest("name is required")
	}
	if address != nil {
		trimmed := strings.TrimSpace(*address)
		if trimmed == "" {
			address = nil
		} else {
			address = &trimmed
		}
	}
	p := &domain.Property{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           name,
		Address:        address,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := s.repos.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PropertyService) Update(ctx context.Context, orgID, id uuid.UUID, name string, address *string) (*domain.Property, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, apperror.BadRequest("name is required")
	}
	if address != nil {
		trimmed := strings.TrimSpace(*address)
		if trimmed == "" {
			address = nil
		} else {
			address = &trimmed
		}
	}
	return s.repos.Update(ctx, orgID, id, name, address)
}
