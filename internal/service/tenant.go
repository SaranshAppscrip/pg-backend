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

type TenantService struct {
	repos repository.Store
}

func NewTenantService(repos repository.Store) *TenantService {
	return &TenantService{repos: repos}
}

func (s *TenantService) List(ctx context.Context, orgID uuid.UUID) ([]domain.Tenant, error) {
	return s.repos.Tenants.List(ctx, orgID)
}

type CreateTenantInput struct {
	Name       string
	Email      string
	Password   string
	Phone      string
	RoomID     uuid.UUID
	MonthlyFee float64
	JoinDate   time.Time
}

func (s *TenantService) Create(ctx context.Context, orgID uuid.UUID, in CreateTenantInput) (*domain.Tenant, error) {
	email := strings.TrimSpace(strings.ToLower(in.Email))
	if in.Name == "" || email == "" || len(in.Password) < 6 || in.MonthlyFee < 0 {
		return nil, apperror.BadRequest("name, email, password (min 6 chars), and monthly_fee are required")
	}

	room, err := s.repos.Rooms.GetByID(ctx, orgID, in.RoomID)
	if err != nil {
		return nil, err
	}

	count, err := s.repos.Tenants.CountActiveInRoom(ctx, orgID, in.RoomID)
	if err != nil {
		return nil, err
	}
	if count >= room.Capacity {
		return nil, apperror.Conflict("room is at capacity")
	}

	var phone *string
	if in.Phone != "" {
		normalized := normalizePhone(in.Phone)
		phone = &normalized
	}

	tenant := &domain.Tenant{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           in.Name,
		Email:          email,
		PasswordHash:   hashPassword(in.Password),
		Phone:          phone,
		RoomID:         &in.RoomID,
		MonthlyFee:     in.MonthlyFee,
		JoinDate:       in.JoinDate,
		Active:         true,
		CreatedAt:      time.Now(),
	}

	if err := s.repos.Tenants.Create(ctx, tenant); err != nil {
		return nil, err
	}
	tenant.PasswordHash = ""
	return tenant, nil
}

func (s *TenantService) MoveOut(ctx context.Context, orgID, id uuid.UUID) (*domain.Tenant, error) {
	return s.repos.Tenants.MoveOut(ctx, orgID, id)
}

func normalizePhone(phone string) string {
	var b strings.Builder
	for _, r := range phone {
		if r >= '0' && r <= '9' || r == '+' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
