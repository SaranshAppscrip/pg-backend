package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/repository"
	"github.com/nivas/server/pkg/apperror"
)

func (s *Store) ListPayments(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID) ([]domain.Payment, error) {
	query := `
		SELECT p.id, p.tenant_id, p.amount, p.date, p.for_month, p.mode, p.created_at
		FROM payments p
		JOIN tenants t ON t.id = p.tenant_id
		LEFT JOIN rooms r ON r.id = t.room_id
		WHERE p.organization_id = $1 AND p.deleted_at IS NULL`
	args := []any{orgID}
	if propertyID != nil {
		query += ` AND r.property_id = $2`
		args = append(args, *propertyID)
	}
	query += ` ORDER BY p.date DESC, p.created_at DESC`

	rows, err := s.pool.Query(ctx, query, args...)
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

func (s *Store) CreatePayment(ctx context.Context, orgID uuid.UUID, payment *domain.Payment) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO payments (id, organization_id, tenant_id, amount, date, for_month, mode)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, payment.ID, orgID, payment.TenantID, payment.Amount, payment.Date, payment.ForMonth, payment.Mode)
	return mapPgError(err, "")
}

func (s *Store) SoftDeletePayment(ctx context.Context, orgID, id uuid.UUID) (*domain.Payment, error) {
	var p domain.Payment
	err := s.pool.QueryRow(ctx, `
		UPDATE payments SET deleted_at = NOW()
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL
		RETURNING id, tenant_id, amount, date, for_month, mode, created_at
	`, id, orgID).Scan(&p.ID, &p.TenantID, &p.Amount, &p.Date, &p.ForMonth, &p.Mode, &p.CreatedAt)
	if err != nil {
		return nil, mapPgError(err, "payment not found")
	}
	return &p, nil
}

func (s *Store) ListPaymentsByTenant(ctx context.Context, orgID, tenantID uuid.UUID) ([]domain.Payment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, tenant_id, amount, date, for_month, mode, created_at
		FROM payments
		WHERE organization_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
		ORDER BY date DESC
	`, orgID, tenantID)
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

func (s *Store) ListPaymentsForExport(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID) ([]repository.PaymentExportRow, error) {
	query := `
		SELECT p.id, p.date::text, t.name, COALESCE(r.room_number, ''), COALESCE(pr.name, ''), p.for_month, p.amount, p.mode::text
		FROM payments p
		JOIN tenants t ON t.id = p.tenant_id
		LEFT JOIN rooms r ON r.id = t.room_id
		LEFT JOIN properties pr ON pr.id = r.property_id
		WHERE p.organization_id = $1 AND p.deleted_at IS NULL`
	args := []any{orgID}
	if propertyID != nil {
		query += ` AND r.property_id = $2`
		args = append(args, *propertyID)
	}
	query += ` ORDER BY p.date DESC, p.created_at DESC`

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, mapPgError(err, "")
	}
	defer rows.Close()

	var list []repository.PaymentExportRow
	for rows.Next() {
		var row repository.PaymentExportRow
		if err := rows.Scan(&row.ID, &row.Date, &row.TenantName, &row.RoomNumber, &row.PropertyName, &row.ForMonth, &row.Amount, &row.Mode); err != nil {
			return nil, apperror.Internal("scan payment export", err)
		}
		list = append(list, row)
	}
	return list, rows.Err()
}

func (s *Store) GetPaymentReceiptData(ctx context.Context, orgID, paymentID uuid.UUID) (*repository.PaymentReceiptData, error) {
	var d repository.PaymentReceiptData
	err := s.pool.QueryRow(ctx, `
		SELECT p.id, p.amount, p.date::text, p.for_month, p.mode::text,
		       t.name, t.email, COALESCE(r.room_number, ''), COALESCE(pr.name, ''), o.name
		FROM payments p
		JOIN tenants t ON t.id = p.tenant_id
		JOIN organizations o ON o.id = p.organization_id
		LEFT JOIN rooms r ON r.id = t.room_id
		LEFT JOIN properties pr ON pr.id = r.property_id
		WHERE p.id = $1 AND p.organization_id = $2 AND p.deleted_at IS NULL
	`, paymentID, orgID).Scan(
		&d.PaymentID, &d.Amount, &d.Date, &d.ForMonth, &d.Mode,
		&d.TenantName, &d.TenantEmail, &d.RoomNumber, &d.PropertyName, &d.OrganizationName,
	)
	if err != nil {
		return nil, mapPgError(err, "payment not found")
	}
	return &d, nil
}
