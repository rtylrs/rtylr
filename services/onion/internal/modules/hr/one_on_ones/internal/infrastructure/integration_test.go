package infrastructure_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/one_on_ones/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/one_on_ones/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListOneOnOnes_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
}

func TestListOneOnOnes_UnknownOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected empty list for owner_id=0, got %d items", len(items))
	}
}

func TestListOneOnOnes_EachHasNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for i, item := range items {
		if item.ID == 0 {
			t.Errorf("one_on_one[%d] has zero ID", i)
		}
	}
}

func TestListOneOnOnes_EachOwnerIDMatchesFilter(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for i, item := range items {
		if item.OwnerID != testutil.OwnerID {
			t.Errorf("one_on_one[%d] has OwnerID=%d, want %d", i, item.OwnerID, testutil.OwnerID)
		}
	}
}

func TestListOneOnOnes_EachHasNonZeroManagerID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for i, item := range items {
		if item.ManagerID == 0 {
			t.Errorf("one_on_one[%d] (id=%d) has zero ManagerID", i, item.ID)
		}
	}
}

func TestCreateOneOnOne_InsertsRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	input := domain.OneOnOneInput{
		ManagerID:   4101,
		StaffID:     4102,
		ScheduledAt: "2026-06-01T10:00:00Z",
		Notes:       "integration test meeting",
		ActionItems: `["follow up"]`,
		Status:      "scheduled",
	}
	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero inserted ID")
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_one_on_ones WHERE id = ?", id)
	})

	after, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Errorf("expected count to grow by 1: before=%d, after=%d", len(before), len(after))
	}
}

func TestCreateOneOnOne_DefaultsStatusToScheduled(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.OneOnOneInput{
		ManagerID:   4101,
		StaffID:     4102,
		ScheduledAt: "2026-06-02T10:00:00Z",
		// Status intentionally empty
	}
	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_one_on_ones WHERE id = ?", id)
	})

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for _, item := range items {
		if item.ID == id {
			if item.Status != "scheduled" {
				t.Errorf("expected status=scheduled, got %q", item.Status)
			}
			return
		}
	}
	t.Errorf("inserted record id=%d not found in List results", id)
}

func TestListOneOnOnes_TwiceReturnsSameCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("first List: %v", err)
	}
	second, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("second List: %v", err)
	}
	if len(first) != len(second) {
		t.Errorf("count changed between calls: %d vs %d", len(first), len(second))
	}
}

func TestListOneOnOnes_StaffIDNonZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for i, item := range items {
		if item.StaffID == 0 {
			t.Errorf("one_on_one[%d] (id=%d) has zero StaffID", i, item.ID)
		}
	}
}

func TestCreateOneOnOne_NotesPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wantNotes := "check in about project alpha"
	input := domain.OneOnOneInput{
		ManagerID:   4101,
		StaffID:     4102,
		ScheduledAt: "2026-06-03T10:00:00Z",
		Notes:       wantNotes,
		Status:      "scheduled",
	}
	id, err := repo.Create(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_one_on_ones WHERE id = ?", id)
	})

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for _, item := range items {
		if item.ID == id {
			if item.Notes != wantNotes {
				t.Errorf("Notes mismatch: got %q, want %q", item.Notes, wantNotes)
			}
			return
		}
	}
	t.Errorf("inserted record id=%d not found in List results", id)
}

func TestListOneOnOnes_ScheduledAtIsNonZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for i, item := range items {
		if item.ScheduledAt.IsZero() {
			t.Errorf("one_on_one[%d] (id=%d) has zero ScheduledAt", i, item.ID)
		}
	}
}

func TestCreateOneOnOne_PositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.OneOnOneInput{
		ManagerID:   4101,
		StaffID:     4102,
		ScheduledAt: "2026-07-01T09:00:00Z",
		Status:      "scheduled",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_one_on_ones WHERE id = ?", id) })
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreateOneOnOne_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.OneOnOneInput{
		ManagerID:   4101,
		StaffID:     4102,
		ScheduledAt: "2026-07-02T09:00:00Z",
		Status:      "scheduled",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_one_on_ones WHERE id = ?", id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for _, item := range items {
		if item.ID == id {
			return
		}
	}
	t.Errorf("created record id=%d not found in List", id)
}

func TestCreateOneOnOne_ManagerIDStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const want = int64(4201)
	id, err := repo.Create(ctx, testutil.OwnerID, domain.OneOnOneInput{
		ManagerID:   want,
		StaffID:     4202,
		ScheduledAt: "2026-07-03T09:00:00Z",
		Status:      "scheduled",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_one_on_ones WHERE id = ?", id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for _, item := range items {
		if item.ID == id {
			if item.ManagerID != want {
				t.Errorf("ManagerID = %d, want %d", item.ManagerID, want)
			}
			return
		}
	}
	t.Errorf("created record id=%d not found", id)
}

