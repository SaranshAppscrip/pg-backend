package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
)

type SettingsRepository interface {
	Get(ctx context.Context, orgID uuid.UUID) (*domain.Organization, error)
	UpdateName(ctx context.Context, orgID uuid.UUID, name string) (*domain.Organization, error)
}

type StaffRepository interface {
	Create(ctx context.Context, staff *domain.Staff) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Staff, error)
	GetByEmailAndOrg(ctx context.Context, orgID uuid.UUID, email string) (*domain.Staff, error)
	List(ctx context.Context, orgID uuid.UUID) ([]domain.StaffProfile, error)
	Delete(ctx context.Context, orgID, id uuid.UUID) error
}

type RoomRepository interface {
	List(ctx context.Context, orgID uuid.UUID) ([]domain.Room, error)
	Create(ctx context.Context, room *domain.Room) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Room, error)
	GetByRoomNumber(ctx context.Context, orgID uuid.UUID, roomNumber string) (*domain.Room, error)
	Delete(ctx context.Context, orgID, id uuid.UUID) error
	CountActiveTenants(ctx context.Context, orgID, roomID uuid.UUID) (int, error)
}

type TenantRepository interface {
	List(ctx context.Context, orgID uuid.UUID) ([]domain.Tenant, error)
	Create(ctx context.Context, tenant *domain.Tenant) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Tenant, error)
	GetByEmailAndOrg(ctx context.Context, orgID uuid.UUID, email string) (*domain.Tenant, error)
	GetByPhoneAndOrg(ctx context.Context, orgID uuid.UUID, phone string) (*domain.Tenant, error)
	GetByNameAndOrg(ctx context.Context, orgID uuid.UUID, name string) (*domain.Tenant, error)
	MoveOut(ctx context.Context, orgID, id uuid.UUID) (*domain.Tenant, error)
	CountActiveInRoom(ctx context.Context, orgID, roomID uuid.UUID) (int, error)
	GetProfile(ctx context.Context, orgID, id uuid.UUID) (*domain.TenantProfile, error)
}

type PaymentRepository interface {
	List(ctx context.Context, orgID uuid.UUID) ([]domain.Payment, error)
	Create(ctx context.Context, orgID uuid.UUID, payment *domain.Payment) error
	Delete(ctx context.Context, orgID, id uuid.UUID) error
	ListByTenant(ctx context.Context, orgID, tenantID uuid.UUID) ([]domain.Payment, error)
}

type ExpenseRepository interface {
	List(ctx context.Context, orgID uuid.UUID) ([]domain.Expense, error)
	Create(ctx context.Context, expense *domain.Expense) error
	Delete(ctx context.Context, orgID, id uuid.UUID) error
}

type KitchenRepository interface {
	ListItems(ctx context.Context, orgID uuid.UUID) ([]domain.KitchenItem, error)
	CreateItem(ctx context.Context, item *domain.KitchenItem) error
	GetItem(ctx context.Context, orgID, id uuid.UUID) (*domain.KitchenItem, error)
	GetByName(ctx context.Context, orgID uuid.UUID, name string) (*domain.KitchenItem, error)
	UpdateItemQty(ctx context.Context, orgID, id uuid.UUID, qty float64) error
	CreateLog(ctx context.Context, orgID uuid.UUID, log *domain.KitchenLog) error
	ListLog(ctx context.Context, orgID uuid.UUID, limit int) ([]domain.KitchenLog, error)
}

// Store aggregates all repositories.
type Store struct {
	Settings SettingsRepository
	Staff    StaffRepository
	Rooms    RoomRepository
	Tenants  TenantRepository
	Payments PaymentRepository
	Expenses ExpenseRepository
	Kitchen  KitchenRepository
}
