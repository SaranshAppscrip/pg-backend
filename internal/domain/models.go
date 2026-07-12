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
	PropertyID     uuid.UUID `json:"property_id"`
	RoomNumber     string    `json:"room_number"`
	Capacity       int       `json:"capacity"`
	CreatedAt      time.Time `json:"created_at"`
}

type Property struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Name           string    `json:"name"`
	Address        *string   `json:"address"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
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
	DeletedAt *time.Time  `json:"deleted_at,omitempty"`
}

type Expense struct {
	ID             uuid.UUID       `json:"id"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	Category       ExpenseCategory `json:"category"`
	Amount         float64         `json:"amount"`
	Date           time.Time       `json:"date"`
	Note           *string         `json:"note"`
	CreatedAt      time.Time       `json:"created_at"`
	DeletedAt      *time.Time      `json:"deleted_at,omitempty"`
}

// ── Audit log ────────────────────────────────────────────────────────────────

type AuditEntityType string

const (
	AuditEntityPayment AuditEntityType = "payment"
	AuditEntityExpense AuditEntityType = "expense"
	AuditEntityTenant  AuditEntityType = "tenant"
)

type AuditAction string

const (
	AuditActionCreate  AuditAction = "create"
	AuditActionDelete  AuditAction = "delete"
	AuditActionMoveOut AuditAction = "move_out"
)

