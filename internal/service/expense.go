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
	audit *AuditService
}

func NewExpenseService(repos repository.ExpenseRepository, audit *AuditService) *ExpenseService {
	return &ExpenseService{repos: repos, audit: audit}
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

func (s *ExpenseService) Create(ctx context.Context, orgID, staffID uuid.UUID, in CreateExpenseInput) (*domain.Expense, error) {
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
	meta := map[string]any{
		"amount":   expense.Amount,
		"category": string(expense.Category),
		"date":     expense.Date.Format("2006-01-02"),
	}
	if expense.Note != nil {
		meta["note"] = *expense.Note
	}
	_ = s.audit.Log(ctx, orgID, staffID, domain.AuditEntityExpense, expense.ID, domain.AuditActionCreate, meta)
	return expense, nil
}

func (s *ExpenseService) Delete(ctx context.Context, orgID, staffID, id uuid.UUID) error {
	expense, err := s.repos.SoftDelete(ctx, orgID, id)
	if err != nil {
		return err
	}
	meta := map[string]any{
		"amount":   expense.Amount,
		"category": string(expense.Category),
		"date":     expense.Date.Format("2006-01-02"),
	}
	if expense.Note != nil {
		meta["note"] = *expense.Note
	}
	_ = s.audit.Log(ctx, orgID, staffID, domain.AuditEntityExpense, expense.ID, domain.AuditActionDelete, meta)
	return nil
}
