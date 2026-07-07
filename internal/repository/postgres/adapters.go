package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/repository"
)

type settingsRepo struct{ *Store }
type staffRepo struct{ *Store }
type passwordResetRepo struct{ *Store }
type tenantPasswordResetRepo struct{ *Store }
type refreshTokenRepo struct{ *Store }
type roomRepo struct{ *Store }
type tenantRepo struct{ *Store }
type paymentRepo struct{ *Store }
type expenseRepo struct{ *Store }
type kitchenRepo struct{ *Store }
type auditRepo struct{ *Store }

// NewStoreBundle wires all repository interfaces to PostgreSQL.
func NewStoreBundle(pool *pgxpool.Pool) repository.Store {
	s := NewStore(pool)
	return repository.Store{
		Settings:      &settingsRepo{s},
		Staff:         &staffRepo{s},
		PasswordReset:       &passwordResetRepo{s},
		TenantPasswordReset: &tenantPasswordResetRepo{s},
		RefreshTokens:       &refreshTokenRepo{s},
		Rooms:               &roomRepo{s},
		Tenants:  &tenantRepo{s},
		Payments: &paymentRepo{s},
		Expenses: &expenseRepo{s},
		Kitchen:             &kitchenRepo{s},
		Audit:               &auditRepo{s},
	}
}

// Settings
func (r *settingsRepo) Get(ctx context.Context, orgID uuid.UUID) (*domain.Organization, error) {
	return r.GetSettings(ctx, orgID)
}
func (r *settingsRepo) UpdateName(ctx context.Context, orgID uuid.UUID, name string) (*domain.Organization, error) {
	return r.UpdateOrgName(ctx, orgID, name)
}

// Staff
func (r *staffRepo) Create(ctx context.Context, staff *domain.Staff) error {
	return r.CreateStaff(ctx, staff)
}
func (r *staffRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Staff, error) {
	return r.GetStaffByID(ctx, orgID, id)
}
func (r *staffRepo) GetByEmailAndOrg(ctx context.Context, orgID uuid.UUID, email string) (*domain.Staff, error) {
	return r.GetStaffByEmailAndOrg(ctx, orgID, email)
}
func (r *staffRepo) ListByEmail(ctx context.Context, email string) ([]domain.Staff, error) {
	return r.ListStaffByEmail(ctx, email)
}
func (r *staffRepo) List(ctx context.Context, orgID uuid.UUID) ([]domain.StaffProfile, error) {
	return r.ListStaff(ctx, orgID)
}
func (r *staffRepo) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	return r.DeleteStaff(ctx, orgID, id)
}
func (r *staffRepo) UpdatePassword(ctx context.Context, staffID uuid.UUID, passwordHash string) error {
	return r.UpdateStaffPassword(ctx, staffID, passwordHash)
}

// Password reset
func (r *passwordResetRepo) InvalidateUnusedForStaff(ctx context.Context, staffID uuid.UUID) error {
	return r.InvalidateUnusedPasswordResetTokens(ctx, staffID)
}
func (r *passwordResetRepo) Create(ctx context.Context, staffID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	return r.CreatePasswordResetToken(ctx, staffID, tokenHash, expiresAt)
}
func (r *passwordResetRepo) GetValidByTokenHash(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	return r.GetValidPasswordResetByTokenHash(ctx, tokenHash)
}
func (r *passwordResetRepo) MarkUsed(ctx context.Context, tokenHash string) error {
	return r.MarkPasswordResetTokenUsed(ctx, tokenHash)
}

// Tenant password reset
func (r *tenantPasswordResetRepo) InvalidateUnusedForTenant(ctx context.Context, tenantID uuid.UUID) error {
	return r.InvalidateUnusedTenantPasswordResetTokens(ctx, tenantID)
}
func (r *tenantPasswordResetRepo) Create(ctx context.Context, tenantID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	return r.CreateTenantPasswordResetToken(ctx, tenantID, tokenHash, expiresAt)
}
func (r *tenantPasswordResetRepo) GetValidByTokenHash(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	return r.GetValidTenantPasswordResetByTokenHash(ctx, tokenHash)
}
func (r *tenantPasswordResetRepo) MarkUsed(ctx context.Context, tokenHash string) error {
	return r.MarkTenantPasswordResetTokenUsed(ctx, tokenHash)
}

// Refresh tokens
func (r *refreshTokenRepo) Create(ctx context.Context, userType domain.TokenType, userID, orgID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	return r.CreateRefreshToken(ctx, userType, userID, orgID, tokenHash, expiresAt)
}
func (r *refreshTokenRepo) GetValidByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	return r.GetValidRefreshTokenByHash(ctx, tokenHash)
}
func (r *refreshTokenRepo) Revoke(ctx context.Context, tokenHash string) error {
	return r.RevokeRefreshToken(ctx, tokenHash)
}
func (r *refreshTokenRepo) RevokeAllForUser(ctx context.Context, userType domain.TokenType, userID uuid.UUID) error {
	return r.RevokeAllRefreshTokensForUser(ctx, userType, userID)
}

// Rooms
func (r *roomRepo) List(ctx context.Context, orgID uuid.UUID) ([]domain.Room, error) {
	return r.ListRooms(ctx, orgID)
}
func (r *roomRepo) Create(ctx context.Context, room *domain.Room) error {
	return r.CreateRoom(ctx, room)
}
func (r *roomRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Room, error) {
	return r.GetRoomByID(ctx, orgID, id)
}
func (r *roomRepo) GetByRoomNumber(ctx context.Context, orgID uuid.UUID, roomNumber string) (*domain.Room, error) {
	return r.GetRoomByNumber(ctx, orgID, roomNumber)
}
func (r *roomRepo) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	return r.DeleteRoom(ctx, orgID, id)
}
func (r *roomRepo) CountActiveTenants(ctx context.Context, orgID, roomID uuid.UUID) (int, error) {
	return r.CountActiveTenantsInRoom(ctx, orgID, roomID)
}

