package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
)

type SettingsRepository interface {
	Get(ctx context.Context, orgID uuid.UUID) (*domain.Organization, error)
	UpdateName(ctx context.Context, orgID uuid.UUID, name string) (*domain.Organization, error)
}

type PropertyRepository interface {
	List(ctx context.Context, orgID uuid.UUID) ([]domain.Property, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Property, error)
	Create(ctx context.Context, property *domain.Property) error
	Update(ctx context.Context, orgID, id uuid.UUID, name string, address *string) (*domain.Property, error)
}

type StaffRepository interface {
	Create(ctx context.Context, staff *domain.Staff) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Staff, error)
	GetByEmailAndOrg(ctx context.Context, orgID uuid.UUID, email string) (*domain.Staff, error)
	ListByEmail(ctx context.Context, email string) ([]domain.Staff, error)
	List(ctx context.Context, orgID uuid.UUID) ([]domain.StaffProfile, error)
	Delete(ctx context.Context, orgID, id uuid.UUID) error
	UpdatePassword(ctx context.Context, staffID uuid.UUID, passwordHash string) error
}

type PasswordResetRepository interface {
	InvalidateUnusedForStaff(ctx context.Context, staffID uuid.UUID) error
	Create(ctx context.Context, staffID uuid.UUID, tokenHash string, expiresAt time.Time) error
	GetValidByTokenHash(ctx context.Context, tokenHash string) (staffID uuid.UUID, err error)
	MarkUsed(ctx context.Context, tokenHash string) error
}

type RoomRepository interface {
	List(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID) ([]domain.Room, error)
	Create(ctx context.Context, room *domain.Room) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Room, error)
	GetByRoomNumber(ctx context.Context, propertyID uuid.UUID, roomNumber string) (*domain.Room, error)
	Delete(ctx context.Context, orgID, id uuid.UUID) error
	CountActiveTenants(ctx context.Context, orgID, roomID uuid.UUID) (int, error)
}

type TenantRepository interface {
	List(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID) ([]domain.Tenant, error)
	Create(ctx context.Context, tenant *domain.Tenant) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Tenant, error)
	GetByEmailAndOrg(ctx context.Context, orgID uuid.UUID, email string) (*domain.Tenant, error)
	ListByEmail(ctx context.Context, email string) ([]domain.Tenant, error)
	GetByPhoneAndOrg(ctx context.Context, orgID uuid.UUID, phone string) (*domain.Tenant, error)
	GetByNameAndOrg(ctx context.Context, orgID uuid.UUID, name string) (*domain.Tenant, error)
	MoveOut(ctx context.Context, orgID, id uuid.UUID) (*domain.Tenant, error)
	CountActiveInRoom(ctx context.Context, orgID, roomID uuid.UUID) (int, error)
	GetProfile(ctx context.Context, orgID, id uuid.UUID) (*domain.TenantProfile, error)
	UpdatePassword(ctx context.Context, tenantID uuid.UUID, passwordHash string) error
}

type TenantPasswordResetRepository interface {
	InvalidateUnusedForTenant(ctx context.Context, tenantID uuid.UUID) error
	Create(ctx context.Context, tenantID uuid.UUID, tokenHash string, expiresAt time.Time) error
	GetValidByTokenHash(ctx context.Context, tokenHash string) (tenantID uuid.UUID, err error)
	MarkUsed(ctx context.Context, tokenHash string) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, userType domain.TokenType, userID, orgID uuid.UUID, tokenHash string, expiresAt time.Time) error
	GetValidByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	Revoke(ctx context.Context, tokenHash string) error
	RevokeAllForUser(ctx context.Context, userType domain.TokenType, userID uuid.UUID) error
}

type PaymentRepository interface {
	List(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID) ([]domain.Payment, error)
	Create(ctx context.Context, orgID uuid.UUID, payment *domain.Payment) error
	SoftDelete(ctx context.Context, orgID, id uuid.UUID) (*domain.Payment, error)
	ListByTenant(ctx context.Context, orgID, tenantID uuid.UUID) ([]domain.Payment, error)
}

type ExpenseRepository interface {
	List(ctx context.Context, orgID uuid.UUID) ([]domain.Expense, error)
	Create(ctx context.Context, expense *domain.Expense) error
	SoftDelete(ctx context.Context, orgID, id uuid.UUID) (*domain.Expense, error)
}

type ReminderRepository interface {
	ListActiveTenantsWithDues(ctx context.Context, orgID *uuid.UUID) ([]ReminderTenantRow, error)
	HasRentReminder(ctx context.Context, tenantID uuid.UUID, forMonth, reminderType string) (bool, error)
	CreateRentReminder(ctx context.Context, tenantID uuid.UUID, forMonth, reminderType string) error
	ListPaymentsForTenantMonth(ctx context.Context, tenantID uuid.UUID, forMonth string) ([]domain.Payment, error)
}

