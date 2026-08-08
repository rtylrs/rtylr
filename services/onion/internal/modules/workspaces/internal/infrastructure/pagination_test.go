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
		OwnerID: testutil.OwnerID, Name: "ProjectPageWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.OwnerID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "ProjectPage Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	for i := 0; i < 5; i++ {
		taskID, err := repo.CreateTask(ctx, domain.TaskInput{
			ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID,
			Title: fmt.Sprintf("ProjectPage Task %d", i), ReporterID: testutil.OwnerID,
		})
		if err != nil {
			t.Fatalf("CreateTask %d: %v", i, err)
		}
		t.Cleanup(func(id int64) func() { return func() { _ = repo.DeleteTask(ctx, id, testutil.OwnerID) } }(taskID))
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
		OwnerID: testutil.OwnerID, Name: "MyTasksPageWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.OwnerID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "MyTasksPage Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	// Sentinel assignee ID unlikely to collide with tasks created by other tests.
	const assignee = int64(919191919)
	for i := 0; i < 5; i++ {
		taskID, err := repo.CreateTask(ctx, domain.TaskInput{
			ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID,
			Title: fmt.Sprintf("MyTasksPage Task %d", i), ReporterID: testutil.OwnerID,
			AssigneeID: int64Ptr(assignee),
		})
		if err != nil {
			t.Fatalf("CreateTask %d: %v", i, err)
		}
		t.Cleanup(func(id int64) func() { return func() { _ = repo.DeleteTask(ctx, id, testutil.OwnerID) } }(taskID))
	}

	page1, total1, err := repo.GetMyTasks(ctx, assignee, 2, 0)
	if err != nil {
		t.Fatalf("GetMyTasks page1: %v", err)
	}
	if total1 != 5 {
		t.Errorf("expected total=5, got %d", total1)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 tasks on page1, got %d", len(page1))
	}

	page2, total2, err := repo.GetMyTasks(ctx, assignee, 2, 2)
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

	page3, total3, err := repo.GetMyTasks(ctx, assignee, 2, 4)
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

func TestWorkspaces_SearchTasks_PaginationAndTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, Name: "SearchPageWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.OwnerID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "SearchPage Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	const searchToken = "zzqsearchpaginationtoken"
	for i := 0; i < 5; i++ {
		taskID, err := repo.CreateTask(ctx, domain.TaskInput{
			ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID,
			Title:      fmt.Sprintf("%s task %d", searchToken, i),
			ReporterID: testutil.OwnerID,
		})
		if err != nil {
			t.Fatalf("CreateTask %d: %v", i, err)
		}
		t.Cleanup(func(id int64) func() { return func() { _ = repo.DeleteTask(ctx, id, testutil.OwnerID) } }(taskID))
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
		OwnerID: testutil.OwnerID, Name: "SubtaskPageWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.OwnerID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "SubtaskPage Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	parentID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID,
		Title: "Subtask Page Parent", ReporterID: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask parent: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(ctx, parentID, testutil.OwnerID) })

	for i := 0; i < 5; i++ {
		subtaskID, err := repo.CreateTask(ctx, domain.TaskInput{
			ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID,
			Title: fmt.Sprintf("Subtask Page Child %d", i), ReporterID: testutil.OwnerID,
			ParentTaskID: int64Ptr(parentID),
		})
		if err != nil {
			t.Fatalf("CreateTask child %d: %v", i, err)
		}
		t.Cleanup(func(id int64) func() { return func() { _ = repo.DeleteTask(ctx, id, testutil.OwnerID) } }(subtaskID))
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