// Tenants
func (r *tenantRepo) List(ctx context.Context, orgID uuid.UUID) ([]domain.Tenant, error) {
	return r.ListTenants(ctx, orgID)
}
func (r *tenantRepo) Create(ctx context.Context, tenant *domain.Tenant) error {
	return r.CreateTenant(ctx, tenant)
}
func (r *tenantRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Tenant, error) {
	return r.GetTenantByID(ctx, orgID, id)
}
func (r *tenantRepo) GetByEmailAndOrg(ctx context.Context, orgID uuid.UUID, email string) (*domain.Tenant, error) {
	return r.GetTenantByEmailAndOrg(ctx, orgID, email)
}
func (r *tenantRepo) ListByEmail(ctx context.Context, email string) ([]domain.Tenant, error) {
	return r.ListTenantsByEmail(ctx, email)
}
func (r *tenantRepo) GetByPhoneAndOrg(ctx context.Context, orgID uuid.UUID, phone string) (*domain.Tenant, error) {
	return r.GetTenantByPhoneAndOrg(ctx, orgID, phone)
}
func (r *tenantRepo) GetByNameAndOrg(ctx context.Context, orgID uuid.UUID, name string) (*domain.Tenant, error) {
	return r.GetTenantByNameAndOrg(ctx, orgID, name)
}
func (r *tenantRepo) MoveOut(ctx context.Context, orgID, id uuid.UUID) (*domain.Tenant, error) {
	return r.MoveOutTenant(ctx, orgID, id)
}
func (r *tenantRepo) CountActiveInRoom(ctx context.Context, orgID, roomID uuid.UUID) (int, error) {
	return r.CountActiveTenantsInRoom(ctx, orgID, roomID)
}
func (r *tenantRepo) GetProfile(ctx context.Context, orgID, id uuid.UUID) (*domain.TenantProfile, error) {
	return r.GetTenantProfile(ctx, orgID, id)
}
func (r *tenantRepo) UpdatePassword(ctx context.Context, tenantID uuid.UUID, passwordHash string) error {
	return r.UpdateTenantPassword(ctx, tenantID, passwordHash)
}

// Payments
func (r *paymentRepo) List(ctx context.Context, orgID uuid.UUID) ([]domain.Payment, error) {
	return r.ListPayments(ctx, orgID)
}
func (r *paymentRepo) Create(ctx context.Context, orgID uuid.UUID, payment *domain.Payment) error {
	return r.CreatePayment(ctx, orgID, payment)
}
func (r *paymentRepo) SoftDelete(ctx context.Context, orgID, id uuid.UUID) (*domain.Payment, error) {
	return r.SoftDeletePayment(ctx, orgID, id)
}
func (r *paymentRepo) ListByTenant(ctx context.Context, orgID, tenantID uuid.UUID) ([]domain.Payment, error) {
	return r.ListPaymentsByTenant(ctx, orgID, tenantID)
}

// Expenses
func (r *expenseRepo) List(ctx context.Context, orgID uuid.UUID) ([]domain.Expense, error) {
	return r.ListExpenses(ctx, orgID)
}
func (r *expenseRepo) Create(ctx context.Context, expense *domain.Expense) error {
	return r.CreateExpense(ctx, expense)
}
func (r *expenseRepo) SoftDelete(ctx context.Context, orgID, id uuid.UUID) (*domain.Expense, error) {
	return r.SoftDeleteExpense(ctx, orgID, id)
}

// Kitchen
func (r *kitchenRepo) ListItems(ctx context.Context, orgID uuid.UUID) ([]domain.KitchenItem, error) {
	return r.ListKitchenItems(ctx, orgID)
}
func (r *kitchenRepo) CreateItem(ctx context.Context, item *domain.KitchenItem) error {
	return r.CreateKitchenItem(ctx, item)
}
func (r *kitchenRepo) GetItem(ctx context.Context, orgID, id uuid.UUID) (*domain.KitchenItem, error) {
	return r.GetKitchenItem(ctx, orgID, id)
}
func (r *kitchenRepo) GetByName(ctx context.Context, orgID uuid.UUID, name string) (*domain.KitchenItem, error) {
	return r.GetKitchenItemByName(ctx, orgID, name)
}
func (r *kitchenRepo) UpdateItemQty(ctx context.Context, orgID, id uuid.UUID, qty float64) error {
	return r.UpdateKitchenItemQty(ctx, orgID, id, qty)
}
func (r *kitchenRepo) CreateLog(ctx context.Context, orgID uuid.UUID, log *domain.KitchenLog) error {
	return r.CreateKitchenLog(ctx, orgID, log)
}
func (r *kitchenRepo) ListLog(ctx context.Context, orgID uuid.UUID, limit int) ([]domain.KitchenLog, error) {
	return r.ListKitchenLog(ctx, orgID, limit)
}

// Audit
func (r *auditRepo) Create(ctx context.Context, orgID uuid.UUID, staffID *uuid.UUID, entry *domain.StaffAuditLog) error {
	return r.CreateStaffAuditLog(ctx, orgID, staffID, entry)
}
func (r *auditRepo) List(ctx context.Context, orgID uuid.UUID, limit int) ([]domain.StaffAuditLog, error) {
	return r.ListStaffAuditLog(ctx, orgID, limit)
}
