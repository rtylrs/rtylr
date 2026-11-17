package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/infrastructure"
	flowreferences "github.com/voxire/onion/internal/modules/flow/references"
	"github.com/voxire/onion/internal/testutil"
)

// invoiceFingerprint is every stored column a project re-tag must not touch,
// plus the item count, as one comparable string.
func invoiceFingerprint(t *testing.T, db *sql.DB, invoiceID int64) string {
	t.Helper()
	var fingerprint string
	err := db.QueryRowContext(context.Background(), `
		SELECT CONCAT_WS('|', invoice_number, client_name, client_email, status, issue_date, due_date,
		       currency, subtotal, discount_type, discount_value, discount_amount, tax_rate, tax_amount,
		       total, amount_paid, amount_due, IFNULL(customer_id, '-'), IFNULL(notes, '-'),
		       (SELECT COUNT(*) FROM finance_invoice_items WHERE invoice_id = finance_invoices.id))
		FROM finance_invoices WHERE id = ?`, invoiceID).Scan(&fingerprint)
	if err != nil {
		t.Fatalf("fingerprint invoice %d: %v", invoiceID, err)
	}
	return fingerprint
}

func TestSetInvoiceProject_WritesOnlyTheTag(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	businessID := int64(994571)
	otherBusinessID := int64(994572)
	cleanupProjectLinkInvoices(t, db, businessID, otherBusinessID)
	ownProject := insertProjectLinkFlowProject(t, db, businessID, "Billed project")
	foreignProject := insertProjectLinkFlowProject(t, db, otherBusinessID, "Someone else's project")

	repo := infrastructure.NewRepository(db)
	service := application.NewService(repo, nil)
	service.SetProjectReferenceReader(flowreferences.NewReader(db))

	created, err := service.CreateInvoice(ctx, businessID, domain.InvoiceInput{
		ClientName: "Sent Client",
		IssueDate:  "2026-11-01",
		DueDate:    "2026-11-30",
		Currency:   "USD",
		TaxRate:    10,
		Items:      []domain.InvoiceItemInput{{Description: "Build", Quantity: 2, UnitPrice: 125}},
	})
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		"UPDATE finance_invoices SET status = 'sent', amount_paid = 40, amount_due = total - 40, updated_at = '2020-01-01 00:00:00' WHERE id = ?",
		created.ID); err != nil {
		t.Fatalf("mark sent: %v", err)
	}
	before := invoiceFingerprint(t, db, created.ID)

	if _, err := service.SetInvoiceProject(ctx, businessID, created.ID, foreignProject); !errors.Is(err, application.ErrProjectNotFound) {
		t.Fatalf("SetInvoiceProject(foreign) error = %v, want ErrProjectNotFound", err)
	}

	tagged, err := service.SetInvoiceProject(ctx, businessID, created.ID, ownProject)
	if err != nil {
		t.Fatalf("SetInvoiceProject(own): %v", err)
	}
	if tagged.ProjectID == nil || *tagged.ProjectID != ownProject {
		t.Fatalf("project = %v, want %d", tagged.ProjectID, ownProject)
	}
	if after := invoiceFingerprint(t, db, created.ID); after != before {
		t.Fatalf("re-tag changed the invoice:\nbefore %s\nafter  %s", before, after)
	}
	var updatedAt time.Time
	if err := db.QueryRowContext(ctx, "SELECT updated_at FROM finance_invoices WHERE id = ?", created.ID).Scan(&updatedAt); err != nil {
		t.Fatalf("read updated_at: %v", err)
	}
	if !updatedAt.After(time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("updated_at = %s, want it bumped by the re-tag", updatedAt)
	}

	byProject, err := repo.ListInvoicesForProject(ctx, businessID, ownProject, "", "", "", "", 50, 0)
	if err != nil || len(byProject) != 1 || byProject[0].ID != created.ID {
		t.Fatalf("ListInvoicesForProject = %+v, %v; want only invoice %d", byProject, err, created.ID)
	}

	if err := repo.SetInvoiceProject(ctx, businessID, created.ID, nil); err != nil {
		t.Fatalf("SetInvoiceProject(clear): %v", err)
	}
	cleared, err := repo.GetInvoiceByID(ctx, businessID, created.ID)
	if err != nil {
		t.Fatalf("GetInvoiceByID: %v", err)
	}
	if cleared.ProjectID != nil {
		t.Fatalf("project after clear = %d, want none", *cleared.ProjectID)
	}
	if after := invoiceFingerprint(t, db, created.ID); after != before {
		t.Fatalf("clear changed the invoice:\nbefore %s\nafter  %s", before, after)
	}

	// The repository is scoped by business as well as by id.
	tag := ownProject
	if err := repo.SetInvoiceProject(ctx, otherBusinessID, created.ID, &tag); !errors.Is(err, domain.ErrInvoiceNotFound) {
		t.Fatalf("SetInvoiceProject(other business) error = %v, want ErrInvoiceNotFound", err)
	}
}

func TestSetInvoiceProject_CancelledAndVoidAreFrozen(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	businessID := int64(994573)
	cleanupProjectLinkInvoices(t, db, businessID)
	repo := infrastructure.NewRepository(db)
	project := int64(515151)

	for _, status := range []string{domain.StatusCancelled, domain.StatusVoid} {
		t.Run(status, func(t *testing.T) {
			inv, err := repo.CreateInvoiceWithItems(ctx, businessID, domain.Invoice{
				InvoiceNumber: "SET-PROJECT-" + status,
				ClientName:    "Closed Client",
				ProjectID:     &project,
				Status:        status,
				IssueDate:     time.Now(),
				DueDate:       time.Now().Add(24 * time.Hour),
				Currency:      "USD",
			}, nil)
			if err != nil {
				t.Fatalf("CreateInvoiceWithItems: %v", err)
			}
			if err := repo.SetInvoiceProject(ctx, businessID, inv.ID, nil); !errors.Is(err, domain.ErrInvoiceProjectLocked) {
				t.Fatalf("SetInvoiceProject(%s) error = %v, want ErrInvoiceProjectLocked", status, err)
			}
			stored, err := repo.GetInvoiceByID(ctx, businessID, inv.ID)
			if err != nil {
				t.Fatalf("GetInvoiceByID: %v", err)
			}
			if stored.ProjectID == nil || *stored.ProjectID != project {
				t.Fatalf("%s invoice project = %v, want %d unchanged", status, stored.ProjectID, project)
			}
		})
	}
}
