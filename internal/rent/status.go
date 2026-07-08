package rent

import (
	"time"

	"github.com/nivas/server/internal/domain"
)

// Status for a tenant's rent in a given month.
type Status struct {
	Paid  float64
	Due   float64
	State string // paid, partial, unpaid
}

func TenantStatus(monthlyFee float64, payments []domain.Payment, forMonth string) Status {
	var paid float64
	for _, p := range payments {
		if p.ForMonth == forMonth {
			paid += p.Amount
		}
	}
	due := monthlyFee - paid
	if due < 0 {
		due = 0
	}
	state := "unpaid"
	if paid >= monthlyFee {
		state = "paid"
	} else if paid > 0 {
		state = "partial"
	}
	return Status{Paid: paid, Due: due, State: state}
}

func CurrentMonth() string {
	return time.Now().Format("2006-01")
}
