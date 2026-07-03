package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/repository"
	"github.com/nivas/server/pkg/apperror"
)

type SettingsService struct {
	repos repository.SettingsRepository
}

func NewSettingsService(repos repository.SettingsRepository) *SettingsService {
	return &SettingsService{repos: repos}
}

func (s *SettingsService) Get(ctx context.Context, orgID uuid.UUID) (map[string]string, error) {
	org, err := s.repos.Get(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return map[string]string{"pg_name": org.Name}, nil
}

func (s *SettingsService) Update(ctx context.Context, orgID uuid.UUID, pgName string) (map[string]string, error) {
	if pgName == "" {
		return nil, apperror.BadRequest("pg_name is required")
	}
	org, err := s.repos.UpdateName(ctx, orgID, pgName)
	if err != nil {
		return nil, err
	}
	return map[string]string{"pg_name": org.Name}, nil
}
