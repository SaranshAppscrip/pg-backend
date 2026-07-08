package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/export"
	"github.com/nivas/server/internal/repository"
)

type ExportService struct {
	repos repository.ExportRepository
}

func NewExportService(repos repository.ExportRepository) *ExportService {
	return &ExportService{repos: repos}
}

type ExportResult struct {
	Filename    string
	ContentType string
	Data        []byte
}

func (s *ExportService) Payments(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID, format string) (*ExportResult, error) {
	rows, err := s.repos.ListPaymentsForExport(ctx, orgID, propertyID)
	if err != nil {
		return nil, err
	}
	return buildExport("payments", format, func() ([]byte, error) { return export.PaymentsCSV(rows) }, func() ([]byte, error) { return export.PaymentsXLSX(rows) })
}

func (s *ExportService) Tenants(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID, format string) (*ExportResult, error) {
	rows, err := s.repos.ListTenantsForExport(ctx, orgID, propertyID)
	if err != nil {
		return nil, err
	}
	return buildExport("tenants", format, func() ([]byte, error) { return export.TenantsCSV(rows) }, func() ([]byte, error) { return export.TenantsXLSX(rows) })
}

func (s *ExportService) Expenses(ctx context.Context, orgID uuid.UUID, format string) (*ExportResult, error) {
	rows, err := s.repos.ListExpensesForExport(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return buildExport("expenses", format, func() ([]byte, error) { return export.ExpensesCSV(rows) }, func() ([]byte, error) { return export.ExpensesXLSX(rows) })
}

func buildExport(base, format string, csvFn, xlsxFn func() ([]byte, error)) (*ExportResult, error) {
	if format == "xlsx" {
		data, err := xlsxFn()
		if err != nil {
			return nil, err
		}
		return &ExportResult{
			Filename:    export.Filename(base, "xlsx"),
			ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			Data:        data,
		}, nil
	}
	data, err := csvFn()
	if err != nil {
		return nil, err
	}
	return &ExportResult{
		Filename:    export.Filename(base, "csv"),
		ContentType: "text/csv",
		Data:        data,
	}, nil
}
