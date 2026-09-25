package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

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
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), "DELETE FROM flow_projects WHERE id = ?", projID) })
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
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), "DELETE FROM flow_projects WHERE id = ?", projID) })

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
	id1, err := repo.CreatePersonalWorkspace(context.Background(), testutil.OwnerID, testutil.BusinessID)
	if err != nil {
		t.Fatalf("CreatePersonalWorkspace first call: %v", err)
	}
	id2, err := repo.CreatePersonalWorkspace(context.Background(), testutil.OwnerID, testutil.BusinessID)
	if err != nil {
		t.Fatalf("CreatePersonalWorkspace second call: %v", err)
	}
	if id1 != id2 {
		t.Errorf("expected idempotent ID, got %d then %d", id1, id2)
	}

	// business_id used to be omitted from the INSERT entirely, leaving every
	// personal workspace stuck at the schema's DEFAULT 0. Confirm it's
	// findable by its real business, not just present with some value.
	workspaces, _, err := repo.GetWorkspacesByOwner(context.Background(), testutil.BusinessID, 50, 0)
	if err != nil {
		t.Fatalf("GetWorkspacesByOwner: %v", err)
	}
	found := false
	for _, w := range workspaces {
		if w.ID == id1 {
			found = true
		}
	}
	if !found {
		t.Errorf("personal workspace %d not found when listed by its real business id %d - business_id was likely left at 0",
			id1, testutil.BusinessID)
	}
}

// The idempotency lookup behind CreatePersonalWorkspace read owner_id only, so
// a user who belongs to two businesses got the first business's personal
// workspace handed back for the second, and the second business never got one.
// That is also why the test above failed intermittently: any earlier test that
// left a personal or default workspace for the same owner under a different
// business won the lookup.
//
// Seeds the other tenant first, deliberately, so an unscoped lookup would find
// it before the caller's own.
func TestWorkspaces_CreatePersonalWorkspace_IsPerBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const sharedOwner int64 = 970100

	theirs, err := repo.CreatePersonalWorkspace(ctx, sharedOwner, testutil.OtherBusinessID)
	if err != nil {
		t.Fatalf("CreatePersonalWorkspace for the other business: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM flow_workspaces WHERE id = ?", theirs)
	})

	mine, err := repo.CreatePersonalWorkspace(ctx, sharedOwner, testutil.BusinessID)
	if err != nil {
		t.Fatalf("CreatePersonalWorkspace for this business: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM flow_workspaces WHERE id = ?", mine)
	})

	if mine == theirs {
		t.Fatalf("returned the other business's workspace %d instead of creating one for business %d",
			theirs, testutil.BusinessID)
	}

	// And it must still be idempotent within a business.
	again, err := repo.CreatePersonalWorkspace(ctx, sharedOwner, testutil.BusinessID)
	if err != nil {
		t.Fatalf("CreatePersonalWorkspace second call: %v", err)
	}
	if again != mine {
		t.Errorf("not idempotent within a business: got %d then %d", mine, again)
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
	const countTestBusinessID int64 = 881001

	before, err := repo.CountWorkspacesByOwner(ctx, countTestBusinessID)
	if err != nil {
		t.Fatalf("count before: %v", err)
	}

	id, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: countTestBusinessID, Name: "Count Increase WS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, id, countTestBusinessID) })

	after, err := repo.CountWorkspacesByOwner(ctx, countTestBusinessID)
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
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

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
