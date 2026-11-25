package application

import (
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

func TestPlannedWorkloadScopeAndFilters(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	shared := createV2Project(ctx, t, svc, "planned-shared", domain.VisibilityPrivate, v2UserA)
	if err := svc.AddProjectMember(ctx, domain.AddProjectMemberInput{ProjectID: shared.ID, UserID: v2UserB, Role: domain.FlowRoleMember, AddedBy: v2UserA}); err != nil {
		t.Fatalf("AddProjectMember: %v", err)
	}
	privateToB := createV2Project(ctx, t, svc, "planned-b-only", domain.VisibilityPrivate, v2UserB)

	start, _ := time.Parse("2006-01-02", "2026-11-30")
	due, _ := time.Parse("2006-01-02", "2026-12-01")
	newTask := func(projectID, assignee int64, title string, effort int32) int64 {
		input := domain.TaskInput{
			ProjectID: projectID, BusinessID: v2Business, OwnerID: v2UserA, ReporterID: v2UserA,
			Title: title, AssigneeID: &assignee, StartDate: &start, DueDate: &due,
		}
		if effort > 0 {
			input.EffortMinutes = &effort
		}
		task, err := svc.CreateTask(ctx, input)
		if err != nil {
			t.Fatalf("CreateTask(%s): %v", title, err)
		}
		return task.ID
	}
	forB := newTask(shared.ID, v2UserB, "B's work", 240)
	forA := newTask(shared.ID, v2UserA, "A's work", 120)
	finished := newTask(shared.ID, v2UserB, "Done already", 60)
	newTask(shared.ID, v2UserB, "No estimate", 0)
	if err := svc.UpdateTaskStatus(ctx, domain.UpdateTaskStatusInput{ID: finished, OwnerID: v2UserA, BusinessID: v2Business, Status: domain.TaskStatusDone}, v2UserA); err != nil {
		t.Fatalf("UpdateTaskStatus: %v", err)
	}
	hidden := newTask(privateToB.ID, v2UserB, "B's private work", 60)

	from, _ := time.Parse("2006-01-02", "2026-11-30")
	to, _ := time.Parse("2006-01-02", "2026-12-06")
	ids := func(tasks []domain.PlannedTask) map[int64]bool {
		out := map[int64]bool{}
		for _, task := range tasks {
			out[task.ID] = true
		}
		return out
	}

	// A administers the shared project: both people's work there, not B's private project.
	days, tasks, people, err := svc.PlannedWorkload(ctx, v2UserA, v2Business, from, to, nil)
	if err != nil {
		t.Fatalf("PlannedWorkload(A): %v", err)
	}
	seen := ids(tasks)
	if !seen[forB] || !seen[forA] || seen[finished] || seen[hidden] || len(seen) != 2 {
		t.Fatalf("A sees %v", seen)
	}
	if len(people) != 2 {
		t.Fatalf("A's rows = %v, want A and B", people)
	}
	var bMonday int64
	for _, d := range days {
		if d.UserID == v2UserB && d.Date.Format("2006-01-02") == "2026-11-30" {
			bMonday = d.Minutes
		}
	}
	if bMonday != 120 {
		t.Fatalf("B's Monday = %d minutes, want 120 (240 over two days)", bMonday)
	}

	// B is a plain member of the shared project: only B's own work, anywhere.
	_, tasks, people, err = svc.PlannedWorkload(ctx, v2UserB, v2Business, from, to, nil)
	if err != nil {
		t.Fatalf("PlannedWorkload(B): %v", err)
	}
	seen = ids(tasks)
	if !seen[forB] || !seen[hidden] || seen[forA] || len(seen) != 2 {
		t.Fatalf("B sees %v", seen)
	}
	if len(people) != 1 || people[0] != v2UserB {
		t.Fatalf("B's rows = %v, want only B", people)
	}

	tooLong := from.AddDate(0, 0, 90)
	if _, _, _, err := svc.PlannedWorkload(ctx, v2UserA, v2Business, from, tooLong, nil); !errors.Is(err, domain.ErrInvalidPlannedRange) {
		t.Fatalf("90-day range: err = %v", err)
	}
}
