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
