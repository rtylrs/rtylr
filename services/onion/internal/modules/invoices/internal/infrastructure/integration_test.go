package infrastructure_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	"github.com/voxire/onion/internal/modules/invoices/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListInvoices_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	invoices, err := repo.ListInvoices(context.Background(), testutil.OwnerID, "", "", "", "", 50, 0)
	if err != nil {
		t.Fatalf("ListInvoices error: %v", err)
	}
	// Fixture may have no invoices; just assert the call succeeds.
	_ = invoices
}

func TestSumOverdueAmountUsesDueDateForSentInvoices(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	businessID := int64(994101)

	_, err := db.ExecContext(ctx, `
		INSERT INTO finance_invoices
			(owner_id, business_id, invoice_number, client_name, status,
			 issue_date, due_date, total, amount_due)
		VALUES (?, ?, 'OVERDUE-REGRESSION-994101', 'Overdue Client', 'sent',
			'2020-01-01', '2020-01-02', 125.50, 125.50)`, businessID, businessID)
	if err != nil {
		t.Fatalf("insert overdue invoice: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM finance_invoices WHERE business_id = ?", businessID)
	})

	got, err := repo.SumOverdueAmount(ctx, businessID)
	if err != nil {
		t.Fatalf("SumOverdueAmount: %v", err)
	}
	if got != 125.50 {
		t.Fatalf("overdue amount = %v, want 125.50", got)
	}
}

func TestListFinanceExpenses_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	expenses, err := repo.ListExpenses(context.Background(), testutil.OwnerID, 0, "", "", "", 50, 0)
	if err != nil {
		t.Fatalf("ListExpenses error: %v", err)
	}
	// Fixture may have no expenses; just assert the call succeeds.
	_ = expenses
}

func TestListExpenseCategories_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	cats, err := repo.ListExpenseCategories(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListExpenseCategories error: %v", err)
	}
	_ = cats
}

func TestListInvoices_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	invoices, err := repo.ListInvoices(ctx, 0, "", "", "", "", 100, 0)
	if err != nil {
		t.Fatalf("ListInvoices owner=0: %v", err)
	}
	if len(invoices) != 0 {
		t.Errorf("expected 0 invoices for owner_id=0, got %d", len(invoices))
	}
}

func TestListInvoices_AllStatuses(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	invoices, err := repo.ListInvoices(ctx, testutil.OwnerID, "", "", "", "", 200, 0)
	if err != nil {
		t.Fatalf("ListInvoices all statuses: %v", err)
	}
	for _, inv := range invoices {
		if inv.Status == "" {
			t.Errorf("invoice id=%d has empty Status", inv.ID)
		}
	}
}

func TestListInvoices_StatusFilter_Paid(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	invoices, err := repo.ListInvoices(ctx, testutil.OwnerID, "paid", "", "", "", 100, 0)
	if err != nil {
		t.Fatalf("ListInvoices status=paid: %v", err)
	}
	for _, inv := range invoices {
		if inv.Status != "paid" {
			t.Errorf("invoice id=%d has status=%q, want 'paid'", inv.ID, inv.Status)
		}
	}
}

func TestListInvoices_DateRange(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	invoices, err := repo.ListInvoices(ctx, testutil.OwnerID, "", "", "2026-01-01", "2026-12-31", 100, 0)
	if err != nil {
		t.Fatalf("ListInvoices date range: %v", err)
	}
	_ = invoices
}

func TestCreateInvoice_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateInvoice(ctx, testutil.OwnerID, domain.Invoice{
		InvoiceNumber: "TEST-001",
		ClientName:    "Acme Corp",
		IssueDate:     time.Now(),
		DueDate:       time.Now().Add(30 * 24 * time.Hour),
		Currency:      "USD",
		Status:        "draft",
		Total:         1000.00,
		AmountDue:     1000.00,
	})
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM finance_invoices WHERE id = ?", id)
	})
}

