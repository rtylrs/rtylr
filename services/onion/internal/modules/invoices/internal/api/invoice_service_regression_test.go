package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
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
	svc := application.NewService(ctx, db)
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
