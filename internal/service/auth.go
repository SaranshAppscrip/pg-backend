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
	repos    repository.Store
	tokens   *auth.TokenService
	jwtCfg   config.JWTConfig
	email    notification.EmailSender
	emailCfg config.EmailConfig
}

func NewAuthService(repos repository.Store, tokens *auth.TokenService, jwtCfg config.JWTConfig, email notification.EmailSender, emailCfg config.EmailConfig) *AuthService {
	return &AuthService{repos: repos, tokens: tokens, jwtCfg: jwtCfg, email: email, emailCfg: emailCfg}
}

func (s *AuthService) StaffLogin(ctx context.Context, email, plainPassword string, orgID *uuid.UUID) (*domain.AuthResponse, error) {
	defer normalizeAuthLatency(time.Now())
	log := logger.FromContext(ctx)
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || len(plainPassword) < 6 {
		log.Warn("staff login validation failed", "email", email)
		return nil, apperror.BadRequest("email and password (min 6 chars) are required")
	}

	staff, err := s.resolveStaffLogin(ctx, email, plainPassword, orgID)
	if err != nil {
		if appErr, ok := apperror.As(err); ok && appErr.Code == apperror.CodeUnauthorized {
			log.Warn("staff login failed", "email", email, "reason", "invalid_credentials")
		}
		return nil, err
	}

	tokenPair, err := s.issueStaffTokenPair(ctx, staff)
	if err != nil {
		return nil, err
	}

	log.Info("staff login succeeded",
		"organization_id", staff.OrganizationID,
		"user_id", staff.ID,
		"email", email,
		"is_owner", staff.IsOwner,
	)

	return &domain.AuthResponse{
		TokenPair: *tokenPair,
		User:      staffToProfile(staff),
	}, nil
}

func (s *AuthService) StaffMe(ctx context.Context, orgID, staffID uuid.UUID) (*domain.StaffProfile, error) {
	staff, err := s.repos.Staff.GetByID(ctx, orgID, staffID)
	if err != nil {
		return nil, err
	}
	return staffToProfile(staff), nil
}

func (s *AuthService) TenantLogin(ctx context.Context, email, plainPassword string, orgID *uuid.UUID) (*domain.TenantAuthResponse, error) {
	defer normalizeAuthLatency(time.Now())
	log := logger.FromContext(ctx)
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || len(plainPassword) < 6 {
		log.Warn("tenant login validation failed", "email", email)
		return nil, apperror.BadRequest("email and password (min 6 chars) are required")
	}

	tenant, err := s.resolveTenantLogin(ctx, email, plainPassword, orgID)
	if err != nil {
		if appErr, ok := apperror.As(err); ok && appErr.Code == apperror.CodeUnauthorized {
			log.Warn("tenant login failed", "email", email, "reason", "invalid_credentials")
		}
		return nil, err
	}

	tokenPair, err := s.issueTenantTokenPair(ctx, tenant)
	if err != nil {
		return nil, err
	}

	profile, err := s.repos.Tenants.GetProfile(ctx, tenant.OrganizationID, tenant.ID)
	if err != nil {
		return nil, err
	}

	log.Info("tenant login succeeded", "organization_id", tenant.OrganizationID, "user_id", tenant.ID, "email", email)

	return &domain.TenantAuthResponse{TokenPair: *tokenPair, User: profile}, nil
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

func (s *AuthService) StaffForgotPassword(ctx context.Context, email string, orgID *uuid.UUID) error {
	defer normalizeAuthLatency(time.Now())
	log := logger.FromContext(ctx)
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return apperror.BadRequest("email is required")
	}

	staff, err := s.resolveStaffForReset(ctx, email, orgID)
	if err != nil {
		if appErr, ok := apperror.As(err); ok && appErr.Code == apperror.CodeMultipleOrganizations {
			return err
		}
		return nil
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
		log.Warn("password reset email failed", "organization_id", staff.OrganizationID, "email", email, "error", err)
		return nil
	}

	log.Info("password reset requested", "organization_id", staff.OrganizationID, "user_id", staff.ID, "email", email)
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
	if err := s.repos.RefreshTokens.RevokeAllForUser(ctx, domain.TokenTypeStaff, staffID); err != nil {
		return err
	}

	log.Info("password reset completed", "user_id", staffID)
	return nil
}

