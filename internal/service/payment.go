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
	audit *AuditService
}

func NewPaymentService(repos repository.PaymentRepository, audit *AuditService) *PaymentService {
	return &PaymentService{repos: repos, audit: audit}
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

func (s *PaymentService) Create(ctx context.Context, orgID, staffID uuid.UUID, in CreatePaymentInput) (*domain.Payment, error) {
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
	_ = s.audit.Log(ctx, orgID, staffID, domain.AuditEntityPayment, payment.ID, domain.AuditActionCreate, map[string]any{
		"amount":    payment.Amount,
		"for_month": payment.ForMonth,
		"tenant_id": payment.TenantID.String(),
		"mode":      string(payment.Mode),
		"date":      payment.Date.Format("2006-01-02"),
	})
	return payment, nil
}

func (s *PaymentService) Delete(ctx context.Context, orgID, staffID, id uuid.UUID) error {
	payment, err := s.repos.SoftDelete(ctx, orgID, id)
	if err != nil {
		return err
	}
	_ = s.audit.Log(ctx, orgID, staffID, domain.AuditEntityPayment, payment.ID, domain.AuditActionDelete, map[string]any{
		"amount":    payment.Amount,
		"for_month": payment.ForMonth,
		"tenant_id": payment.TenantID.String(),
		"mode":      string(payment.Mode),
		"date":      payment.Date.Format("2006-01-02"),
	})
	return nil
}

func (s *PaymentService) ListByTenant(ctx context.Context, orgID, tenantID uuid.UUID) ([]domain.Payment, error) {
	return s.repos.ListByTenant(ctx, orgID, tenantID)
}
