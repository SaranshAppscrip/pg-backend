package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/auth"
	"github.com/nivas/server/internal/config"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/notification"
	"github.com/nivas/server/internal/password"
	"github.com/nivas/server/internal/repository"
	"github.com/nivas/server/pkg/apperror"
	"github.com/nivas/server/pkg/logger"
)

type AuthService struct {
	repos  repository.Store
	tokens *auth.TokenService
	email  notification.EmailSender
	emailCfg config.EmailConfig
}

func NewAuthService(repos repository.Store, tokens *auth.TokenService, email notification.EmailSender, emailCfg config.EmailConfig) *AuthService {
	return &AuthService{repos: repos, tokens: tokens, email: email, emailCfg: emailCfg}
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

	if password.Compare(staff.PasswordHash, plainPassword) != nil {
		log.Warn("staff login failed", "organization_id", orgID, "email", email, "reason", "invalid_credentials")
		return nil, apperror.Unauthorized("invalid email or password")
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

	org, err := s.repos.Settings.Get(ctx, orgID)
	orgName := "your organization"
	if err == nil && org != nil {
		orgName = org.Name
	}

	if err := s.email.SendStaffInvite(ctx, notification.StaffInviteParams{
		To:               email,
		OrganizationName: orgName,
		OrganizationID:   orgID.String(),
		Email:            email,
		TempPassword:     plainPassword,
	}); err != nil {
		if delErr := s.repos.Staff.Delete(ctx, orgID, staff.ID); delErr != nil {
			log.Error("staff invite rollback failed", "organization_id", orgID, "user_id", staff.ID, "error", delErr)
		}
		log.Warn("staff invite email failed", "organization_id", orgID, "email", email, "error", err)
		msg := "invite email could not be sent"
		if resendMsg := notification.ResendErrorMessage(err); resendMsg != "" {
			msg = resendMsg
		}
		return nil, apperror.BadGateway(msg, err)
	}

	log.Info("staff invited", "organization_id", orgID, "user_id", staff.ID, "email", email)
	return staffToProfile(staff), nil
}

func (s *AuthService) StaffForgotPassword(ctx context.Context, orgID uuid.UUID, email string) error {
	log := logger.FromContext(ctx)
	email = strings.TrimSpace(strings.ToLower(email))
	if orgID == uuid.Nil || email == "" {
		return apperror.BadRequest("organization_id and email are required")
	}

	staff, err := s.repos.Staff.GetByEmailAndOrg(ctx, orgID, email)
	if err != nil {
		if apperror.IsNotFound(err) {
			return nil
		}
		return err
	}

	token, tokenHash, err := generateResetToken()
	if err != nil {
		return apperror.Internal("generate reset token", err)
	}

	expiresAt := time.Now().Add(s.emailCfg.PasswordResetTTL)

	if err := s.repos.PasswordReset.InvalidateUnusedForStaff(ctx, staff.ID); err != nil {
		return err
	}
	if err := s.repos.PasswordReset.Create(ctx, staff.ID, tokenHash, expiresAt); err != nil {
		return err
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s",
		strings.TrimRight(s.emailCfg.FrontendURL, "/"), token)

	if err := s.email.SendPasswordReset(ctx, notification.PasswordResetParams{
		To:       email,
		ResetURL: resetURL,
	}); err != nil {
		log.Warn("password reset email failed", "organization_id", orgID, "email", email, "error", err)
		return nil
	}

	log.Info("password reset requested", "organization_id", orgID, "user_id", staff.ID, "email", email)
	return nil
}

func (s *AuthService) StaffResetPassword(ctx context.Context, token, newPassword string) error {
	log := logger.FromContext(ctx)
	token = strings.TrimSpace(token)
	if token == "" || len(newPassword) < 6 {
		return apperror.BadRequest("token and password (min 6 chars) are required")
	}

	tokenHash := hashToken(token)
	staffID, err := s.repos.PasswordReset.GetValidByTokenHash(ctx, tokenHash)
	if err != nil {
		return apperror.BadRequest("invalid or expired reset token")
	}

	hash, err := password.Hash(newPassword)
	if err != nil {
		return apperror.Internal("hash password", err)
	}

	if err := s.repos.Staff.UpdatePassword(ctx, staffID, hash); err != nil {
		return err
	}
	if err := s.repos.PasswordReset.MarkUsed(ctx, tokenHash); err != nil {
		return err
	}

	log.Info("password reset completed", "user_id", staffID)
	return nil
}

func generateResetToken() (plain, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	plain = base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
	return plain, hashToken(plain), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
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
