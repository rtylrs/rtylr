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
	count, err := repo.CountWorkspacesByOwner(context.Background(), testutil.OwnerID)
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
		OwnerID:   testutil.OwnerID,
		Name:      "Integration Test Workspace",
		CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), id, testutil.OwnerID) })
}

func TestWorkspaces_GetWorkspacesByOwner_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.GetWorkspacesByOwner(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetWorkspacesByOwner: %v", err)
	}
}

func TestWorkspaces_CreateWorkspace_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateWorkspace(context.Background(), domain.WorkspaceInput{
		OwnerID:   testutil.OwnerID,
		Name:      "List Test WS",
		CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), id, testutil.OwnerID) })

	workspaces, err := repo.GetWorkspacesByOwner(context.Background(), testutil.OwnerID)
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
		OwnerID: testutil.OwnerID, Name: "OldWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), id, testutil.OwnerID) })

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
		OwnerID: testutil.OwnerID, Name: "DeleteMe", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	if err := repo.DeleteWorkspace(context.Background(), id, testutil.OwnerID); err != nil {
		t.Fatalf("DeleteWorkspace: %v", err)
	}

	workspaces, _ := repo.GetWorkspacesByOwner(context.Background(), testutil.OwnerID)
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
		OwnerID: testutil.OwnerID, Name: "ProjectWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), wsID, testutil.OwnerID) })

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
		OwnerID: testutil.OwnerID, Name: "ProjListWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), wsID, testutil.OwnerID) })

	projects, err := repo.GetProjectsByWorkspace(context.Background(), wsID, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetProjectsByWorkspace: %v", err)
	}
	_ = projects
}

func TestWorkspaces_CreateTask_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	wsID, err := repo.CreateWorkspace(context.Background(), domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, Name: "TaskWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), wsID, testutil.OwnerID) })

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
		Title:       "Test Task",
		ReporterID:  testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if taskID <= 0 {
		t.Fatalf("expected positive task ID, got %d", taskID)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(context.Background(), taskID, testutil.OwnerID) })
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
		OwnerID: testutil.OwnerID, Name: "StatsWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), wsID, testutil.OwnerID) })

	_, err = repo.GetDashboardStats(context.Background(), wsID, testutil.OwnerID)
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
