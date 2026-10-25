package infrastructure_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

func TestActualHoursCombinesFlowAndHRTimeThroughTheOwnerPort(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := testsupport.NewRepository(db)
	store := newTestHRTimeStore(ctx, db)
	repo.SetWorkDayLineStore(store)
	taskID, businessID := seedArchivedTestTask(t, db, repo, "ActualHoursOwnerBoundary")
	clientKey := fmt.Sprintf("actual-hours-%d", time.Now().UnixNano())

	entry, err := store.CreateWorkDayLine(ctx, domain.WorkDayLineInput{
		BusinessID: businessID, EmployeeID: testutil.OwnerID,
		WorkDayID: taskID, Date: "2035-01-01", ProjectID: taskID, TaskID: taskID,
		DurationMinutes: 30, ClientKey: clientKey, LoggedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("create HR-owned work-day entry: %v", err)
	}
	t.Cleanup(func() {
		if _, err := store.DeleteWorkDayLine(ctx, entry.ID, businessID); err != nil {
			t.Errorf("delete HR-owned entry fixture: %v", err)
		}
	})

	logID, err := repo.CreateTimeLog(ctx, domain.TimeLogInput{
		OwnerID: testutil.OwnerID, BusinessID: businessID, TaskID: taskID,
		UserID: testutil.OwnerID, DurationMinutes: 45, Description: "Flow time log",
	})
	if err != nil {
		t.Fatalf("create Flow time log: %v", err)
	}
	t.Cleanup(func() {
		if err := repo.DeleteTimeLog(ctx, logID, testutil.OwnerID); err != nil {
			t.Errorf("delete Flow time log fixture: %v", err)
		}
	})

	if err := repo.RecalcActualHours(ctx, taskID, testutil.OwnerID); err != nil {
		t.Fatalf("recalculate task hours: %v", err)
	}
	var actualHours sql.NullFloat64
	if err := db.QueryRowContext(ctx, "SELECT actual_hours FROM flow_tasks WHERE id = ?", taskID).Scan(&actualHours); err != nil {
		t.Fatalf("read task hours: %v", err)
	}
	if !actualHours.Valid || actualHours.Float64 != 1.25 {
		t.Fatalf("actual_hours = %+v, want 1.25 from 30 HR minutes plus 45 Flow minutes", actualHours)
	}

	day, err := repo.UpsertWorkDay(ctx, domain.WorkDayInput{
		BusinessID: businessID, EmployeeID: testutil.OwnerID,
		Date: "2035-01-01", Timezone: "UTC", WorkMode: domain.WorkModeOnsite, LoggedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("create Flow work day: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(ctx, "DELETE FROM flow_work_days WHERE id = ?", day.ID); err != nil {
			t.Errorf("delete Flow work day fixture: %v", err)
		}
	})

	service := application.NewService(ctx, repo)
	line, err := service.CreateWorkDayLine(ctx, domain.WorkDayLineInput{
		BusinessID: businessID, EmployeeID: testutil.OwnerID,
		WorkDayID: day.ID, Date: "2035-01-01", ProjectID: taskID, TaskID: taskID,
		DurationMinutes: 15, ClientKey: clientKey + "-flow-line", LoggedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("create Flow work-day line: %v", err)
	}
	assertActualHours := func(want float64) {
		t.Helper()
		if err := db.QueryRowContext(ctx, "SELECT actual_hours FROM flow_tasks WHERE id = ?", taskID).Scan(&actualHours); err != nil {
			t.Fatalf("read task hours: %v", err)
		}
		if !actualHours.Valid || actualHours.Float64 != want {
			t.Fatalf("actual_hours = %+v, want %.2f", actualHours, want)
		}
	}
	assertActualHours(1.50)

	line, err = service.UpdateWorkDayLine(ctx, domain.WorkDayLineInput{
		ID: line.ID, BusinessID: businessID,
		ProjectID: taskID, TaskID: taskID, DurationMinutes: 30,
		Description: "updated Flow work-day line", LoggedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("update Flow work-day line: %v", err)
	}
	assertActualHours(1.75)

	deleted, err := service.DeleteWorkDayLine(ctx, line.ID, businessID, testutil.OwnerID)
	if err != nil || !deleted {
		t.Fatalf("delete Flow work-day line = %t, err=%v", deleted, err)
	}
	assertActualHours(1.25)
}