type StaffAuditLog struct {
	ID             uuid.UUID       `json:"id"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	StaffID        *uuid.UUID      `json:"staff_id"`
	StaffEmail     string          `json:"staff_email,omitempty"`
	EntityType     AuditEntityType `json:"entity_type"`
	EntityID       uuid.UUID       `json:"entity_id"`
	Action         AuditAction     `json:"action"`
	Metadata       map[string]any  `json:"metadata"`
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

// ── Documents ────────────────────────────────────────────────────────────────

type TenantDocumentType string

const (
	TenantDocIDProof            TenantDocumentType = "id_proof"
	TenantDocLeaseAgreement     TenantDocumentType = "lease_agreement"
	TenantDocPoliceVerification TenantDocumentType = "police_verification"
	TenantDocPhoto              TenantDocumentType = "photo"
	TenantDocOther              TenantDocumentType = "other"
)

type OrganizationDocumentType string

const (
	OrgDocPGRegistration  OrganizationDocumentType = "pg_registration"
	OrgDocFireSafetyNOC   OrganizationDocumentType = "fire_safety_noc"
	OrgDocPolicePermission OrganizationDocumentType = "police_permission"
	OrgDocTradeLicense    OrganizationDocumentType = "trade_license"
	OrgDocPropertyTax     OrganizationDocumentType = "property_tax"
	OrgDocOther           OrganizationDocumentType = "other"
)

type TenantDocument struct {
	ID               uuid.UUID          `json:"id"`
	OrganizationID   uuid.UUID          `json:"organization_id"`
	TenantID         uuid.UUID          `json:"tenant_id"`
	DocumentType     TenantDocumentType `json:"document_type"`
	Title            *string            `json:"title"`
	OriginalFilename string             `json:"original_filename"`
	ContentType      string             `json:"content_type"`
	SizeBytes        int64              `json:"size_bytes"`
	UploadedBy       *uuid.UUID         `json:"uploaded_by"`
	ExpiresAt        *time.Time         `json:"expires_at"`
	CreatedAt        time.Time          `json:"created_at"`
}

type OrganizationDocument struct {
	ID               uuid.UUID                `json:"id"`
	OrganizationID   uuid.UUID                `json:"organization_id"`
	PropertyID       *uuid.UUID               `json:"property_id"`
	DocumentType     OrganizationDocumentType `json:"document_type"`
	Title            *string                  `json:"title"`
	OriginalFilename string                   `json:"original_filename"`
	ContentType      string                   `json:"content_type"`
	SizeBytes        int64                    `json:"size_bytes"`
	UploadedBy       *uuid.UUID               `json:"uploaded_by"`
	ExpiresAt        *time.Time               `json:"expires_at"`
	CreatedAt        time.Time                `json:"created_at"`
}

// ── Portal (announcements, maintenance, visitors) ────────────────────────────

type AnnouncementCategory string

const (
	AnnouncementMaintenance AnnouncementCategory = "maintenance"
	AnnouncementHoliday     AnnouncementCategory = "holiday"
	AnnouncementRules       AnnouncementCategory = "rules"
	AnnouncementGeneral     AnnouncementCategory = "general"
)

type Announcement struct {
	ID             uuid.UUID            `json:"id"`
	OrganizationID uuid.UUID            `json:"organization_id"`
	PropertyID     *uuid.UUID           `json:"property_id"`
	Title          string               `json:"title"`
	Body           string               `json:"body"`
	Category       AnnouncementCategory `json:"category"`
	Pinned         bool                 `json:"pinned"`
	Published      bool                 `json:"published"`
	ExpiresAt      *time.Time           `json:"expires_at"`
	CreatedBy      *uuid.UUID           `json:"created_by"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

type MaintenanceCategory string

const (
	MaintenanceElectrical MaintenanceCategory = "electrical"
	MaintenancePlumbing   MaintenanceCategory = "plumbing"
	MaintenanceWifi       MaintenanceCategory = "wifi"
	MaintenanceCleaning     MaintenanceCategory = "cleaning"
	MaintenanceOther        MaintenanceCategory = "other"
)

type MaintenanceStatus string

const (
	MaintenanceOpen       MaintenanceStatus = "open"
	MaintenanceInProgress MaintenanceStatus = "in_progress"
	MaintenanceResolved   MaintenanceStatus = "resolved"
	MaintenanceClosed     MaintenanceStatus = "closed"
)

type MaintenancePriority string

const (
	MaintenancePriorityLow    MaintenancePriority = "low"
	MaintenancePriorityMedium MaintenancePriority = "medium"
	MaintenancePriorityHigh   MaintenancePriority = "high"
	MaintenancePriorityUrgent MaintenancePriority = "urgent"
)

type MaintenanceRequest struct {
	ID             uuid.UUID           `json:"id"`
	OrganizationID uuid.UUID           `json:"organization_id"`
	TenantID       uuid.UUID           `json:"tenant_id"`
	TenantName     string              `json:"tenant_name,omitempty"`
	RoomNumber     *string             `json:"room_number,omitempty"`
	Category       MaintenanceCategory `json:"category"`
	Title          string              `json:"title"`
	Description    string              `json:"description"`
	Status         MaintenanceStatus   `json:"status"`
	Priority       MaintenancePriority `json:"priority"`
	AssignedTo     *uuid.UUID          `json:"assigned_to"`
	AssignedToName *string             `json:"assigned_to_name,omitempty"`
	StaffNote      *string             `json:"staff_note"`
	ResolvedAt     *time.Time          `json:"resolved_at"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

type MaintenanceUpdate struct {
	Status     MaintenanceStatus
	Priority   MaintenancePriority
	AssignedTo *uuid.UUID
	StaffNote  *string
}

type VisitorLogEntry struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	PropertyID     uuid.UUID  `json:"property_id"`
	PropertyName   string     `json:"property_name,omitempty"`
	TenantID       *uuid.UUID `json:"tenant_id"`
	TenantName     *string    `json:"tenant_name,omitempty"`
	RoomNumber     *string    `json:"room_number,omitempty"`
	VisitorName    string     `json:"visitor_name"`
	VisitorPhone   *string    `json:"visitor_phone"`
	Purpose            *string `json:"purpose"`
	IDType             *string `json:"id_type"`
	IDNumber           *string `json:"id_number"` // masked display (last 4 only)
	IDNumberEncrypted  *string `json:"-"`
	IDNumberLast4      *string `json:"-"`
	EntryAt        time.Time  `json:"entry_at"`
	ExitAt         *time.Time `json:"exit_at"`
	LoggedBy       *uuid.UUID `json:"logged_by"`
	Notes          *string    `json:"notes"`
	CreatedAt      time.Time  `json:"created_at"`
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
