package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/recurring_invoices/internal/domain"
	"github.com/voxire/onion/internal/modules/recurring_invoices/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListRecurringInvoices_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("recurring_invoices.List: unexpected error: %v", err)
	}
	if items == nil {
		t.Fatal("recurring_invoices.List: expected non-nil slice, got nil")
	}
}

func TestListRecurringInvoices_WrongOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("recurring_invoices.List(ownerID=0): unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("recurring_invoices.List(ownerID=0): expected 0 rows, got %d", len(items))
	}
}

func seedRecurringInvoice(t *testing.T, ctx context.Context, repo *infrastructure.Repository, ownerID int64, freq string) (int64, error) {
	t.Helper()
	// We need a template_invoice_id -- use 0 and let the DB accept it.
	input := domain.RecurringInvoiceInput{
		TemplateInvoiceID: 0,
		Frequency:         freq,
		NextInvoiceDate:   "2026-06-01",
		IsActive:          true,
	}
	return repo.Create(ctx, ownerID, input)
}

func TestCreateRecurringInvoice_ReturnsInsertedID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := seedRecurringInvoice(t, ctx, repo, testutil.OwnerID, "monthly")
	if err != nil {
		t.Fatalf("recurring_invoices.Create: unexpected error: %v", err)
	}
	if id <= 0 {
		t.Fatalf("recurring_invoices.Create: expected positive ID, got %d", id)
	}
}

func TestCreateRecurringInvoice_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := seedRecurringInvoice(t, ctx, repo, testutil.OwnerID, "weekly")
	if err != nil {
		t.Fatalf("recurring_invoices.Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("recurring_invoices.List: %v", err)
	}

	var found bool
	for _, ri := range items {
		if ri.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created recurring invoice id=%d not found in list", id)
	}
}

func TestCreateRecurringInvoice_FrequencyPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := seedRecurringInvoice(t, ctx, repo, testutil.OwnerID, "monthly")
	if err != nil {
		t.Fatalf("recurring_invoices.Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("recurring_invoices.List: %v", err)
	}

	for _, ri := range items {
		if ri.ID == id {
			if ri.Frequency != "monthly" {
				t.Errorf("Frequency mismatch: got %q, want monthly", ri.Frequency)
			}
			return
		}
	}
	t.Fatalf("created recurring invoice id=%d not found in list", id)
}

func TestCreateRecurringInvoice_IsActivePersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.RecurringInvoiceInput{
		TemplateInvoiceID: 0,
		Frequency:         "monthly",
		NextInvoiceDate:   "2026-07-01",
		IsActive:          true,
	})
	if err != nil {
		t.Fatalf("recurring_invoices.Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("recurring_invoices.List: %v", err)
	}

	for _, ri := range items {
		if ri.ID == id {
			if !ri.IsActive {
				t.Errorf("expected IsActive=true, got false")
			}
			return
		}
	}
	t.Fatalf("created recurring invoice id=%d not found in list", id)
}

func TestSetActive_DisablesRecurringInvoice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := seedRecurringInvoice(t, ctx, repo, testutil.OwnerID, "weekly")
	if err != nil {
		t.Fatalf("recurring_invoices.Create: %v", err)
	}

	err = repo.SetActive(ctx, testutil.OwnerID, id, false)
	if err != nil {
		t.Fatalf("recurring_invoices.SetActive(false): %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("recurring_invoices.List after SetActive: %v", err)
	}

	for _, ri := range items {
		if ri.ID == id {
			if ri.IsActive {
				t.Errorf("expected IsActive=false after SetActive(false), got true")
			}
			return
		}
	}
	t.Fatalf("recurring invoice id=%d not found after SetActive", id)
}

func TestSetActive_NotFound_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.SetActive(ctx, testutil.OwnerID, 0, true)
	if err == nil {
		t.Fatal("recurring_invoices.SetActive(id=0): expected error, got nil")
	}
}

func TestUpdateRecurringInvoice_FrequencyChanges(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := seedRecurringInvoice(t, ctx, repo, testutil.OwnerID, "monthly")
	if err != nil {
		t.Fatalf("recurring_invoices.Create: %v", err)
	}

	err = repo.Update(ctx, testutil.OwnerID, id, domain.RecurringInvoiceInput{
		TemplateInvoiceID: 0,
		Frequency:         "yearly",
		NextInvoiceDate:   "2027-01-01",
		IsActive:          true,
	})
	if err != nil {
		t.Fatalf("recurring_invoices.Update: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("recurring_invoices.List after update: %v", err)
	}

	for _, ri := range items {
		if ri.ID == id {
			if ri.Frequency != "yearly" {
				t.Errorf("expected frequency=yearly after update, got %q", ri.Frequency)
			}
			return
		}
	}
	t.Fatalf("recurring invoice id=%d not found after update", id)
}

func TestListRecurringInvoices_EachHasPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := seedRecurringInvoice(t, ctx, repo, testutil.OwnerID, "monthly")
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("recurring_invoices.List: %v", err)
	}

	for _, ri := range items {
		if ri.ID <= 0 {
			t.Errorf("recurring invoice has non-positive ID: %d", ri.ID)
		}
	}
}

func TestListRecurringInvoices_ScopedToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := seedRecurringInvoice(t, ctx, repo, testutil.OwnerID, "monthly")
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	other, err := repo.List(ctx, testutil.OwnerID+9999)
	if err != nil {
		t.Fatalf("recurring_invoices.List(other owner): %v", err)
	}
	for _, ri := range other {
		if ri.OwnerID == testutil.OwnerID {
			t.Errorf("other-owner list contains ownerID=%d record", testutil.OwnerID)
		}
	}
}

func TestListRecurringInvoices_EachHasNonZeroOwnerID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := seedRecurringInvoice(t, ctx, repo, testutil.OwnerID, "weekly")
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("recurring_invoices.List: %v", err)
	}

	for _, ri := range items {
		if ri.OwnerID == 0 {
			t.Errorf("recurring invoice id=%d has zero OwnerID", ri.ID)
		}
	}
}
