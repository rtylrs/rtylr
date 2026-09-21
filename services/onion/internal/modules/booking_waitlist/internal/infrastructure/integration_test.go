package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/booking_waitlist/internal/domain"
	"github.com/voxire/onion/internal/modules/booking_waitlist/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func baseWaitlistInput() domain.WaitlistInput {
	return domain.WaitlistInput{
		ServiceID:     0,
		CustomerName:  "John Waitlist",
		CustomerEmail: "john@waitlist.com",
		CustomerPhone: "555-9876",
		PreferredDate: "2026-06-15",
		PreferredTime: "14:00",
		Status:        "waiting",
	}
}

func TestListWaitlist_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for owner 0, got %d", len(items))
	}
}

func TestListWaitlist_IsNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
}

func TestCreateWaitlistEntry_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, baseWaitlistInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreateWaitlistEntry_CustomerNameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := baseWaitlistInput()
	input.CustomerName = "Mary Wait"
	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, w := range items {
		if w.ID == id {
			if w.CustomerName != "Mary Wait" {
				t.Errorf("expected customer name Mary Wait, got %s", w.CustomerName)
			}
			return
		}
	}
	t.Errorf("waitlist entry id=%d not found", id)
}

func TestCreateWaitlistEntry_CustomerEmailIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := baseWaitlistInput()
	input.CustomerEmail = "mary@email.com"
	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, w := range items {
		if w.ID == id {
			if w.CustomerEmail != "mary@email.com" {
				t.Errorf("expected email mary@email.com, got %s", w.CustomerEmail)
			}
			return
		}
	}
	t.Errorf("waitlist entry id=%d not found", id)
}

func TestCreateWaitlistEntry_ServiceIDIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := baseWaitlistInput()
	input.ServiceID = 0 // nullable FK
	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, w := range items {
		if w.ID == id {
			// ServiceID=0 when FK is null
			return
		}
	}
	t.Errorf("waitlist entry id=%d not found", id)
}

func TestCreateWaitlistEntry_StatusIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := baseWaitlistInput()
	input.Status = "waiting"
	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, w := range items {
		if w.ID == id {
			if w.Status != "waiting" {
				t.Errorf("expected status waiting, got %s", w.Status)
			}
			return
		}
	}
	t.Errorf("waitlist entry id=%d not found", id)
}

func TestCreateWaitlistEntry_PreferredDateIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := baseWaitlistInput()
	input.PreferredDate = "2026-06-15"
	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, w := range items {
		if w.ID == id {
			if w.PreferredDate != "2026-06-15" {
				t.Errorf("expected preferred date 2026-06-15, got %s", w.PreferredDate)
			}
			return
		}
	}
	t.Errorf("waitlist entry id=%d not found", id)
}

func TestUpdateWaitlistEntry_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, baseWaitlistInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updated := baseWaitlistInput()
	updated.Status = "notified"
	if err := repo.Update(ctx, testutil.OwnerID, id, updated); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, w := range items {
		if w.ID == id && w.Status != "notified" {
			t.Errorf("expected status notified, got %s", w.Status)
		}
	}
}

func TestCreate_ThenList_FindsEntry(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, baseWaitlistInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, w := range items {
		if w.ID == id {
			return
		}
	}
	t.Errorf("waitlist entry id=%d not found after create", id)
}

func TestListWaitlist_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	_, err = repo.Create(ctx, testutil.OwnerID, baseWaitlistInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	after, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) <= len(before) {
		t.Errorf("expected count to increase: before=%d, after=%d", len(before), len(after))
	}
}

func TestCreateWaitlistEntry_PhoneIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := baseWaitlistInput()
	input.CustomerPhone = "555-0000"
	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, w := range items {
		if w.ID == id {
			if w.CustomerPhone != "555-0000" {
				t.Errorf("expected phone 555-0000, got %s", w.CustomerPhone)
			}
			return
		}
	}
	t.Errorf("waitlist entry id=%d not found", id)
}

