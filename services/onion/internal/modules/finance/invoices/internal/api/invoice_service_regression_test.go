package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	financeinvoicespaymentsv1 "github.com/voxire/voxie/gen/go/finance/invoices/payments/v1"
	financeinvoicesv1 "github.com/voxire/voxie/gen/go/finance/invoices/v1"
)

type invoiceRegressionSnapshots struct {
	businessID int64
}

func (p invoiceRegressionSnapshots) CustomerSnapshot(_ context.Context, _, _ int64) (string, string, string, error) {
	return "Regression Client", fmt.Sprintf("billing-%d@example.test", p.businessID), "1 Regression Way", nil
}

func TestFinanceDashboardCountsPastDueSentInvoice(t *testing.T) {
	api, ctx, db := newInvoiceRegressionAPI(t, 995915)
	const businessID int64 = 995915

	_, err := db.ExecContext(ctx, `
		INSERT INTO finance_invoices
			(owner_id, business_id, invoice_number, client_name, status,
			 issue_date, due_date, total, amount_due)
		VALUES (?, ?, 'DASHBOARD-OVERDUE-995915', 'Dashboard Client', 'sent',
			'2020-01-01', '2020-01-02', 75.00, 75.00)`, businessID, businessID)
	if err != nil {
		t.Fatalf("insert past-due invoice: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM finance_invoices WHERE business_id = ?", businessID)
	})

	dashboard, err := api.Service.GetDashboard(ctx, businessID)
	if err != nil {
		t.Fatalf("GetDashboard: %v", err)
	}
	if dashboard.OverdueAmount != 75.00 {
		t.Fatalf("overdue_amount = %v, want 75.00", dashboard.OverdueAmount)
	}
}

func TestCreateInvoiceReturnsPersistedTimestamps(t *testing.T) {
	api, ctx, db := newInvoiceRegressionAPI(t, 995911)

	// Create through the live API so this assertion exercises
	// the response path rather than the fixture helper's domain object.
	resp, err := api.CreateInvoice(ctx, invoiceRegressionRequest())
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	createdAt := resp.GetInvoice().GetCreatedAt()
	updatedAt := resp.GetInvoice().GetUpdatedAt()
	if _, err := time.Parse("2006-01-02T15:04:05Z", createdAt); err != nil {
		t.Fatalf("created_at = %q: %v", createdAt, err)
	}
	if _, err := time.Parse("2006-01-02T15:04:05Z", updatedAt); err != nil {
		t.Fatalf("updated_at = %q: %v", updatedAt, err)
	}
	cleanupInvoiceRegressionID(t, db, resp.GetInvoice().GetId())
}

func TestDeleteInvoiceRejectsMissingAndNonDraft(t *testing.T) {
	api, ctx, db := newInvoiceRegressionAPI(t, 995912)

	_, err := api.DeleteInvoice(ctx, &financeinvoicesv1.DeleteInvoiceRequest{Id: 999999999})
	if got := financePermissionStatus(err); got != http.StatusNotFound {
		t.Fatalf("missing invoice status = %d, want %d (err = %v)", got, http.StatusNotFound, err)
	}

	invoiceID := createInvoiceRegressionFixture(t, api, db, ctx)
	if _, err := db.ExecContext(ctx, "UPDATE finance_invoices SET status = 'sent' WHERE id = ?", invoiceID); err != nil {
		t.Fatalf("mark invoice sent: %v", err)
	}
	_, err = api.DeleteInvoice(ctx, &financeinvoicesv1.DeleteInvoiceRequest{Id: invoiceID})
	if got := financePermissionStatus(err); got != http.StatusConflict {
		t.Fatalf("non-draft invoice status = %d, want %d (err = %v)", got, http.StatusConflict, err)
	}

	deletableID := createInvoiceRegressionFixture(t, api, db, ctx)
	if _, err := api.DeleteInvoice(ctx, &financeinvoicesv1.DeleteInvoiceRequest{Id: deletableID}); err != nil {
		t.Fatalf("DeleteInvoice(draft): %v", err)
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM finance_invoices WHERE id = ?", deletableID).Scan(&count); err != nil {
		t.Fatalf("count deleted invoice: %v", err)
	}
	if count != 0 {
		t.Fatalf("deleted invoice still exists")
	}
}

