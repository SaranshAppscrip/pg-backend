package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/pkg/apperror"
)

func (s *Store) CreateStaffAuditLog(ctx context.Context, orgID uuid.UUID, staffID *uuid.UUID, entry *domain.StaffAuditLog) error {
	meta, err := json.Marshal(entry.Metadata)
	if err != nil {
		return apperror.Internal("marshal audit metadata", err)
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO staff_audit_log (id, organization_id, staff_id, entity_type, entity_id, action, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, entry.ID, orgID, staffID, entry.EntityType, entry.EntityID, entry.Action, meta)
	return mapPgError(err, "")
}

func (s *Store) ListStaffAuditLog(ctx context.Context, orgID uuid.UUID, limit int) ([]domain.StaffAuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	rows, err := s.pool.Query(ctx, `
		SELECT a.id, a.organization_id, a.staff_id, COALESCE(s.email, ''), a.entity_type, a.entity_id, a.action, a.metadata, a.created_at
		FROM staff_audit_log a
		LEFT JOIN staff s ON s.id = a.staff_id
		WHERE a.organization_id = $1
		ORDER BY a.created_at DESC
		LIMIT $2
	`, orgID, limit)
	if err != nil {
		return nil, mapPgError(err, "")
	}
	defer rows.Close()

	var list []domain.StaffAuditLog
	for rows.Next() {
		var entry domain.StaffAuditLog
		var metaJSON []byte
		if err := rows.Scan(
			&entry.ID, &entry.OrganizationID, &entry.StaffID, &entry.StaffEmail,
			&entry.EntityType, &entry.EntityID, &entry.Action, &metaJSON, &entry.CreatedAt,
		); err != nil {
			return nil, apperror.Internal("scan audit log", err)
		}
		entry.Metadata = map[string]any{}
		if len(metaJSON) > 0 {
			_ = json.Unmarshal(metaJSON, &entry.Metadata)
		}
		list = append(list, entry)
	}
	return list, rows.Err()
}
