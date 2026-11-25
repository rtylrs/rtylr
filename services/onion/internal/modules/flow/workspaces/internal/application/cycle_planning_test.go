package application

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

type planningFixture struct {
	svc     *Service
	ctx     context.Context
	db      *sql.DB
	project domain.Project
}

func newPlanningFixture(t *testing.T) planningFixture {
	t.Helper()
	svc, ctx, db := v2Svc(t)
	p := createV2Project(ctx, t, svc, "cycle-planning", domain.VisibilityPrivate, v2UserA)
	return planningFixture{svc: svc, ctx: ctx, db: db, project: p}
}

func (f planningFixture) setStatus(t *testing.T, taskID int64, status string) {
	t.Helper()
	if err := f.svc.UpdateTaskStatus(f.ctx, domain.UpdateTaskStatusInput{
		ID: taskID, OwnerID: v2UserA, BusinessID: v2Business, Status: status,
	}, v2UserA); err != nil {
		t.Fatalf("UpdateTaskStatus(%d, %s): %v", taskID, status, err)
	}
}

func (f planningFixture) archive(t *testing.T, taskID int64) {
	t.Helper()
	if _, err := f.db.Exec("UPDATE flow_tasks SET archived = TRUE WHERE id = ?", taskID); err != nil {
		t.Fatalf("archive task %d: %v", taskID, err)
	}
}

func (f planningFixture) cycle(t *testing.T, name, start, end string) int64 {
	t.Helper()
	c, err := f.svc.CreateCycleV2(f.ctx, domain.CycleInput{
		BusinessID: v2Business, ProjectID: &f.project.ID, Name: name, Goal: name + " goal", CreatedBy: v2UserA,
	}, start, end)
	if err != nil {
		t.Fatalf("CreateCycleV2(%s): %v", name, err)
	}
	if c.Status != domain.CycleStatusUpcoming {
		t.Fatalf("a new cycle is %q, want upcoming until it is started", c.Status)
	}
	return c.ID
}

func (f planningFixture) task(t *testing.T, title string, assignee *int64, effort int32) int64 {
	t.Helper()
	task, err := f.svc.CreateTask(f.ctx, domain.TaskInput{
		ProjectID: f.project.ID, BusinessID: v2Business, OwnerID: v2UserA, ReporterID: v2UserA,
		Title: title, AssigneeID: assignee, EffortMinutes: &effort,
	})
	if err != nil {
		t.Fatalf("CreateTask(%s): %v", title, err)
	}
	return task.ID
}

func (f planningFixture) move(t *testing.T, taskID, cycleID, before, after int64) {
	t.Helper()
	if _, err := f.svc.MoveCycleTask(f.ctx, domain.MoveCycleTaskInput{
		BusinessID: v2Business, TaskID: taskID, CycleID: cycleID, BeforeTaskID: before, AfterTaskID: after,
	}); err != nil {
		t.Fatalf("MoveCycleTask(task %d to cycle %d): %v", taskID, cycleID, err)
	}
}

// order returns the ids in cycleID (0 for the backlog) by plan rank, the way
// the Backlog screen sorts them.
func (f planningFixture) order(t *testing.T, cycleID int64, ids ...int64) []int64 {
	t.Helper()
	tasks := make([]domain.Task, 0, len(ids))
	for _, id := range ids {
		tasks = append(tasks, domain.Task{ID: id})
	}
	if err := f.svc.AttachTaskCycles(f.ctx, tasks); err != nil {
		t.Fatalf("AttachTaskCycles: %v", err)
	}
	var in []domain.Task
	for _, task := range tasks {
		var got int64
		if task.CycleID != nil {
			got = *task.CycleID
		}
		if got == cycleID {
			in = append(in, task)
		}
	}
	sort.SliceStable(in, func(i, j int) bool {
		a, b := in[i].PlanRank, in[j].PlanRank
		if a == nil || b == nil {
			return a != nil
		}
		return *a < *b
	})
	out := make([]int64, 0, len(in))
	for _, task := range in {
		out = append(out, task.ID)
	}
	return out
}

