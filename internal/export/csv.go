package export

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"

	"github.com/nivas/server/internal/repository"
)

func PaymentsCSV(rows []repository.PaymentExportRow) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Date", "Tenant", "Room", "Property", "Month", "Amount", "Mode"})
	for _, r := range rows {
		_ = w.Write([]string{
			r.Date, r.TenantName, r.RoomNumber, r.PropertyName, r.ForMonth,
			strconv.FormatFloat(r.Amount, 'f', 2, 64), r.Mode,
		})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func TenantsCSV(rows []repository.TenantExportRow) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Name", "Email", "Phone", "Property", "Room", "Monthly Fee", "Join Date", "Active"})
	for _, r := range rows {
		phone := ""
		if r.Phone != nil {
			phone = *r.Phone
		}
		room := ""
		if r.RoomNumber != nil {
			room = *r.RoomNumber
		}
		_ = w.Write([]string{
			r.Name, r.Email, phone, r.PropertyName, room,
			strconv.FormatFloat(r.MonthlyFee, 'f', 2, 64), r.JoinDate,
			strconv.FormatBool(r.Active),
		})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func ExpensesCSV(rows []repository.ExpenseExportRow) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Date", "Category", "Amount", "Note"})
	for _, r := range rows {
		note := ""
		if r.Note != nil {
			note = *r.Note
		}
		_ = w.Write([]string{
			r.Date, r.Category,
			strconv.FormatFloat(r.Amount, 'f', 2, 64), note,
		})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func Filename(base, format string) string {
	if format == "xlsx" {
		return fmt.Sprintf("%s.xlsx", base)
	}
	return fmt.Sprintf("%s.csv", base)
}
