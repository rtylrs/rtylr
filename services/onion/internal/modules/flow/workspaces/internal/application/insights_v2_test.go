package application

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

type progressTasks struct {
	done, customDone, open, cancelled, archived, subtask int64
}

// seedProgress gives the fixture's project one task of each kind project
// progress must tell apart. Only done, customDone and open count, and two of
// them are done.
func seedProgress(t *testing.T, f planningFixture) progressTasks {
	t.Helper()
	var p progressTasks
	p.done = f.task(t, "done", nil, 0)
	p.customDone = f.task(t, "custom done", nil, 0)
	p.open = f.task(t, "open", nil, 0)
	p.cancelled = f.task(t, "cancelled", nil, 0)
	p.archived = f.task(t, "archived", nil, 0)
	f.setStatus(t, p.done, domain.TaskStatusDone)
	f.setStatus(t, p.cancelled, domain.TaskStatusCancelled)
	f.setStatus(t, p.archived, domain.TaskStatusDone)
	f.archive(t, p.archived)

	sub, err := f.svc.CreateSubtask(f.ctx, domain.TaskInput{
		ProjectID: f.project.ID, BusinessID: v2Business, OwnerID: v2UserA, ReporterID: v2UserA,
		Title: "subtask", ParentTaskID: &p.open,
	})
	if err != nil {
		t.Fatalf("CreateSubtask: %v", err)
	}
	p.subtask = sub.ID
	f.setStatus(t, p.subtask, domain.TaskStatusDone)

	// A custom status in the done bucket decides, whatever the legacy status.
	res, err := f.db.Exec(`INSERT INTO flow_statuses (business_id, project_id, name, bucket, legacy_status, created_by)
VALUES (?, ?, 'Shipped', 'done', 'done', ?)`, v2Business, f.project.ID, v2UserA)
	if err != nil {
		t.Fatalf("insert status: %v", err)
	}
	statusID, _ := res.LastInsertId()
	if _, err := f.db.Exec("UPDATE flow_tasks SET status_id = ?, status = 'todo', completed_at = NOW() WHERE id = ?",
		statusID, p.customDone); err != nil {
		t.Fatalf("set custom status: %v", err)
	}
	return p
}

func TestProjectProgressCountsTopLevelUnarchivedByBucket(t *testing.T) {
	f := newPlanningFixture(t)
	seedProgress(t, f)

	check := func(source string, total, done int32) {
		t.Helper()
		if total != 3 || done != 2 {
			t.Errorf("%s: %d of %d done, want 2 of 3", source, done, total)
		}
	}
	listed, _, err := f.svc.ListProjectsV2(f.ctx, v2UserA, v2Business, "", "", "", 50, 0)
	if err != nil || len(listed) != 1 {
		t.Fatalf("ListProjectsV2 = %d projects, %v", len(listed), err)
	}
	check("ListProjectsV2", listed[0].TaskCount, listed[0].CompletedCount)

	one, err := f.svc.GetProjectV2(f.ctx, f.project.ID, v2UserA, v2Business)
	if err != nil {
		t.Fatalf("GetProjectV2: %v", err)
	}
	check("GetProjectV2", one.TaskCount, one.CompletedCount)

	counts, err := f.svc.GetProjectTaskCountsByProjectIDs(f.ctx, []int64{f.project.ID})
	if err != nil {
		t.Fatalf("GetProjectTaskCountsByProjectIDs: %v", err)
	}
	c := counts[f.project.ID]
	check("GetProjectTaskCountsByProjectIDs", int32(c.Total), int32(c.Done)) // #nosec G115 -- small test counts.
}

func TestInsightsV2SeriesOverVisibleProjects(t *testing.T) {
	f := newPlanningFixture(t)
	p := seedProgress(t, f)
	now := time.Now().UTC()

	// A closed cycle that finished one of its two tasks.
	start := time.Date(now.Year()+1, 1, 4, 0, 0, 0, 0, time.UTC)
	closed := f.cycle(t, "Sprint 1", start.Format("2006-01-02"), start.AddDate(0, 0, 11).Format("2006-01-02"))
	f.move(t, p.done, closed, 0, 0)
	f.move(t, p.open, closed, 0, 0)
	if _, err := f.svc.StartCycle(f.ctx, domain.StartCycleInput{ID: closed, BusinessID: v2Business, Today: start}); err != nil {
		t.Fatalf("StartCycle: %v", err)
	}
	if _, err := f.svc.CompleteCycle(f.ctx, domain.CompleteCycleInput{ID: closed, BusinessID: v2Business, ActorID: v2UserA}); err != nil {
		t.Fatalf("CompleteCycle: %v", err)
	}

	// The running cycle holds one open and one finished task, plus a cancelled
	// one that is out of its scope.
	today := now.Format("2006-01-02")
	running := f.cycle(t, "Sprint 2", today, now.AddDate(0, 0, 7).Format("2006-01-02"))
	for _, id := range []int64{p.open, p.customDone, p.cancelled} {
		f.move(t, id, running, 0, 0)
	}
	if _, err := f.svc.StartCycle(f.ctx, domain.StartCycleInput{
		ID: running, BusinessID: v2Business, LengthDays: 7, Today: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("StartCycle: %v", err)
	}

	ins, err := f.svc.GetInsightsV2(f.ctx, v2UserA, v2Business, now)
	if err != nil {
		t.Fatalf("GetInsightsV2: %v", err)
	}
	if ins.TotalTasks != 3 || ins.DoneTasks != 2 {
		t.Errorf("totals = %d of %d done, want 2 of 3", ins.DoneTasks, ins.TotalTasks)
	}
	completed := 0
	for _, w := range ins.Throughput {
		completed += w.Completed
	}
	if len(ins.Throughput) != domain.InsightsThroughputWeeks || completed != 2 {
		t.Errorf("throughput = %+v, want %d weeks holding 2 completions", ins.Throughput, domain.InsightsThroughputWeeks)
	}
	if len(ins.Velocity) != 1 || ins.Velocity[0].Cycle != "Sprint 1" || ins.Velocity[0].Completed != 1 ||
		ins.Velocity[0].Total != 2 || ins.Velocity[0].CompletionRate != 50 {
		t.Errorf("velocity = %+v, want Sprint 1 at 1 of 2 (50%%)", ins.Velocity)
	}
	if len(ins.Burndown) != 2 || ins.Burndown[0].Ideal != 2 || ins.Burndown[0].Actual != 1 {
		t.Errorf("burndown = %+v, want two points starting at ideal 2, actual 1", ins.Burndown)
	}

	stranger, err := f.svc.GetInsightsV2(f.ctx, v2UserB, v2Business, now)
	if err != nil {
		t.Fatalf("GetInsightsV2 stranger: %v", err)
	}
	if stranger.TotalTasks != 0 || len(stranger.Velocity) != 0 || len(stranger.Burndown) != 0 {
		t.Errorf("a non-member sees %+v of a private project", stranger)
	}
}
