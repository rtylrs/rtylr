package infrastructure_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestWorkspaces_GetTasksByProject_PaginationAndTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "ProjectPageWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "ProjectPage Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	for i := 0; i < 5; i++ {
		taskID, err := repo.CreateTask(ctx, domain.TaskInput{
			ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
			Title: fmt.Sprintf("ProjectPage Task %d", i), ReporterID: testutil.OwnerID,
		})
		if err != nil {
			t.Fatalf("CreateTask %d: %v", i, err)
		}
		t.Cleanup(func(id int64) func() { return func() { _ = repo.DeleteTask(ctx, id, testutil.BusinessID) } }(taskID))
	}

	page1, total1, err := repo.GetTasksByProject(ctx, domain.ListTasksFilter{ProjectID: projID, WorkspaceID: wsID, Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("GetTasksByProject page1: %v", err)
	}
	if total1 != 5 {
		t.Errorf("expected total=5, got %d", total1)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 tasks on page1, got %d", len(page1))
	}

	page2, total2, err := repo.GetTasksByProject(ctx, domain.ListTasksFilter{ProjectID: projID, WorkspaceID: wsID, Limit: 2, Offset: 2})
	if err != nil {
		t.Fatalf("GetTasksByProject page2: %v", err)
	}
	if total2 != 5 {
		t.Errorf("expected total=5 on page2, got %d", total2)
	}
	if len(page2) != 2 {
		t.Fatalf("expected 2 tasks on page2, got %d", len(page2))
	}
	for _, p1 := range page1 {
		for _, p2 := range page2 {
			if p1.ID == p2.ID {
				t.Errorf("task %d appeared on both page1 and page2", p1.ID)
			}
		}
	}

	page3, total3, err := repo.GetTasksByProject(ctx, domain.ListTasksFilter{ProjectID: projID, WorkspaceID: wsID, Limit: 2, Offset: 4})
	if err != nil {
		t.Fatalf("GetTasksByProject page3: %v", err)
	}
	if total3 != 5 {
		t.Errorf("expected total=5 on page3, got %d", total3)
	}
	if len(page3) != 1 {
		t.Errorf("expected 1 remainder task on page3, got %d", len(page3))
	}
}

func TestWorkspaces_GetMyTasks_PaginationAndTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "MyTasksPageWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "MyTasksPage Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	// Sentinel assignee ID unlikely to collide with tasks created by other tests.
	const assignee = int64(919191919)
	for i := 0; i < 5; i++ {
		taskID, err := repo.CreateTask(ctx, domain.TaskInput{
			ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
			Title: fmt.Sprintf("MyTasksPage Task %d", i), ReporterID: testutil.OwnerID,
			AssigneeID: int64Ptr(assignee),
		})
		if err != nil {
			t.Fatalf("CreateTask %d: %v", i, err)
		}
		t.Cleanup(func(id int64) func() { return func() { _ = repo.DeleteTask(ctx, id, testutil.BusinessID) } }(taskID))
	}

	page1, total1, err := repo.GetMyTasks(ctx, assignee, 0, 2, 0)
	if err != nil {
		t.Fatalf("GetMyTasks page1: %v", err)
	}
	if total1 != 5 {
		t.Errorf("expected total=5, got %d", total1)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 tasks on page1, got %d", len(page1))
	}

	page2, total2, err := repo.GetMyTasks(ctx, assignee, 0, 2, 2)
	if err != nil {
		t.Fatalf("GetMyTasks page2: %v", err)
	}
	if total2 != 5 {
		t.Errorf("expected total=5 on page2, got %d", total2)
	}
	if len(page2) != 2 {
		t.Fatalf("expected 2 tasks on page2, got %d", len(page2))
	}
	for _, p1 := range page1 {
		for _, p2 := range page2 {
			if p1.ID == p2.ID {
				t.Errorf("task %d appeared on both page1 and page2", p1.ID)
			}
		}
	}

	page3, total3, err := repo.GetMyTasks(ctx, assignee, 0, 2, 4)
	if err != nil {
		t.Fatalf("GetMyTasks page3: %v", err)
	}
	if total3 != 5 {
		t.Errorf("expected total=5 on page3, got %d", total3)
	}
	if len(page3) != 1 {
		t.Errorf("expected 1 remainder task on page3, got %d", len(page3))
	}
}

