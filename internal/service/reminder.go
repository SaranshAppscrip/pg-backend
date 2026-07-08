package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/config"
	"github.com/nivas/server/internal/notification"
	"github.com/nivas/server/internal/rent"
	"github.com/nivas/server/internal/repository"
	"github.com/nivas/server/pkg/logger"
)

type ReminderService struct {
	repos repository.ReminderRepository
	email notification.EmailSender
	cfg   config.ReminderConfig
}

func NewReminderService(repos repository.ReminderRepository, email notification.EmailSender, cfg config.ReminderConfig) *ReminderService {
	return &ReminderService{repos: repos, email: email, cfg: cfg}
}

func (s *ReminderService) Run(ctx context.Context) error {
	log := logger.FromContext(ctx)
	now := time.Now()
	day := now.Day()
	forMonth := rent.CurrentMonth()

	var reminderType string
	switch day {
	case s.cfg.DueDay:
		reminderType = "due"
	case s.cfg.OverdueDay:
		reminderType = "overdue"
	default:
		log.Info("rent reminders skipped", "day", day)
		return nil
	}

	tenants, err := s.repos.ListActiveTenantsWithDues(ctx, nil)
	if err != nil {
		return err
	}

	sent, err := s.sendReminders(ctx, tenants, forMonth, reminderType, false)
	if err != nil {
		return err
	}

	log.Info("rent reminders completed", "sent", sent, "type", reminderType, "for_month", forMonth)
	return nil
}

// RunForOrg sends rent reminders for one organization (owner-triggered, for testing).
func (s *ReminderService) RunForOrg(ctx context.Context, orgID uuid.UUID, reminderType string, force bool) (int, error) {
	if reminderType != "due" && reminderType != "overdue" {
		return 0, fmt.Errorf("invalid reminder type: %s", reminderType)
	}
	tenants, err := s.repos.ListActiveTenantsWithDues(ctx, &orgID)
	if err != nil {
		return 0, err
	}
	return s.sendReminders(ctx, tenants, rent.CurrentMonth(), reminderType, force)
}

func (s *ReminderService) sendReminders(ctx context.Context, tenants []repository.ReminderTenantRow, forMonth, reminderType string, force bool) (int, error) {
	log := logger.FromContext(ctx)
	sent := 0
	for _, t := range tenants {
		payments, err := s.repos.ListPaymentsForTenantMonth(ctx, t.TenantID, forMonth)
		if err != nil {
			return sent, err
		}
		status := rent.TenantStatus(t.MonthlyFee, payments, forMonth)
		if status.State == "paid" {
			continue
		}

		if !force {
			already, err := s.repos.HasRentReminder(ctx, t.TenantID, forMonth, reminderType)
			if err != nil {
				return sent, err
			}
			if already {
				continue
			}
		}

		if err := s.email.SendRentReminder(ctx, notification.RentReminderParams{
			To:           t.Email,
			TenantName:   t.Name,
			PropertyName: t.PropertyName,
			ForMonth:     forMonth,
			MonthlyFee:   t.MonthlyFee,
			Paid:         status.Paid,
			Due:          status.Due,
			ReminderType: reminderType,
		}); err != nil {
			log.Warn("rent reminder email failed", "tenant_id", t.TenantID, "error", err)
			continue
		}
		if err := s.repos.CreateRentReminder(ctx, t.TenantID, forMonth, reminderType); err != nil {
			return sent, err
		}
		sent++
	}
	return sent, nil
}
