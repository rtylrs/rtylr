package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// ptr takes the address of a literal, which the optional filters need.
func ptr[T any](v T) *T { return &v }

func TestListTimeEntries_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	entries, total, err := repo.ListTimeEntries(ctx, testutil.OwnerID, domain.ListTimeEntriesParams{
		FromDate: ptr("2024-01-01"),
		ToDate:   ptr("2025-12-31"),
		Limit:    50,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("ListTimeEntries: %v", err)
	}
	if total < 0 {
		t.Errorf("total count is negative: %d", total)
	}
	_ = entries
}

func TestListTimeEntries_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	entries, _, err := repo.ListTimeEntries(ctx, testutil.OwnerID, domain.ListTimeEntriesParams{
		FromDate: ptr("2024-01-01"),
		ToDate:   ptr("2025-12-31"),
		Limit:    100,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("ListTimeEntries: %v", err)
	}
	for _, e := range entries {
		if e.OwnerID != testutil.OwnerID {
			t.Errorf("entry %d has owner_id %d, want %d", e.ID, e.OwnerID, testutil.OwnerID)
		}
	}
}

func TestListTimeEntries_FilterByEmployee_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _, err := repo.ListTimeEntries(ctx, testutil.OwnerID, domain.ListTimeEntriesParams{
		EmployeeID: ptr[int64](4101),
		FromDate:   ptr("2024-01-01"),
		ToDate:     ptr("2025-12-31"),
		Limit:      50,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("ListTimeEntries with employee filter: %v", err)
	}
}

func TestGetTimeEntry_NotFound_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetTimeEntry(ctx, testutil.OwnerID, 9999999)
	if err == nil {
		t.Fatal("expected error for non-existent time entry, got nil")
	}
}

func TestCreateTimeEntry_AndGet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.TimeEntryRecord{
		EmployeeID:      4101,
		Date:            "2025-06-01",
		DurationMinutes: 480,
		IsBillable:      false,
		Status:          "draft",
		Source:          "manual",
	}

	created, err := repo.CreateTimeEntry(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("CreateTimeEntry: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("created time entry has no ID")
	}

	got, err := repo.GetTimeEntry(ctx, testutil.OwnerID, created.ID)
	if err != nil {
		t.Fatalf("GetTimeEntry after create: %v", err)
	}
	if got.DurationMinutes != 480 {
		t.Errorf("duration_minutes = %d, want 480", got.DurationMinutes)
	}
}

func TestDeleteTimeEntry_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.TimeEntryRecord{
		EmployeeID:      4101,
		Date:            "2025-06-02",
		DurationMinutes: 120,
		Status:          "draft",
		Source:          "manual",
	}
	created, err := repo.CreateTimeEntry(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("CreateTimeEntry for delete test: %v", err)
	}

	if err := repo.DeleteTimeEntry(ctx, testutil.OwnerID, created.ID); err != nil {
		t.Fatalf("DeleteTimeEntry: %v", err)
	}
}

func TestListTimesheets_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	sheets, total, err := repo.ListTimesheets(ctx, testutil.OwnerID, domain.ListTimesheetsParams{
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListTimesheets: %v", err)
	}
	if total < 0 {
		t.Errorf("total is negative: %d", total)
	}
	_ = sheets
}

func TestUpsertTimesheet_AndGet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ts := domain.TimesheetRecord{
		EmployeeID:           4101,
		PeriodStart:          "2025-05-01",
		PeriodEnd:            "2025-05-31",
		Status:               "draft",
		CurrentApprovalLevel: 0,
		TotalMinutes:         0,
	}

	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_timesheets WHERE owner_id = ? AND employee_id = ? AND period_start = ?", testutil.OwnerID, int64(4101), "2025-05-01")
	})

	created, err := repo.UpsertTimesheet(ctx, testutil.OwnerID, ts)
	if err != nil {
		t.Fatalf("UpsertTimesheet: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("created timesheet has no ID")
	}

	got, err := repo.GetTimesheet(ctx, testutil.OwnerID, created.ID)
	if err != nil {
		t.Fatalf("GetTimesheet after upsert: %v", err)
	}
	if got.EmployeeID != 4101 {
		t.Errorf("timesheet employee_id = %d, want 4101", got.EmployeeID)
	}
}

func TestGetTimesheetForPeriod_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// May return nil if no timesheet exists for this period -- that is fine.
	ts, err := repo.GetTimesheetForPeriod(ctx, testutil.OwnerID, 4101, "2025-03-01")
	if err != nil {
		t.Fatalf("GetTimesheetForPeriod: %v", err)
	}
	_ = ts
}

func TestListUnsubmittedEmployees_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ids, err := repo.ListUnsubmittedEmployees(ctx, testutil.OwnerID, "2025-05-01", "2025-05-31")
	if err != nil {
		t.Fatalf("ListUnsubmittedEmployees: %v", err)
	}
	_ = ids
}

// hr_time_entries is the table #871 calls out as both cross-tenant-exposable and
// an input to payroll, and it is where Flow's work-day lines are stored.
//
// Every test above passes a single tenant id in and reads it back with the same
// id, so they hold whether or not business_id scopes anything:
// TestListTimeEntries_AllBelongToOwner asserts a property of the rows returned,
// which is satisfied by returning the caller's rows and also by returning
// nothing. This one writes under two tenants and asserts the second is absent.
//
// The id passed to these methods lands in business_id: the parameter is named
// businessID on ListTimeEntries and CreateTimeEntry. The tests above pass
// testutil.OwnerID into it, which is the owner/business conflation #871
// documents; this test passes business ids so the two tenants are genuinely
// distinct whichever convention a reader expects.
func TestListTimeEntries_DoesNotReturnAnotherTenantsEntries(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Same employee id for both rows on purpose. What separates them is the
	// tenant column, which is the only thing under test here.
	entry := func(date string) domain.TimeEntryRecord {
		return domain.TimeEntryRecord{
			EmployeeID:      4101,
			Date:            date,
			DurationMinutes: 60,
			Status:          "draft",
			Source:          "manual",
		}
	}

	mine, err := repo.CreateTimeEntry(ctx, testutil.BusinessID, entry("2025-07-15"))
	if err != nil {
		t.Fatalf("CreateTimeEntry for own tenant: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTimeEntry(context.Background(), testutil.BusinessID, mine.ID) })

	theirs, err := repo.CreateTimeEntry(ctx, testutil.OtherBusinessID, entry("2025-07-15"))
	if err != nil {
		t.Fatalf("CreateTimeEntry for the other tenant: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTimeEntry(context.Background(), testutil.OtherBusinessID, theirs.ID) })

	// EmployeeID is mandatory, not optional: ListTimeEntries filters
	// `employee_id = ?` unconditionally, so omitting it filters employee_id = 0
	// and returns nothing. That is why TestListTimeEntries_AllBelongToOwner above
	// passes without ever entering its loop.
	entries, _, err := repo.ListTimeEntries(ctx, testutil.BusinessID, domain.ListTimeEntriesParams{
		EmployeeID: ptr[int64](4101),
		FromDate:   ptr("2025-07-01"),
		ToDate:     ptr("2025-07-31"),
		Limit:      100,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("ListTimeEntries: %v", err)
	}

	got := make([]int64, 0, len(entries))
	for _, e := range entries {
		got = append(got, e.ID)
	}

	testutil.AssertTenantIsolation(t, got, []int64{mine.ID}, []int64{theirs.ID})
}
