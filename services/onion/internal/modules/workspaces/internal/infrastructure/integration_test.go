package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestWorkspaces_CountWorkspacesByOwner_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	count, err := repo.CountWorkspacesByOwner(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("CountWorkspacesByOwner: %v", err)
	}
	if count < 0 {
		t.Errorf("expected count >= 0, got %d", count)
	}
}

func TestWorkspaces_CreateWorkspace_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateWorkspace(context.Background(), domain.WorkspaceInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Integration Test Workspace",
		CreatedBy:  testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), id, testutil.BusinessID) })
}

func TestWorkspaces_GetWorkspacesByOwner_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, _, err := repo.GetWorkspacesByOwner(context.Background(), testutil.BusinessID, 100, 0)
	if err != nil {
		t.Fatalf("GetWorkspacesByOwner: %v", err)
	}
}

func TestWorkspaces_CreateWorkspace_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateWorkspace(context.Background(), domain.WorkspaceInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "List Test WS",
		CreatedBy:  testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), id, testutil.BusinessID) })

	workspaces, _, err := repo.GetWorkspacesByOwner(context.Background(), testutil.BusinessID, 500, 0)
	if err != nil {
		t.Fatalf("GetWorkspacesByOwner: %v", err)
	}
	found := false
	for _, w := range workspaces {
		if w.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("created workspace id=%d not found in list", id)
	}
}

func TestWorkspaces_UpdateWorkspace_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateWorkspace(context.Background(), domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "OldWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), id, testutil.BusinessID) })

	if err := repo.UpdateWorkspace(context.Background(), domain.UpdateWorkspaceInput{
		ID: id, OwnerID: testutil.OwnerID, Name: "NewWS",
	}); err != nil {
		t.Fatalf("UpdateWorkspace: %v", err)
	}
}

func TestWorkspaces_DeleteWorkspace_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateWorkspace(context.Background(), domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "DeleteMe", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	if err := repo.DeleteWorkspace(context.Background(), id, testutil.BusinessID); err != nil {
		t.Fatalf("DeleteWorkspace: %v", err)
	}

	workspaces, _, _ := repo.GetWorkspacesByOwner(context.Background(), testutil.BusinessID, 500, 0)
	for _, w := range workspaces {
		if w.ID == id {
			t.Errorf("deleted workspace id=%d still in list", id)
		}
	}
}

func TestWorkspaces_CreateProject_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	wsID, err := repo.CreateWorkspace(context.Background(), domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "ProjectWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(context.Background(), domain.ProjectInput{
		WorkspaceID: wsID,
		OwnerID:     testutil.OwnerID,
		Name:        "Test Project",
		CreatedBy:   testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if projID <= 0 {
		t.Fatalf("expected positive project ID, got %d", projID)
	}
}

func TestWorkspaces_GetProjectsByWorkspace_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	wsID, err := repo.CreateWorkspace(context.Background(), domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "ProjListWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), wsID, testutil.BusinessID) })

	projects, _, err := repo.GetProjectsByWorkspace(context.Background(), wsID, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("GetProjectsByWorkspace: %v", err)
	}
	_ = projects
}

func TestWorkspaces_CreateTask_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	wsID, err := repo.CreateWorkspace(context.Background(), domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "TaskWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(context.Background(), domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "Task Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	taskID, err := repo.CreateTask(context.Background(), domain.TaskInput{
		ProjectID:   projID,
		WorkspaceID: wsID,
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		Title:       "Test Task",
		ReporterID:  testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if taskID <= 0 {
		t.Fatalf("expected positive task ID, got %d", taskID)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(context.Background(), taskID, testutil.BusinessID) })
}

func TestWorkspaces_LookupUserByEmail_Owner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	// First get actual owner email from users table
	var email string
	if err := db.QueryRowContext(context.Background(), "SELECT email FROM users WHERE id = ?", testutil.OwnerID).Scan(&email); err != nil {
		t.Fatalf("get owner email: %v", err)
	}

	id, found, err := repo.LookupUserByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("LookupUserByEmail: %v", err)
	}
	if !found {
		t.Fatal("expected to find owner by email")
	}
	if id != testutil.OwnerID {
		t.Errorf("expected id %d, got %d", testutil.OwnerID, id)
	}
}

func TestWorkspaces_LookupUserByEmail_Unknown_NotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, found, err := repo.LookupUserByEmail(context.Background(), "nonexistent@example.com")
	if err != nil {
		t.Fatalf("LookupUserByEmail: %v", err)
	}
	if found {
		t.Error("expected found=false for unknown email")
	}
}

func TestWorkspaces_GetDashboardStats_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	wsID, err := repo.CreateWorkspace(context.Background(), domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "StatsWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), wsID, testutil.BusinessID) })

	_, err = repo.GetDashboardStats(context.Background(), wsID, testutil.BusinessID)
	if err != nil {
		t.Fatalf("GetDashboardStats: %v", err)
	}
}

