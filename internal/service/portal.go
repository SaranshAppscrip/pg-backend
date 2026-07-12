package service

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	fieldcrypto "github.com/nivas/server/internal/crypto"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/notification"
	"github.com/nivas/server/internal/repository"
	"github.com/nivas/server/pkg/apperror"
)

type PortalService struct {
	repos      repository.Store
	email      notification.EmailSender
	frontendURL string
	encryptor  *fieldcrypto.FieldEncryptor
	log        *slog.Logger
}

func NewPortalService(repos repository.Store, email notification.EmailSender, frontendURL, jwtSecret string, log *slog.Logger) *PortalService {
	enc, err := fieldcrypto.NewFieldEncryptor(jwtSecret)
	if err != nil && log != nil {
		log.Warn("visitor ID encryption disabled", "error", err)
	}
	return &PortalService{
		repos:       repos,
		email:       email,
		frontendURL: strings.TrimRight(frontendURL, "/"),
		encryptor:   enc,
		log:         log,
	}
}

// ── Announcements ────────────────────────────────────────────────────────────

func (s *PortalService) ListAnnouncements(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID) ([]domain.Announcement, error) {
	list, err := s.repos.Portal.ListAnnouncements(ctx, orgID, propertyID, false)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []domain.Announcement{}
	}
	return list, nil
}

func (s *PortalService) ListAnnouncementsForTenant(ctx context.Context, orgID, tenantID uuid.UUID) ([]domain.Announcement, error) {
	list, err := s.repos.Portal.ListAnnouncementsForTenant(ctx, orgID, tenantID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []domain.Announcement{}
	}
	return list, nil
}

type AnnouncementInput struct {
	PropertyID *uuid.UUID
	Title      string
	Body       string
	Category   domain.AnnouncementCategory
	Pinned     bool
	Published  bool
	ExpiresAt  *time.Time
}

func (s *PortalService) CreateAnnouncement(ctx context.Context, orgID, staffID uuid.UUID, in AnnouncementInput) (*domain.Announcement, error) {
	title := strings.TrimSpace(in.Title)
	body := strings.TrimSpace(in.Body)
	if title == "" || body == "" {
		return nil, apperror.BadRequest("title and body are required")
	}
	if in.PropertyID != nil {
		if _, err := s.repos.Properties.GetByID(ctx, orgID, *in.PropertyID); err != nil {
			return nil, err
		}
	}
	a := &domain.Announcement{
		ID:             uuid.New(),
		OrganizationID: orgID,
		PropertyID:     in.PropertyID,
		Title:          title,
		Body:           body,
		Category:       in.Category,
		Pinned:         in.Pinned,
		Published:      in.Published,
		ExpiresAt:      in.ExpiresAt,
		CreatedBy:      &staffID,
	}
	if err := s.repos.Portal.CreateAnnouncement(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *PortalService) UpdateAnnouncement(ctx context.Context, orgID, id uuid.UUID, in AnnouncementInput) (*domain.Announcement, error) {
	if _, err := s.repos.Portal.GetAnnouncement(ctx, orgID, id); err != nil {
		return nil, err
	}
	title := strings.TrimSpace(in.Title)
	body := strings.TrimSpace(in.Body)
	if title == "" || body == "" {
		return nil, apperror.BadRequest("title and body are required")
	}
	if in.PropertyID != nil {
		if _, err := s.repos.Properties.GetByID(ctx, orgID, *in.PropertyID); err != nil {
			return nil, err
		}
	}
	return s.repos.Portal.UpdateAnnouncement(ctx, orgID, id, &domain.Announcement{
		PropertyID: in.PropertyID, Title: title, Body: body, Category: in.Category,
		Pinned: in.Pinned, Published: in.Published, ExpiresAt: in.ExpiresAt,
	})
}

func (s *PortalService) DeleteAnnouncement(ctx context.Context, orgID, id uuid.UUID) error {
	return s.repos.Portal.DeleteAnnouncement(ctx, orgID, id)
}

// ── Maintenance ──────────────────────────────────────────────────────────────

func (s *PortalService) ListMaintenance(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID, status *domain.MaintenanceStatus) ([]domain.MaintenanceRequest, error) {
	list, err := s.repos.Portal.ListMaintenanceRequests(ctx, orgID, propertyID, status)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []domain.MaintenanceRequest{}
	}
	return list, nil
}

func (s *PortalService) ListMaintenanceForTenant(ctx context.Context, orgID, tenantID uuid.UUID) ([]domain.MaintenanceRequest, error) {
	list, err := s.repos.Portal.ListMaintenanceByTenant(ctx, orgID, tenantID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []domain.MaintenanceRequest{}
	}
	return list, nil
}

type MaintenanceInput struct {
	Category    domain.MaintenanceCategory
	Title       string
	Description string
}

func (s *PortalService) CreateMaintenance(ctx context.Context, orgID, tenantID uuid.UUID, in MaintenanceInput) (*domain.MaintenanceRequest, error) {
	tenant, err := s.repos.Tenants.GetByID(ctx, orgID, tenantID)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(in.Title)
	desc := strings.TrimSpace(in.Description)
	if title == "" || desc == "" {
		return nil, apperror.BadRequest("title and description are required")
	}
	req := &domain.MaintenanceRequest{
		ID:             uuid.New(),
		OrganizationID: orgID,
		TenantID:       tenantID,
		Category:       in.Category,
		Title:          title,
		Description:    desc,
		Status:         domain.MaintenanceOpen,
		Priority:       domain.MaintenancePriorityMedium,
	}
	if err := s.repos.Portal.CreateMaintenanceRequest(ctx, req); err != nil {
		return nil, err
	}
	created, err := s.repos.Portal.GetMaintenanceRequest(ctx, orgID, req.ID)
	if err != nil {
		return nil, err
	}
	s.notifyStaffNewMaintenance(ctx, orgID, tenant.Name, created)
	return created, nil
}

