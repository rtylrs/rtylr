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

	schedules, err := repo.ListWorkSchedules(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListWorkSchedules: %v", err)
	}
	_ = schedules
}

func TestListWorkSchedules_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	schedules, err := repo.ListWorkSchedules(ctx, testutil.OwnerID, 100, 0)
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

	before, err := repo.ListWorkSchedules(ctx, testutil.OwnerID, 100, 0)
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

	after, err := repo.ListWorkSchedules(ctx, testutil.OwnerID, 100, 0)
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

// hr_work_schedules sits in a schema where business_id is BIGINT NOT NULL
// DEFAULT 0, so a query that omits the tenant filter returns tenant 0 rather
// than failing. #868 calls this test mandatory for that reason.
//
// The tests above pass one tenant id and assert a property of the rows returned,
// which holds whether or not business_id scopes anything. This one seeds two
// tenants and asserts the second is absent, and asserts the caller's own row is
// present so the absence is not vacuous.
func TestListWorkSchedules_DoesNotReturnAnotherTenantsSchedules(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	mine, err := repo.CreateWorkSchedule(ctx, testutil.BusinessID, domain.WorkScheduleRecord{
		Name:         "Iso Own Schedule",
		ScheduleData: `{"contracted_hours_per_week":32}`,
	})
	if err != nil {
		t.Fatalf("CreateWorkSchedule for own tenant: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_work_schedules WHERE id = ?", mine.ID)
	})

	theirs, err := repo.CreateWorkSchedule(ctx, testutil.OtherBusinessID, domain.WorkScheduleRecord{
		Name:         "Iso Other Schedule",
		ScheduleData: `{"contracted_hours_per_week":40}`,
	})
	if err != nil {
		t.Fatalf("CreateWorkSchedule for the other tenant: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_work_schedules WHERE id = ?", theirs.ID)
	})

	schedules, err := repo.ListWorkSchedules(ctx, testutil.BusinessID, 100, 0)
	if err != nil {
		t.Fatalf("ListWorkSchedules: %v", err)
	}

	var sawMine, sawTheirs bool
	for _, ws := range schedules {
		switch ws.ID {
		case mine.ID:
			sawMine = true
		case theirs.ID:
			sawTheirs = true
		}
	}
	if !sawMine {
		t.Errorf("own schedule %d missing, so the absence below proves nothing", mine.ID)
	}
	if sawTheirs {
		t.Errorf("cross-tenant leak: schedule %d belongs to business %d", theirs.ID, testutil.OtherBusinessID)
	}
}

// The count backs the list response's total, so it needs the same filter. A
// total that counts every tenant's rows would tell a paginating client to keep
// asking for pages that are not theirs.
func TestCountWorkSchedules_IsTenantScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.CountWorkSchedules(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("CountWorkSchedules before: %v", err)
	}

	theirs, err := repo.CreateWorkSchedule(ctx, testutil.OtherBusinessID, domain.WorkScheduleRecord{
		Name: "Count Other Schedule",
	})
	if err != nil {
		t.Fatalf("CreateWorkSchedule for the other tenant: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_work_schedules WHERE id = ?", theirs.ID)
	})

	after, err := repo.CountWorkSchedules(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("CountWorkSchedules after: %v", err)
	}
	if after != before {
		t.Errorf("count changed from %d to %d after another tenant added a schedule", before, after)
	}
}
