package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/repository"
)

type StaffService struct {
	repos repository.StaffRepository
	auth  *AuthService
}

func NewStaffService(repos repository.StaffRepository, auth *AuthService) *StaffService {
	return &StaffService{repos: repos, auth: auth}
}

func (s *StaffService) List(ctx context.Context, orgID uuid.UUID) ([]domain.StaffProfile, error) {
	return s.repos.List(ctx, orgID)
}

func (s *StaffService) Invite(ctx context.Context, orgID uuid.UUID, email, password string) (*domain.StaffProfile, error) {
	return s.auth.InviteStaff(ctx, orgID, email, password)
}

func (s *StaffService) Remove(ctx context.Context, orgID, id uuid.UUID) error {
	return s.repos.Delete(ctx, orgID, id)
}
