package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/repository"
	"github.com/nivas/server/pkg/apperror"
)

func (s *Store) ListExpenses(ctx context.Context, orgID uuid.UUID) ([]domain.Expense, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, category, amount, date, note, created_at
		FROM expenses
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY date DESC, created_at DESC
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

func (s *Store) SoftDeleteExpense(ctx context.Context, orgID, id uuid.UUID) (*domain.Expense, error) {
	var e domain.Expense
	err := s.pool.QueryRow(ctx, `
		UPDATE expenses SET deleted_at = NOW()
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL
		RETURNING id, organization_id, category, amount, date, note, created_at
	`, id, orgID).Scan(&e.ID, &e.OrganizationID, &e.Category, &e.Amount, &e.Date, &e.Note, &e.CreatedAt)
	if err != nil {
		return nil, mapPgError(err, "expense not found")
	}
	return &e, nil
}

func (s *Store) ListExpensesForExport(ctx context.Context, orgID uuid.UUID) ([]repository.ExpenseExportRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, date::text, category::text, amount, note
		FROM expenses
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY date DESC, created_at DESC
	`, orgID)
	if err != nil {
		return nil, mapPgError(err, "")
	}
	defer rows.Close()

	var list []repository.ExpenseExportRow
	for rows.Next() {
		var row repository.ExpenseExportRow
		if err := rows.Scan(&row.ID, &row.Date, &row.Category, &row.Amount, &row.Note); err != nil {
			return nil, apperror.Internal("scan expense export", err)
		}
		list = append(list, row)
	}
	return list, rows.Err()
}