func TestSendInvoiceMapsMissingAndRejectsTerminal(t *testing.T) {
	api, ctx, db := newInvoiceRegressionAPI(t, 995913)

	_, err := api.SendInvoice(ctx, &financeinvoicesv1.SendInvoiceRequest{Id: 999999999})
	if got := financePermissionStatus(err); got != http.StatusNotFound {
		t.Fatalf("missing invoice status = %d, want %d (err = %v)", got, http.StatusNotFound, err)
	}

	invoiceID := createInvoiceRegressionFixture(t, api, db, ctx)
	if _, err := db.ExecContext(ctx, "UPDATE finance_invoices SET status = 'paid' WHERE id = ?", invoiceID); err != nil {
		t.Fatalf("mark invoice paid: %v", err)
	}
	_, err = api.SendInvoice(ctx, &financeinvoicesv1.SendInvoiceRequest{Id: invoiceID})
	if got := financePermissionStatus(err); got != http.StatusConflict {
		t.Fatalf("terminal invoice status = %d, want %d (err = %v)", got, http.StatusConflict, err)
	}
}

func TestRecordPaymentNormalizesBankTransferLabel(t *testing.T) {
	api, ctx, db := newInvoiceRegressionAPI(t, 995914)
	invoiceID := createInvoiceRegressionFixture(t, api, db, ctx)

	resp, err := NewInvoicePaymentAPI(api.Service).RecordPayment(ctx, &financeinvoicespaymentsv1.RecordPaymentRequest{
		InvoiceId:   invoiceID,
		Amount:      25,
		Currency:    "USD",
		PaymentDate: "2026-06-20",
		Method:      "Bank Transfer",
		Reference:   "regression-bank-transfer",
	})
	if err != nil {
		t.Fatalf("RecordPayment(Bank Transfer): %v", err)
	}
	if got := resp.GetPayment().GetMethod(); got != "bank_transfer" {
		t.Fatalf("payment method = %q, want bank_transfer", got)
	}

	var method string
	if err := db.QueryRowContext(ctx, "SELECT method FROM finance_payments WHERE id = ?", resp.GetPayment().GetId()).Scan(&method); err != nil {
		t.Fatalf("read stored payment: %v", err)
	}
	if method != "bank_transfer" {
		t.Fatalf("stored payment method = %q, want bank_transfer", method)
	}
}

func TestNormalizeInvoicePaymentMethod(t *testing.T) {
	cases := map[string]string{
		"":              "cash",
		"Bank Transfer": "bank_transfer",
		"Credit Card":   "card",
		"Check":         "cheque",
		"ACH":           "bank_transfer",
		"PayPal":        "online",
	}
	for input, want := range cases {
		got, ok := normalizeInvoicePaymentMethod(input)
		if !ok || got != want {
			t.Errorf("normalize(%q) = (%q, %v), want (%q, true)", input, got, ok, want)
		}
	}
	if _, ok := normalizeInvoicePaymentMethod("crypto"); ok {
		t.Error("unsupported payment method was accepted")
	}
}

