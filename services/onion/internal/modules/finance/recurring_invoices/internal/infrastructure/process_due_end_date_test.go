package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/finance/recurring_invoices/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/recurring_invoices/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestProcessDueIncludesEndDateAfterMidnight(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const invoiceNumber = "recurring-end-date-regression"
	const invoiceDate = "2026-10-31"

	result, err := db.ExecContext(ctx, `
		INSERT INTO finance_invoices
			(owner_id, business_id, invoice_number, client_name, status, issue_date, due_date, total, amount_due)
		VALUES (?, ?, ?, 'Recurring test', 'draft', ?, ?, 25.00, 25.00)`,
		testutil.OwnerID, testutil.OwnerID, invoiceNumber, invoiceDate, invoiceDate)
	if err != nil {
		t.Fatalf("insert template invoice: %v", err)
	}
	templateID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("template invoice id: %v", err)
	}
	var scheduleID int64
	t.Cleanup(func() {
		if scheduleID != 0 {
			generatedNumber := fmt.Sprintf("REC-%d-20261031", scheduleID)
			_, _ = db.ExecContext(context.Background(), "DELETE FROM finance_invoices WHERE business_id = ? AND invoice_number = ?", testutil.OwnerID, generatedNumber)
			_, _ = db.ExecContext(context.Background(), "DELETE FROM finance_recurring_invoices WHERE id = ? AND business_id = ?", scheduleID, testutil.OwnerID)
		}
		_, _ = db.ExecContext(context.Background(), "DELETE FROM finance_invoices WHERE id = ? AND business_id = ?", templateID, testutil.OwnerID)
	})

	scheduleID, err = repo.Create(ctx, testutil.OwnerID, domain.RecurringInvoiceInput{
		TemplateInvoiceID: &templateID,
		Frequency:         "monthly",
		NextInvoiceDate:   invoiceDate,
		EndDate:           stringPointer(invoiceDate),
	})
	if err != nil {
		t.Fatalf("create bounded recurring schedule: %v", err)
	}

	noonOnEndDate := time.Date(2026, time.October, 31, 12, 0, 0, 0, time.UTC)
	if _, err := repo.ProcessDue(ctx, noonOnEndDate); err != nil {
		t.Fatalf("process due schedules: %v", err)
	}

	generatedNumber := fmt.Sprintf("REC-%d-20261031", scheduleID)
	var count int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM finance_invoices WHERE business_id = ? AND invoice_number = ?",
		testutil.OwnerID, generatedNumber,
	).Scan(&count); err != nil {
		t.Fatalf("count generated invoice: %v", err)
	}
	if count != 1 {
		t.Fatalf("generated invoices for inclusive end date = %d, want 1", count)
	}
}

func stringPointer(value string) *string { return &value }
