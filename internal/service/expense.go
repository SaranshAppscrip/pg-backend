package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/repository"
	"github.com/nivas/server/pkg/apperror"
)

type ExpenseService struct {
	repos repository.ExpenseRepository
}

func NewExpenseService(repos repository.ExpenseRepository) *ExpenseService {
	return &ExpenseService{repos: repos}
}

func (s *ExpenseService) List(ctx context.Context, orgID uuid.UUID) ([]domain.Expense, error) {
	return s.repos.List(ctx, orgID)
}

type CreateExpenseInput struct {
	Category domain.ExpenseCategory
	Amount   float64
	Date     time.Time
	Note     *string
}

func (s *ExpenseService) Create(ctx context.Context, orgID uuid.UUID, in CreateExpenseInput) (*domain.Expense, error) {
	if in.Amount <= 0 {
		return nil, apperror.BadRequest("amount must be positive")
	}
	expense := &domain.Expense{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Category:       in.Category,
		Amount:         in.Amount,
		Date:           in.Date,
		Note:           in.Note,
		CreatedAt:      time.Now(),
	}
	if err := s.repos.Create(ctx, expense); err != nil {
		return nil, err
	}
	return expense, nil
}

func (s *ExpenseService) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	return s.repos.Delete(ctx, orgID, id)
}