// An authenticated finance caller must not be able to turn an unpaid invoice
// into revenue with a status write: "paid" adds total to revenue and removes
// the invoice from receivables, so the books would net to zero with both sides
// wrong. Only a recorded payment may set it.
func TestUpdateInvoiceStatusRejectsPaidAndKeepsRevenueAndReceivables(t *testing.T) {
	api, ctx, db := newInvoiceRegressionAPI(t, 995916)
	const businessID int64 = 995916

	invoiceID := createInvoiceRegressionFixture(t, api, db, ctx)
	if _, err := db.ExecContext(ctx, `
		UPDATE finance_invoices
		SET status = 'sent', total = 1085.00, amount_paid = 0, amount_due = 1085.00
		WHERE id = ?`, invoiceID); err != nil {
		t.Fatalf("stage sent invoice: %v", err)
	}

	receivableBefore, revenueBefore := invoiceStatusAggregates(t, db, ctx, businessID)

	for _, status := range []string{"paid", "partial"} {
		_, err := api.UpdateInvoiceStatus(ctx, &financeinvoicesv1.UpdateInvoiceStatusRequest{
			Id:   invoiceID,
			Body: &financeinvoicesv1.UpdateInvoiceStatusBody{Status: status},
		})
		if got := financePermissionStatus(err); got != http.StatusBadRequest {
			t.Fatalf("status %q = %d, want %d (err = %v)", status, got, http.StatusBadRequest, err)
		}
	}

	_, err := api.UpdateInvoiceStatus(ctx, &financeinvoicesv1.UpdateInvoiceStatusRequest{
		Id:   invoiceID,
		Body: &financeinvoicesv1.UpdateInvoiceStatusBody{Status: "settled"},
	})
	if got := financePermissionStatus(err); got != http.StatusBadRequest {
		t.Fatalf("unknown status = %d, want %d (err = %v)", got, http.StatusBadRequest, err)
	}

	var (
		storedStatus string
		amountPaid   float64
		paidDate     sql.NullTime
	)
	if err := db.QueryRowContext(ctx,
		"SELECT status, amount_paid, paid_date FROM finance_invoices WHERE id = ?", invoiceID,
	).Scan(&storedStatus, &amountPaid, &paidDate); err != nil {
		t.Fatalf("read invoice after rejected writes: %v", err)
	}
	if storedStatus != "sent" {
		t.Errorf("stored status = %q after rejected writes, want sent", storedStatus)
	}
	if amountPaid != 0 {
		t.Errorf("amount_paid = %v, want 0", amountPaid)
	}
	if paidDate.Valid {
		t.Errorf("paid_date = %v, want NULL", paidDate.Time)
	}

	receivableAfter, revenueAfter := invoiceStatusAggregates(t, db, ctx, businessID)
	if receivableAfter != receivableBefore || revenueAfter != revenueBefore {
		t.Errorf("receivables/revenue = %v/%v after rejected writes, want %v/%v",
			receivableAfter, revenueAfter, receivableBefore, revenueBefore)
	}
}

// invoiceStatusAggregates reads the two figures a status write corrupts: the
// receivable it would leave and the revenue it would invent. Same predicates as
// SumTotalReceivable and SumPaidInvoices.
func invoiceStatusAggregates(t *testing.T, db *sql.DB, ctx context.Context, businessID int64) (receivable, revenue float64) {
	t.Helper()
	if err := db.QueryRowContext(ctx, `
		SELECT
			COALESCE((SELECT SUM(amount_due) FROM finance_invoices
				WHERE business_id = ? AND status IN ('sent','viewed','partial','overdue')), 0),
			COALESCE((SELECT SUM(total) FROM finance_invoices
				WHERE business_id = ? AND status = 'paid'), 0)`, businessID, businessID,
	).Scan(&receivable, &revenue); err != nil {
		t.Fatalf("read invoice aggregates: %v", err)
	}
	return receivable, revenue
}

// "overdue" is a delivery/ageing label an invoice can reach without money, so
// the status endpoint must still accept it.
func TestUpdateInvoiceStatusAllowsOverdueFromSent(t *testing.T) {
	api, ctx, db := newInvoiceRegressionAPI(t, 995917)

	invoiceID := createInvoiceRegressionFixture(t, api, db, ctx)
	if _, err := db.ExecContext(ctx, "UPDATE finance_invoices SET status = 'sent' WHERE id = ?", invoiceID); err != nil {
		t.Fatalf("stage sent invoice: %v", err)
	}

	if _, err := api.UpdateInvoiceStatus(ctx, &financeinvoicesv1.UpdateInvoiceStatusRequest{
		Id:   invoiceID,
		Body: &financeinvoicesv1.UpdateInvoiceStatusBody{Status: "overdue"},
	}); err != nil {
		t.Fatalf("sent -> overdue: %v", err)
	}

	var status string
	if err := db.QueryRowContext(ctx, "SELECT status FROM finance_invoices WHERE id = ?", invoiceID).Scan(&status); err != nil {
		t.Fatalf("read invoice status: %v", err)
	}
	if status != "overdue" {
		t.Errorf("stored status = %q, want overdue", status)
	}
}