// TestWorkspaces_GetMyTasks_ScopedByWorkspace guards against regressing to the
// bug in rtylr issue #783: the dashboard's "My tasks" widget called GetMyTasks
// without a workspace filter, so it rendered tasks the same user had assigned
// to them in a completely different workspace. GetMyTasks must exclude a task
// from workspace B when queried with workspace A's ID, while workspaceID=0
// (the cross-workspace "My tasks" board) must still return both.
func TestWorkspaces_GetMyTasks_ScopedByWorkspace(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Sentinel assignee ID unlikely to collide with tasks created by other tests.
	const assignee = int64(929292929)

	wsA, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "ScopedMyTasksWS-A", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace A: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsA, testutil.BusinessID) })

	wsB, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "ScopedMyTasksWS-B", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace B: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsB, testutil.BusinessID) })

	projA, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsA, OwnerID: testutil.OwnerID, Name: "ScopedMyTasks Project A", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject A: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projA) })

	projB, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsB, OwnerID: testutil.OwnerID, Name: "ScopedMyTasks Project B", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject B: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projB) })

	taskA, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projA, WorkspaceID: wsA, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Title: "ScopedMyTasks Task A", ReporterID: testutil.OwnerID,
		AssigneeID: int64Ptr(assignee),
	})
	if err != nil {
		t.Fatalf("CreateTask A: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(ctx, taskA, testutil.BusinessID) })

	taskB, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projB, WorkspaceID: wsB, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Title: "ScopedMyTasks Task B", ReporterID: testutil.OwnerID,
		AssigneeID: int64Ptr(assignee),
	})
	if err != nil {
		t.Fatalf("CreateTask B: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(ctx, taskB, testutil.BusinessID) })

	scopedToA, totalA, err := repo.GetMyTasks(ctx, assignee, wsA, 10, 0)
	if err != nil {
		t.Fatalf("GetMyTasks scoped to A: %v", err)
	}
	if totalA != 1 {
		t.Fatalf("expected 1 task when scoped to workspace A, got %d", totalA)
	}
	if len(scopedToA) != 1 || scopedToA[0].ID != taskA {
		t.Errorf("expected only task A when scoped to workspace A, got %+v", scopedToA)
	}

	scopedToB, totalB, err := repo.GetMyTasks(ctx, assignee, wsB, 10, 0)
	if err != nil {
		t.Fatalf("GetMyTasks scoped to B: %v", err)
	}
	if totalB != 1 {
		t.Fatalf("expected 1 task when scoped to workspace B, got %d", totalB)
	}
	if len(scopedToB) != 1 || scopedToB[0].ID != taskB {
		t.Errorf("expected only task B when scoped to workspace B, got %+v", scopedToB)
	}

	unscoped, totalAll, err := repo.GetMyTasks(ctx, assignee, 0, 10, 0)
	if err != nil {
		t.Fatalf("GetMyTasks unscoped: %v", err)
	}
	if totalAll != 2 {
		t.Fatalf("expected 2 tasks across both workspaces when unscoped, got %d", totalAll)
	}
	if len(unscoped) != 2 {
		t.Errorf("expected both tasks when unscoped, got %+v", unscoped)
	}
}

func TestWorkspaces_SearchTasks_PaginationAndTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "SearchPageWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "SearchPage Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	const searchToken = "zzqsearchpaginationtoken"
	for i := 0; i < 5; i++ {
		taskID, err := repo.CreateTask(ctx, domain.TaskInput{
			ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
			Title:      fmt.Sprintf("%s task %d", searchToken, i),
			ReporterID: testutil.OwnerID,
		})
		if err != nil {
			t.Fatalf("CreateTask %d: %v", i, err)
		}
		t.Cleanup(func(id int64) func() { return func() { _ = repo.DeleteTask(ctx, id, testutil.BusinessID) } }(taskID))
	}

	page1, total1, err := repo.SearchTasks(ctx, wsID, searchToken, 2, 0)
	if err != nil {
		t.Fatalf("SearchTasks page1: %v", err)
	}
	if total1 != 5 {
		t.Errorf("expected total=5, got %d", total1)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 tasks on page1, got %d", len(page1))
	}

	page2, total2, err := repo.SearchTasks(ctx, wsID, searchToken, 2, 2)
	if err != nil {
		t.Fatalf("SearchTasks page2: %v", err)
	}
	if total2 != 5 {
		t.Errorf("expected total=5 on page2, got %d", total2)
	}
	if len(page2) != 2 {
		t.Fatalf("expected 2 tasks on page2, got %d", len(page2))
	}
	for _, p1 := range page1 {
		for _, p2 := range page2 {
			if p1.ID == p2.ID {
				t.Errorf("task %d appeared on both page1 and page2", p1.ID)
			}
		}
	}

	page3, total3, err := repo.SearchTasks(ctx, wsID, searchToken, 2, 4)
	if err != nil {
		t.Fatalf("SearchTasks page3: %v", err)
	}
	if total3 != 5 {
		t.Errorf("expected total=5 on page3, got %d", total3)
	}
	if len(page3) != 1 {
		t.Errorf("expected 1 remainder task on page3, got %d", len(page3))
	}
}

