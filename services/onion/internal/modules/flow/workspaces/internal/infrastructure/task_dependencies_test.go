package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

func TestWorkspaces_AddTaskDependency_IsIdempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "DepIdempotent")
	blockerID, _ := seedArchivedTestTask(t, db, repo, "DepIdempotentBlocker")
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_task_dependencies WHERE task_id = ?", taskID)
	})

	in := domain.AddTaskDependencyInput{TaskID: taskID, DependsOnTaskID: blockerID, BusinessID: businessID}
	// The unique key is what makes a repeat add a no-op rather than a duplicate
	// row, so adding twice must leave exactly one edge.
	for i := 0; i < 2; i++ {
		if err := repo.AddTaskDependency(ctx, in); err != nil {
			t.Fatalf("add %d: %v", i, err)
		}
	}

	deps, err := repo.ListTaskDependencies(ctx, []int64{taskID}, businessID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(deps[taskID]) != 1 || deps[taskID][0] != blockerID {
		t.Fatalf("dependencies = %v, want exactly [%d]", deps[taskID], blockerID)
	}
}

func TestWorkspaces_RemoveTaskDependency(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "DepRemove")
	blockerID, _ := seedArchivedTestTask(t, db, repo, "DepRemoveBlocker")
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_task_dependencies WHERE task_id = ?", taskID)
	})

	if err := repo.AddTaskDependency(ctx, domain.AddTaskDependencyInput{
		TaskID: taskID, DependsOnTaskID: blockerID, BusinessID: businessID,
	}); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := repo.RemoveTaskDependency(ctx, domain.RemoveTaskDependencyInput{
		TaskID: taskID, DependsOnTaskID: blockerID, BusinessID: businessID,
	}); err != nil {
		t.Fatalf("remove: %v", err)
	}

	deps, err := repo.ListTaskDependencies(ctx, []int64{taskID}, businessID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(deps[taskID]) != 0 {
		t.Fatalf("dependencies = %v, want empty", deps[taskID])
	}
}

func TestWorkspaces_ListTaskDependencies_ScopedToBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "DepTenancy")
	blockerID, _ := seedArchivedTestTask(t, db, repo, "DepTenancyBlocker")
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_task_dependencies WHERE task_id = ?", taskID)
	})

	if err := repo.AddTaskDependency(ctx, domain.AddTaskDependencyInput{
		TaskID: taskID, DependsOnTaskID: blockerID, BusinessID: businessID,
	}); err != nil {
		t.Fatalf("add: %v", err)
	}

	// Every other Flow query is tenant scoped; a dependency list that ignores
	// business_id would be a cross-tenant read.
	deps, err := repo.ListTaskDependencies(ctx, []int64{taskID}, businessID+999)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(deps[taskID]) != 0 {
		t.Fatalf("dependencies leaked across businesses: %v", deps[taskID])
	}
}