// Recording the money is what flips the status, and it stays correct after the
// status endpoint starts refusing "paid" and "partial".
func TestRecordPaymentStillSetsPaidAndPartial(t *testing.T) {
	api, ctx, db := newInvoiceRegressionAPI(t, 995918)
	payments := NewInvoicePaymentAPI(api.Service)

	invoiceID := createInvoiceRegressionFixture(t, api, db, ctx)
	if _, err := db.ExecContext(ctx, "UPDATE finance_invoices SET total = 100, amount_due = 100 WHERE id = ?", invoiceID); err != nil {
		t.Fatalf("stage invoice total: %v", err)
	}

	if _, err := payments.RecordPayment(ctx, &financeinvoicespaymentsv1.RecordPaymentRequest{
		InvoiceId:   invoiceID,
		Amount:      25,
		Currency:    "USD",
		PaymentDate: "2026-06-20",
		Method:      "card",
		Reference:   "partial-payment-995918",
	}); err != nil {
		t.Fatalf("RecordPayment(partial): %v", err)
	}
	assertInvoiceMoney(t, db, invoiceID, "partial", 25, 75)

	if _, err := payments.RecordPayment(ctx, &financeinvoicespaymentsv1.RecordPaymentRequest{
		InvoiceId:   invoiceID,
		Amount:      75,
		Currency:    "USD",
		PaymentDate: "2026-06-21",
		Method:      "card",
		Reference:   "final-payment-995918",
	}); err != nil {
		t.Fatalf("RecordPayment(final): %v", err)
	}
	assertInvoiceMoney(t, db, invoiceID, "paid", 100, 0)
}

func assertInvoiceMoney(t *testing.T, db *sql.DB, invoiceID int64, wantStatus string, wantPaid, wantDue float64) {
	t.Helper()
	var (
		status     string
		amountPaid float64
		amountDue  float64
	)
	if err := db.QueryRowContext(context.Background(),
		"SELECT status, amount_paid, amount_due FROM finance_invoices WHERE id = ?", invoiceID,
	).Scan(&status, &amountPaid, &amountDue); err != nil {
		t.Fatalf("read invoice money: %v", err)
	}
	if status != wantStatus {
		t.Errorf("status = %q, want %q", status, wantStatus)
	}
	if amountPaid != wantPaid || amountDue != wantDue {
		t.Errorf("amount_paid/amount_due = %v/%v, want %v/%v", amountPaid, amountDue, wantPaid, wantDue)
	}
}

func newInvoiceRegressionAPI(t *testing.T, businessID int64) (*InvoicesAPI, context.Context, *sql.DB) {
	t.Helper()
	db := testutil.OpenDB(t)
	const userID int64 = 7
	testutil.GrantBusinessAccess(t, db, businessID, userID)
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), "DELETE FROM crm_customers WHERE business_id = ?", businessID); err != nil {
			t.Errorf("cleanup CRM customers: %v", err)
		}
	})

	user := &authPkg.AuthUser{
		BusinessID: businessID,
		UserID:     userID,
		EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
			AppID:       authorization.AppFinance,
			Licensed:    true,
			HasAccess:   true,
			Permissions: []string{"finance:invoices:create", "finance:invoices:view", "finance:invoices:send"},
		}}},
	}
	ctx := authPkg.WithAuthUser(context.Background(), user)
	svc := newInvoicesServiceForTest(db)
	svc.SetCustomerSnapshotProvider(invoiceRegressionSnapshots{businessID: businessID})
	return NewInvoicesAPI(svc), ctx, db
}

func invoiceRegressionRequest() *financeinvoicesv1.CreateInvoiceRequest {
	return &financeinvoicesv1.CreateInvoiceRequest{
		CustomerId: testutil.Int64Ptr(5301),
		ClientName: "Regression Client",
		IssueDate:  "2026-06-15",
		DueDate:    "2026-06-30",
		Currency:   "USD",
		Items: []*financeinvoicesv1.InvoiceItem{{
			Description: "Regression item",
			Quantity:    1,
			Unit:        "unit",
			UnitPrice:   100,
		}},
	}
}

func createInvoiceRegressionFixture(t *testing.T, api *InvoicesAPI, db *sql.DB, ctx context.Context) int64 {
	t.Helper()
	resp, err := api.CreateInvoice(ctx, invoiceRegressionRequest())
	if err != nil {
		t.Fatalf("create invoice fixture: %v", err)
	}
	cleanupInvoiceRegressionID(t, db, resp.GetInvoice().GetId())
	return resp.GetInvoice().GetId()
}

func cleanupInvoiceRegressionID(t *testing.T, db *sql.DB, invoiceID int64) {
	t.Helper()
	t.Cleanup(func() {
		ctx := context.Background()
		for _, query := range []string{
			"DELETE FROM finance_payments WHERE invoice_id = ?",
			"DELETE FROM finance_invoice_items WHERE invoice_id = ?",
			"DELETE FROM finance_invoices WHERE id = ?",
		} {
			if _, err := db.ExecContext(ctx, query, invoiceID); err != nil {
				t.Errorf("cleanup %q: %v", query, err)
			}
		}
	})
}