func (s *AuthService) StaffRefresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	return s.refreshTokens(ctx, refreshToken, domain.TokenTypeStaff)
}

func (s *AuthService) StaffLogout(ctx context.Context, refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return apperror.BadRequest("refresh_token is required")
	}
	if err := s.repos.RefreshTokens.Revoke(ctx, hashToken(refreshToken)); err != nil {
		if appErr, ok := apperror.As(err); ok && appErr.Code == apperror.CodeUnauthorized {
			return nil // idempotent logout
		}
		return err
	}
	logger.FromContext(ctx).Info("staff logout")
	return nil
}

func (s *AuthService) RevokeAllStaffSessions(ctx context.Context, staffID uuid.UUID) error {
	return s.repos.RefreshTokens.RevokeAllForUser(ctx, domain.TokenTypeStaff, staffID)
}

func (s *AuthService) TenantForgotPassword(ctx context.Context, email string, orgID *uuid.UUID) error {
	defer normalizeAuthLatency(time.Now())
	log := logger.FromContext(ctx)
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return apperror.BadRequest("email is required")
	}

	tenant, err := s.resolveTenantForReset(ctx, email, orgID)
	if err != nil {
		if appErr, ok := apperror.As(err); ok && appErr.Code == apperror.CodeMultipleOrganizations {
			return err
		}
		return nil
	}

	token, tokenHash, err := generateResetToken()
	if err != nil {
		return apperror.Internal("generate reset token", err)
	}

	expiresAt := time.Now().Add(s.emailCfg.PasswordResetTTL)

	if err := s.repos.TenantPasswordReset.InvalidateUnusedForTenant(ctx, tenant.ID); err != nil {
		return err
	}
	if err := s.repos.TenantPasswordReset.Create(ctx, tenant.ID, tokenHash, expiresAt); err != nil {
		return err
	}

	resetURL := fmt.Sprintf("%s/tenant/reset-password?token=%s",
		strings.TrimRight(s.emailCfg.FrontendURL, "/"), token)

	if err := s.email.SendPasswordReset(ctx, notification.PasswordResetParams{
		To:        email,
		ResetURL:  resetURL,
		ForTenant: true,
	}); err != nil {
		log.Warn("tenant password reset email failed", "organization_id", tenant.OrganizationID, "email", email, "error", err)
		return nil
	}

	log.Info("tenant password reset requested", "organization_id", tenant.OrganizationID, "user_id", tenant.ID, "email", email)
	return nil
}

func (s *AuthService) TenantResetPassword(ctx context.Context, token, newPassword string) error {
	log := logger.FromContext(ctx)
	token = strings.TrimSpace(token)
	if token == "" || len(newPassword) < 6 {
		return apperror.BadRequest("token and password (min 6 chars) are required")
	}

	tokenHash := hashToken(token)
	tenantID, err := s.repos.TenantPasswordReset.GetValidByTokenHash(ctx, tokenHash)
	if err != nil {
		return apperror.BadRequest("invalid or expired reset token")
	}

	hash, err := password.Hash(newPassword)
	if err != nil {
		return apperror.Internal("hash password", err)
	}

	if err := s.repos.Tenants.UpdatePassword(ctx, tenantID, hash); err != nil {
		return err
	}
	if err := s.repos.TenantPasswordReset.MarkUsed(ctx, tokenHash); err != nil {
		return err
	}
	if err := s.repos.RefreshTokens.RevokeAllForUser(ctx, domain.TokenTypeTenant, tenantID); err != nil {
		return err
	}

	log.Info("tenant password reset completed", "user_id", tenantID)
	return nil
}

func (s *AuthService) TenantRefresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	return s.refreshTokens(ctx, refreshToken, domain.TokenTypeTenant)
}

func (s *AuthService) TenantLogout(ctx context.Context, refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return apperror.BadRequest("refresh_token is required")
	}
	if err := s.repos.RefreshTokens.Revoke(ctx, hashToken(refreshToken)); err != nil {
		if appErr, ok := apperror.As(err); ok && appErr.Code == apperror.CodeUnauthorized {
			return nil
		}
		return err
	}
	logger.FromContext(ctx).Info("tenant logout")
	return nil
}

