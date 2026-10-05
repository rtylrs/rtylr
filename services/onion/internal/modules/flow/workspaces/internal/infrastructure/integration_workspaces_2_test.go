package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

func TestWorkspaces_CreateTask_AppearsInProject(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "TaskListWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "Task List Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	taskID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Title: "Appears Task", ReporterID: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(ctx, taskID, testutil.BusinessID) })

	tasks, _, err := repo.GetTasksByProject(ctx, domain.ListTasksFilter{ProjectID: projID, WorkspaceID: wsID, Limit: 100})
	if err != nil {
		t.Fatalf("GetTasksByProject: %v", err)
	}
	found := false
	for _, tk := range tasks {
		if tk.ID == taskID {
			found = true
		}
	}
	if !found {
		t.Errorf("created task id=%d not found in GetTasksByProject", taskID)
	}
}

func TestWorkspaces_DeleteTask_RemovesFromProject(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "DelTaskWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "Del Task Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	taskID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Title: "Delete Me Task", ReporterID: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	if err := repo.DeleteTask(ctx, taskID, testutil.BusinessID); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}

	tasks, _, _ := repo.GetTasksByProject(ctx, domain.ListTasksFilter{ProjectID: projID, WorkspaceID: wsID, Limit: 100})
	for _, tk := range tasks {
		if tk.ID == taskID {
			t.Errorf("deleted task id=%d still in list", taskID)
		}
	}
}

func TestWorkspaces_CreateComment_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "CommentWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "Comment Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	taskID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Title: "Comment Task", ReporterID: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(ctx, taskID, testutil.BusinessID) })

	commentID, err := repo.CreateComment(ctx, taskID, testutil.OwnerID, "Integration test comment")
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}
	if commentID <= 0 {
		t.Errorf("expected positive comment ID, got %d", commentID)
	}
}

func TestWorkspaces_GetCommentsByTask_ReturnsComment(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "CommListWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "CommList Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	taskID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Title: "CommList Task", ReporterID: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(ctx, taskID, testutil.BusinessID) })

	commentID, err := repo.CreateComment(ctx, taskID, testutil.OwnerID, "Check me")
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}

	comments, err := repo.GetCommentsByTask(ctx, taskID)
	if err != nil {
		t.Fatalf("GetCommentsByTask: %v", err)
	}
	found := false
	for _, c := range comments {
		if c.ID == commentID {
			found = true
			if c.ActorName == "" || c.ActorInitials == "" {
				t.Errorf("comment actor identity is empty: name=%q initials=%q", c.ActorName, c.ActorInitials)
			}
		}
	}
	if !found {
		t.Errorf("comment id=%d not found in GetCommentsByTask", commentID)
	}

	if err := repo.LogActivity(ctx, taskID, testutil.OwnerID, "task_updated", "old", "new"); err != nil {
		t.Fatalf("LogActivity: %v", err)
	}
	activities, err := repo.GetActivityByTask(ctx, taskID, 10)
	if err != nil {
		t.Fatalf("GetActivityByTask: %v", err)
	}
	if len(activities) == 0 {
		t.Fatal("expected activity row")
	}
	if activities[0].ActorName == "" || activities[0].ActorInitials == "" {
		t.Fatalf("activity actor identity is empty: name=%q initials=%q", activities[0].ActorName, activities[0].ActorInitials)
	}
}

func TestWorkspaces_CreateLabel_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "LabelWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	labelID, err := repo.CreateLabel(ctx, domain.LabelInput{
		WorkspaceID: wsID, BusinessID: testutil.BusinessID, Name: "Bug", Color: "#ff0000",
	})
	if err != nil {
		t.Fatalf("CreateLabel: %v", err)
	}
	if labelID <= 0 {
		t.Errorf("expected positive label ID, got %d", labelID)
	}
	t.Cleanup(func() { _ = repo.DeleteLabel(ctx, labelID, testutil.OwnerID, testutil.BusinessID) })
}

func TestWorkspaces_GetLabelsByWorkspace_ReturnsLabel(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "LabelListWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	labelID, err := repo.CreateLabel(ctx, domain.LabelInput{
		WorkspaceID: wsID, BusinessID: testutil.BusinessID, Name: "Feature", Color: "#00ff00",
	})
	if err != nil {
		t.Fatalf("CreateLabel: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteLabel(ctx, labelID, testutil.OwnerID, testutil.BusinessID) })

	labels, _, err := repo.GetLabelsByWorkspace(ctx, wsID, 100, 0)
	if err != nil {
		t.Fatalf("GetLabelsByWorkspace: %v", err)
	}
	found := false
	for _, l := range labels {
		if l.ID == labelID {
			found = true
		}
	}
	if !found {
		t.Errorf("label id=%d not found in GetLabelsByWorkspace", labelID)
	}
}

func TestWorkspaces_UpdateTaskStatus_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "StatusWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "Status Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	taskID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Title: "Status Task", ReporterID: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(ctx, taskID, testutil.BusinessID) })

	if err := repo.UpdateTaskStatus(ctx, domain.UpdateTaskStatusInput{
		ID: taskID, OwnerID: testutil.OwnerID, Status: "in_progress",
	}); err != nil {
		t.Fatalf("UpdateTaskStatus: %v", err)
	}
}