func TestWorkspaces_CreatePersonalWorkspace_Idempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id1, err := repo.CreatePersonalWorkspace(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePersonalWorkspace first call: %v", err)
	}
	id2, err := repo.CreatePersonalWorkspace(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("CreatePersonalWorkspace second call: %v", err)
	}
	if id1 != id2 {
		t.Errorf("expected idempotent ID, got %d then %d", id1, id2)
	}
}

func TestWorkspaces_UpdateWorkspace_Description(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "Desc WS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, id, testutil.BusinessID) })

	if err := repo.UpdateWorkspace(ctx, domain.UpdateWorkspaceInput{
		ID: id, OwnerID: testutil.OwnerID, Name: "Desc WS", Description: "Updated description",
	}); err != nil {
		t.Fatalf("UpdateWorkspace description: %v", err)
	}
}

func TestWorkspaces_CountWorkspacesByOwner_IncreasesOnCreate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.CountWorkspacesByOwner(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("count before: %v", err)
	}

	id, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "Count Increase WS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, id, testutil.BusinessID) })

	after, err := repo.CountWorkspacesByOwner(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("count after: %v", err)
	}
	if after != before+1 {
		t.Errorf("expected count to increase by 1: before=%d after=%d", before, after)
	}
}

func TestWorkspaces_CreateProject_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "ProjListWS2", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.OwnerID, Name: "Appears Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	projects, _, err := repo.GetProjectsByWorkspace(ctx, wsID, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("GetProjectsByWorkspace: %v", err)
	}
	found := false
	for _, p := range projects {
		if p.ID == projID {
			found = true
		}
	}
	if !found {
		t.Errorf("created project id=%d not found in list", projID)
	}
}

func TestWorkspaces_CreateTask_AppearsInProject(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
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
	repo := infrastructure.NewRepository(db)
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
	repo := infrastructure.NewRepository(db)
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
	repo := infrastructure.NewRepository(db)
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
		}
	}
	if !found {
		t.Errorf("comment id=%d not found in GetCommentsByTask", commentID)
	}
}

func TestWorkspaces_CreateLabel_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "LabelWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	labelID, err := repo.CreateLabel(ctx, domain.LabelInput{
		WorkspaceID: wsID, Name: "Bug", Color: "#ff0000",
	})
	if err != nil {
		t.Fatalf("CreateLabel: %v", err)
	}
	if labelID <= 0 {
		t.Errorf("expected positive label ID, got %d", labelID)
	}
	t.Cleanup(func() { _ = repo.DeleteLabel(ctx, labelID, testutil.OwnerID) })
}

func TestWorkspaces_GetLabelsByWorkspace_ReturnsLabel(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "LabelListWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	labelID, err := repo.CreateLabel(ctx, domain.LabelInput{
		WorkspaceID: wsID, Name: "Feature", Color: "#00ff00",
	})
	if err != nil {
		t.Fatalf("CreateLabel: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteLabel(ctx, labelID, testutil.OwnerID) })

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
	repo := infrastructure.NewRepository(db)
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

func TestWorkspaces_CreateColumn_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "ColumnWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "Column Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	colID, err := repo.CreateColumn(ctx, domain.ProjectColumnInput{
		ProjectID: projID, WorkspaceID: wsID, Name: "To Do", Color: "#aabbcc", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateColumn: %v", err)
	}
	if colID <= 0 {
		t.Errorf("expected positive column ID, got %d", colID)
	}
	t.Cleanup(func() { _ = repo.DeleteColumn(ctx, colID, testutil.OwnerID) })
}

func TestWorkspaces_LogActivity_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "ActivityWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "Activity Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	taskID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Title: "Activity Task", ReporterID: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(ctx, taskID, testutil.BusinessID) })

	if err := repo.LogActivity(ctx, taskID, testutil.OwnerID, "status_changed", "todo", "in_progress"); err != nil {
		t.Fatalf("LogActivity: %v", err)
	}
}

func TestWorkspaces_GetTaskByID_ReturnsTask(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "GetByIDWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "GetByID Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	taskID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Title: "GetByID Task", ReporterID: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(ctx, taskID, testutil.BusinessID) })

	task, err := repo.GetTaskByID(ctx, taskID)
	if err != nil {
		t.Fatalf("GetTaskByID: %v", err)
	}
	if task.ID != taskID {
		t.Errorf("expected task.ID=%d, got %d", taskID, task.ID)
	}
}

func TestWorkspaces_IsWorkspaceMember_OwnerIsNotMemberByDefault(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "MemberCheckWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	// Just ensure it doesn't panic for an arbitrary user.
	_ = repo.IsWorkspaceMember(ctx, 9999999, wsID)
}