func TestCreateInvoice_StatusIsDraft(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateInvoice(ctx, testutil.OwnerID, domain.Invoice{
		InvoiceNumber: "TEST-002",
		ClientName:    "Draft Corp",
		IssueDate:     time.Now(),
		DueDate:       time.Now().Add(30 * 24 * time.Hour),
		Currency:      "USD",
		Status:        "draft",
	})
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM finance_invoices WHERE id = ?", id)
	})

	inv, err := repo.GetInvoiceByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetInvoiceByID: %v", err)
	}
	if inv.Status != "draft" {
		t.Errorf("Status = %q, want 'draft'", inv.Status)
	}
}

func TestCreateInvoice_TotalAmountIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wantTotal := 2500.00

	id, err := repo.CreateInvoice(ctx, testutil.OwnerID, domain.Invoice{
		InvoiceNumber: "TEST-003",
		ClientName:    "Total Corp",
		IssueDate:     time.Now(),
		DueDate:       time.Now().Add(30 * 24 * time.Hour),
		Currency:      "USD",
		Status:        "draft",
		Total:         wantTotal,
		AmountDue:     wantTotal,
	})
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM finance_invoices WHERE id = ?", id)
	})

	inv, err := repo.GetInvoiceByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetInvoiceByID: %v", err)
	}
	if inv.Total != wantTotal {
		t.Errorf("Total = %g, want %g", inv.Total, wantTotal)
	}
}

func TestGetInvoiceByID_ByID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateInvoice(ctx, testutil.OwnerID, domain.Invoice{
		InvoiceNumber: "TEST-004",
		ClientName:    "Get Corp",
		IssueDate:     time.Now(),
		DueDate:       time.Now().Add(30 * 24 * time.Hour),
		Currency:      "USD",
		Status:        "draft",
	})
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM finance_invoices WHERE id = ?", id)
	})

	inv, err := repo.GetInvoiceByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetInvoiceByID: %v", err)
	}
	if inv.ID != id {
		t.Errorf("ID = %d, want %d", inv.ID, id)
	}
	if inv.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", inv.OwnerID, testutil.OwnerID)
	}
}

func TestGetInvoiceByID_NotFound_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetInvoiceByID(ctx, testutil.OwnerID, 9999999)
	if err == nil {
		t.Fatal("expected error for non-existent invoice, got nil")
	}
}

func TestCreateInvoiceItems_PersistsItems(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	invID, err := repo.CreateInvoice(ctx, testutil.OwnerID, domain.Invoice{
		InvoiceNumber: "TEST-005",
		ClientName:    "Item Corp",
		IssueDate:     time.Now(),
		DueDate:       time.Now().Add(30 * 24 * time.Hour),
		Currency:      "USD",
		Status:        "draft",
	})
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM finance_invoice_items WHERE invoice_id = ?", invID)
		db.ExecContext(ctx, "DELETE FROM finance_invoices WHERE id = ?", invID)
	})

	if err := repo.CreateInvoiceItems(ctx, invID, []domain.InvoiceItem{
		{Description: "Widget A", Quantity: 2, UnitPrice: 50, Total: 100},
		{Description: "Widget B", Quantity: 1, UnitPrice: 200, Total: 200},
	}); err != nil {
		t.Fatalf("CreateInvoiceItems: %v", err)
	}

	items, err := repo.GetInvoiceItems(ctx, invID)
	if err != nil {
		t.Fatalf("GetInvoiceItems: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestGetInvoiceItems_ReturnsItems(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	invID, err := repo.CreateInvoice(ctx, testutil.OwnerID, domain.Invoice{
		InvoiceNumber: "TEST-006",
		ClientName:    "Items Get Corp",
		IssueDate:     time.Now(),
		DueDate:       time.Now().Add(30 * 24 * time.Hour),
		Currency:      "USD",
		Status:        "draft",
	})
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM finance_invoice_items WHERE invoice_id = ?", invID)
		db.ExecContext(ctx, "DELETE FROM finance_invoices WHERE id = ?", invID)
	})

	if err := repo.CreateInvoiceItems(ctx, invID, []domain.InvoiceItem{
		{Description: "Service fee", Quantity: 1, UnitPrice: 500, Total: 500},
	}); err != nil {
		t.Fatalf("CreateInvoiceItems: %v", err)
	}

	items, err := repo.GetInvoiceItems(ctx, invID)
	if err != nil {
		t.Fatalf("GetInvoiceItems: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}
	if items[0].Description != "Service fee" {
		t.Errorf("Description = %q, want 'Service fee'", items[0].Description)
	}
}

func TestUpdateInvoiceStatus_ToPaid(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	invID, err := repo.CreateInvoice(ctx, testutil.OwnerID, domain.Invoice{
		InvoiceNumber: "TEST-007",
		ClientName:    "Pay Corp",
		IssueDate:     time.Now(),
		DueDate:       time.Now().Add(30 * 24 * time.Hour),
		Currency:      "USD",
		Status:        "draft",
	})
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM finance_invoices WHERE id = ?", invID)
	})

	if err := repo.UpdateInvoiceStatus(ctx, testutil.OwnerID, invID, "paid"); err != nil {
		t.Fatalf("UpdateInvoiceStatus: %v", err)
	}

	inv, err := repo.GetInvoiceByID(ctx, testutil.OwnerID, invID)
	if err != nil {
		t.Fatalf("GetInvoiceByID after status update: %v", err)
	}
	if inv.Status != "paid" {
		t.Errorf("Status = %q after UpdateInvoiceStatus, want 'paid'", inv.Status)
	}
}