func TestCreateOneOnOne_StaffIDStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const want = int64(4302)
	id, err := repo.Create(ctx, testutil.OwnerID, domain.OneOnOneInput{
		ManagerID:   4301,
		StaffID:     want,
		ScheduledAt: "2026-07-04T09:00:00Z",
		Status:      "scheduled",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_one_on_ones WHERE id = ?", id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for _, item := range items {
		if item.ID == id {
			if item.StaffID != want {
				t.Errorf("StaffID = %d, want %d", item.StaffID, want)
			}
			return
		}
	}
	t.Errorf("created record id=%d not found", id)
}

func TestCreateOneOnOne_ActionItemsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const want = `["review Q2 goals","update roadmap"]`
	id, err := repo.Create(ctx, testutil.OwnerID, domain.OneOnOneInput{
		ManagerID:   4101,
		StaffID:     4102,
		ScheduledAt: "2026-07-05T09:00:00Z",
		ActionItems: want,
		Status:      "scheduled",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_one_on_ones WHERE id = ?", id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for _, item := range items {
		if item.ID == id {
			var gotSlice, wantSlice []string
			if err := json.Unmarshal([]byte(item.ActionItems), &gotSlice); err != nil {
				t.Errorf("ActionItems is not valid JSON: %q", item.ActionItems)
				return
			}
			if err := json.Unmarshal([]byte(want), &wantSlice); err != nil {
				t.Fatalf("want is not valid JSON: %q", want)
			}
			if !reflect.DeepEqual(gotSlice, wantSlice) {
				t.Errorf("ActionItems = %q, want %q", item.ActionItems, want)
			}
			return
		}
	}
	t.Errorf("created record id=%d not found", id)
}

func TestListOneOnOnes_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	id, err := repo.Create(ctx, testutil.OwnerID, domain.OneOnOneInput{
		ManagerID:   4101,
		StaffID:     4102,
		ScheduledAt: "2026-07-06T09:00:00Z",
		Status:      "scheduled",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_one_on_ones WHERE id = ?", id) })

	after, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Errorf("expected count +1: before=%d after=%d", len(before), len(after))
	}
}

func TestCreateOneOnOne_IDsAreUnique(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.Create(ctx, testutil.OwnerID, domain.OneOnOneInput{
		ManagerID:   4101,
		StaffID:     4102,
		ScheduledAt: "2026-07-07T09:00:00Z",
		Status:      "scheduled",
	})
	if err != nil {
		t.Fatalf("Create A: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_one_on_ones WHERE id = ?", id1) })

	id2, err := repo.Create(ctx, testutil.OwnerID, domain.OneOnOneInput{
		ManagerID:   4101,
		StaffID:     4102,
		ScheduledAt: "2026-07-08T09:00:00Z",
		Status:      "scheduled",
	})
	if err != nil {
		t.Fatalf("Create B: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_one_on_ones WHERE id = ?", id2) })

	if id1 == id2 {
		t.Errorf("expected unique IDs, both got %d", id1)
	}
}

func TestListOneOnOnes_IsNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestCreateOneOnOne_StatusCompleted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.OneOnOneInput{
		ManagerID:   4101,
		StaffID:     4102,
		ScheduledAt: "2026-05-01T10:00:00Z",
		Status:      "completed",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_one_on_ones WHERE id = ?", id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for _, item := range items {
		if item.ID == id {
			if item.Status != "completed" {
				t.Errorf("Status = %q, want completed", item.Status)
			}
			return
		}
	}
	t.Errorf("created record id=%d not found", id)
}

func TestCreateOneOnOne_OwnerScopedInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.OneOnOneInput{
		ManagerID:   4101,
		StaffID:     4102,
		ScheduledAt: "2026-07-10T09:00:00Z",
		Status:      "scheduled",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_one_on_ones WHERE id = ?", id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for _, item := range items {
		if item.OwnerID != testutil.OwnerID {
			t.Errorf("item id=%d has OwnerID=%d, want %d", item.ID, item.OwnerID, testutil.OwnerID)
		}
	}
}

func TestCreateOneOnOne_ZeroOwnerNotInOwnerList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for owner_id=0, got %d", len(items))
	}
}

func TestCreateOneOnOne_EmptyNotesIsOk(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.OneOnOneInput{
		ManagerID:   4101,
		StaffID:     4102,
		ScheduledAt: "2026-07-11T09:00:00Z",
		Notes:       "",
		Status:      "scheduled",
	})
	if err != nil {
		t.Fatalf("Create with empty notes: %v", err)
	}
	t.Cleanup(func() { db.ExecContext(ctx, "DELETE FROM hr_one_on_ones WHERE id = ?", id) })
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}
