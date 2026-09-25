package infrastructure_test

import (
	"context"
	"errors"
	"testing"
	"time"

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

func seedRecurringInvoice(ctx context.Context, t *testing.T, repo *infrastructure.Repository, ownerID int64, freq string) (int64, error) {
	t.Helper()
	// We need a template_invoice_id -- use 0 and let the DB accept it.
	input := domain.RecurringInvoiceInput{
		TemplateInvoiceID: testutil.Int64Ptr(0),
		Frequency:         freq,
		NextInvoiceDate:   "2026-06-01",
		IsActive:          boolPtr(true),
	}
	return repo.Create(ctx, ownerID, input)
}

func boolPtr(value bool) *bool {
	return &value
}

func stringPtr(value string) *string {
	return &value
}

func findRecurringInvoice(t *testing.T, repo *infrastructure.Repository, ownerID, id int64) domain.RecurringInvoice {
	t.Helper()
	items, err := repo.List(context.Background(), ownerID)
	if err != nil {
		t.Fatalf("recurring_invoices.List: %v", err)
	}
	for _, ri := range items {
		if ri.ID == id {
			return ri
		}
	}
	t.Fatalf("recurring invoice id=%d not found", id)
	return domain.RecurringInvoice{}
}

func TestCreateRecurringInvoice_ReturnsInsertedID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := seedRecurringInvoice(ctx, t, repo, testutil.OwnerID, "monthly")
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

	id, err := seedRecurringInvoice(ctx, t, repo, testutil.OwnerID, "weekly")
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

	id, err := seedRecurringInvoice(ctx, t, repo, testutil.OwnerID, "monthly")
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
		TemplateInvoiceID: testutil.Int64Ptr(0),
		Frequency:         "monthly",
		NextInvoiceDate:   "2026-07-01",
		IsActive:          boolPtr(true),
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

	id, err := seedRecurringInvoice(ctx, t, repo, testutil.OwnerID, "weekly")
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

	id, err := seedRecurringInvoice(ctx, t, repo, testutil.OwnerID, "monthly")
	if err != nil {
		t.Fatalf("recurring_invoices.Create: %v", err)
	}

	err = repo.Update(ctx, testutil.OwnerID, id, domain.RecurringInvoiceInput{
		TemplateInvoiceID: testutil.Int64Ptr(0),
		Frequency:         "yearly",
		NextInvoiceDate:   "2027-01-01",
		IsActive:          boolPtr(true),
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

	_, err := seedRecurringInvoice(ctx, t, repo, testutil.OwnerID, "monthly")
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

	_, err := seedRecurringInvoice(ctx, t, repo, testutil.OwnerID, "monthly")
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

	_, err := seedRecurringInvoice(ctx, t, repo, testutil.OwnerID, "weekly")
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

// Regression: pausing an already-paused schedule (or resuming an active one)
// changes no column value, so MySQL reports zero affected rows. That must not
// be mistaken for a missing row.
func TestSetActive_AlreadyInTargetState_Succeeds(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := seedRecurringInvoice(ctx, t, repo, testutil.OwnerID, "monthly")
	if err != nil {
		t.Fatalf("recurring_invoices.Create: %v", err)
	}

	if err := repo.SetActive(ctx, testutil.OwnerID, id, false); err != nil {
		t.Fatalf("first SetActive(false): %v", err)
	}
	if err := repo.SetActive(ctx, testutil.OwnerID, id, false); err != nil {
		t.Fatalf("second SetActive(false) on an already paused schedule: %v", err)
	}
	if err := repo.SetActive(ctx, testutil.OwnerID, id, true); err != nil {
		t.Fatalf("SetActive(true): %v", err)
	}
	if err := repo.SetActive(ctx, testutil.OwnerID, id, true); err != nil {
		t.Fatalf("second SetActive(true) on an already active schedule: %v", err)
	}
}

func TestSetActive_UnknownID_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.SetActive(ctx, testutil.OwnerID, 999999999, false)
	if !errors.Is(err, domain.ErrRecurringInvoiceNotFound) {
		t.Fatalf("SetActive(unknown id) = %v, want ErrRecurringInvoiceNotFound", err)
	}
}

// Regression: an update that omits the optional is_active/end_date fields must
// leave the stored values untouched; collapsing them to false/"" paused the
// schedule and cleared its bound.
func TestUpdateRecurringInvoice_AbsentOptionalFields_KeepStoredValues(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.RecurringInvoiceInput{
		TemplateInvoiceID: testutil.Int64Ptr(0),
		Frequency:         "monthly",
		NextInvoiceDate:   "2026-06-01",
		EndDate:           stringPtr("2026-12-31"),
		IsActive:          boolPtr(true),
	})
	if err != nil {
		t.Fatalf("recurring_invoices.Create: %v", err)
	}

	if err := repo.Update(ctx, testutil.OwnerID, id, domain.RecurringInvoiceInput{
		TemplateInvoiceID: testutil.Int64Ptr(0),
		Frequency:         "quarterly",
		NextInvoiceDate:   "2026-09-01",
	}); err != nil {
		t.Fatalf("recurring_invoices.Update: %v", err)
	}

	stored := findRecurringInvoice(t, repo, testutil.OwnerID, id)
	if stored.Frequency != "quarterly" {
		t.Errorf("Frequency = %q, want quarterly", stored.Frequency)
	}
	if !stored.IsActive {
		t.Error("IsActive = false after an update that omitted is_active, want true")
	}
	want := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	if !stored.EndDate.Equal(want) {
		t.Errorf("EndDate = %v after an update that omitted end_date, want %v", stored.EndDate, want)
	}
}

// An explicitly empty end_date still clears a previously bounded schedule.
func TestUpdateRecurringInvoice_EmptyEndDate_ClearsBound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.RecurringInvoiceInput{
		TemplateInvoiceID: testutil.Int64Ptr(0),
		Frequency:         "monthly",
		NextInvoiceDate:   "2026-06-01",
		EndDate:           stringPtr("2026-12-31"),
		IsActive:          boolPtr(true),
	})
	if err != nil {
		t.Fatalf("recurring_invoices.Create: %v", err)
	}

	if err := repo.Update(ctx, testutil.OwnerID, id, domain.RecurringInvoiceInput{
		TemplateInvoiceID: testutil.Int64Ptr(0),
		Frequency:         "monthly",
		NextInvoiceDate:   "2026-06-01",
		EndDate:           stringPtr(""),
	}); err != nil {
		t.Fatalf("recurring_invoices.Update with empty end_date: %v", err)
	}

	stored := findRecurringInvoice(t, repo, testutil.OwnerID, id)
	if !stored.EndDate.IsZero() {
		t.Errorf("EndDate = %v after an explicit clear, want the zero time", stored.EndDate)
	}
	if !stored.IsActive {
		t.Error("IsActive = false after an update that omitted is_active, want true")
	}
}

