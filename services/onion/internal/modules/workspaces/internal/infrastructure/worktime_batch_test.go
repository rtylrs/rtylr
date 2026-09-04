package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestWorkspaces_SumWorkDayLineMinutesByDayIDs_ReturnsEachDayTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const (
		employeeA   int64 = 991101
		employeeB   int64 = 991102
		dateA             = "2026-08-20"
		dateB             = "2026-08-21"
		clientKeyA1       = "batch-workday-a-1"
		clientKeyA2       = "batch-workday-a-2"
		clientKeyB1       = "batch-workday-b-1"
	)

	t.Cleanup(func() {
		if _, err := db.ExecContext(ctx,
			"DELETE FROM hr_time_entries WHERE owner_id = ? AND business_id = ? AND client_key IN (?, ?, ?)",
			testutil.OwnerID, testutil.BusinessID, clientKeyA1, clientKeyA2, clientKeyB1,
		); err != nil {
			t.Errorf("delete test time entries: %v", err)
		}
		if _, err := db.ExecContext(ctx,
			"DELETE FROM flow_work_days WHERE owner_id = ? AND business_id = ? AND employee_id IN (?, ?)",
			testutil.OwnerID, testutil.BusinessID, employeeA, employeeB,
		); err != nil {
			t.Errorf("delete test work days: %v", err)
		}
	})

	dayA, err := repo.UpsertWorkDay(ctx, domain.WorkDayInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, EmployeeID: employeeA,
		Date: dateA, Timezone: "UTC", WorkMode: "onsite", LoggedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("UpsertWorkDay day A: %v", err)
	}
	dayB, err := repo.UpsertWorkDay(ctx, domain.WorkDayInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, EmployeeID: employeeB,
		Date: dateB, Timezone: "UTC", WorkMode: "onsite", LoggedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("UpsertWorkDay day B: %v", err)
	}

	for _, line := range []domain.WorkDayLineInput{
		{OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, EmployeeID: employeeA, WorkDayID: dayA.ID, Date: dateA, DurationMinutes: 45, ClientKey: clientKeyA1, LoggedBy: testutil.OwnerID},
		{OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, EmployeeID: employeeA, WorkDayID: dayA.ID, Date: dateA, DurationMinutes: 30, ClientKey: clientKeyA2, LoggedBy: testutil.OwnerID},
		{OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, EmployeeID: employeeB, WorkDayID: dayB.ID, Date: dateB, DurationMinutes: 90, ClientKey: clientKeyB1, LoggedBy: testutil.OwnerID},
	} {
		if _, err := repo.CreateWorkDayLine(ctx, line); err != nil {
			t.Fatalf("CreateWorkDayLine %s: %v", line.ClientKey, err)
		}
	}

	totals, err := repo.SumWorkDayLineMinutesByDayIDs(ctx, testutil.OwnerID, testutil.BusinessID, []int64{dayA.ID, dayB.ID})
	if err != nil {
		t.Fatalf("SumWorkDayLineMinutesByDayIDs: %v", err)
	}
	if len(totals) != 2 {
		t.Fatalf("totals = %+v, want one result per populated day", totals)
	}
	if totals[dayA.ID] != 75 {
		t.Errorf("day A total = %d, want 75", totals[dayA.ID])
	}
	if totals[dayB.ID] != 90 {
		t.Errorf("day B total = %d, want 90", totals[dayB.ID])
	}
}
