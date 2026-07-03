package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/pkg/apperror"
)

func (s *Store) ListPayments(ctx context.Context, orgID uuid.UUID) ([]domain.Payment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, tenant_id, amount, date, for_month, mode, created_at
		FROM payments WHERE organization_id = $1 ORDER BY date DESC, created_at DESC
	`, orgID)
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

func (s *Store) DeletePayment(ctx context.Context, orgID, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM payments WHERE id = $1 AND organization_id = $2`, id, orgID)
	if err != nil {
		return mapPgError(err, "")
	}
	if tag.RowsAffected() == 0 {
		return apperror.NotFound("payment not found")
	}
	return nil
}

func (s *Store) ListPaymentsByTenant(ctx context.Context, orgID, tenantID uuid.UUID) ([]domain.Payment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, tenant_id, amount, date, for_month, mode, created_at
		FROM payments WHERE organization_id = $1 AND tenant_id = $2 ORDER BY date DESC
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
