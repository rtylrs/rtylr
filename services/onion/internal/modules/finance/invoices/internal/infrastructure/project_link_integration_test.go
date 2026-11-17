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

func insertProjectLinkFlowProject(t *testing.T, db *sql.DB, businessID int64, name string) int64 {
	t.Helper()
	ctx := context.Background()
	result, err := db.ExecContext(ctx,
		`INSERT INTO flow_projects (workspace_id, owner_id, business_id, name, status, created_by)
		 VALUES (0, ?, ?, ?, 'active', ?)`,
		businessID, businessID, name, businessID,
	)
	if err != nil {
		t.Fatalf("insert flow_projects(%s): %v", name, err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("flow_projects id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM flow_projects WHERE id = ?", id)
	})
	return id
}

func cleanupProjectLinkInvoices(t *testing.T, db *sql.DB, businessIDs ...int64) {
	t.Helper()
	t.Cleanup(func() {
		ctx := context.Background()
		for _, businessID := range businessIDs {
			_, _ = db.ExecContext(ctx, "DELETE FROM finance_invoice_items WHERE invoice_id IN (SELECT id FROM finance_invoices WHERE business_id = ?)", businessID)
			_, _ = db.ExecContext(ctx, "DELETE FROM finance_invoices WHERE business_id = ?", businessID)
		}
	})
}

func TestInvoiceProjectID_RoundTripsThroughRepository(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	businessID := int64(994567)
	cleanupProjectLinkInvoices(t, db, businessID)
	projectID := int64(424242)

	tagged, err := repo.CreateInvoiceWithItems(ctx, businessID, domain.Invoice{
		InvoiceNumber: "PROJECT-LINK-1",
		ClientName:    "Tagged Client",
		ProjectID:     &projectID,
		Status:        domain.StatusDraft,
		IssueDate:     time.Now(),
		DueDate:       time.Now().Add(24 * time.Hour),
		Currency:      "USD",
	}, []domain.InvoiceItem{{Description: "Design", Quantity: 1, UnitPrice: 100, Total: 100}})
	if err != nil {
		t.Fatalf("CreateInvoiceWithItems(tagged): %v", err)
	}
	untagged, err := repo.CreateInvoiceWithItems(ctx, businessID, domain.Invoice{
		InvoiceNumber: "PROJECT-LINK-2",
		ClientName:    "Untagged Client",
		Status:        domain.StatusDraft,
		IssueDate:     time.Now(),
		DueDate:       time.Now().Add(24 * time.Hour),
		Currency:      "USD",
	}, nil)
	if err != nil {
		t.Fatalf("CreateInvoiceWithItems(untagged): %v", err)
	}

	got, err := repo.GetInvoiceByID(ctx, businessID, tagged.ID)
	if err != nil {
		t.Fatalf("GetInvoiceByID: %v", err)
	}
	if got.ProjectID == nil || *got.ProjectID != projectID {
		t.Fatalf("GetInvoiceByID project = %v, want %d", got.ProjectID, projectID)
	}

	all, err := repo.ListInvoices(ctx, businessID, "", "", "", "", 50, 0)
	if err != nil {
		t.Fatalf("ListInvoices: %v", err)
	}
	seen := map[int64]*int64{}
	for _, inv := range all {
		seen[inv.ID] = inv.ProjectID
	}
	if p, ok := seen[tagged.ID]; !ok || p == nil || *p != projectID {
		t.Fatalf("ListInvoices project for tagged invoice = %v, want %d", p, projectID)
	}
	if p, ok := seen[untagged.ID]; !ok || p != nil {
		t.Fatalf("ListInvoices project for untagged invoice = %v (present %v), want none", p, ok)
	}

	byProject, err := repo.ListInvoicesForProject(ctx, businessID, projectID, "", "", "", "", 50, 0)
	if err != nil {
		t.Fatalf("ListInvoicesForProject: %v", err)
	}
	if len(byProject) != 1 || byProject[0].ID != tagged.ID {
		t.Fatalf("ListInvoicesForProject = %+v, want only invoice %d", byProject, tagged.ID)
	}

	got.ProjectID = nil
	if err := repo.UpdateInvoiceWithItems(ctx, businessID, *got, nil); err != nil {
		t.Fatalf("UpdateInvoiceWithItems(clear project): %v", err)
	}
	cleared, err := repo.GetInvoiceByID(ctx, businessID, tagged.ID)
	if err != nil {
		t.Fatalf("GetInvoiceByID after clear: %v", err)
	}
	if cleared.ProjectID != nil {
		t.Fatalf("project after clear = %d, want none", *cleared.ProjectID)
	}
}

func TestCreateInvoice_RejectsProjectFromAnotherBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	businessID := int64(994568)
	otherBusinessID := int64(994569)
	cleanupProjectLinkInvoices(t, db, businessID, otherBusinessID)
	ownProject := insertProjectLinkFlowProject(t, db, businessID, "Own project")
	foreignProject := insertProjectLinkFlowProject(t, db, otherBusinessID, "Foreign project")

	service := application.NewService(infrastructure.NewRepository(db), nil)
	service.SetProjectReferenceReader(flowreferences.NewReader(db))
	input := func(projectID int64) domain.InvoiceInput {
		return domain.InvoiceInput{
			ClientName: "Project Client",
			ProjectID:  projectID,
			IssueDate:  "2026-11-01",
			DueDate:    "2026-11-30",
			Currency:   "USD",
			Items:      []domain.InvoiceItemInput{{Description: "Build", Quantity: 1, UnitPrice: 250}},
		}
	}

	if _, err := service.CreateInvoice(ctx, businessID, input(foreignProject)); !errors.Is(err, application.ErrProjectNotFound) {
		t.Fatalf("CreateInvoice(foreign project) error = %v, want ErrProjectNotFound", err)
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM finance_invoices WHERE business_id = ?", businessID).Scan(&count); err != nil {
		t.Fatalf("count invoices: %v", err)
	}
	if count != 0 {
		t.Fatalf("invoices written for a foreign project = %d, want 0", count)
	}

	created, err := service.CreateInvoice(ctx, businessID, input(ownProject))
	if err != nil {
		t.Fatalf("CreateInvoice(own project): %v", err)
	}
	stored, err := service.GetInvoice(ctx, businessID, created.ID)
	if err != nil {
		t.Fatalf("GetInvoice: %v", err)
	}
	if stored.ProjectID == nil || *stored.ProjectID != ownProject {
		t.Fatalf("stored project = %v, want %d", stored.ProjectID, ownProject)
	}

	foreign := foreignProject
	if _, err := service.UpdateInvoice(ctx, businessID, created.ID, domain.InvoiceUpdateInput{
		ClientName: "Project Client",
		ProjectID:  &foreign,
		IssueDate:  "2026-11-01",
		DueDate:    "2026-11-30",
		Currency:   "USD",
		Items:      []domain.InvoiceItemInput{{Description: "Build", Quantity: 1, UnitPrice: 250}},
	}); !errors.Is(err, application.ErrProjectNotFound) {
		t.Fatalf("UpdateInvoice(foreign project) error = %v, want ErrProjectNotFound", err)
	}
	unchanged, err := service.GetInvoice(ctx, businessID, created.ID)
	if err != nil {
		t.Fatalf("GetInvoice after rejected update: %v", err)
	}
	if unchanged.ProjectID == nil || *unchanged.ProjectID != ownProject {
		t.Fatalf("project after rejected update = %v, want %d", unchanged.ProjectID, ownProject)
	}
}
