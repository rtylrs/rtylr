package infrastructure_test

import (
	"context"
	"database/sql"
	"math"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// The insights endpoint used to report cycleTime 0 for every workspace because
// AVG(DATEDIFF(...)) arrives from MySQL as DECIMAL -> []byte and the code
// asserted float64. Velocity total/completionRate and the burndown actual line
// were arithmetic on constants rather than measured history. These pin the
// measured values against a real database.

const insightOwnerID int64 = 980701

func insertInsightTask(t *testing.T, db *sql.DB, workspaceID int64, title string, createdAt time.Time, completedAt *time.Time) {
	t.Helper()

	status := "todo"
	var completed interface{}
	if completedAt != nil {
		status = "done"
		completed = completedAt.UTC()
	}
	if _, err := db.ExecContext(context.Background(),
		`INSERT INTO flow_tasks (project_id, workspace_id, owner_id, business_id, title, reporter_id, status, created_at, completed_at)
		 VALUES (0, ?, ?, ?, ?, ?, ?, ?, ?)`,
		workspaceID, insightOwnerID, testutil.BusinessID, title, insightOwnerID, status, createdAt.UTC(), completed,
	); err != nil {
		t.Fatalf("insert task %q: %v", title, err)
	}
}

func newInsightsFixture(t *testing.T, workspaceID int64) (*infrastructure.Repository, *sql.DB) {
	t.Helper()

	db := testutil.OpenDB(t)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM flow_tasks WHERE workspace_id = ?", workspaceID)
	})
	return testsupport.NewRepository(db), db
}

// A five-day average is what the database stores; the response must carry it
// rather than the 0 the failed []byte assertion produced.
func TestGetInsights_CycleTimeIsMeasured(t *testing.T) {
	const workspaceID int64 = 980711
	repo, db := newInsightsFixture(t, workspaceID)
	ctx := context.Background()

	now := time.Now().UTC()
	completed := now.Add(-time.Hour)
	insertInsightTask(t, db, workspaceID, "five day task", now.AddDate(0, 0, -5), &completed)

	ins, err := repo.GetInsights(ctx, workspaceID)
	if err != nil {
		t.Fatalf("GetInsights: %v", err)
	}
	if ins.CycleTime != 5 {
		t.Errorf("CycleTime = %v, want 5", ins.CycleTime)
	}
}

// Velocity total is the number of task events recorded in the week (creations
// plus completions), not completed*1.3.
func TestGetInsights_VelocityTotalIsMeasured(t *testing.T) {
	const workspaceID int64 = 980721
	repo, db := newInsightsFixture(t, workspaceID)
	ctx := context.Background()

	now := time.Now().UTC()
	completed := now.Add(-time.Hour)
	// Two completed tasks and one still-open task, all created today.
	insertInsightTask(t, db, workspaceID, "done 1", now.Add(-2*time.Hour), &completed)
	insertInsightTask(t, db, workspaceID, "done 2", now.Add(-3*time.Hour), &completed)
	insertInsightTask(t, db, workspaceID, "open 1", now.Add(-4*time.Hour), nil)

	ins, err := repo.GetInsights(ctx, workspaceID)
	if err != nil {
		t.Fatalf("GetInsights: %v", err)
	}

	weekday := (int(now.Weekday()) + 6) % 7 // Monday = 0, matching WEEKDAY().
	monday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -weekday)
	key := monday.Format("2006-01-02")

	var point *domain.InsightsVelocityPoint
	for i := range ins.Velocity {
		if ins.Velocity[i].Cycle == key {
			point = &ins.Velocity[i]
			break
		}
	}
	if point == nil {
		t.Fatalf("no velocity point for the current week %s (got %+v)", key, ins.Velocity)
	}
	if point.Completed != 2 || point.Total != 5 {
		t.Errorf("velocity point = %+v, want Completed 2 and Total 5 (3 creations + 2 completions)", *point)
	}
	if math.Abs(point.CompletionRate-0.4) > 1e-9 {
		t.Errorf("completion rate = %v, want 0.4", point.CompletionRate)
	}
}

// The burndown actual line is measured: remaining work rises when a task is
// created and falls when one completes. The old line was a straight synthetic
// burn that ignored both dates.
func TestGetInsights_BurndownActualFollowsTaskHistory(t *testing.T) {
	const workspaceID int64 = 980731
	repo, db := newInsightsFixture(t, workspaceID)
	ctx := context.Background()

	now := time.Now().UTC()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.UTC)
	// Created ten days ago, completed three days ago.
	completedAt := dayStart.AddDate(0, 0, -3)
	insertInsightTask(t, db, workspaceID, "old done", dayStart.AddDate(0, 0, -10), &completedAt)
	// Created two days ago and still open.
	insertInsightTask(t, db, workspaceID, "new open", dayStart.AddDate(0, 0, -2), nil)

	ins, err := repo.GetInsights(ctx, workspaceID)
	if err != nil {
		t.Fatalf("GetInsights: %v", err)
	}
	if len(ins.Burndown) != 15 {
		t.Fatalf("burndown points = %d, want 15", len(ins.Burndown))
	}

	// Day 0 is 14 days ago, so the ten-days-ago creation lands on day 4, the
	// three-days-ago completion on day 11 and the two-days-ago creation on
	// day 12.
	wantActual := map[int]float64{4: 1, 10: 1, 11: 0, 14: 1}
	for day, want := range wantActual {
		if got := ins.Burndown[day].Actual; got != want {
			t.Errorf("burndown[%d].Actual = %v, want %v (full: %+v)", day, got, want, ins.Burndown)
		}
	}
	if ins.Burndown[0].Ideal != ins.Burndown[0].Actual {
		t.Errorf("ideal line starts at %v but actual starts at %v", ins.Burndown[0].Ideal, ins.Burndown[0].Actual)
	}
	if ins.Burndown[14].Ideal != 0 {
		t.Errorf("ideal line ends at %v, want 0", ins.Burndown[14].Ideal)
	}
}