func TestUpdateRecurringInvoice_UnknownID_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.Update(ctx, testutil.OwnerID, 999999999, domain.RecurringInvoiceInput{
		TemplateInvoiceID: testutil.Int64Ptr(0),
		Frequency:         "monthly",
		NextInvoiceDate:   "2026-06-01",
		IsActive:          boolPtr(true),
	})
	if !errors.Is(err, domain.ErrRecurringInvoiceNotFound) {
		t.Fatalf("Update(unknown id) = %v, want ErrRecurringInvoiceNotFound", err)
	}
}

// Regression: a malformed end_date was silently stored as NULL, turning a
// bounded schedule into one that generates invoices forever.
func TestCreateRecurringInvoice_MalformedEndDate_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, testutil.OwnerID, domain.RecurringInvoiceInput{
		TemplateInvoiceID: testutil.Int64Ptr(0),
		Frequency:         "monthly",
		NextInvoiceDate:   "2026-06-01",
		EndDate:           stringPtr("31/12/2026"),
		IsActive:          boolPtr(true),
	})
	if err == nil {
		t.Fatal("Create with malformed end_date: expected an error, got nil")
	}
}

func TestUpdateRecurringInvoice_MalformedEndDate_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.RecurringInvoiceInput{
		TemplateInvoiceID: testutil.Int64Ptr(0),
		Frequency:         "monthly",
		NextInvoiceDate:   "2026-06-01",
		EndDate:           stringPtr("2026-12-31"),
		IsActive:          boolPtr(true),
	})
	if err != nil {
		t.Fatalf("recurring_invoices.Create: %v", err)
	}

	err = repo.Update(ctx, testutil.OwnerID, id, domain.RecurringInvoiceInput{
		TemplateInvoiceID: testutil.Int64Ptr(0),
		Frequency:         "monthly",
		NextInvoiceDate:   "2026-06-01",
		EndDate:           stringPtr("2026-12-31T00:00:00Z"),
	})
	if err == nil {
		t.Fatal("Update with malformed end_date: expected an error, got nil")
	}

	stored := findRecurringInvoice(t, repo, testutil.OwnerID, id)
	want := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	if !stored.EndDate.Equal(want) {
		t.Errorf("EndDate = %v after a rejected update, want the stored %v", stored.EndDate, want)
	}
}
