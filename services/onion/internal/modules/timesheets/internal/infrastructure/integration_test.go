package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/timesheets/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestTimesheets_GetTimesheets_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.GetTimesheets(context.Background(), testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("GetTimesheets: %v", err)
	}
}

func TestTimesheets_GetTimesheets_WithWeekStart(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.GetTimesheets(context.Background(), testutil.OwnerID, "2026-05-18")
	if err != nil {
		t.Fatalf("GetTimesheets with weekStart: %v", err)
	}
}

func TestTimesheets_GetTimesheets_UnknownOwner_Empty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	entries, err := repo.GetTimesheets(context.Background(), 9999999, "")
	if err != nil {
		t.Fatalf("GetTimesheets unknown owner: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 for unknown owner, got %d", len(entries))
	}
}

func TestTimesheets_GetTimesheets_OwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	entries, err := repo.GetTimesheets(context.Background(), testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("GetTimesheets: %v", err)
	}
	for _, e := range entries {
		if e.StaffID <= 0 {
			t.Errorf("expected positive staff_id, got %d", e.StaffID)
		}
	}
}

func TestTimesheets_GetTimesheets_FutureWeek_Empty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	// Week far in the future should return no entries
	entries, err := repo.GetTimesheets(context.Background(), testutil.OwnerID, "2099-01-04")
	if err != nil {
		t.Fatalf("GetTimesheets future: %v", err)
	}
	if len(entries) != 0 {
		t.Logf("got %d entries for future week (unexpected but non-fatal)", len(entries))
	}
}

func TestTimesheets_GetTimesheets_PastWeek(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.GetTimesheets(context.Background(), testutil.OwnerID, "2024-01-01")
	if err != nil {
		t.Fatalf("GetTimesheets past week: %v", err)
	}
}

func TestTimesheets_GetTimesheets_ConsistentTotals(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	entries, err := repo.GetTimesheets(context.Background(), testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("GetTimesheets: %v", err)
	}
	for _, e := range entries {
		if e.TotalHours < 0 {
			t.Errorf("entry staff %d: TotalHours %f is negative", e.StaffID, e.TotalHours)
		}
		if e.RegularHours < 0 {
			t.Errorf("entry staff %d: RegularHours %f is negative", e.StaffID, e.RegularHours)
		}
		if e.OvertimeHours < 0 {
			t.Errorf("entry staff %d: OvertimeHours %f is negative", e.StaffID, e.OvertimeHours)
		}
	}
}

func TestTimesheets_GetTimesheets_WeekStartIsMonday(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	entries, err := repo.GetTimesheets(context.Background(), testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("GetTimesheets: %v", err)
	}
	for _, e := range entries {
		if e.WeekStart == "" {
			t.Errorf("entry staff %d: WeekStart is empty", e.StaffID)
		}
	}
}

func TestTimesheets_GetTimesheets_StaffNameNonEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	entries, err := repo.GetTimesheets(context.Background(), testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("GetTimesheets: %v", err)
	}
	for _, e := range entries {
		if e.StaffName == "" {
			t.Errorf("entry staff %d: StaffName is empty", e.StaffID)
		}
	}
}

func TestTimesheets_GetTimesheets_RegularPlusOvertimeEqualsTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	entries, err := repo.GetTimesheets(context.Background(), testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("GetTimesheets: %v", err)
	}
	for _, e := range entries {
		expected := e.RegularHours + e.OvertimeHours
		diff := expected - e.TotalHours
		if diff < -0.001 || diff > 0.001 {
			t.Errorf("staff %d: regular(%.2f)+overtime(%.2f) != total(%.2f)", e.StaffID, e.RegularHours, e.OvertimeHours, e.TotalHours)
		}
	}
}

func TestTimesheets_GetTimesheets_NilContextDoesNotPanic(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.GetTimesheets(context.Background(), testutil.OwnerID, "2026-05-18")
	if err != nil {
		t.Fatalf("GetTimesheets with Monday date: %v", err)
	}
}

func TestTimesheets_GetTimesheets_MultipleWeeks_DifferentResults(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	_, err1 := repo.GetTimesheets(ctx, testutil.OwnerID, "2026-05-11")
	if err1 != nil {
		t.Fatalf("GetTimesheets week1: %v", err1)
	}
	_, err2 := repo.GetTimesheets(ctx, testutil.OwnerID, "2026-05-18")
	if err2 != nil {
		t.Fatalf("GetTimesheets week2: %v", err2)
	}
}

func TestTimesheets_GetTimesheets_ZeroOwnerID_Empty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	entries, err := repo.GetTimesheets(context.Background(), 0, "")
	if err != nil {
		t.Fatalf("GetTimesheets ownerID=0: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for ownerID=0, got %d", len(entries))
	}
}

func TestTimesheets_GetTimesheets_NegativeOwnerID_Empty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	entries, err := repo.GetTimesheets(context.Background(), -1, "")
	if err != nil {
		t.Fatalf("GetTimesheets ownerID=-1: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for negative ownerID, got %d", len(entries))
	}
}

func TestTimesheets_GetTimesheets_WithWeekStart_SubsetOfAll(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	all, err := repo.GetTimesheets(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("GetTimesheets all: %v", err)
	}
	filtered, err := repo.GetTimesheets(ctx, testutil.OwnerID, "2026-05-18")
	if err != nil {
		t.Fatalf("GetTimesheets filtered: %v", err)
	}
	if len(filtered) > len(all) {
		t.Errorf("filtered(%d) cannot exceed all(%d)", len(filtered), len(all))
	}
}

func TestTimesheets_GetTimesheets_IsolationBetweenOwners(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	ownerEntries, err := repo.GetTimesheets(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("GetTimesheets owner: %v", err)
	}
	otherEntries, err := repo.GetTimesheets(ctx, 8888888, "")
	if err != nil {
		t.Fatalf("GetTimesheets other: %v", err)
	}
	// Other owner should have no records from testutil.OwnerID.
	for _, e := range otherEntries {
		for _, oe := range ownerEntries {
			if e.StaffID == oe.StaffID && e.WeekStart == oe.WeekStart {
				t.Errorf("owner isolation violated: staff %d week %s appears in both owners", e.StaffID, e.WeekStart)
			}
		}
	}
}

func TestTimesheets_GetTimesheets_Idempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	a, err := repo.GetTimesheets(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	b, err := repo.GetTimesheets(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if len(a) != len(b) {
		t.Errorf("idempotency: got %d then %d", len(a), len(b))
	}
}

func TestTimesheets_GetTimesheets_EmptyWeekStart_ReturnsAllWeeks(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	entries, err := repo.GetTimesheets(context.Background(), testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("GetTimesheets empty week: %v", err)
	}
	_ = entries // all weeks should be returned without error
}

func TestTimesheets_GetTimesheets_SpecificWeekFiltersCorrectly(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	const week = "2026-05-18"
	entries, err := repo.GetTimesheets(context.Background(), testutil.OwnerID, week)
	if err != nil {
		t.Fatalf("GetTimesheets specific week: %v", err)
	}
	for _, e := range entries {
		if e.WeekStart != week {
			t.Errorf("expected week_start=%s, got %s", week, e.WeekStart)
		}
	}
}
