package postgres

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	fieldcrypto "github.com/nivas/server/internal/crypto"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/pkg/apperror"
)

func (s *Store) ListAnnouncements(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID, activeOnly bool) ([]domain.Announcement, error) {
	query := `
		SELECT id, organization_id, property_id, title, body, category, pinned, published,
		       expires_at, created_by, created_at, updated_at
		FROM announcements WHERE organization_id = $1`
	args := []any{orgID}
	idx := 2
	if propertyID != nil {
		query += ` AND (property_id IS NULL OR property_id = $` + strconv.Itoa(idx) + `)`
		args = append(args, *propertyID)
		idx++
	}
	if activeOnly {
		query += ` AND published = true AND (expires_at IS NULL OR expires_at > NOW())`
	}
	query += ` ORDER BY pinned DESC, created_at DESC`

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, mapPgError(err, "")
	}
	defer rows.Close()
	return scanAnnouncements(rows)
}

func (s *Store) ListAnnouncementsForTenant(ctx context.Context, orgID, tenantID uuid.UUID) ([]domain.Announcement, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.id, a.organization_id, a.property_id, a.title, a.body, a.category, a.pinned, a.published,
		       a.expires_at, a.created_by, a.created_at, a.updated_at
		FROM announcements a
		LEFT JOIN tenants t ON t.id = $2 AND t.organization_id = $1
		LEFT JOIN rooms r ON r.id = t.room_id
		WHERE a.organization_id = $1
		  AND a.published = true
		  AND (a.expires_at IS NULL OR a.expires_at > NOW())
		  AND (a.property_id IS NULL OR a.property_id = r.property_id)
		ORDER BY a.pinned DESC, a.created_at DESC
	`, orgID, tenantID)
	if err != nil {
		return nil, mapPgError(err, "")
	}
	defer rows.Close()
	return scanAnnouncements(rows)
}

func (s *Store) GetAnnouncement(ctx context.Context, orgID, id uuid.UUID) (*domain.Announcement, error) {
	var a domain.Announcement
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, property_id, title, body, category, pinned, published,
		       expires_at, created_by, created_at, updated_at
		FROM announcements WHERE organization_id = $1 AND id = $2
	`, orgID, id).Scan(
		&a.ID, &a.OrganizationID, &a.PropertyID, &a.Title, &a.Body, &a.Category, &a.Pinned, &a.Published,
		&a.ExpiresAt, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, mapPgError(err, "announcement")
	}
	return &a, nil
}

func (s *Store) CreateAnnouncement(ctx context.Context, a *domain.Announcement) error {
	err := s.pool.QueryRow(ctx, `
		INSERT INTO announcements (id, organization_id, property_id, title, body, category, pinned, published, expires_at, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING created_at, updated_at
	`, a.ID, a.OrganizationID, a.PropertyID, a.Title, a.Body, a.Category, a.Pinned, a.Published, a.ExpiresAt, a.CreatedBy,
	).Scan(&a.CreatedAt, &a.UpdatedAt)
	return mapPgError(err, "announcement")
}

func (s *Store) UpdateAnnouncement(ctx context.Context, orgID, id uuid.UUID, a *domain.Announcement) (*domain.Announcement, error) {
	var out domain.Announcement
	err := s.pool.QueryRow(ctx, `
		UPDATE announcements SET
			property_id = $3, title = $4, body = $5, category = $6, pinned = $7, published = $8, expires_at = $9,
			updated_at = NOW()
		WHERE organization_id = $1 AND id = $2
		RETURNING id, organization_id, property_id, title, body, category, pinned, published,
		          expires_at, created_by, created_at, updated_at
	`, orgID, id, a.PropertyID, a.Title, a.Body, a.Category, a.Pinned, a.Published, a.ExpiresAt,
	).Scan(
		&out.ID, &out.OrganizationID, &out.PropertyID, &out.Title, &out.Body, &out.Category, &out.Pinned, &out.Published,
		&out.ExpiresAt, &out.CreatedBy, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return nil, mapPgError(err, "announcement")
	}
	return &out, nil
}

