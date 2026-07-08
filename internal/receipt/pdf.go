package receipt

import (
	"bytes"
	"fmt"

	"github.com/jung-kurt/gofpdf"
	"github.com/nivas/server/internal/repository"
)

func GeneratePDF(d *repository.PaymentReceiptData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 12, "Payment Receipt")
	pdf.Ln(14)

	pdf.SetFont("Arial", "", 11)
	lines := []string{
		fmt.Sprintf("Organization: %s", d.OrganizationName),
		fmt.Sprintf("Property: %s", d.PropertyName),
		fmt.Sprintf("Tenant: %s", d.TenantName),
		fmt.Sprintf("Room: %s", d.RoomNumber),
		fmt.Sprintf("Receipt ID: %s", d.PaymentID.String()),
		fmt.Sprintf("Date: %s", d.Date),
		fmt.Sprintf("For month: %s", d.ForMonth),
		fmt.Sprintf("Amount: Rs. %.2f", d.Amount),
		fmt.Sprintf("Mode: %s", d.Mode),
	}
	for _, line := range lines {
		pdf.Cell(0, 8, line)
		pdf.Ln(8)
	}

	pdf.Ln(8)
	pdf.SetFont("Arial", "I", 10)
	pdf.Cell(0, 8, "Thank you for your payment.")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