type ReminderTenantRow struct {
	TenantID       uuid.UUID
	OrganizationID uuid.UUID
	PropertyID     uuid.UUID
	Name           string
	Email          string
	MonthlyFee     float64
	PropertyName   string
}

type ExportRepository interface {
	ListPaymentsForExport(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID) ([]PaymentExportRow, error)
	ListTenantsForExport(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID) ([]TenantExportRow, error)
	ListExpensesForExport(ctx context.Context, orgID uuid.UUID) ([]ExpenseExportRow, error)
	GetPaymentReceiptData(ctx context.Context, orgID, paymentID uuid.UUID) (*PaymentReceiptData, error)
}

type PaymentExportRow struct {
	ID           uuid.UUID
	Date         string
	TenantName   string
	RoomNumber   string
	PropertyName string
	ForMonth     string
	Amount       float64
	Mode         string
}

type TenantExportRow struct {
	ID           uuid.UUID
	Name         string
	Email        string
	Phone        *string
	PropertyName string
	RoomNumber   *string
	MonthlyFee   float64
	JoinDate     string
	Active       bool
}

type ExpenseExportRow struct {
	ID       uuid.UUID
	Date     string
	Category string
	Amount   float64
	Note     *string
}

type PaymentReceiptData struct {
	PaymentID        uuid.UUID
	Amount           float64
	Date             string
	ForMonth         string
	Mode             string
	TenantName       string
	TenantEmail      string
	RoomNumber       string
	PropertyName     string
	OrganizationName string
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

type AuditRepository interface {
	Create(ctx context.Context, orgID uuid.UUID, staffID *uuid.UUID, entry *domain.StaffAuditLog) error
	List(ctx context.Context, orgID uuid.UUID, limit int) ([]domain.StaffAuditLog, error)
}

type DocumentRepository interface {
	CreateTenantDocument(ctx context.Context, doc *domain.TenantDocument, storageKey string) error
	ListTenantDocuments(ctx context.Context, orgID, tenantID uuid.UUID) ([]domain.TenantDocument, error)
	GetTenantDocument(ctx context.Context, orgID, id uuid.UUID) (*domain.TenantDocument, string, error)
	SoftDeleteTenantDocument(ctx context.Context, orgID, id uuid.UUID) (*domain.TenantDocument, string, error)

	CreateOrganizationDocument(ctx context.Context, doc *domain.OrganizationDocument, storageKey string) error
	ListOrganizationDocuments(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID) ([]domain.OrganizationDocument, error)
	GetOrganizationDocument(ctx context.Context, orgID, id uuid.UUID) (*domain.OrganizationDocument, string, error)
	SoftDeleteOrganizationDocument(ctx context.Context, orgID, id uuid.UUID) (*domain.OrganizationDocument, string, error)
}

type PortalRepository interface {
	ListAnnouncements(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID, activeOnly bool) ([]domain.Announcement, error)
	GetAnnouncement(ctx context.Context, orgID, id uuid.UUID) (*domain.Announcement, error)
	CreateAnnouncement(ctx context.Context, a *domain.Announcement) error
	UpdateAnnouncement(ctx context.Context, orgID, id uuid.UUID, a *domain.Announcement) (*domain.Announcement, error)
	DeleteAnnouncement(ctx context.Context, orgID, id uuid.UUID) error
	ListAnnouncementsForTenant(ctx context.Context, orgID, tenantID uuid.UUID) ([]domain.Announcement, error)

	ListMaintenanceRequests(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID, status *domain.MaintenanceStatus) ([]domain.MaintenanceRequest, error)
	ListMaintenanceByTenant(ctx context.Context, orgID, tenantID uuid.UUID) ([]domain.MaintenanceRequest, error)
	GetMaintenanceRequest(ctx context.Context, orgID, id uuid.UUID) (*domain.MaintenanceRequest, error)
	CreateMaintenanceRequest(ctx context.Context, req *domain.MaintenanceRequest) error
	UpdateMaintenanceRequest(ctx context.Context, orgID, id uuid.UUID, upd domain.MaintenanceUpdate) (*domain.MaintenanceRequest, error)

	ListVisitorLog(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID, limit int) ([]domain.VisitorLogEntry, error)
	CreateVisitorEntry(ctx context.Context, entry *domain.VisitorLogEntry) error
	RecordVisitorExit(ctx context.Context, orgID, id uuid.UUID, exitAt time.Time) (*domain.VisitorLogEntry, error)
}

// Store aggregates all repositories.
type Store struct {
	Settings            SettingsRepository
	Properties          PropertyRepository
	Staff               StaffRepository
	PasswordReset       PasswordResetRepository
	TenantPasswordReset TenantPasswordResetRepository
	RefreshTokens       RefreshTokenRepository
	Rooms               RoomRepository
	Tenants             TenantRepository
	Payments            PaymentRepository
	Expenses            ExpenseRepository
	Kitchen             KitchenRepository
	Audit               AuditRepository
	Reminders           ReminderRepository
	Export              ExportRepository
	Documents           DocumentRepository
	Portal              PortalRepository
}
