package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/auth"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/repository"
	"github.com/nivas/server/pkg/apperror"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repos  repository.Store
	tokens *auth.TokenService
}

func NewAuthService(repos repository.Store, tokens *auth.TokenService) *AuthService {
	return &AuthService{repos: repos, tokens: tokens}
}

func (s *AuthService) StaffLogin(ctx context.Context, orgID uuid.UUID, email, password string) (*domain.AuthResponse, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if orgID == uuid.Nil || email == "" || len(password) < 6 {
		return nil, apperror.BadRequest("organization_id, email, and password (min 6 chars) are required")
	}

	staff, err := s.repos.Staff.GetByEmailAndOrg(ctx, orgID, email)
	if err != nil {
		return nil, apperror.Unauthorized("invalid email or password")
	}

	if bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash), []byte(password)) != nil {
		return nil, apperror.Unauthorized("invalid email or password")
	}

	token, err := s.tokens.GenerateStaffToken(staff)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token: token,
		User:  staffToProfile(staff),
	}, nil
}

func (s *AuthService) StaffMe(ctx context.Context, orgID, staffID uuid.UUID) (*domain.StaffProfile, error) {
	staff, err := s.repos.Staff.GetByID(ctx, orgID, staffID)
	if err != nil {
		return nil, err
	}
	return staffToProfile(staff), nil
}

func (s *AuthService) TenantLogin(ctx context.Context, orgID uuid.UUID, email, password string) (*domain.TenantAuthResponse, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if orgID == uuid.Nil || email == "" || len(password) < 6 {
		return nil, apperror.BadRequest("organization_id, email, and password (min 6 chars) are required")
	}

	tenant, err := s.repos.Tenants.GetByEmailAndOrg(ctx, orgID, email)
	if err != nil {
		return nil, apperror.Unauthorized("invalid email or password")
	}
	if !tenant.Active {
		return nil, apperror.Unauthorized("account is inactive")
	}

	if bcrypt.CompareHashAndPassword([]byte(tenant.PasswordHash), []byte(password)) != nil {
		return nil, apperror.Unauthorized("invalid email or password")
	}

	token, err := s.tokens.GenerateTenantToken(tenant)
	if err != nil {
		return nil, err
	}

	profile, err := s.repos.Tenants.GetProfile(ctx, orgID, tenant.ID)
	if err != nil {
		return nil, err
	}

	return &domain.TenantAuthResponse{Token: token, User: profile}, nil
}

func (s *AuthService) TenantMe(ctx context.Context, orgID, tenantID uuid.UUID) (*domain.TenantProfile, error) {
	return s.repos.Tenants.GetProfile(ctx, orgID, tenantID)
}

func (s *AuthService) InviteStaff(ctx context.Context, orgID uuid.UUID, email, password string) (*domain.StaffProfile, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || len(password) < 6 {
		return nil, apperror.BadRequest("email and password (min 6 chars) are required")
	}

	staff := &domain.Staff{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Email:          email,
		PasswordHash:   hashPassword(password),
		IsOwner:        false,
	}

	if err := s.repos.Staff.Create(ctx, staff); err != nil {
		return nil, err
	}

	return staffToProfile(staff), nil
}

func hashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func staffToProfile(staff *domain.Staff) *domain.StaffProfile {
	return &domain.StaffProfile{
		ID:             staff.ID,
		OrganizationID: staff.OrganizationID,
		Email:          staff.Email,
		FullName:       staff.FullName,
		IsOwner:        staff.IsOwner,
		CreatedAt:      staff.CreatedAt,
	}
}
