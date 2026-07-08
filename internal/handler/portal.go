package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/middleware"
	"github.com/nivas/server/internal/service"
	"github.com/nivas/server/pkg/apperror"
	"github.com/nivas/server/pkg/response"
)

type PortalHandler struct {
	svc *service.PortalService
}

func NewPortalHandler(svc *service.PortalService) *PortalHandler {
	return &PortalHandler{svc: svc}
}

func parseOptionalTime(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05Z07:00", raw)
	}
	if err != nil {
		t, err = time.Parse("2006-01-02", raw)
	}
	if err != nil {
		return nil, apperror.BadRequest("invalid datetime format")
	}
	return &t, nil
}

// ── Announcements (staff) ────────────────────────────────────────────────────

func (h *PortalHandler) ListAnnouncements(c *gin.Context) {
	propertyID, err := middleware.OptionalPropertyID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	list, err := h.svc.ListAnnouncements(c.Request.Context(), middleware.GetOrganizationID(c), propertyID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

type announcementRequest struct {
	PropertyID *string `json:"property_id"`
	Title      string  `json:"title"`
	Body       string  `json:"body"`
	Category   string  `json:"category"`
	Pinned     bool    `json:"pinned"`
	Published  bool    `json:"published"`
	ExpiresAt  *string `json:"expires_at"`
}

func (h *PortalHandler) parseAnnouncementReq(req announcementRequest) (service.AnnouncementInput, error) {
	var propertyID *uuid.UUID
	if req.PropertyID != nil && strings.TrimSpace(*req.PropertyID) != "" {
		id, err := uuid.Parse(*req.PropertyID)
		if err != nil {
			return service.AnnouncementInput{}, apperror.BadRequest("invalid property_id")
		}
		propertyID = &id
	}
	var expiresAt *time.Time
	if req.ExpiresAt != nil && strings.TrimSpace(*req.ExpiresAt) != "" {
		t, err := parseOptionalTime(*req.ExpiresAt)
		if err != nil {
			return service.AnnouncementInput{}, err
		}
		expiresAt = t
	}
	return service.AnnouncementInput{
		PropertyID: propertyID,
		Title:      req.Title,
		Body:       req.Body,
		Category:   domain.AnnouncementCategory(req.Category),
		Pinned:     req.Pinned,
		Published:  req.Published,
		ExpiresAt:  expiresAt,
	}, nil
}

func (h *PortalHandler) CreateAnnouncement(c *gin.Context) {
	var req announcementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	in, err := h.parseAnnouncementReq(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	a, err := h.svc.CreateAnnouncement(c.Request.Context(), middleware.GetOrganizationID(c), middleware.GetUserID(c), in)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, a)
}

func (h *PortalHandler) UpdateAnnouncement(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	var req announcementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	in, err := h.parseAnnouncementReq(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	a, err := h.svc.UpdateAnnouncement(c.Request.Context(), middleware.GetOrganizationID(c), id, in)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, a)
}

func (h *PortalHandler) DeleteAnnouncement(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteAnnouncement(c.Request.Context(), middleware.GetOrganizationID(c), id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

func (h *PortalHandler) TenantAnnouncements(c *gin.Context) {
	list, err := h.svc.ListAnnouncementsForTenant(c.Request.Context(), middleware.GetOrganizationID(c), middleware.GetUserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

// ── Maintenance ──────────────────────────────────────────────────────────────

func (h *PortalHandler) ListMaintenance(c *gin.Context) {
	propertyID, err := middleware.OptionalPropertyID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var status *domain.MaintenanceStatus
	if s := c.Query("status"); s != "" {
		st := domain.MaintenanceStatus(s)
		status = &st
	}
	list, err := h.svc.ListMaintenance(c.Request.Context(), middleware.GetOrganizationID(c), propertyID, status)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

func (h *PortalHandler) TenantMaintenance(c *gin.Context) {
	list, err := h.svc.ListMaintenanceForTenant(c.Request.Context(), middleware.GetOrganizationID(c), middleware.GetUserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

type maintenanceCreateRequest struct {
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (h *PortalHandler) CreateMaintenance(c *gin.Context) {
	var req maintenanceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	m, err := h.svc.CreateMaintenance(c.Request.Context(), middleware.GetOrganizationID(c), middleware.GetUserID(c), service.MaintenanceInput{
		Category: domain.MaintenanceCategory(req.Category), Title: req.Title, Description: req.Description,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, m)
}

type maintenanceUpdateRequest struct {
	Status    string `json:"status"`
	StaffNote string `json:"staff_note"`
}

func (h *PortalHandler) UpdateMaintenance(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	var req maintenanceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	m, err := h.svc.UpdateMaintenance(c.Request.Context(), middleware.GetOrganizationID(c), id, domain.MaintenanceStatus(req.Status), req.StaffNote)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, m)
}

// ── Visitors ───────────────────────────────────────────────────────────────────

func (h *PortalHandler) ListVisitors(c *gin.Context) {
	propertyID, err := middleware.OptionalPropertyID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	list, err := h.svc.ListVisitors(c.Request.Context(), middleware.GetOrganizationID(c), propertyID, limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

type visitorEntryRequest struct {
	PropertyID   string  `json:"property_id"`
	TenantID     *string `json:"tenant_id"`
	VisitorName  string  `json:"visitor_name"`
	VisitorPhone string  `json:"visitor_phone"`
	Purpose      string  `json:"purpose"`
	IDType       string  `json:"id_type"`
	IDNumber     string  `json:"id_number"`
	EntryAt      string  `json:"entry_at"`
	Notes        string  `json:"notes"`
}

func (h *PortalHandler) CreateVisitorEntry(c *gin.Context) {
	var req visitorEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	propertyID, err := uuid.Parse(req.PropertyID)
	if err != nil {
		response.Error(c, apperror.BadRequest("invalid property_id"))
		return
	}
	var tenantID *uuid.UUID
	if req.TenantID != nil && strings.TrimSpace(*req.TenantID) != "" {
		id, err := uuid.Parse(*req.TenantID)
		if err != nil {
			response.Error(c, apperror.BadRequest("invalid tenant_id"))
			return
		}
		tenantID = &id
	}
	var entryAt time.Time
	if req.EntryAt != "" {
		t, err := parseOptionalTime(req.EntryAt)
		if err != nil {
			response.Error(c, err)
			return
		}
		if t != nil {
			entryAt = *t
		}
	}
	entry, err := h.svc.LogVisitorEntry(c.Request.Context(), middleware.GetOrganizationID(c), middleware.GetUserID(c), service.VisitorEntryInput{
		PropertyID: propertyID, TenantID: tenantID, VisitorName: req.VisitorName,
		VisitorPhone: req.VisitorPhone, Purpose: req.Purpose, IDType: req.IDType,
		IDNumber: req.IDNumber, EntryAt: entryAt, Notes: req.Notes,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, entry)
}

func (h *PortalHandler) RecordVisitorExit(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	entry, err := h.svc.LogVisitorExit(c.Request.Context(), middleware.GetOrganizationID(c), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, entry)
}
