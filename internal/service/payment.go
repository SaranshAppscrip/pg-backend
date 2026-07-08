package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/notification"
	"github.com/nivas/server/internal/receipt"
	"github.com/nivas/server/internal/repository"
	"github.com/nivas/server/pkg/apperror"
	"github.com/nivas/server/pkg/logger"
)

type PaymentService struct {
	repos  repository.PaymentRepository
	export repository.ExportRepository
	audit  *AuditService
	email  notification.EmailSender
}

func NewPaymentService(repos repository.PaymentRepository, export repository.ExportRepository, audit *AuditService, email notification.EmailSender) *PaymentService {
	return &PaymentService{repos: repos, export: export, audit: audit, email: email}
}

func (s *PaymentService) List(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID) ([]domain.Payment, error) {
	return s.repos.List(ctx, orgID, propertyID)
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
	s.sendReceipt(ctx, orgID, payment.ID)
	return payment, nil
}

func (s *PaymentService) sendReceipt(ctx context.Context, orgID, paymentID uuid.UUID) {
	log := logger.FromContext(ctx)
	data, err := s.export.GetPaymentReceiptData(ctx, orgID, paymentID)
	if err != nil || data.TenantEmail == "" {
		return
	}
	pdf, err := receipt.GeneratePDF(data)
	if err != nil {
		log.Warn("payment receipt pdf failed", "payment_id", paymentID, "error", err)
		pdf = nil
	}
	subject, html := notification.PaymentReceiptHTML(struct {
		TenantName, PropertyName, OrganizationName, Date, ForMonth, Mode string
		Amount                                                             float64
		PaymentID                                                          string
	}{
		TenantName: data.TenantName, PropertyName: data.PropertyName, OrganizationName: data.OrganizationName,
		Date: data.Date, ForMonth: data.ForMonth, Mode: data.Mode, Amount: data.Amount, PaymentID: data.PaymentID.String(),
	})
	text := fmt.Sprintf("Payment of Rs. %.0f received for %s. Receipt ID: %s", data.Amount, data.ForMonth, data.PaymentID)
	if err := s.email.SendPaymentReceipt(ctx, notification.PaymentReceiptParams{
		To: data.TenantEmail, TenantName: data.TenantName, Subject: subject, Text: text, HTML: html,
		PDF: pdf, Filename: fmt.Sprintf("receipt-%s.pdf", data.PaymentID.String()),
	}); err != nil {
		log.Warn("payment receipt email failed", "payment_id", paymentID, "error", err)
	}
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