func TestInvoiceLifecycle_PreventsReopeningAndEditingPaidInvoice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	invID, err := repo.CreateInvoice(ctx, testutil.OwnerID, domain.Invoice{
		InvoiceNumber: "TEST-LIFECYCLE",
		ClientName:    "Lifecycle Corp",
		IssueDate:     time.Now(),
		DueDate:       time.Now().Add(30 * 24 * time.Hour),
		Currency:      "USD",
		Status:        domain.StatusDraft,
		Total:         100,
		AmountDue:     100,
	})
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM finance_invoice_items WHERE invoice_id = ?", invID)
		db.ExecContext(ctx, "DELETE FROM finance_invoices WHERE id = ?", invID)
	})

	if err := repo.UpdateInvoiceStatus(ctx, testutil.OwnerID, invID, domain.StatusPaid); err != nil {
		t.Fatalf("mark invoice paid: %v", err)
	}
	if err := repo.UpdateInvoiceStatus(ctx, testutil.OwnerID, invID, domain.StatusDraft); err != nil {
		t.Fatalf("attempt to reopen paid invoice: %v", err)
	}

	paid, err := repo.GetInvoiceByID(ctx, testutil.OwnerID, invID)
	if err != nil {
		t.Fatalf("GetInvoiceByID after reopen attempt: %v", err)
	}
	if paid.Status != domain.StatusPaid {
		t.Fatalf("paid invoice status changed to %q after reopen attempt", paid.Status)
	}

	paid.ClientName = "Tampered Corp"
	paid.Total = 999
	if err := repo.UpdateInvoice(ctx, testutil.OwnerID, *paid); !errors.Is(err, domain.ErrInvoiceNotDraft) {
		t.Fatalf("expected paid invoice edit to return ErrInvoiceNotDraft, got %v", err)
	}

	unchanged, err := repo.GetInvoiceByID(ctx, testutil.OwnerID, invID)
	if err != nil {
		t.Fatalf("GetInvoiceByID after edit attempt: %v", err)
	}
	if unchanged.ClientName != "Lifecycle Corp" || unchanged.Total != 100 {
		t.Fatalf("paid invoice was edited: client=%q total=%g", unchanged.ClientName, unchanged.Total)
	}
}

