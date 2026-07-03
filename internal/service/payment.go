package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/repository"
	"github.com/nivas/server/pkg/apperror"
)

type PaymentService struct {
	repos repository.PaymentRepository
}

func NewPaymentService(repos repository.PaymentRepository) *PaymentService {
	return &PaymentService{repos: repos}
}

func (s *PaymentService) List(ctx context.Context, orgID uuid.UUID) ([]domain.Payment, error) {
	return s.repos.List(ctx, orgID)
}

type CreatePaymentInput struct {
	TenantID uuid.UUID
	Amount   float64
	Date     time.Time
	ForMonth string
	Mode     domain.PaymentMode
}

func (s *PaymentService) Create(ctx context.Context, orgID uuid.UUID, in CreatePaymentInput) (*domain.Payment, error) {
	if in.Amount <= 0 || in.ForMonth == "" {
		return nil, apperror.BadRequest("amount and for_month are required")
	}
	payment := &domain.Payment{
		ID:        uuid.New(),
		TenantID:  in.TenantID,
		Amount:    in.Amount,
		Date:      in.Date,
		ForMonth:  in.ForMonth,
		Mode:      in.Mode,
		CreatedAt: time.Now(),
	}
	if err := s.repos.Create(ctx, orgID, payment); err != nil {
		return nil, err
	}
	return payment, nil
}

func (s *PaymentService) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	return s.repos.Delete(ctx, orgID, id)
}

func (s *PaymentService) ListByTenant(ctx context.Context, orgID, tenantID uuid.UUID) ([]domain.Payment, error) {
	return s.repos.ListByTenant(ctx, orgID, tenantID)
}
