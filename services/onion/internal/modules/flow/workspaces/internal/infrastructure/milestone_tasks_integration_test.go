package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"
	"github.com/voxire/onion/internal/testutil"
)

func TestWorkspaces_TaskMilestoneLink(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "MilestoneLinkWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "Milestone Link Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	newMilestone := func(name string) int64 {
		t.Helper()
		id, err := repo.CreateMilestoneV2(ctx, domain.MilestoneInput{
			BusinessID: testutil.BusinessID, ProjectID: projID, Name: name, DueDate: time.Now().AddDate(0, 0, 14),
		})
		if err != nil {
			t.Fatalf("CreateMilestoneV2: %v", err)
		}
		t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_milestones WHERE id = ?", id) })
		return id
	}
	first := newMilestone("M-1")
	second := newMilestone("M-2")

	newTask := func(title string, due *time.Time) int64 {
		t.Helper()
		id, err := repo.CreateTask(ctx, domain.TaskInput{
			ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
			Title: title, ReporterID: testutil.OwnerID, DueDate: due,
		})
		if err != nil {
			t.Fatalf("CreateTask: %v", err)
		}
		t.Cleanup(func() { _ = repo.DeleteTask(ctx, id, testutil.BusinessID) })
		return id
	}
	lastWeek := time.Now().AddDate(0, 0, -7)
	overdue := newTask("Overdue task", &lastWeek)
	done := newTask("Done task", nil)
	if err := repo.UpdateTaskStatus(ctx, domain.UpdateTaskStatusInput{
		ID: done, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Status: domain.TaskStatusDone,
	}); err != nil {
		t.Fatalf("UpdateTaskStatus: %v", err)
	}

	if got, err := repo.MilestoneProjectID(ctx, first, testutil.BusinessID); err != nil || got != projID {
		t.Fatalf("MilestoneProjectID = %d, %v; want %d", got, err, projID)
	}
	if got, err := repo.MilestoneProjectID(ctx, first, testutil.BusinessID+1); err != nil || got != 0 {
		t.Fatalf("MilestoneProjectID from another business = %d, %v; want 0", got, err)
	}

	for _, task := range []int64{overdue, done} {
		if err := repo.SetTaskMilestone(ctx, task, &second); err != nil {
			t.Fatalf("SetTaskMilestone: %v", err)
		}
		// Moving the task to another milestone must not leave it in both.
		if err := repo.SetTaskMilestone(ctx, task, &first); err != nil {
			t.Fatalf("SetTaskMilestone: %v", err)
		}
	}

	links, err := repo.TaskMilestoneIDs(ctx, []int64{overdue, done})
	if err != nil {
		t.Fatalf("TaskMilestoneIDs: %v", err)
	}
	if links[overdue] != first || links[done] != first {
		t.Fatalf("TaskMilestoneIDs = %v; want both tasks in milestone %d", links, first)
	}

	counts := milestoneCounts(t, repo, ctx, projID)
	if got := counts[first]; got != (domain.Milestone{TaskCount: 2, DoneCount: 1, OverdueCount: 1}) {
		t.Errorf("milestone %d counts = %+v; want 2 tasks, 1 done, 1 overdue", first, got)
	}
	if got := counts[second]; got.TaskCount != 0 {
		t.Errorf("milestone %d still counts %d tasks after they moved", second, got.TaskCount)
	}

	if err := repo.SetTaskMilestone(ctx, done, nil); err != nil {
		t.Fatalf("SetTaskMilestone(nil): %v", err)
	}
	if err := repo.DeleteMilestoneV2(ctx, first, testutil.OwnerID); err != nil {
		t.Fatalf("DeleteMilestoneV2: %v", err)
	}
	links, err = repo.TaskMilestoneIDs(ctx, []int64{overdue, done})
	if err != nil {
		t.Fatalf("TaskMilestoneIDs: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("TaskMilestoneIDs after clear and delete = %v; want none", links)
	}
}

func milestoneCounts(t *testing.T, repo interface {
	ListMilestonesV2(context.Context, int64, *int64, int64, bool, int64, int64) ([]domain.Milestone, int64, error)
}, ctx context.Context, projID int64) map[int64]domain.Milestone {
	t.Helper()
	milestones, _, err := repo.ListMilestonesV2(ctx, testutil.BusinessID, &projID, testutil.OwnerID, true, 50, 0)
	if err != nil {
		t.Fatalf("ListMilestonesV2: %v", err)
	}
	out := make(map[int64]domain.Milestone, len(milestones))
	for _, m := range milestones {
		out[m.ID] = domain.Milestone{TaskCount: m.TaskCount, DoneCount: m.DoneCount, OverdueCount: m.OverdueCount}
	}
	return out
}
