package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetMyTasksForBusinessDoesNotCrossTenant(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()
	assignee := int64(930001)

	type tenantRows struct {
		workspaceID int64
		projectID   int64
		taskID      int64
	}
	created := make([]tenantRows, 0, 2)
	for i, businessID := range []int64{testutil.BusinessID, testutil.BusinessID + 1000} {
		workspaceID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
			OwnerID: testutil.OwnerID, BusinessID: businessID,
			Name: "my-tasks-business-scope", CreatedBy: testutil.OwnerID,
		}, false)
		if err != nil {
			t.Fatalf("CreateWorkspace %d: %v", i, err)
		}
		projectID, err := repo.CreateProject(ctx, domain.ProjectInput{
			WorkspaceID: workspaceID, OwnerID: testutil.OwnerID, BusinessID: businessID,
			Name: "my-tasks-business-scope", CreatedBy: testutil.OwnerID,
		})
		if err != nil {
			t.Fatalf("CreateProject %d: %v", i, err)
		}
		taskID, err := repo.CreateTask(ctx, domain.TaskInput{
			ProjectID: projectID, WorkspaceID: workspaceID, OwnerID: testutil.OwnerID,
			BusinessID: businessID, Title: "tenant task", ReporterID: testutil.OwnerID,
			AssigneeID: &assignee,
		})
		if err != nil {
			t.Fatalf("CreateTask %d: %v", i, err)
		}
		created = append(created, tenantRows{workspaceID, projectID, taskID})
	}
	t.Cleanup(func() {
		for _, row := range created {
			_, _ = db.ExecContext(ctx, "DELETE FROM flow_tasks WHERE id = ?", row.taskID)
			_, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", row.projectID)
			_, _ = db.ExecContext(ctx, "DELETE FROM flow_workspaces WHERE id = ?", row.workspaceID)
		}
	})

	rows, total, err := repo.GetMyTasksForBusiness(ctx, assignee, testutil.BusinessID, 0, 10, 0)
	if err != nil {
		t.Fatalf("GetMyTasksForBusiness: %v", err)
	}
	if total != 1 || len(rows) != 1 || rows[0].ID != created[0].taskID {
		t.Fatalf("business-scoped rows = %+v (total %d), want only task %d", rows, total, created[0].taskID)
	}
}