func TestDeleteInvoiceItems_ClearsItems(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	invID, err := repo.CreateInvoice(ctx, testutil.OwnerID, domain.Invoice{
		InvoiceNumber: "TEST-008",
		ClientName:    "Delete Items Corp",
		IssueDate:     time.Now(),
		DueDate:       time.Now().Add(30 * 24 * time.Hour),
		Currency:      "USD",
		Status:        "draft",
	})
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM finance_invoices WHERE id = ?", invID)
	})

	if err := repo.CreateInvoiceItems(ctx, invID, []domain.InvoiceItem{
		{Description: "Temporary item", Quantity: 1, UnitPrice: 100, Total: 100},
	}); err != nil {
		t.Fatalf("CreateInvoiceItems: %v", err)
	}

	if err := repo.DeleteInvoiceItems(ctx, invID); err != nil {
		t.Fatalf("DeleteInvoiceItems: %v", err)
	}

	items, err := repo.GetInvoiceItems(ctx, invID)
	if err != nil {
		t.Fatalf("GetInvoiceItems after delete: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items after delete, got %d", len(items))
	}
}

func TestListInvoices_SearchByCustomer(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	invID, err := repo.CreateInvoice(ctx, testutil.OwnerID, domain.Invoice{
		InvoiceNumber: "TEST-009",
		ClientName:    "Searchable Ltd",
		IssueDate:     time.Now(),
		DueDate:       time.Now().Add(30 * 24 * time.Hour),
		Currency:      "USD",
		Status:        "draft",
	})
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM finance_invoices WHERE id = ?", invID)
	})

	results, err := repo.ListInvoices(ctx, testutil.OwnerID, "", "Searchable", "", "", 50, 0)
	if err != nil {
		t.Fatalf("ListInvoices search: %v", err)
	}

	found := false
	for _, inv := range results {
		if inv.ID == invID {
			found = true
		}
	}
	if !found {
		t.Errorf("created invoice id=%d not found when searching by client name", invID)
	}
}

func TestCountInvoicesByStatus_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	counts, err := repo.CountInvoicesByStatus(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CountInvoicesByStatus: %v", err)
	}
	// Returned map may be empty if owner has no invoices; just assert no error.
	_ = counts
}

func TestCountInvoicesByStatus_ZeroOwnerReturnsEmptyMap(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	counts, err := repo.CountInvoicesByStatus(ctx, 0)
	if err != nil {
		t.Fatalf("CountInvoicesByStatus(ownerID=0): %v", err)
	}
	if len(counts) != 0 {
		t.Errorf("expected empty map for ownerID=0, got %v", counts)
	}
}

func TestCountInvoicesByStatus_ReflectsCreatedInvoice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateInvoice(ctx, testutil.OwnerID, domain.Invoice{
		InvoiceNumber: "CNT-001",
		ClientName:    "Count Corp",
		IssueDate:     time.Now(),
		DueDate:       time.Now().Add(30 * 24 * time.Hour),
		Currency:      "USD",
		Status:        "draft",
	})
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM finance_invoices WHERE id = ?", id)
	})

	counts, err := repo.CountInvoicesByStatus(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CountInvoicesByStatus: %v", err)
	}
	if counts["draft"] <= 0 {
		t.Errorf("expected at least 1 draft invoice, got %d", counts["draft"])
	}
}

func TestGetOverdueInvoices_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	overdue, err := repo.GetOverdueInvoices(ctx, testutil.OwnerID, "", 100, 0)
	if err != nil {
		t.Fatalf("GetOverdueInvoices: %v", err)
	}
	_ = overdue
}

func TestGetOverdueInvoices_ZeroOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	overdue, err := repo.GetOverdueInvoices(ctx, 0, "", 100, 0)
	if err != nil {
		t.Fatalf("GetOverdueInvoices(ownerID=0): %v", err)
	}
	if len(overdue) != 0 {
		t.Errorf("expected empty slice for ownerID=0, got %d", len(overdue))
	}
}

func TestGetOverdueInvoices_FutureCutoffReturnsNone(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// A past cutoff means nothing is "overdue before year 2000".
	overdue, err := repo.GetOverdueInvoices(ctx, testutil.OwnerID, "2000-01-01", 100, 0)
	if err != nil {
		t.Fatalf("GetOverdueInvoices past cutoff: %v", err)
	}
	if len(overdue) != 0 {
		t.Errorf("expected 0 overdue before year 2000, got %d", len(overdue))
	}
}