type MaintenanceUpdateInput struct {
	Status     domain.MaintenanceStatus
	Priority   domain.MaintenancePriority
	AssignedTo *uuid.UUID
	StaffNote  string
}

func (s *PortalService) UpdateMaintenance(ctx context.Context, orgID, id uuid.UUID, in MaintenanceUpdateInput) (*domain.MaintenanceRequest, error) {
	if _, err := s.repos.Portal.GetMaintenanceRequest(ctx, orgID, id); err != nil {
		return nil, err
	}
	if in.AssignedTo != nil {
		if _, err := s.repos.Staff.GetByID(ctx, orgID, *in.AssignedTo); err != nil {
			return nil, apperror.BadRequest("invalid assigned_to staff member")
		}
	}
	var note *string
	if strings.TrimSpace(in.StaffNote) != "" {
		n := strings.TrimSpace(in.StaffNote)
		note = &n
	}
	priority := in.Priority
	if priority == "" {
		priority = domain.MaintenancePriorityMedium
	}
	return s.repos.Portal.UpdateMaintenanceRequest(ctx, orgID, id, domain.MaintenanceUpdate{
		Status: in.Status, Priority: priority, AssignedTo: in.AssignedTo, StaffNote: note,
	})
}

func (s *PortalService) notifyStaffNewMaintenance(ctx context.Context, orgID uuid.UUID, tenantName string, req *domain.MaintenanceRequest) {
	if s.email == nil {
		return
	}
	staffList, err := s.repos.Staff.List(ctx, orgID)
	if err != nil {
		if s.log != nil {
			s.log.Warn("list staff for maintenance alert", "error", err)
		}
		return
	}
	opsURL := s.frontendURL + "/operations"
	for _, st := range staffList {
		if err := s.email.SendMaintenanceAlert(ctx, notification.MaintenanceAlertParams{
			To:         st.Email,
			TenantName: tenantName,
			Title:      req.Title,
			Category:   string(req.Category),
			RoomNumber: req.RoomNumber,
			OpsURL:     opsURL,
		}); err != nil && s.log != nil {
			s.log.Warn("send maintenance alert", "to", st.Email, "error", err)
		}
	}
}

// ── Visitors ─────────────────────────────────────────────────────────────────

func (s *PortalService) ListVisitors(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID, limit int) ([]domain.VisitorLogEntry, error) {
	list, err := s.repos.Portal.ListVisitorLog(ctx, orgID, propertyID, limit)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []domain.VisitorLogEntry{}
	}
	return list, nil
}

type VisitorEntryInput struct {
	PropertyID   uuid.UUID
	TenantID     *uuid.UUID
	VisitorName  string
	VisitorPhone string
	Purpose      string
	IDType       string
	IDNumber     string
	EntryAt      time.Time
	Notes        string
}

func (s *PortalService) LogVisitorEntry(ctx context.Context, orgID, staffID uuid.UUID, in VisitorEntryInput) (*domain.VisitorLogEntry, error) {
	name := strings.TrimSpace(in.VisitorName)
	if name == "" {
		return nil, apperror.BadRequest("visitor_name is required")
	}
	if _, err := s.repos.Properties.GetByID(ctx, orgID, in.PropertyID); err != nil {
		return nil, err
	}
	if in.TenantID != nil {
		if _, err := s.repos.Tenants.GetByID(ctx, orgID, *in.TenantID); err != nil {
			return nil, err
		}
	}
	entry := &domain.VisitorLogEntry{
		ID:             uuid.New(),
		OrganizationID: orgID,
		PropertyID:     in.PropertyID,
		TenantID:       in.TenantID,
		VisitorName:    name,
		EntryAt:        in.EntryAt,
		LoggedBy:       &staffID,
	}
	if in.VisitorPhone != "" {
		p := strings.TrimSpace(in.VisitorPhone)
		entry.VisitorPhone = &p
	}
	if in.Purpose != "" {
		p := strings.TrimSpace(in.Purpose)
		entry.Purpose = &p
	}
	if in.IDType != "" {
		t := strings.TrimSpace(in.IDType)
		entry.IDType = &t
	}
	if rawID := strings.TrimSpace(in.IDNumber); rawID != "" {
		last4 := fieldcrypto.Last4(rawID)
		if last4 != "" {
			entry.IDNumberLast4 = &last4
		}
		if s.encryptor != nil {
			enc, err := s.encryptor.Encrypt(rawID)
			if err != nil {
				return nil, apperror.Internal("encrypt visitor ID", err)
			}
			entry.IDNumberEncrypted = &enc
		}
		masked := fieldcrypto.MaskIDNumber(last4)
		if masked != "" {
			entry.IDNumber = &masked
		}
	}
	if in.Notes != "" {
		n := strings.TrimSpace(in.Notes)
		entry.Notes = &n
	}
	if entry.EntryAt.IsZero() {
		entry.EntryAt = time.Now()
	}
	if err := s.repos.Portal.CreateVisitorEntry(ctx, entry); err != nil {
		return nil, err
	}
	list, err := s.repos.Portal.ListVisitorLog(ctx, orgID, &in.PropertyID, 1)
	if err == nil && len(list) > 0 && list[0].ID == entry.ID {
		return &list[0], nil
	}
	return entry, nil
}

func (s *PortalService) LogVisitorExit(ctx context.Context, orgID, id uuid.UUID) (*domain.VisitorLogEntry, error) {
	return s.repos.Portal.RecordVisitorExit(ctx, orgID, id, time.Now())
}
