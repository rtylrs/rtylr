package infrastructure_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

func TestMoveTaskToColumnRejectsColumnFromAnotherProject(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Name: "cross-project-move-ws", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_workspaces WHERE id = ?", wsID) })

	projectA, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Name: "cross-project-move-a", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject A: %v", err)
	}
	projectB, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Name: "cross-project-move-b", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject B: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id IN (?, ?)", projectA, projectB)
	})

	columnA, err := repo.CreateColumn(ctx, domain.ProjectColumnInput{
		ProjectID: projectA, WorkspaceID: 0, Name: "A column", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateColumn A: %v", err)
	}
	columnB, err := repo.CreateColumn(ctx, domain.ProjectColumnInput{
		ProjectID: projectB, WorkspaceID: 0, Name: "B column", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateColumn B: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_project_columns WHERE id IN (?, ?)", columnA, columnB)
	})

	taskID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projectA, WorkspaceID: 0, OwnerID: testutil.OwnerID,
		BusinessID: testutil.BusinessID, Title: "keep on A", ReporterID: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if _, err := db.ExecContext(ctx, "UPDATE flow_tasks SET section_id = ? WHERE id = ?", columnA, taskID); err != nil {
		t.Fatalf("seed task section: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(ctx, taskID, testutil.BusinessID) })

	if err := repo.MoveTaskToColumn(ctx, taskID, columnB, testutil.OwnerID); err == nil {
		t.Fatal("moving a task to a column from another project unexpectedly succeeded")
	}

	var section sql.NullInt64
	if err := db.QueryRowContext(ctx, "SELECT section_id FROM flow_tasks WHERE id = ?", taskID).Scan(&section); err != nil {
		t.Fatalf("read task section: %v", err)
	}
	if !section.Valid || section.Int64 != columnA {
		t.Fatalf("section_id = %+v after rejected move, want %d", section, columnA)
	}
}
