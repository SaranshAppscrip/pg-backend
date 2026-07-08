package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/repository"
	"github.com/nivas/server/pkg/apperror"
)

type TenantDueRow = repository.ReminderTenantRow

func listActiveTenantsWithDues(s *Store, ctx context.Context, orgID *uuid.UUID) ([]repository.ReminderTenantRow, error) {
	query := `
		SELECT t.id, t.organization_id, r.property_id, t.name, t.email, t.monthly_fee, p.name
		FROM tenants t
		JOIN rooms r ON r.id = t.room_id
		JOIN properties p ON p.id = r.property_id
		WHERE t.active = true AND t.email <> ''
	`
	args := []any{}
	if orgID != nil {
		query += ` AND t.organization_id = $1`
		args = append(args, *orgID)
	}
	query += ` ORDER BY t.organization_id, p.name, t.name`

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, mapPgError(err, "")
	}
	defer rows.Close()

	var list []repository.ReminderTenantRow
	for rows.Next() {
		var row repository.ReminderTenantRow
		if err := rows.Scan(&row.TenantID, &row.OrganizationID, &row.PropertyID, &row.Name, &row.Email, &row.MonthlyFee, &row.PropertyName); err != nil {
			return nil, apperror.Internal("scan tenant due row", err)
		}
		list = append(list, row)
	}
	return list, rows.Err()
}

func (s *Store) HasRentReminder(ctx context.Context, tenantID uuid.UUID, forMonth, reminderType string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM rent_reminder_log
			WHERE tenant_id = $1 AND for_month = $2 AND reminder_type = $3
		)
	`, tenantID, forMonth, reminderType).Scan(&exists)
	return exists, mapPgError(err, "")
}

func (s *Store) CreateRentReminder(ctx context.Context, tenantID uuid.UUID, forMonth, reminderType string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO rent_reminder_log (tenant_id, for_month, reminder_type)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`, tenantID, forMonth, reminderType)
	return mapPgError(err, "")
}

func (s *Store) ListPaymentsForTenantMonth(ctx context.Context, tenantID uuid.UUID, forMonth string) ([]domain.Payment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, tenant_id, amount, date, for_month, mode, created_at
		FROM payments
		WHERE tenant_id = $1 AND for_month = $2 AND deleted_at IS NULL
	`, tenantID, forMonth)
	if err != nil {
		return nil, mapPgError(err, "")
	}
	defer rows.Close()

	var list []domain.Payment
	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Amount, &p.Date, &p.ForMonth, &p.Mode, &p.CreatedAt); err != nil {
			return nil, apperror.Internal("scan payment", err)
		}
		list = append(list, p)
	}
	return list, rows.Err()
}