func TestWorkspaces_GetSubtasks_PaginationAndTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "SubtaskPageWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "SubtaskPage Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	parentID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Title: "Subtask Page Parent", ReporterID: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask parent: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(ctx, parentID, testutil.BusinessID) })

	for i := 0; i < 5; i++ {
		subtaskID, err := repo.CreateTask(ctx, domain.TaskInput{
			ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
			Title: fmt.Sprintf("Subtask Page Child %d", i), ReporterID: testutil.OwnerID,
			ParentTaskID: int64Ptr(parentID),
		})
		if err != nil {
			t.Fatalf("CreateTask child %d: %v", i, err)
		}
		t.Cleanup(func(id int64) func() { return func() { _ = repo.DeleteTask(ctx, id, testutil.BusinessID) } }(subtaskID))
	}

	page1, total1, err := repo.GetSubtasks(ctx, parentID, 2, 0)
	if err != nil {
		t.Fatalf("GetSubtasks page1: %v", err)
	}
	if total1 != 5 {
		t.Errorf("expected total=5, got %d", total1)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 subtasks on page1, got %d", len(page1))
	}

	page2, total2, err := repo.GetSubtasks(ctx, parentID, 2, 2)
	if err != nil {
		t.Fatalf("GetSubtasks page2: %v", err)
	}
	if total2 != 5 {
		t.Errorf("expected total=5 on page2, got %d", total2)
	}
	if len(page2) != 2 {
		t.Fatalf("expected 2 subtasks on page2, got %d", len(page2))
	}
	for _, p1 := range page1 {
		for _, p2 := range page2 {
			if p1.ID == p2.ID {
				t.Errorf("subtask %d appeared on both page1 and page2", p1.ID)
			}
		}
	}

	page3, total3, err := repo.GetSubtasks(ctx, parentID, 2, 4)
	if err != nil {
		t.Fatalf("GetSubtasks page3: %v", err)
	}
	if total3 != 5 {
		t.Errorf("expected total=5 on page3, got %d", total3)
	}
	if len(page3) != 1 {
		t.Errorf("expected 1 remainder subtask on page3, got %d", len(page3))
	}
}

func int64Ptr(id int64) *int64 { return &id }
