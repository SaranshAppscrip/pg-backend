package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/password"
	"github.com/nivas/server/internal/repository"
	"github.com/nivas/server/pkg/apperror"
	"github.com/nivas/server/pkg/logger"
)

type TenantService struct {
	repos repository.Store
	audit *AuditService
}

func NewTenantService(repos repository.Store, audit *AuditService) *TenantService {
	return &TenantService{repos: repos, audit: audit}
}

func (s *TenantService) List(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID) ([]domain.Tenant, error) {
	return s.repos.Tenants.List(ctx, orgID, propertyID)
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

func (s *TenantService) Create(ctx context.Context, orgID, staffID uuid.UUID, in CreateTenantInput) (*domain.Tenant, error) {
	log := logger.FromContext(ctx)
	name := strings.TrimSpace(in.Name)
	email := strings.TrimSpace(strings.ToLower(in.Email))
	if name == "" || email == "" || len(in.Password) < 6 || in.MonthlyFee < 0 {
		return nil, apperror.BadRequest("name, email, password (min 6 chars), and monthly_fee are required")
	}

	if _, err := s.repos.Tenants.GetByEmailAndOrg(ctx, orgID, email); err == nil {
		log.Warn("tenant create rejected", "organization_id", orgID, "email", email, "reason", "duplicate_email")
		return nil, apperror.DuplicateEmail("Tenant")
	} else if !apperror.IsNotFound(err) {
		return nil, err
	}

	if _, err := s.repos.Tenants.GetByNameAndOrg(ctx, orgID, name); err == nil {
		log.Warn("tenant create rejected", "organization_id", orgID, "name", name, "reason", "duplicate_name")
		return nil, apperror.DuplicateName("Tenant")
	} else if !apperror.IsNotFound(err) {
		return nil, err
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
		return nil, apperror.RoomAtCapacity()
	}

	var phone *string
	if in.Phone != "" {
		normalized := normalizePhone(in.Phone)
		if _, err := s.repos.Tenants.GetByPhoneAndOrg(ctx, orgID, normalized); err == nil {
			log.Warn("tenant create rejected", "organization_id", orgID, "phone", normalized, "reason", "duplicate_phone")
			return nil, apperror.DuplicatePhone("Tenant")
		} else if !apperror.IsNotFound(err) {
			return nil, err
		}
		phone = &normalized
	}

	hash, err := password.Hash(in.Password)
	if err != nil {
		return nil, apperror.Internal("hash password", err)
	}

	tenant := &domain.Tenant{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           name,
		Email:          email,
		PasswordHash:   hash,
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
	log.Info("tenant created",
		"organization_id", orgID,
		"tenant_id", tenant.ID,
		"email", email,
		"room_id", in.RoomID,
	)
	meta := map[string]any{
		"name":        tenant.Name,
		"email":       tenant.Email,
		"monthly_fee": tenant.MonthlyFee,
		"room_id":     in.RoomID.String(),
		"join_date":   tenant.JoinDate.Format("2006-01-02"),
	}
	if tenant.Phone != nil {
		meta["phone"] = *tenant.Phone
	}
	_ = s.audit.Log(ctx, orgID, staffID, domain.AuditEntityTenant, tenant.ID, domain.AuditActionCreate, meta)
	return tenant, nil
}

func (s *TenantService) MoveOut(ctx context.Context, orgID, staffID, id uuid.UUID) (*domain.Tenant, error) {
	tenant, err := s.repos.Tenants.MoveOut(ctx, orgID, id)
	if err != nil {
		return nil, err
	}
	_ = s.audit.Log(ctx, orgID, staffID, domain.AuditEntityTenant, tenant.ID, domain.AuditActionMoveOut, map[string]any{
		"name":  tenant.Name,
		"email": tenant.Email,
	})
	return tenant, nil
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
