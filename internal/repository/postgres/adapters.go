package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/repository"
)

type settingsRepo struct{ *Store }
type staffRepo struct{ *Store }
type roomRepo struct{ *Store }
type tenantRepo struct{ *Store }
type paymentRepo struct{ *Store }
type expenseRepo struct{ *Store }
type kitchenRepo struct{ *Store }

// NewStoreBundle wires all repository interfaces to PostgreSQL.
func NewStoreBundle(pool *pgxpool.Pool) repository.Store {
	s := NewStore(pool)
	return repository.Store{
		Settings: &settingsRepo{s},
		Staff:    &staffRepo{s},
		Rooms:    &roomRepo{s},
		Tenants:  &tenantRepo{s},
		Payments: &paymentRepo{s},
		Expenses: &expenseRepo{s},
		Kitchen:  &kitchenRepo{s},
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
func (r *staffRepo) List(ctx context.Context, orgID uuid.UUID) ([]domain.StaffProfile, error) {
	return r.ListStaff(ctx, orgID)
}
func (r *staffRepo) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	return r.DeleteStaff(ctx, orgID, id)
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
func (r *tenantRepo) MoveOut(ctx context.Context, orgID, id uuid.UUID) (*domain.Tenant, error) {
	return r.MoveOutTenant(ctx, orgID, id)
}
func (r *tenantRepo) CountActiveInRoom(ctx context.Context, orgID, roomID uuid.UUID) (int, error) {
	return r.CountActiveTenantsInRoom(ctx, orgID, roomID)
}
func (r *tenantRepo) GetProfile(ctx context.Context, orgID, id uuid.UUID) (*domain.TenantProfile, error) {
	return r.GetTenantProfile(ctx, orgID, id)
}

// Payments
func (r *paymentRepo) List(ctx context.Context, orgID uuid.UUID) ([]domain.Payment, error) {
	return r.ListPayments(ctx, orgID)
}
func (r *paymentRepo) Create(ctx context.Context, orgID uuid.UUID, payment *domain.Payment) error {
	return r.CreatePayment(ctx, orgID, payment)
}
func (r *paymentRepo) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	return r.DeletePayment(ctx, orgID, id)
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
func (r *expenseRepo) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	return r.DeleteExpense(ctx, orgID, id)
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
func (r *kitchenRepo) UpdateItemQty(ctx context.Context, orgID, id uuid.UUID, qty float64) error {
	return r.UpdateKitchenItemQty(ctx, orgID, id, qty)
}
func (r *kitchenRepo) CreateLog(ctx context.Context, orgID uuid.UUID, log *domain.KitchenLog) error {
	return r.CreateKitchenLog(ctx, orgID, log)
}
func (r *kitchenRepo) ListLog(ctx context.Context, orgID uuid.UUID, limit int) ([]domain.KitchenLog, error) {
	return r.ListKitchenLog(ctx, orgID, limit)
}