func sameIDs(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestCyclePlanningOrderMoveAndMembership(t *testing.T) {
	f := newPlanningFixture(t)
	sprint := f.cycle(t, "Sprint 1", "2026-12-07", "2026-12-18")
	t1 := f.task(t, "one", nil, 60)
	t2 := f.task(t, "two", nil, 60)
	t3 := f.task(t, "three", nil, 60)
	t4 := f.task(t, "four", nil, 60)

	for _, id := range []int64{t1, t2, t3} {
		f.move(t, id, sprint, 0, 0)
	}
	if got := f.order(t, sprint, t1, t2, t3, t4); !sameIDs(got, []int64{t1, t2, t3}) {
		t.Fatalf("appended order = %v, want %v", got, []int64{t1, t2, t3})
	}
	f.move(t, t3, sprint, t1, 0)
	if got := f.order(t, sprint, t1, t2, t3, t4); !sameIDs(got, []int64{t3, t1, t2}) {
		t.Fatalf("after moving three before one = %v", got)
	}
	f.move(t, t2, sprint, 0, t3)
	if got := f.order(t, sprint, t1, t2, t3, t4); !sameIDs(got, []int64{t3, t2, t1}) {
		t.Fatalf("after moving two after three = %v", got)
	}

	// Back to the backlog: a moved task goes to the top, ahead of tasks that
	// were never placed.
	f.move(t, t1, 0, 0, 0)
	if got := f.order(t, 0, t1, t2, t3, t4); !sameIDs(got, []int64{t1, t4}) {
		t.Fatalf("backlog order = %v, want [one four]", got)
	}
	f.move(t, t4, 0, t1, 0)
	if got := f.order(t, 0, t1, t2, t3, t4); !sameIDs(got, []int64{t4, t1}) {
		t.Fatalf("backlog after reorder = %v, want [four one]", got)
	}

	// A task is in one open cycle at a time.
	other := f.cycle(t, "Sprint 2", "2026-12-21", "2027-01-01")
	f.move(t, t2, other, 0, 0)
	tasks := []domain.Task{{ID: t2}}
	if err := f.svc.AttachTaskCycles(f.ctx, tasks); err != nil {
		t.Fatalf("AttachTaskCycles: %v", err)
	}
	if tasks[0].CycleID == nil || *tasks[0].CycleID != other || tasks[0].CycleName != "Sprint 2" {
		t.Fatalf("task two cycle = %v %q, want Sprint 2", tasks[0].CycleID, tasks[0].CycleName)
	}
	if got := f.order(t, sprint, t1, t2, t3, t4); !sameIDs(got, []int64{t3}) {
		t.Fatalf("Sprint 1 still holds %v after task two moved out", got)
	}

	// The task panel's write appends to the end of the cycle.
	if err := f.svc.SetTaskCycle(f.ctx, v2Business, t4, sprint); err != nil {
		t.Fatalf("SetTaskCycle: %v", err)
	}
	if got := f.order(t, sprint, t1, t2, t3, t4); !sameIDs(got, []int64{t3, t4}) {
		t.Fatalf("after SetTaskCycle = %v, want [three four]", got)
	}

	// An anchor from another list is refused.
	_, err := f.svc.MoveCycleTask(f.ctx, domain.MoveCycleTaskInput{BusinessID: v2Business, TaskID: t1, CycleID: sprint, BeforeTaskID: t2})
	if !errors.Is(err, domain.ErrCycleRankAnchor) {
		t.Fatalf("anchor outside the list: err = %v, want ErrCycleRankAnchor", err)
	}
}

func TestCyclePlanningStartAndComplete(t *testing.T) {
	f := newPlanningFixture(t)
	sprint := f.cycle(t, "Sprint 1", "2026-12-07", "2026-12-18")
	next := f.cycle(t, "Sprint 2", "2026-12-21", "2027-01-01")
	assignee := v2UserA
	done := f.task(t, "done", &assignee, 120)
	carry := f.task(t, "carry", &assignee, 60)
	cancel := f.task(t, "cancel", nil, 30)
	leftover := f.task(t, "leftover", nil, 0)
	for _, id := range []int64{done, carry, cancel, leftover} {
		f.move(t, id, sprint, 0, 0)
	}
	if err := f.svc.UpdateTaskStatus(f.ctx, domain.UpdateTaskStatusInput{
		ID: done, OwnerID: v2UserA, BusinessID: v2Business, Status: domain.TaskStatusDone,
	}, v2UserA); err != nil {
		t.Fatalf("UpdateTaskStatus: %v", err)
	}

	goal := "Ship the backlog screen"
	started, err := f.svc.StartCycle(f.ctx, domain.StartCycleInput{
		ID: sprint, BusinessID: v2Business, LengthDays: 7, Goal: &goal,
		Today: time.Date(2026, 12, 7, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("StartCycle: %v", err)
	}
	if started.Status != domain.CycleStatusActive || started.Goal != goal || started.StartedAt == nil {
		t.Fatalf("started cycle = %+v", started)
	}
	if got := started.EndDate.Format("2006-01-02"); got != "2026-12-14" {
		t.Fatalf("end date = %s, want one week after the start", got)
	}
	if started.Scope != 4 || started.Totals.Done != 1 || started.Totals.Backlog != 3 || started.EstimateMinutes != 210 {
		t.Fatalf("totals = scope %d %+v estimate %d", started.Scope, started.Totals, started.EstimateMinutes)
	}

	if _, err := f.svc.StartCycle(f.ctx, domain.StartCycleInput{ID: next, BusinessID: v2Business}); !errors.Is(err, domain.ErrCycleAlreadyActive) {
		t.Fatalf("second active cycle: err = %v, want ErrCycleAlreadyActive", err)
	}
	if _, err := f.svc.StartCycle(f.ctx, domain.StartCycleInput{ID: sprint, BusinessID: v2Business}); !errors.Is(err, domain.ErrCycleNotUpcoming) {
		t.Fatalf("restart: err = %v, want ErrCycleNotUpcoming", err)
	}

	// A decision for a finished task is rejected and nothing is saved.
	_, err = f.svc.CompleteCycle(f.ctx, domain.CompleteCycleInput{
		ID: sprint, BusinessID: v2Business, ActorID: v2UserA,
		Decisions: []domain.CycleTaskDecision{{TaskID: done, Decision: domain.CycleDecisionBacklog}},
	})
	var decisionErr *domain.CycleDecisionError
	if !errors.As(err, &decisionErr) || decisionErr.TaskID != done {
		t.Fatalf("decision for a done task: err = %v", err)
	}

	completed, err := f.svc.CompleteCycle(f.ctx, domain.CompleteCycleInput{
		ID: sprint, BusinessID: v2Business, ActorID: v2UserA,
		Decisions: []domain.CycleTaskDecision{
			{TaskID: carry, Decision: domain.CycleDecisionMove, ToCycleID: next},
			{TaskID: cancel, Decision: domain.CycleDecisionCancel},
		},
	})
	if err != nil {
		t.Fatalf("CompleteCycle: %v", err)
	}
	if completed.Status != domain.CycleStatusCompleted || completed.CompletedAt == nil {
		t.Fatalf("completed cycle = %+v", completed)
	}

	receiving, err := f.svc.GetCycleV2(f.ctx, v2Business, next)
	if err != nil {
		t.Fatalf("GetCycleV2: %v", err)
	}
	if len(receiving.CarriedOver) != 1 || receiving.CarriedOver[0].FromCycleID != sprint ||
		receiving.CarriedOver[0].FromCycleName != "Sprint 1" || !sameIDs(receiving.CarriedOver[0].TaskIDs, []int64{carry}) {
		t.Fatalf("carried over = %+v", receiving.CarriedOver)
	}

	tasks := []domain.Task{{ID: done}, {ID: carry}, {ID: cancel}, {ID: leftover}}
	if err := f.svc.AttachTaskCycles(f.ctx, tasks); err != nil {
		t.Fatalf("AttachTaskCycles: %v", err)
	}
	cycleOf := func(task domain.Task) int64 {
		if task.CycleID == nil {
			return 0
		}
		return *task.CycleID
	}
	if cycleOf(tasks[0]) != sprint || cycleOf(tasks[1]) != next || cycleOf(tasks[2]) != sprint || cycleOf(tasks[3]) != 0 {
		t.Fatalf("cycles after completion: done %d carry %d cancel %d leftover %d",
			cycleOf(tasks[0]), cycleOf(tasks[1]), cycleOf(tasks[2]), cycleOf(tasks[3]))
	}
	cancelled, err := f.svc.GetTaskDetail(f.ctx, cancel)
	if err != nil || cancelled.Task.Status != domain.TaskStatusCancelled {
		t.Fatalf("cancelled task status = %q, %v", cancelled.Task.Status, err)
	}

	// A completed cycle takes no more tasks and cannot be completed twice.
	if _, err := f.svc.MoveCycleTask(f.ctx, domain.MoveCycleTaskInput{BusinessID: v2Business, TaskID: leftover, CycleID: sprint}); !errors.Is(err, domain.ErrCycleCompleted) {
		t.Fatalf("move into a completed cycle: err = %v", err)
	}
	if _, err := f.svc.CompleteCycle(f.ctx, domain.CompleteCycleInput{ID: sprint, BusinessID: v2Business, ActorID: v2UserA}); !errors.Is(err, domain.ErrCycleCompleted) {
		t.Fatalf("complete twice: err = %v", err)
	}
}

func TestCyclePlanningIsScopedByBusiness(t *testing.T) {
	f := newPlanningFixture(t)
	sprint := f.cycle(t, "Sprint 1", "2026-12-07", "2026-12-18")
	task := f.task(t, "mine", nil, 0)
	other := v2Business + 1

	if _, err := f.svc.GetCycleV2(f.ctx, other, sprint); !errors.Is(err, domain.ErrCycleNotFound) {
		t.Fatalf("GetCycleV2 from another business: err = %v", err)
	}
	if _, err := f.svc.CycleProjectID(f.ctx, other, sprint); !errors.Is(err, domain.ErrCycleNotFound) {
		t.Fatalf("CycleProjectID from another business: err = %v", err)
	}
	if _, err := f.svc.MoveCycleTask(f.ctx, domain.MoveCycleTaskInput{BusinessID: other, TaskID: task, CycleID: sprint}); !errors.Is(err, domain.ErrCycleTaskNotFound) {
		t.Fatalf("MoveCycleTask from another business: err = %v", err)
	}
	if _, err := f.svc.StartCycle(f.ctx, domain.StartCycleInput{ID: sprint, BusinessID: other}); !errors.Is(err, domain.ErrCycleNotFound) {
		t.Fatalf("StartCycle from another business: err = %v", err)
	}
	if _, err := f.svc.CompleteCycle(f.ctx, domain.CompleteCycleInput{ID: sprint, BusinessID: other}); !errors.Is(err, domain.ErrCycleNotFound) {
		t.Fatalf("CompleteCycle from another business: err = %v", err)
	}

	// A cycle of another project is not a valid target.
	p2 := createV2Project(f.ctx, t, f.svc, "other-project", domain.VisibilityPrivate, v2UserA)
	c2, err := f.svc.CreateCycleV2(f.ctx, domain.CycleInput{BusinessID: v2Business, ProjectID: &p2.ID, Name: "Elsewhere"}, "2026-12-07", "2026-12-18")
	if err != nil {
		t.Fatalf("CreateCycleV2: %v", err)
	}
	if _, err := f.svc.MoveCycleTask(f.ctx, domain.MoveCycleTaskInput{BusinessID: v2Business, TaskID: task, CycleID: c2.ID}); !errors.Is(err, domain.ErrCycleNotFound) {
		t.Fatalf("move into another project's cycle: err = %v", err)
	}
}

type fakeCycleLeave struct {
	dates map[int64]map[string]bool
	err   error
	asked []int64
}

func (f *fakeCycleLeave) ApprovedLeaveDates(_ context.Context, _ int64, userIDs []int64, _, _ time.Time) (map[int64]map[string]bool, error) {
	f.asked = append(f.asked, userIDs...)
	return f.dates, f.err
}

func TestCycleCapacityTakesLeaveFromAvailableTime(t *testing.T) {
	f := newPlanningFixture(t)
	// Monday 7 to Friday 18 December: ten working days.
	sprint := f.cycle(t, "Sprint 1", "2026-12-07", "2026-12-18")
	assignee := v2UserA
	for _, effort := range []int32{600, 300} {
		f.move(t, f.task(t, "assigned", &assignee, effort), sprint, 0, 0)
	}
	f.move(t, f.task(t, "unassigned", nil, 90), sprint, 0, 0)

	leave := &fakeCycleLeave{dates: map[int64]map[string]bool{
		v2UserA: {"2026-12-08": true, "2026-12-12": true, "2026-12-15": true},
	}}
	f.svc.SetCycleLeaveSource(leave)
	capacity, err := f.svc.GetCycleCapacity(f.ctx, v2Business, sprint)
	if err != nil {
		t.Fatalf("GetCycleCapacity: %v", err)
	}
	if !capacity.LeaveApplied || capacity.MinutesPerDay != CycleMinutesPerDay {
		t.Fatalf("capacity flags = %+v", capacity)
	}
	if capacity.UnassignedMinutes != 90 || capacity.UnassignedCount != 1 {
		t.Fatalf("unassigned = %d minutes, %d tasks", capacity.UnassignedMinutes, capacity.UnassignedCount)
	}
	var me *domain.CycleMemberCapacity
	for i := range capacity.Members {
		if capacity.Members[i].UserID == v2UserA {
			me = &capacity.Members[i]
		}
	}
	if me == nil {
		t.Fatalf("the assignee is missing from %+v", capacity.Members)
	}
	// The Saturday leave day does not count against working days.
	if me.WorkingDays != 10 || me.LeaveDays != 2 || me.AvailableMinutes != 8*CycleMinutesPerDay ||
		me.CommittedMinutes != 900 || me.TaskCount != 2 {
		t.Fatalf("member capacity = %+v", *me)
	}

	// A leave read that fails falls back to working days only and says so.
	f.svc.SetCycleLeaveSource(&fakeCycleLeave{err: errors.New("hr unavailable")})
	capacity, err = f.svc.GetCycleCapacity(f.ctx, v2Business, sprint)
	if err != nil {
		t.Fatalf("GetCycleCapacity: %v", err)
	}
	if capacity.LeaveApplied {
		t.Fatal("leave_applied is true although the leave read failed")
	}
	for _, m := range capacity.Members {
		if m.UserID == v2UserA && m.AvailableMinutes != 10*CycleMinutesPerDay {
			t.Fatalf("available without leave = %d", m.AvailableMinutes)
		}
	}
}

func TestDeleteCycleReturnsItsTasksToTheBacklog(t *testing.T) {
	f := newPlanningFixture(t)
	sprint := f.cycle(t, "Sprint 1", "2026-12-07", "2026-12-18")
	next := f.cycle(t, "Sprint 2", "2026-12-21", "2027-01-01")
	carry := f.task(t, "carry", nil, 30)
	f.move(t, carry, sprint, 0, 0)
	if _, err := f.svc.StartCycle(f.ctx, domain.StartCycleInput{ID: sprint, BusinessID: v2Business}); err != nil {
		t.Fatalf("StartCycle: %v", err)
	}
	if _, err := f.svc.CompleteCycle(f.ctx, domain.CompleteCycleInput{
		ID: sprint, BusinessID: v2Business, ActorID: v2UserA,
		Decisions: []domain.CycleTaskDecision{{TaskID: carry, Decision: domain.CycleDecisionMove, ToCycleID: next}},
	}); err != nil {
		t.Fatalf("CompleteCycle: %v", err)
	}

	if err := f.svc.DeleteCycle(f.ctx, v2Business+1, next); !errors.Is(err, domain.ErrCycleNotFound) {
		t.Fatalf("DeleteCycle from another business: err = %v", err)
	}
	if err := f.svc.DeleteCycle(f.ctx, v2Business, next); err != nil {
		t.Fatalf("DeleteCycle: %v", err)
	}
	if _, err := f.svc.GetCycleV2(f.ctx, v2Business, next); !errors.Is(err, domain.ErrCycleNotFound) {
		t.Fatalf("deleted cycle still readable: err = %v", err)
	}
	if got := f.order(t, 0, carry); !sameIDs(got, []int64{carry}) {
		t.Fatalf("backlog after delete = %v, want the carried task back in it", got)
	}

	// The closed cycle that sent work into it still loads.
	if _, err := f.svc.GetCycleV2(f.ctx, v2Business, sprint); err != nil {
		t.Fatalf("GetCycleV2 on the sending cycle: %v", err)
	}
	if err := f.svc.DeleteCycle(f.ctx, v2Business, sprint); err != nil {
		t.Fatalf("DeleteCycle on a completed cycle: %v", err)
	}
}

func TestCycleTotalsLeaveArchivedOutAndCancelledOutOfScope(t *testing.T) {
	f := newPlanningFixture(t)
	sprint := f.cycle(t, "Sprint 1", "2026-12-07", "2026-12-18")
	assignee := v2UserA
	done := f.task(t, "done", &assignee, 60)
	open := f.task(t, "open", &assignee, 30)
	cancelled := f.task(t, "cancelled", &assignee, 45)
	archived := f.task(t, "archived", &assignee, 90)
	for _, id := range []int64{done, open, cancelled, archived} {
		f.move(t, id, sprint, 0, 0)
	}
	f.setStatus(t, done, domain.TaskStatusDone)
	f.setStatus(t, cancelled, domain.TaskStatusCancelled)
	f.setStatus(t, archived, domain.TaskStatusDone)
	f.archive(t, archived)

	c, err := f.svc.GetCycleV2(f.ctx, v2Business, sprint)
	if err != nil {
		t.Fatalf("GetCycleV2: %v", err)
	}
	want := domain.CycleTotals{Backlog: 1, Done: 1, Cancelled: 1}
	if c.Scope != 2 || c.Completed != 1 || c.Totals != want || c.EstimateMinutes != 90 {
		t.Fatalf("scope %d completed %d totals %+v estimate %d; want scope 2, completed 1, %+v, estimate 90",
			c.Scope, c.Completed, c.Totals, c.EstimateMinutes, want)
	}

	capacity, err := f.svc.GetCycleCapacity(f.ctx, v2Business, sprint)
	if err != nil {
		t.Fatalf("GetCycleCapacity: %v", err)
	}
	found := false
	for _, m := range capacity.Members {
		if m.UserID != v2UserA {
			continue
		}
		found = true
		if m.CommittedMinutes != 90 || m.TaskCount != 2 {
			t.Fatalf("committed %d minutes over %d tasks, want 90 over 2", m.CommittedMinutes, m.TaskCount)
		}
	}
	if !found {
		t.Fatalf("the assignee is missing from %+v", capacity.Members)
	}
}

func TestDeleteTaskRemovesItsCycleMembership(t *testing.T) {
	f := newPlanningFixture(t)
	sprint := f.cycle(t, "Sprint 1", "2026-12-07", "2026-12-18")
	next := f.cycle(t, "Sprint 2", "2026-12-21", "2027-01-01")
	carry := f.task(t, "carry", nil, 30)
	stay := f.task(t, "stay", nil, 30)
	f.move(t, carry, sprint, 0, 0)
	if _, err := f.svc.StartCycle(f.ctx, domain.StartCycleInput{ID: sprint, BusinessID: v2Business}); err != nil {
		t.Fatalf("StartCycle: %v", err)
	}
	if _, err := f.svc.CompleteCycle(f.ctx, domain.CompleteCycleInput{
		ID: sprint, BusinessID: v2Business, ActorID: v2UserA,
		Decisions: []domain.CycleTaskDecision{{TaskID: carry, Decision: domain.CycleDecisionMove, ToCycleID: next}},
	}); err != nil {
		t.Fatalf("CompleteCycle: %v", err)
	}

	rows := func(taskID int64) (memberships, carryovers int) {
		t.Helper()
		if err := f.db.QueryRow("SELECT COUNT(*) FROM flow_cycle_tasks WHERE task_id = ?", taskID).Scan(&memberships); err != nil {
			t.Fatalf("count memberships: %v", err)
		}
		if err := f.db.QueryRow("SELECT COUNT(*) FROM flow_cycle_carryovers WHERE task_id = ?", taskID).Scan(&carryovers); err != nil {
			t.Fatalf("count carryovers: %v", err)
		}
		return memberships, carryovers
	}
	f.move(t, stay, next, 0, 0)
	if m, c := rows(carry); m == 0 || c == 0 {
		t.Fatalf("setup: carried task has %d memberships and %d carry-overs", m, c)
	}

	// Another business cannot delete it, so nothing of it may go.
	if err := f.svc.DeleteTask(f.ctx, carry, v2Business+1); err != nil {
		t.Fatalf("DeleteTask from another business: %v", err)
	}
	if m, c := rows(carry); m == 0 || c == 0 {
		t.Fatalf("a foreign delete removed %d/%d cycle rows", m, c)
	}

	if err := f.svc.DeleteTask(f.ctx, carry, v2Business); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}
	if m, c := rows(carry); m != 0 || c != 0 {
		t.Fatalf("deleted task left %d memberships and %d carry-overs", m, c)
	}
	if m, _ := rows(stay); m != 1 {
		t.Fatalf("the other task has %d memberships, want 1", m)
	}
	receiving, err := f.svc.GetCycleV2(f.ctx, v2Business, next)
	if err != nil {
		t.Fatalf("GetCycleV2: %v", err)
	}
	if receiving.Scope != 1 || len(receiving.CarriedOver) != 0 {
		t.Fatalf("receiving cycle scope %d carried %+v after the delete", receiving.Scope, receiving.CarriedOver)
	}
}
