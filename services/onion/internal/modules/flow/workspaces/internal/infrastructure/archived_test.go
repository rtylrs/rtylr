package infrastructure_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// seedArchivedTestTask creates a throwaway workspace/project/task and returns
// the task id plus the business id the task was written under, so callers can
// scope UpdateTask the same way the API layer does.
func seedArchivedTestTask(t *testing.T, db *sql.DB, repo *infrastructure.Repository, name string) (taskID, businessID int64) {
	t.Helper()
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: name + " WS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: name + " Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	taskID, err = repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID,
		BusinessID: testutil.BusinessID, Title: name + " Task", ReporterID: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(ctx, taskID, testutil.BusinessID) })

	return taskID, testutil.BusinessID
}

func TestWorkspaces_UpdateTask_ArchivedTrueIsPersistedAndRead(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "ArchiveRoundTrip")

	created, err := repo.GetTaskByID(ctx, taskID)
	if err != nil {
		t.Fatalf("GetTaskByID before archive: %v", err)
	}
	if created.Archived {
		t.Fatalf("expected a freshly created task to be unarchived, got archived=true")
	}

	archived := true
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, OwnerID: testutil.OwnerID, BusinessID: businessID, Archived: &archived,
	}); err != nil {
		t.Fatalf("UpdateTask archiving task: %v", err)
	}

	got, err := repo.GetTaskByID(ctx, taskID)
	if err != nil {
		t.Fatalf("GetTaskByID after archive: %v", err)
	}
	if !got.Archived {
		t.Errorf("expected archived=true on the read model after archiving, got false")
	}
}

func TestWorkspaces_UpdateTask_OmittedArchivedKeepsExistingValue(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "ArchiveCoalesce")

	archived := true
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, OwnerID: testutil.OwnerID, BusinessID: businessID, Archived: &archived,
	}); err != nil {
		t.Fatalf("UpdateTask archiving task: %v", err)
	}

	// A partial update that omits archived entirely must leave it alone —
	// this is the COALESCE(?, archived) contract in the SQL update path.
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, OwnerID: testutil.OwnerID, BusinessID: businessID, Title: "Renamed but still archived",
	}); err != nil {
		t.Fatalf("UpdateTask omitting archived: %v", err)
	}

	got, err := repo.GetTaskByID(ctx, taskID)
	if err != nil {
		t.Fatalf("GetTaskByID after partial update: %v", err)
	}
	if got.Title != "Renamed but still archived" {
		t.Errorf("expected the partial update to apply the new title, got %q", got.Title)
	}
	if !got.Archived {
		t.Errorf("expected archived=true to survive an update that omitted archived, got false")
	}
}