func (s *Store) DeleteAnnouncement(ctx context.Context, orgID, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM announcements WHERE organization_id = $1 AND id = $2`, orgID, id)
	if err != nil {
		return mapPgError(err, "announcement")
	}
	if tag.RowsAffected() == 0 {
		return apperror.NotFound("announcement")
	}
	return nil
}

func maintenanceOrderClause() string {
	return ` ORDER BY CASE m.priority
		WHEN 'urgent' THEN 4 WHEN 'high' THEN 3 WHEN 'medium' THEN 2 WHEN 'low' THEN 1
	END DESC, m.created_at DESC`
}

func maintenanceSelectColumns() string {
	return `m.id, m.organization_id, m.tenant_id, t.name, rm.room_number,
		m.category, m.title, m.description, m.status, m.priority, m.assigned_to, s.full_name,
		m.staff_note, m.resolved_at, m.created_at, m.updated_at`
}

func maintenanceJoins() string {
	return `FROM maintenance_requests m
		JOIN tenants t ON t.id = m.tenant_id
		LEFT JOIN rooms rm ON rm.id = t.room_id
		LEFT JOIN staff s ON s.id = m.assigned_to`
}

func (s *Store) ListMaintenanceRequests(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID, status *domain.MaintenanceStatus) ([]domain.MaintenanceRequest, error) {
	query := `SELECT ` + maintenanceSelectColumns() + `
		` + maintenanceJoins() + `
		WHERE m.organization_id = $1`
	args := []any{orgID}
	idx := 2
	if propertyID != nil {
		query += ` AND rm.property_id = $` + strconv.Itoa(idx)
		args = append(args, *propertyID)
		idx++
	}
	if status != nil {
		query += ` AND m.status = $` + strconv.Itoa(idx)
		args = append(args, *status)
		idx++
	}
	query += maintenanceOrderClause()

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, mapPgError(err, "")
	}
	defer rows.Close()
	return scanMaintenance(rows)
}

func (s *Store) ListMaintenanceByTenant(ctx context.Context, orgID, tenantID uuid.UUID) ([]domain.MaintenanceRequest, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+maintenanceSelectColumns()+`
		`+maintenanceJoins()+`
		WHERE m.organization_id = $1 AND m.tenant_id = $2
		ORDER BY m.created_at DESC
	`, orgID, tenantID)
	if err != nil {
		return nil, mapPgError(err, "")
	}
	defer rows.Close()
	return scanMaintenance(rows)
}

func (s *Store) GetMaintenanceRequest(ctx context.Context, orgID, id uuid.UUID) (*domain.MaintenanceRequest, error) {
	var m domain.MaintenanceRequest
	err := s.pool.QueryRow(ctx, `
		SELECT `+maintenanceSelectColumns()+`
		`+maintenanceJoins()+`
		WHERE m.organization_id = $1 AND m.id = $2
	`, orgID, id).Scan(
		&m.ID, &m.OrganizationID, &m.TenantID, &m.TenantName, &m.RoomNumber,
		&m.Category, &m.Title, &m.Description, &m.Status, &m.Priority, &m.AssignedTo, &m.AssignedToName,
		&m.StaffNote, &m.ResolvedAt, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, mapPgError(err, "maintenance request")
	}
	return &m, nil
}

func (s *Store) CreateMaintenanceRequest(ctx context.Context, req *domain.MaintenanceRequest) error {
	if req.Priority == "" {
		req.Priority = domain.MaintenancePriorityMedium
	}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO maintenance_requests (id, organization_id, tenant_id, category, title, description, status, priority)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING created_at, updated_at
	`, req.ID, req.OrganizationID, req.TenantID, req.Category, req.Title, req.Description, req.Status, req.Priority,
	).Scan(&req.CreatedAt, &req.UpdatedAt)
	return mapPgError(err, "maintenance request")
}

func (s *Store) UpdateMaintenanceRequest(ctx context.Context, orgID, id uuid.UUID, upd domain.MaintenanceUpdate) (*domain.MaintenanceRequest, error) {
	var resolvedAt *time.Time
	if upd.Status == domain.MaintenanceResolved || upd.Status == domain.MaintenanceClosed {
		now := time.Now()
		resolvedAt = &now
	}
	var m domain.MaintenanceRequest
	err := s.pool.QueryRow(ctx, `
		UPDATE maintenance_requests SET
			status = $3, priority = $4, assigned_to = $5, staff_note = $6, resolved_at = $7, updated_at = NOW()
		WHERE organization_id = $1 AND id = $2
		RETURNING id, organization_id, tenant_id,
		          (SELECT name FROM tenants WHERE id = maintenance_requests.tenant_id),
		          (SELECT r.room_number FROM tenants t JOIN rooms r ON r.id = t.room_id WHERE t.id = maintenance_requests.tenant_id),
		          category, title, description, status, priority, assigned_to,
		          (SELECT full_name FROM staff WHERE id = maintenance_requests.assigned_to),
		          staff_note, resolved_at, created_at, updated_at
	`, orgID, id, upd.Status, upd.Priority, upd.AssignedTo, upd.StaffNote, resolvedAt).Scan(
		&m.ID, &m.OrganizationID, &m.TenantID, &m.TenantName, &m.RoomNumber,
		&m.Category, &m.Title, &m.Description, &m.Status, &m.Priority, &m.AssignedTo, &m.AssignedToName,
		&m.StaffNote, &m.ResolvedAt, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, mapPgError(err, "maintenance request")
	}
	return &m, nil
}

