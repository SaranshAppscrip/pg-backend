package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/pkg/apperror"
)

func (s *Store) ListExpenses(ctx context.Context, orgID uuid.UUID) ([]domain.Expense, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, category, amount, date, note, created_at
		FROM expenses WHERE organization_id = $1 ORDER BY date DESC, created_at DESC
	`, orgID)
	if err != nil {
		return nil, mapPgError(err, "")
	}
	defer rows.Close()

	var list []domain.Expense
	for rows.Next() {
		var e domain.Expense
		if err := rows.Scan(&e.ID, &e.OrganizationID, &e.Category, &e.Amount, &e.Date, &e.Note, &e.CreatedAt); err != nil {
			return nil, apperror.Internal("scan expense", err)
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func (s *Store) CreateExpense(ctx context.Context, expense *domain.Expense) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO expenses (id, organization_id, category, amount, date, note)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, expense.ID, expense.OrganizationID, expense.Category, expense.Amount, expense.Date, expense.Note)
	return mapPgError(err, "")
}

func (s *Store) DeleteExpense(ctx context.Context, orgID, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM expenses WHERE id = $1 AND organization_id = $2`, id, orgID)
	if err != nil {
		return mapPgError(err, "")
	}
	if tag.RowsAffected() == 0 {
		return apperror.NotFound("expense not found")
	}
	return nil
}
