package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_core/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateWorkSchedule_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ws, err := repo.CreateWorkSchedule(ctx, testutil.OwnerID, domain.WorkScheduleRecord{
		Name:         "Name Stored Schedule",
		ScheduleData: `{}`,
	})
	if err != nil {
		t.Fatalf("CreateWorkSchedule: %v", err)
	}
	if ws.Name != "Name Stored Schedule" {
		t.Errorf("Name = %q, want 'Name Stored Schedule'", ws.Name)
	}
}

func TestCreateWorkSchedule_OwnerIDSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ws, err := repo.CreateWorkSchedule(ctx, testutil.OwnerID, domain.WorkScheduleRecord{
		Name:         "OwnerID Schedule",
		ScheduleData: `{}`,
	})
	if err != nil {
		t.Fatalf("CreateWorkSchedule: %v", err)
	}
	if ws.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", ws.OwnerID, testutil.OwnerID)
	}
}

func TestListWorkSchedules_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	schedules, err := repo.ListWorkSchedules(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListWorkSchedules: %v", err)
	}
	_ = schedules
}

func TestListWorkSchedules_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	schedules, err := repo.ListWorkSchedules(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListWorkSchedules: %v", err)
	}
	for _, ws := range schedules {
		if ws.OwnerID != testutil.OwnerID {
			t.Errorf("schedule %d has OwnerID=%d, want %d", ws.ID, ws.OwnerID, testutil.OwnerID)
		}
	}
}

func TestListWorkSchedules_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.ListWorkSchedules(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListWorkSchedules before: %v", err)
	}

	_, err = repo.CreateWorkSchedule(ctx, testutil.OwnerID, domain.WorkScheduleRecord{
		Name:         "Count Increase Schedule",
		ScheduleData: `{}`,
	})
	if err != nil {
		t.Fatalf("CreateWorkSchedule: %v", err)
	}

	after, err := repo.ListWorkSchedules(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListWorkSchedules after: %v", err)
	}
	if len(after) <= len(before) {
		t.Errorf("expected count to increase from %d, got %d", len(before), len(after))
	}
}

func TestUpdateWorkSchedule_NameChanges(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateWorkSchedule(ctx, testutil.OwnerID, domain.WorkScheduleRecord{
		Name:         "Before Schedule",
		ScheduleData: `{}`,
	})
	if err != nil {
		t.Fatalf("CreateWorkSchedule: %v", err)
	}

	created.Name = "After Schedule"
	updated, err := repo.UpdateWorkSchedule(ctx, testutil.OwnerID, *created)
	if err != nil {
		t.Fatalf("UpdateWorkSchedule: %v", err)
	}
	if updated.Name != "After Schedule" {
		t.Errorf("Name = %q, want 'After Schedule'", updated.Name)
	}
}

func TestGetWorkScheduleForEmployee_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ws, err := repo.GetWorkScheduleForEmployee(ctx, testutil.OwnerID, 4101)
	if err != nil {
		t.Fatalf("GetWorkScheduleForEmployee: %v", err)
	}
	// may be nil if not assigned
	_ = ws
}