func (s *AuthService) RevokeAllTenantSessions(ctx context.Context, tenantID uuid.UUID) error {
	return s.repos.RefreshTokens.RevokeAllForUser(ctx, domain.TokenTypeTenant, tenantID)
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

func (s *AuthService) issueStaffTokenPair(ctx context.Context, staff *domain.Staff) (*domain.TokenPair, error) {
	access, err := s.tokens.GenerateStaffToken(staff)
	if err != nil {
		return nil, err
	}
	refresh, err := s.createRefreshToken(ctx, domain.TokenTypeStaff, staff.ID, staff.OrganizationID, s.jwtCfg.StaffRefreshTTL)
	if err != nil {
		return nil, err
	}
	return &domain.TokenPair{AccessToken: access, RefreshToken: refresh, Token: access}, nil
}

func (s *AuthService) issueTenantTokenPair(ctx context.Context, tenant *domain.Tenant) (*domain.TokenPair, error) {
	access, err := s.tokens.GenerateTenantToken(tenant)
	if err != nil {
		return nil, err
	}
	refresh, err := s.createRefreshToken(ctx, domain.TokenTypeTenant, tenant.ID, tenant.OrganizationID, s.jwtCfg.TenantRefreshTTL)
	if err != nil {
		return nil, err
	}
	return &domain.TokenPair{AccessToken: access, RefreshToken: refresh, Token: access}, nil
}

func (s *AuthService) createRefreshToken(ctx context.Context, userType domain.TokenType, userID, orgID uuid.UUID, ttl time.Duration) (string, error) {
	plain, tokenHash, err := generateResetToken()
	if err != nil {
		return "", apperror.Internal("generate refresh token", err)
	}
	if err := s.repos.RefreshTokens.Create(ctx, userType, userID, orgID, tokenHash, time.Now().Add(ttl)); err != nil {
		return "", err
	}
	return plain, nil
}

func (s *AuthService) refreshTokens(ctx context.Context, refreshToken string, expectedType domain.TokenType) (*domain.TokenPair, error) {
	log := logger.FromContext(ctx)
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, apperror.BadRequest("refresh_token is required")
	}

	tokenHash := hashToken(refreshToken)
	stored, err := s.repos.RefreshTokens.GetValidByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if stored.UserType != expectedType {
		return nil, apperror.Unauthorized("invalid or expired refresh token")
	}

	if err := s.repos.RefreshTokens.Revoke(ctx, tokenHash); err != nil {
		return nil, err
	}

	var pair *domain.TokenPair
	switch expectedType {
	case domain.TokenTypeStaff:
		staff, err := s.repos.Staff.GetByID(ctx, stored.OrganizationID, stored.UserID)
		if err != nil {
			return nil, err
		}
		pair, err = s.issueStaffTokenPair(ctx, staff)
		if err != nil {
			return nil, err
		}
	case domain.TokenTypeTenant:
		tenant, err := s.repos.Tenants.GetByID(ctx, stored.OrganizationID, stored.UserID)
		if err != nil {
			return nil, err
		}
		if !tenant.Active {
			return nil, apperror.Unauthorized("invalid or expired refresh token")
		}
		pair, err = s.issueTenantTokenPair(ctx, tenant)
		if err != nil {
			return nil, err
		}
	default:
		return nil, apperror.Unauthorized("invalid or expired refresh token")
	}

	log.Info("token refreshed", "user_type", expectedType, "user_id", stored.UserID)
	return pair, nil
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

func (s *AuthService) resolveStaffLogin(ctx context.Context, email, plainPassword string, orgID *uuid.UUID) (*domain.Staff, error) {
	candidates, err := s.repos.Staff.ListByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	candidates = filterStaffByOrg(candidates, orgID)
	if len(candidates) == 0 {
		return nil, apperror.Unauthorized("invalid email or password")
	}

	var matched []domain.Staff
	for _, st := range candidates {
		if password.Compare(st.PasswordHash, plainPassword) == nil {
			matched = append(matched, st)
		}
	}
	if len(matched) == 0 {
		return nil, apperror.Unauthorized("invalid email or password")
	}
	if len(matched) == 1 {
		return &matched[0], nil
	}
	return nil, s.multipleOrganizationsError(ctx, staffOrgIDs(matched))
}

func (s *AuthService) resolveTenantLogin(ctx context.Context, email, plainPassword string, orgID *uuid.UUID) (*domain.Tenant, error) {
	candidates, err := s.repos.Tenants.ListByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	candidates = filterTenantsByOrg(candidates, orgID)
	if len(candidates) == 0 {
		return nil, apperror.Unauthorized("invalid email or password")
	}

	var matched []domain.Tenant
	for _, t := range candidates {
		if !t.Active {
			continue
		}
		if password.Compare(t.PasswordHash, plainPassword) == nil {
			matched = append(matched, t)
		}
	}
	if len(matched) == 0 {
		return nil, apperror.Unauthorized("invalid email or password")
	}
	if len(matched) == 1 {
		return &matched[0], nil
	}
	return nil, s.multipleOrganizationsError(ctx, tenantOrgIDs(matched))
}

func (s *AuthService) resolveStaffForReset(ctx context.Context, email string, orgID *uuid.UUID) (*domain.Staff, error) {
	candidates, err := s.repos.Staff.ListByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	candidates = filterStaffByOrg(candidates, orgID)
	if len(candidates) == 0 {
		return nil, apperror.NotFound("staff not found")
	}
	if len(candidates) == 1 {
		return &candidates[0], nil
	}
	return nil, s.multipleOrganizationsError(ctx, staffOrgIDs(candidates))
}

func (s *AuthService) resolveTenantForReset(ctx context.Context, email string, orgID *uuid.UUID) (*domain.Tenant, error) {
	candidates, err := s.repos.Tenants.ListByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	candidates = filterActiveTenants(filterTenantsByOrg(candidates, orgID))
	if len(candidates) == 0 {
		return nil, apperror.NotFound("tenant not found")
	}
	if len(candidates) == 1 {
		return &candidates[0], nil
	}
	return nil, s.multipleOrganizationsError(ctx, tenantOrgIDs(candidates))
}

func filterActiveTenants(tenants []domain.Tenant) []domain.Tenant {
	out := make([]domain.Tenant, 0, len(tenants))
	for _, t := range tenants {
		if t.Active {
			out = append(out, t)
		}
	}
	return out
}

func (s *AuthService) multipleOrganizationsError(ctx context.Context, orgIDs []uuid.UUID) error {
	details, err := s.orgChoices(ctx, orgIDs)
	if err != nil {
		return err
	}
	return apperror.MultipleOrganizations(details)
}

func (s *AuthService) orgChoices(ctx context.Context, orgIDs []uuid.UUID) ([]map[string]string, error) {
	seen := make(map[uuid.UUID]bool, len(orgIDs))
	details := make([]map[string]string, 0, len(orgIDs))
	for _, id := range orgIDs {
		if seen[id] {
			continue
		}
		seen[id] = true
		name := "Organization"
		if org, err := s.repos.Settings.Get(ctx, id); err == nil && org != nil {
			name = org.Name
		}
		details = append(details, map[string]string{
			"organization_id": id.String(),
			"name":            name,
		})
	}
	return details, nil
}

func filterStaffByOrg(staff []domain.Staff, orgID *uuid.UUID) []domain.Staff {
	if orgID == nil || *orgID == uuid.Nil {
		return staff
	}
	out := make([]domain.Staff, 0, len(staff))
	for _, st := range staff {
		if st.OrganizationID == *orgID {
			out = append(out, st)
		}
	}
	return out
}

func filterTenantsByOrg(tenants []domain.Tenant, orgID *uuid.UUID) []domain.Tenant {
	if orgID == nil || *orgID == uuid.Nil {
		return tenants
	}
	out := make([]domain.Tenant, 0, len(tenants))
	for _, t := range tenants {
		if t.OrganizationID == *orgID {
			out = append(out, t)
		}
	}
	return out
}

func staffOrgIDs(staff []domain.Staff) []uuid.UUID {
	ids := make([]uuid.UUID, len(staff))
	for i, st := range staff {
		ids[i] = st.OrganizationID
	}
	return ids
}

func tenantOrgIDs(tenants []domain.Tenant) []uuid.UUID {
	ids := make([]uuid.UUID, len(tenants))
	for i, t := range tenants {
		ids[i] = t.OrganizationID
	}
	return ids
}

const minAuthLatency = 200 * time.Millisecond

func normalizeAuthLatency(start time.Time) {
	if d := time.Since(start); d < minAuthLatency {
		time.Sleep(minAuthLatency - d)
	}
}
