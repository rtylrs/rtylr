package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/testutil"
	financeinvoicesv1 "github.com/voxire/voxie/gen/go/finance/invoices/v1"
)

// stubCustomerSnapshots stands in for the customers module port. The create
// path only needs a positive verification plus the snapshot fields.
type stubCustomerSnapshots struct {
	name    string
	email   string
	address string
}

func (s stubCustomerSnapshots) CustomerSnapshot(_ context.Context, _, _ int64) (string, string, string, error) {
	return s.name, s.email, s.address, nil
}

type failingCustomerSnapshots struct{}

func (failingCustomerSnapshots) CustomerSnapshot(_ context.Context, _, _ int64) (string, string, string, error) {
	return "", "", "", errors.New("customer not found")
}

// newCustomerLinkAPI wires the real invoices service to a dedicated tenant so
// the invoice-number sequence cannot collide with the shared fixture tenant.
func newCustomerLinkAPI(t *testing.T, businessID, userID int64, provider application.CustomerSnapshotProvider) (*InvoicesAPI, context.Context, *sql.DB) {
	t.Helper()

	db := testutil.OpenDB(t)
	testutil.GrantBusinessAccess(t, db, businessID, userID)

	user := financeUser("finance:invoices:create", "finance:invoices:view")
	user.BusinessID = businessID
	user.UserID = userID
	ctx := authPkg.WithAuthUser(context.Background(), user)

	svc := application.NewService(ctx, db)
	svc.SetCustomerSnapshotProvider(provider)
	return NewInvoicesAPI(svc), ctx, db
}

func createInvoiceRequest(customerID int64) *financeinvoicesv1.CreateInvoiceRequest {
	return &financeinvoicesv1.CreateInvoiceRequest{
		CustomerId:   &customerID,
		IssueDate:    "2026-06-15",
		DueDate:      "2026-06-30",
		Currency:     "USD",
		DiscountType: "fixed",
		Items: []*financeinvoicesv1.InvoiceItem{
			{Description: "Consulting", Quantity: 1, UnitPrice: 100},
		},
	}
}

// Regression: InvoicesAPI.CreateInvoice dropped the request's customer_id, so
// the ownership verification never ran and the persisted row carried no
// customer link, making per-customer reads unable to return the invoice.
func TestCreateInvoicePersistsCustomerLink(t *testing.T) {
	const businessID int64 = 993101
	const userID int64 = 993101
	apiSvc, ctx, db := newCustomerLinkAPI(t, businessID, userID, stubCustomerSnapshots{
		name: "Acme Ltd", email: "billing@acme.test", address: "1 Test Way",
	})

	customerID := int64(5301)
	resp, err := apiSvc.CreateInvoice(ctx, createInvoiceRequest(customerID))
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}

	invoice := resp.GetInvoice()
	if invoice.GetCustomerId() != customerID {
		t.Fatalf("response customer_id = %d, want %d", invoice.GetCustomerId(), customerID)
	}
	if invoice.GetClientName() != "Acme Ltd" || invoice.GetClientEmail() != "billing@acme.test" {
		t.Fatalf("customer snapshot not applied: name=%q email=%q", invoice.GetClientName(), invoice.GetClientEmail())
	}

	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM finance_invoice_items WHERE invoice_id = ?", invoice.GetId())
		_, _ = db.ExecContext(context.Background(), "DELETE FROM finance_invoices WHERE id = ?", invoice.GetId())
	})

	var storedCustomer int64
	var storedName string
	if err := db.QueryRowContext(ctx, `
		SELECT customer_id, client_name FROM finance_invoices WHERE id = ? AND business_id = ?`,
		invoice.GetId(), businessID).Scan(&storedCustomer, &storedName); err != nil {
		t.Fatalf("read stored invoice: %v", err)
	}
	if storedCustomer != customerID {
		t.Fatalf("stored customer_id = %d, want %d", storedCustomer, customerID)
	}
	if storedName != "Acme Ltd" {
		t.Fatalf("stored client_name = %q, want the snapshotted name", storedName)
	}

	// The per-customer read surface must return the linked invoice.
	listResp, err := apiSvc.ListInvoices(ctx, &financeinvoicesv1.ListInvoicesRequest{CustomerId: &customerID})
	if err != nil {
		t.Fatalf("ListInvoices(customer_id=%d): %v", customerID, err)
	}
	found := false
	for _, summary := range listResp.GetInvoices() {
		if summary.GetId() == invoice.GetId() {
			found = true
		}
	}
	if !found {
		t.Fatalf("invoice %d missing from ListInvoices(customer_id=%d)", invoice.GetId(), customerID)
	}
}

func TestCreateInvoiceUnknownCustomerIsNotFound(t *testing.T) {
	const businessID int64 = 993102
	const userID int64 = 993102
	apiSvc, ctx, db := newCustomerLinkAPI(t, businessID, userID, failingCustomerSnapshots{})

	_, err := apiSvc.CreateInvoice(ctx, createInvoiceRequest(999999))
	if got := financePermissionStatus(err); got != http.StatusNotFound {
		t.Fatalf("status = %d, want %d (err = %v)", got, http.StatusNotFound, err)
	}

	var invoices int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM finance_invoices WHERE business_id = ?", businessID,
	).Scan(&invoices); err != nil {
		t.Fatalf("count invoices: %v", err)
	}
	if invoices != 0 {
		t.Fatalf("unverified customer link persisted %d invoices, want 0", invoices)
	}
}
