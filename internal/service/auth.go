package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/auth"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/password"
	"github.com/nivas/server/internal/repository"
	"github.com/nivas/server/pkg/apperror"
	"github.com/nivas/server/pkg/logger"
)

type AuthService struct {
	repos  repository.Store
	tokens *auth.TokenService
}

func NewAuthService(repos repository.Store, tokens *auth.TokenService) *AuthService {
	return &AuthService{repos: repos, tokens: tokens}
}

func (s *AuthService) StaffLogin(ctx context.Context, orgID uuid.UUID, email, plainPassword string) (*domain.AuthResponse, error) {
	log := logger.FromContext(ctx)
	email = strings.TrimSpace(strings.ToLower(email))
	if orgID == uuid.Nil || email == "" || len(plainPassword) < 6 {
		log.Warn("staff login validation failed", "organization_id", orgID, "email", email)
		return nil, apperror.BadRequest("organization_id, email, and password (min 6 chars) are required")
	}

	staff, err := s.repos.Staff.GetByEmailAndOrg(ctx, orgID, email)
	if err != nil {
		log.Warn("staff login failed", "organization_id", orgID, "email", email, "reason", "invalid_credentials")
		return nil, apperror.Unauthorized("invalid email or password")
	}

	fmt.Printf("staff.PasswordHash: %s\n", staff.PasswordHash)
	fmt.Printf("plainPassword: %s\n", plainPassword)

	if password.Compare(staff.PasswordHash, plainPassword) != nil {
		log.Warn("staff login failed", "organization_id", orgID, "email", email, "reason", "invalid_credentials")
		return nil, apperror.Unauthorized("error comparing passwords")
	}

	token, err := s.tokens.GenerateStaffToken(staff)
	if err != nil {
		return nil, err
	}

	log.Info("staff login succeeded",
		"organization_id", orgID,
		"user_id", staff.ID,
		"email", email,
		"is_owner", staff.IsOwner,
	)

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

func (s *AuthService) TenantLogin(ctx context.Context, orgID uuid.UUID, email, plainPassword string) (*domain.TenantAuthResponse, error) {
	log := logger.FromContext(ctx)
	email = strings.TrimSpace(strings.ToLower(email))
	if orgID == uuid.Nil || email == "" || len(plainPassword) < 6 {
		log.Warn("tenant login validation failed", "organization_id", orgID, "email", email)
		return nil, apperror.BadRequest("organization_id, email, and password (min 6 chars) are required")
	}

	tenant, err := s.repos.Tenants.GetByEmailAndOrg(ctx, orgID, email)
	if err != nil {
		log.Warn("tenant login failed", "organization_id", orgID, "email", email, "reason", "invalid_credentials")
		return nil, apperror.Unauthorized("invalid email or password")
	}
	if !tenant.Active {
		log.Warn("tenant login failed", "organization_id", orgID, "email", email, "reason", "inactive")
		return nil, apperror.Unauthorized("account is inactive")
	}

	if password.Compare(tenant.PasswordHash, plainPassword) != nil {
		log.Warn("tenant login failed", "organization_id", orgID, "email", email, "reason", "invalid_credentials")
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

	log.Info("tenant login succeeded", "organization_id", orgID, "user_id", tenant.ID, "email", email)

	return &domain.TenantAuthResponse{Token: token, User: profile}, nil
}

func (s *AuthService) TenantMe(ctx context.Context, orgID, tenantID uuid.UUID) (*domain.TenantProfile, error) {
	return s.repos.Tenants.GetProfile(ctx, orgID, tenantID)
}

func (s *AuthService) InviteStaff(ctx context.Context, orgID uuid.UUID, email, plainPassword string) (*domain.StaffProfile, error) {
	log := logger.FromContext(ctx)
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || len(plainPassword) < 6 {
		return nil, apperror.BadRequest("email and password (min 6 chars) are required")
	}

	if _, err := s.repos.Staff.GetByEmailAndOrg(ctx, orgID, email); err == nil {
		log.Warn("staff invite rejected", "organization_id", orgID, "email", email, "reason", "duplicate_email")
		return nil, apperror.DuplicateEmail("Staff member")
	} else if !apperror.IsNotFound(err) {
		return nil, err
	}

	hash, err := password.Hash(plainPassword)
	if err != nil {
		return nil, apperror.Internal("hash password", err)
	}

	staff := &domain.Staff{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Email:          email,
		PasswordHash:   hash,
		IsOwner:        false,
	}

	if err := s.repos.Staff.Create(ctx, staff); err != nil {
		return nil, err
	}

	log.Info("staff invited", "organization_id", orgID, "user_id", staff.ID, "email", email)
	return staffToProfile(staff), nil
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
