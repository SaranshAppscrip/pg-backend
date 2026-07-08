package export

import (
	"bytes"
	"strconv"

	"github.com/nivas/server/internal/repository"
	"github.com/xuri/excelize/v2"
)

func writeSheet(f *excelize.File, sheet string, headers []string, rows [][]string) error {
	if _, err := f.NewSheet(sheet); err != nil {
		return err
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	for ri, row := range rows {
		for ci, val := range row {
			cell, _ := excelize.CoordinatesToCellName(ci+1, ri+2)
			_ = f.SetCellValue(sheet, cell, val)
		}
	}
	return nil
}

func writeWorkbook(build func(f *excelize.File, sheet string) error) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Sheet1"
	if err := build(f, sheet); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func PaymentsXLSX(rows []repository.PaymentExportRow) ([]byte, error) {
	return writeWorkbook(func(f *excelize.File, sheet string) error {
		data := make([][]string, len(rows))
		for i, r := range rows {
			data[i] = []string{r.Date, r.TenantName, r.RoomNumber, r.PropertyName, r.ForMonth, strconv.FormatFloat(r.Amount, 'f', 2, 64), r.Mode}
		}
		return writeSheet(f, sheet, []string{"Date", "Tenant", "Room", "Property", "Month", "Amount", "Mode"}, data)
	})
}

func TenantsXLSX(rows []repository.TenantExportRow) ([]byte, error) {
	return writeWorkbook(func(f *excelize.File, sheet string) error {
		data := make([][]string, len(rows))
		for i, r := range rows {
			phone, room := "", ""
			if r.Phone != nil {
				phone = *r.Phone
			}
			if r.RoomNumber != nil {
				room = *r.RoomNumber
			}
			data[i] = []string{r.Name, r.Email, phone, r.PropertyName, room, strconv.FormatFloat(r.MonthlyFee, 'f', 2, 64), r.JoinDate, strconv.FormatBool(r.Active)}
		}
		return writeSheet(f, sheet, []string{"Name", "Email", "Phone", "Property", "Room", "Monthly Fee", "Join Date", "Active"}, data)
	})
}

func ExpensesXLSX(rows []repository.ExpenseExportRow) ([]byte, error) {
	return writeWorkbook(func(f *excelize.File, sheet string) error {
		data := make([][]string, len(rows))
		for i, r := range rows {
			note := ""
			if r.Note != nil {
				note = *r.Note
			}
			data[i] = []string{r.Date, r.Category, strconv.FormatFloat(r.Amount, 'f', 2, 64), note}
		}
		return writeSheet(f, sheet, []string{"Date", "Category", "Amount", "Note"}, data)
	})
}
