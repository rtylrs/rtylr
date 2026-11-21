package api_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/api"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/testutil"
	flowcyclev2 "github.com/voxire/voxie/gen/go/flow/cycle/v2"
	flowtaskv2 "github.com/voxire/voxie/gen/go/flow/task/v2"
)

// cycleFixture adds cycles to the role fixture's project. Cycles are created
// through the repository so each test controls exactly which exist.
func (f *roleFixture) newCycle(t *testing.T, businessID, projectID int64, name string) int64 {
	t.Helper()
	ctx := context.Background()
	start := time.Date(2026, 12, 7, 0, 0, 0, 0, time.UTC)
	id, err := f.repo.CreateCycleV2(ctx, domain.CycleInput{
		BusinessID: businessID, ProjectID: &projectID, Name: name,
		StartDate: start, EndDate: start.AddDate(0, 0, 13), Status: domain.CycleStatusUpcoming,
	})
	if err != nil {
		t.Fatalf("CreateCycleV2: %v", err)
	}
	db := testutil.OpenDB(t)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_cycle_carryovers WHERE cycle_id = ? OR to_cycle_id = ?", id, id)
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_cycle_tasks WHERE cycle_id = ?", id)
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_cycles WHERE id = ?", id)
	})
	return id
}

func assertErrorContains(t *testing.T, err error, want, what string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("%s: expected an error containing %q, got %v", what, want, err)
	}
}

// TestCyclePlanningEndpointRoles is the access matrix for the planning
// endpoints: a project viewer reads, a member plans, a stranger does neither,
// and a cycle of another business is not found.
func TestCyclePlanningEndpointRoles(t *testing.T) {
	f := newRoleFixture(t)
	cycles := api.NewCycleAPIV2(f.svc())
	sprint := f.newCycle(t, testutil.BusinessID, f.projID, "Sprint A")
	next := f.newCycle(t, testutil.BusinessID, f.projID, "Sprint B")
	task := f.newTask(t)

	// Reads: viewer yes, stranger no.
	if _, err := cycles.GetCycle(ctxAs(roleViewerID), &flowcyclev2.GetCycleRequest{Id: sprint}); err != nil {
		t.Fatalf("viewer GetCycle: %v", err)
	}
	if _, err := cycles.GetCycleCapacity(ctxAs(roleViewerID), &flowcyclev2.GetCycleCapacityRequest{Id: sprint}); err != nil {
		t.Fatalf("viewer GetCycleCapacity: %v", err)
	}
	_, err := cycles.GetCycle(ctxAs(roleStrangerID), &flowcyclev2.GetCycleRequest{Id: sprint})
	assertPermissionDenied(t, err, "stranger GetCycle")
	_, err = cycles.GetCycleCapacity(ctxAs(roleStrangerID), &flowcyclev2.GetCycleCapacityRequest{Id: sprint})
	assertPermissionDenied(t, err, "stranger GetCycleCapacity")

	// Writes: viewer and stranger no.
	taskID := task
	for _, caller := range []int64{roleViewerID, roleStrangerID} {
		_, err = cycles.StartCycle(ctxAs(caller), &flowcyclev2.StartCycleRequest{Id: sprint, LengthWeeks: 2})
		assertPermissionDenied(t, err, "StartCycle")
		_, err = cycles.CompleteCycle(ctxAs(caller), &flowcyclev2.CompleteCycleRequest{Id: sprint})
		assertPermissionDenied(t, err, "CompleteCycle")
		_, err = cycles.UpdateCycle(ctxAs(caller), &flowcyclev2.UpdateCycleRequest{Id: sprint, Name: "Renamed"})
		assertPermissionDenied(t, err, "UpdateCycle")
		_, err = cycles.MoveCycleTask(ctxAs(caller), &flowcyclev2.MoveCycleTaskRequest{TaskId: &taskID, CycleId: &sprint})
		assertPermissionDenied(t, err, "MoveCycleTask")
		_, err = cycles.CreateCycle(ctxAs(caller), &flowcyclev2.CreateCycleRequest{ProjectId: &f.projID, Name: "Nope"})
		assertPermissionDenied(t, err, "CreateCycle")
	}

	// Member plans the whole loop.
	member := ctxAs(roleMemberID)
	moved, err := cycles.MoveCycleTask(member, &flowcyclev2.MoveCycleTaskRequest{TaskId: &taskID, CycleId: &sprint})
	if err != nil || moved.GetCycleId() != sprint {
		t.Fatalf("member MoveCycleTask = %+v, %v", moved, err)
	}
	goal := "Finish the backlog"
	started, err := cycles.StartCycle(member, &flowcyclev2.StartCycleRequest{Id: sprint, LengthWeeks: 1, Goal: &goal})
	if err != nil {
		t.Fatalf("member StartCycle: %v", err)
	}
	if started.Cycle.Status != domain.CycleStatusActive || started.Cycle.Goal != goal || started.Cycle.Scope != 1 {
		t.Fatalf("started cycle = %+v", started.Cycle)
	}
	_, err = cycles.StartCycle(member, &flowcyclev2.StartCycleRequest{Id: next})
	assertErrorContains(t, err, "already has an active cycle", "second StartCycle")
	_, err = cycles.StartCycle(member, &flowcyclev2.StartCycleRequest{Id: next, LengthWeeks: 5})
	assertErrorContains(t, err, "length_weeks", "StartCycle with five weeks")

	// A goal-only update keeps the name.
	if _, err := cycles.UpdateCycle(member, &flowcyclev2.UpdateCycleRequest{Id: next, Goal: &goal}); err != nil {
		t.Fatalf("member UpdateCycle: %v", err)
	}
	got, err := cycles.GetCycle(member, &flowcyclev2.GetCycleRequest{Id: next})
	if err != nil || got.Cycle.Name != "Sprint B" || got.Cycle.Goal != goal {
		t.Fatalf("after goal-only update: %+v, %v", got, err)
	}

	completed, err := cycles.CompleteCycle(member, &flowcyclev2.CompleteCycleRequest{
		Id:        sprint,
		Decisions: []*flowcyclev2.CycleTaskDecision{{TaskId: task, Decision: domain.CycleDecisionMove, ToCycleId: &next}},
	})
	if err != nil || completed.Cycle.Status != domain.CycleStatusCompleted {
		t.Fatalf("member CompleteCycle = %+v, %v", completed, err)
	}
	got, err = cycles.GetCycle(member, &flowcyclev2.GetCycleRequest{Id: next})
	if err != nil || len(got.Cycle.CarriedOver) != 1 || got.Cycle.CarriedOver[0].FromCycleName != "Sprint A" {
		t.Fatalf("carried over into Sprint B = %+v, %v", got, err)
	}

	// Another business's cycle is not found, not forbidden.
	foreign := f.newCycle(t, testutil.BusinessID+7, f.projID, "Foreign")
	_, err = cycles.GetCycle(member, &flowcyclev2.GetCycleRequest{Id: foreign})
	assertErrorContains(t, err, "Cycle not found", "GetCycle of another business")
	_, err = cycles.StartCycle(member, &flowcyclev2.StartCycleRequest{Id: foreign})
	assertErrorContains(t, err, "Cycle not found", "StartCycle of another business")
}

