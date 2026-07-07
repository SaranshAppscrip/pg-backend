package domain

import (
	"time"

	"github.com/google/uuid"
)

// ── Enums ───────────────────────────────────────────────────────────────────

type PaymentMode string

const (
	PaymentModeCash         PaymentMode = "Cash"
	PaymentModeUPI          PaymentMode = "UPI"
	PaymentModeBankTransfer PaymentMode = "Bank Transfer"
	PaymentModeOther        PaymentMode = "Other"
)

type ExpenseCategory string

const (
	ExpenseKitchenSupplies ExpenseCategory = "Kitchen Supplies"
	ExpenseMaintenance     ExpenseCategory = "Maintenance"
	ExpenseElectricity     ExpenseCategory = "Electricity"
	ExpenseWater           ExpenseCategory = "Water"
	ExpenseStaffSalary     ExpenseCategory = "Staff Salary"
	ExpenseRent            ExpenseCategory = "Rent"
	ExpenseOther           ExpenseCategory = "Other"
)

type KitchenUnit string

const (
	KitchenUnitKg      KitchenUnit = "kg"
	KitchenUnitLitre   KitchenUnit = "litre"
	KitchenUnitPacket  KitchenUnit = "packet"
	KitchenUnitPiece   KitchenUnit = "piece"
	KitchenUnitDozen   KitchenUnit = "dozen"
)

type KitchenLogType string

const (
	KitchenLogIn  KitchenLogType = "in"
	KitchenLogOut KitchenLogType = "out"
)

// ── Entities ─────────────────────────────────────────────────────────────────

type Organization struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BusinessSettings struct {
	ID        uuid.UUID `json:"id"`
	PGName    string    `json:"pg_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Staff struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Email          string    `json:"email"`
	PasswordHash   string    `json:"-"`
	FullName       *string   `json:"full_name"`
	IsOwner        bool      `json:"is_owner"`
	CreatedAt      time.Time `json:"created_at"`
}

type StaffProfile struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Email          string    `json:"email"`
	FullName       *string   `json:"full_name"`
	IsOwner        bool      `json:"is_owner"`
	CreatedAt      time.Time `json:"created_at"`
}

type Room struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	RoomNumber     string    `json:"room_number"`
	Capacity       int       `json:"capacity"`
	CreatedAt      time.Time `json:"created_at"`
}

type Tenant struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	Name           string     `json:"name"`
	Email          string     `json:"email"`
	PasswordHash   string     `json:"-"`
	Phone          *string    `json:"phone"`
	RoomID         *uuid.UUID `json:"room_id"`
	MonthlyFee     float64    `json:"monthly_fee"`
	JoinDate       time.Time  `json:"join_date"`
	Active         bool       `json:"active"`
	CreatedAt      time.Time  `json:"created_at"`
}

type Payment struct {
	ID        uuid.UUID   `json:"id"`
	TenantID  uuid.UUID   `json:"tenant_id"`
	Amount    float64     `json:"amount"`
	Date      time.Time   `json:"date"`
	ForMonth  string      `json:"for_month"`
	Mode      PaymentMode `json:"mode"`
	CreatedAt time.Time   `json:"created_at"`
}

type Expense struct {
	ID             uuid.UUID       `json:"id"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	Category       ExpenseCategory `json:"category"`
	Amount         float64         `json:"amount"`
	Date           time.Time       `json:"date"`
	Note           *string         `json:"note"`
	CreatedAt      time.Time       `json:"created_at"`
}

type KitchenItem struct {
	ID               uuid.UUID   `json:"id"`
	OrganizationID   uuid.UUID   `json:"organization_id"`
	Name             string      `json:"name"`
	Qty              float64     `json:"qty"`
	Unit             KitchenUnit `json:"unit"`
	ReorderThreshold float64     `json:"reorder_threshold"`
	CreatedAt        time.Time   `json:"created_at"`
}

type KitchenLog struct {
	ID        uuid.UUID      `json:"id"`
	ItemID    uuid.UUID      `json:"item_id"`
	Type      KitchenLogType `json:"type"`
	Qty       float64        `json:"qty"`
	Date      time.Time      `json:"date"`
	Note      *string        `json:"note"`
	CreatedAt time.Time      `json:"created_at"`
}

type TenantProfile struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Phone          *string   `json:"phone"`
	MonthlyFee     float64   `json:"monthly_fee"`
	JoinDate       time.Time `json:"join_date"`
	RoomNumber     *string   `json:"room_number"`
}

// ── Auth claims ──────────────────────────────────────────────────────────────

type TokenType string

const (
	TokenTypeStaff  TokenType = "staff"
	TokenTypeTenant TokenType = "tenant"
)

type AuthClaims struct {
	Type           TokenType `json:"type"`
	OrganizationID uuid.UUID `json:"organization_id"`
	UserID         uuid.UUID `json:"user_id"`
	Email          string    `json:"email,omitempty"`
	IsOwner        bool      `json:"is_owner,omitempty"`
}

type RefreshToken struct {
	ID             uuid.UUID `json:"id"`
	UserType       TokenType `json:"user_type"`
	UserID         uuid.UUID `json:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Token        string `json:"token"` // alias for access_token (backward compat)
}

type AuthResponse struct {
	TokenPair
	User *StaffProfile `json:"user,omitempty"`
}

type TenantAuthResponse struct {
	TokenPair
	User *TenantProfile `json:"user,omitempty"`
}
