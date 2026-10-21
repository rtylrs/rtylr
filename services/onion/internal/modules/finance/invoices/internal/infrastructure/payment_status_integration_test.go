package infrastructure_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// TestRecordPayment_TwoPaymentsMarkInvoicePaid reproduces issue #757: a
// $36.66 invoice paid with two $18.33 payments must flip to "paid" with a
// zero amount due, not stay "partial" on float64 residue.
func TestRecordPayment_RejectsAmountOverRemainingBalance(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	result, err := db.ExecContext(ctx, `
		INSERT INTO finance_invoices
			(owner_id, business_id, invoice_number, client_name, status, issue_date, due_date,
			 currency, total, amount_paid, amount_due)
		VALUES (?, ?, 'TEST-PAYMENT-OVERPAY', 'Acme Corp', 'sent', ?, ?, 'USD', 100, 0, 100)`,
		testutil.OwnerID, testutil.OwnerID, time.Now(), time.Now().Add(30*24*time.Hour))
	if err != nil {
		t.Fatalf("insert invoice: %v", err)
	}
	invoiceID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("invoice id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM finance_payments WHERE invoice_id = ?", invoiceID)
		_, _ = db.ExecContext(ctx, "DELETE FROM finance_invoices WHERE id = ?", invoiceID)
	})

	_, err = repo.RecordPayment(ctx, testutil.OwnerID, domain.PaymentInput{
		InvoiceID:   invoiceID,
		Amount:      100.01,
		PaymentDate: time.Now().Format("2006-01-02"),
		Method:      "card",
	})
	if !errors.Is(err, domain.ErrPaymentExceedsBalance) {
		t.Fatalf("overpayment error = %v, want ErrPaymentExceedsBalance", err)
	}

	var payments int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM finance_payments WHERE invoice_id = ?", invoiceID).Scan(&payments); err != nil {
		t.Fatalf("count payments: %v", err)
	}
	if payments != 0 {
		t.Fatalf("rejected overpayment persisted %d payment rows", payments)
	}
}

func TestRecordPayment_TwoPaymentsMarkInvoicePaid(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	invoiceID, err := repo.CreateInvoice(ctx, testutil.OwnerID, domain.Invoice{
		InvoiceNumber: "TEST-PAID-FLIP",
		ClientName:    "Acme Corp",
		IssueDate:     time.Now(),
		DueDate:       time.Now().Add(30 * 24 * time.Hour),
		Currency:      "USD",
		Status:        "draft",
		Total:         36.66,
		AmountDue:     36.66,
	})
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM finance_payments WHERE invoice_id = ?", invoiceID)
		db.ExecContext(ctx, "DELETE FROM finance_invoices WHERE id = ?", invoiceID)
	})

	for i := 0; i < 2; i++ {
		payment, err := repo.RecordPayment(ctx, testutil.OwnerID, domain.PaymentInput{
			InvoiceID:   invoiceID,
			Amount:      18.33,
			PaymentDate: time.Now().Format("2006-01-02"),
			Method:      "card",
		})
		if err != nil {
			t.Fatalf("RecordPayment #%d: %v", i+1, err)
		}
		if payment == nil || payment.ID <= 0 {
			t.Fatalf("RecordPayment #%d returned no payment", i+1)
		}
	}

	invoice, err := repo.GetInvoiceByID(ctx, testutil.OwnerID, invoiceID)
	if err != nil {
		t.Fatalf("GetInvoiceByID: %v", err)
	}
	if invoice.Status != "paid" {
		t.Errorf("status = %q, want paid", invoice.Status)
	}
	if invoice.AmountDue > 0.001 || invoice.AmountDue < -0.001 {
		t.Errorf("amount due = %f, want 0", invoice.AmountDue)
	}
	if invoice.AmountPaid < 36.659 || invoice.AmountPaid > 36.661 {
		t.Errorf("amount paid = %f, want 36.66", invoice.AmountPaid)
	}
}