// TestTaskCycleFieldIsEditable covers the task panel: a task's cycle can be
// set, read back with its name, cleared, and never pointed at a cycle of
// another project.
func TestTaskCycleFieldIsEditable(t *testing.T) {
	f := newRoleFixture(t)
	tasks := api.NewTaskAPIV2(f.svc())
	sprint := f.newCycle(t, testutil.BusinessID, f.projID, "Sprint A")
	task := f.newTask(t)
	member := ctxAs(roleMemberID)

	if _, err := tasks.UpdateTask(member, &flowtaskv2.UpdateTaskRequest{Id: task, Body: &flowtaskv2.UpdateTaskBody{CycleId: &sprint}}); err != nil {
		t.Fatalf("UpdateTask cycle_id: %v", err)
	}
	detail, err := tasks.GetTaskDetail(member, &flowtaskv2.GetTaskDetailRequest{Id: task})
	if err != nil || detail.Task.GetCycleId() != sprint || detail.Task.CycleName != "Sprint A" || detail.Task.PlanRank == nil {
		t.Fatalf("task after setting its cycle = %+v, %v", detail.GetTask(), err)
	}

	zero := int64(0)
	if _, err := tasks.UpdateTask(member, &flowtaskv2.UpdateTaskRequest{Id: task, Body: &flowtaskv2.UpdateTaskBody{CycleId: &zero}}); err != nil {
		t.Fatalf("UpdateTask cycle_id 0: %v", err)
	}
	detail, err = tasks.GetTaskDetail(member, &flowtaskv2.GetTaskDetailRequest{Id: task})
	if err != nil || detail.Task.CycleId != nil {
		t.Fatalf("task after clearing its cycle = %+v, %v", detail.GetTask(), err)
	}

	otherProject, err := f.repo.CreateProject(context.Background(), domain.ProjectInput{
		WorkspaceID: f.wsID, OwnerID: roleOwnerID, BusinessID: testutil.BusinessID,
		Name: "Other Project", CreatedBy: roleOwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	elsewhere := f.newCycle(t, testutil.BusinessID, otherProject, "Elsewhere")
	_, err = tasks.UpdateTask(member, &flowtaskv2.UpdateTaskRequest{Id: task, Body: &flowtaskv2.UpdateTaskBody{CycleId: &elsewhere}})
	assertErrorContains(t, err, "Cycle must belong to the task project", "UpdateTask into another project's cycle")

	_, err = tasks.UpdateTask(ctxAs(roleViewerID), &flowtaskv2.UpdateTaskRequest{Id: task, Body: &flowtaskv2.UpdateTaskBody{CycleId: &sprint}})
	if err == nil {
		t.Fatal("viewer UpdateTask cycle_id: expected a refusal")
	}
}