func (s *Store) ListVisitorLog(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID, limit int) ([]domain.VisitorLogEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT v.id, v.organization_id, v.property_id, p.name, v.tenant_id, t.name, rm.room_number,
		       v.visitor_name, v.visitor_phone, v.purpose, v.id_type, v.id_number_last4,
		       v.entry_at, v.exit_at, v.logged_by, v.notes, v.created_at
		FROM visitor_log v
		JOIN properties p ON p.id = v.property_id
		LEFT JOIN tenants t ON t.id = v.tenant_id
		LEFT JOIN rooms rm ON rm.id = t.room_id
		WHERE v.organization_id = $1`
	args := []any{orgID}
	idx := 2
	if propertyID != nil {
		query += fmt.Sprintf(` AND v.property_id = $%d`, idx)
		args = append(args, *propertyID)
		idx++
	}
	query += fmt.Sprintf(` ORDER BY v.entry_at DESC LIMIT $%d`, idx)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, mapPgError(err, "")
	}
	defer rows.Close()
	return scanVisitors(rows)
}

func (s *Store) CreateVisitorEntry(ctx context.Context, entry *domain.VisitorLogEntry) error {
	err := s.pool.QueryRow(ctx, `
		INSERT INTO visitor_log (
			id, organization_id, property_id, tenant_id, visitor_name, visitor_phone,
			purpose, id_type, id_number_encrypted, id_number_last4, entry_at, logged_by, notes
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING created_at
	`, entry.ID, entry.OrganizationID, entry.PropertyID, entry.TenantID, entry.VisitorName, entry.VisitorPhone,
		entry.Purpose, entry.IDType, entry.IDNumberEncrypted, entry.IDNumberLast4, entry.EntryAt, entry.LoggedBy, entry.Notes,
	).Scan(&entry.CreatedAt)
	return mapPgError(err, "visitor log")
}

func (s *Store) RecordVisitorExit(ctx context.Context, orgID, id uuid.UUID, exitAt time.Time) (*domain.VisitorLogEntry, error) {
	var v domain.VisitorLogEntry
	var idLast4 *string
	err := s.pool.QueryRow(ctx, `
		UPDATE visitor_log SET exit_at = $3
		WHERE organization_id = $1 AND id = $2 AND exit_at IS NULL
		RETURNING id, organization_id, property_id,
		          (SELECT name FROM properties WHERE id = visitor_log.property_id),
		          tenant_id,
		          (SELECT name FROM tenants WHERE id = visitor_log.tenant_id),
		          (SELECT r.room_number FROM tenants t JOIN rooms r ON r.id = t.room_id WHERE t.id = visitor_log.tenant_id),
		          visitor_name, visitor_phone, purpose, id_type, id_number_last4,
		          entry_at, exit_at, logged_by, notes, created_at
	`, orgID, id, exitAt).Scan(
		&v.ID, &v.OrganizationID, &v.PropertyID, &v.PropertyName, &v.TenantID, &v.TenantName, &v.RoomNumber,
		&v.VisitorName, &v.VisitorPhone, &v.Purpose, &v.IDType, &idLast4,
		&v.EntryAt, &v.ExitAt, &v.LoggedBy, &v.Notes, &v.CreatedAt,
	)
	if err != nil {
		return nil, mapPgError(err, "visitor log")
	}
	v.IDNumber = maskVisitorIDNumber(idLast4)
	return &v, nil
}

func scanAnnouncements(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]domain.Announcement, error) {
	var list []domain.Announcement
	for rows.Next() {
		var a domain.Announcement
		if err := rows.Scan(
			&a.ID, &a.OrganizationID, &a.PropertyID, &a.Title, &a.Body, &a.Category, &a.Pinned, &a.Published,
			&a.ExpiresAt, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, apperror.Internal("scan announcement", err)
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func scanMaintenance(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]domain.MaintenanceRequest, error) {
	var list []domain.MaintenanceRequest
	for rows.Next() {
		var m domain.MaintenanceRequest
		if err := rows.Scan(
			&m.ID, &m.OrganizationID, &m.TenantID, &m.TenantName, &m.RoomNumber,
			&m.Category, &m.Title, &m.Description, &m.Status, &m.Priority, &m.AssignedTo, &m.AssignedToName,
			&m.StaffNote, &m.ResolvedAt, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, apperror.Internal("scan maintenance", err)
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

func scanVisitors(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]domain.VisitorLogEntry, error) {
	var list []domain.VisitorLogEntry
	for rows.Next() {
		var v domain.VisitorLogEntry
		var idLast4 *string
		if err := rows.Scan(
			&v.ID, &v.OrganizationID, &v.PropertyID, &v.PropertyName, &v.TenantID, &v.TenantName, &v.RoomNumber,
			&v.VisitorName, &v.VisitorPhone, &v.Purpose, &v.IDType, &idLast4,
			&v.EntryAt, &v.ExitAt, &v.LoggedBy, &v.Notes, &v.CreatedAt,
		); err != nil {
			return nil, apperror.Internal("scan visitor", err)
		}
		v.IDNumber = maskVisitorIDNumber(idLast4)
		list = append(list, v)
	}
	return list, rows.Err()
}

func maskVisitorIDNumber(last4 *string) *string {
	if last4 == nil || *last4 == "" {
		return nil
	}
	masked := fieldcrypto.MaskIDNumber(*last4)
	return &masked
}