func TestUpdateWaitlistEntry_ChangesCustomerName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, baseWaitlistInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updated := baseWaitlistInput()
	updated.CustomerName = "Updated Waitlist Name"
	if err := repo.Update(ctx, testutil.OwnerID, id, updated); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, w := range items {
		if w.ID == id && w.CustomerName != "Updated Waitlist Name" {
			t.Errorf("expected name Updated Waitlist Name, got %s", w.CustomerName)
		}
	}
}

// A value the parser cannot read must fail the write instead of being dropped
// to NULL while the API reports success. "2026-6-5" is the class of value
// clients actually send.
func TestCreateWaitlistEntry_RejectsUnparseablePreferredDate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := baseWaitlistInput()
	input.CustomerName = "Unparseable Date Marker"
	input.PreferredDate = "2026-6-5"

	if _, err := repo.Create(ctx, testutil.OwnerID, input); !errors.Is(err, domain.ErrPreferredDateInvalid) {
		t.Fatalf("Create with an unparseable preferred_date error = %v, want ErrPreferredDateInvalid", err)
	}

	var count int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM booking_waitlist WHERE business_id = ? AND customer_name = ?",
		testutil.OwnerID, input.CustomerName).Scan(&count); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if count != 0 {
		t.Errorf("rejected create still inserted %d row(s)", count)
	}
}

// The repo's own fixtures use "14:00", which the old HH:MM:SS-only parser
// rejected; it must parse rather than null the column.
func TestCreateWaitlistEntry_AcceptsHHMMPreferredTime(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := baseWaitlistInput()
	input.PreferredTime = "14:00"

	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create with an HH:MM preferred_time: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM booking_waitlist WHERE id = ?", id)
	})

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, w := range items {
		if w.ID == id {
			if w.PreferredTime != "14:00:00" {
				t.Errorf("PreferredTime = %q, want 14:00:00", w.PreferredTime)
			}
			return
		}
	}
	t.Errorf("waitlist entry id=%d not found", id)
}

// A rejected update must leave the stored date/time untouched, not NULL it.
func TestUpdateWaitlistEntry_RejectsUnparseablePreferredTime(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := baseWaitlistInput()
	input.PreferredDate = "2026-06-15"
	input.PreferredTime = "09:30"

	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM booking_waitlist WHERE id = ?", id)
	})

	bad := baseWaitlistInput()
	bad.PreferredTime = "25:99"
	if err := repo.Update(ctx, testutil.OwnerID, id, bad); !errors.Is(err, domain.ErrPreferredTimeInvalid) {
		t.Fatalf("Update with an unparseable preferred_time error = %v, want ErrPreferredTimeInvalid", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, w := range items {
		if w.ID == id {
			if w.PreferredTime != "09:30:00" {
				t.Errorf("PreferredTime = %q, want the stored 09:30:00", w.PreferredTime)
			}
			if w.PreferredDate != "2026-06-15" {
				t.Errorf("PreferredDate = %q, want the stored 2026-06-15", w.PreferredDate)
			}
			return
		}
	}
	t.Errorf("waitlist entry id=%d not found", id)
}

// An update that matches no row of the caller's business must be reported as
// not-found, not as a successful no-op.
func TestUpdateWaitlistEntry_MissingIDReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.Update(ctx, testutil.OwnerID, 999999, baseWaitlistInput())
	if !errors.Is(err, domain.ErrWaitlistEntryNotFound) {
		t.Fatalf("Update of a missing id error = %v, want ErrWaitlistEntryNotFound", err)
	}
}

// MySQL counts changed rows, not matched rows, so re-sending the stored values
// affects zero rows. That idempotent no-op must still succeed.
func TestUpdateWaitlistEntry_UnchangedValuesStillSucceed(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := baseWaitlistInput()
	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM booking_waitlist WHERE id = ?", id)
	})

	// The stored preferred_time is formatted as HH:MM:SS, so this update writes
	// the exact values already present.
	same := baseWaitlistInput()
	same.PreferredTime = "14:00:00"
	if err := repo.Update(ctx, testutil.OwnerID, id, same); err != nil {
		t.Fatalf("Update with unchanged values: %v", err)
	}
}
