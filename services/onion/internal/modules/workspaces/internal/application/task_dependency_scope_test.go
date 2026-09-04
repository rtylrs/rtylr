package application_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// The repository tests for dependencies pass BusinessID in themselves, so they
// exercised the repo correctly and never touched the service, which omitted it.
// Every edge was therefore written with business_id = 0 while
// ListTaskDependencies filters on the caller's real business: adding a
// dependency succeeded and it never appeared anywhere.
//
// This test goes through the service for exactly that reason.
func TestAddTaskDependency_ThroughService_IsReadableBack(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	svc := application.NewService(context.Background(), db)
	ctx := context.Background()

	const owner int64 = 970901
	businessID := testutil.BusinessID

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		OwnerID: owner, BusinessID: businessID, Name: "Dep Scope", CreatedBy: owner,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	mk := func(title string) int64 {
		id, err := repo.CreateTask(ctx, domain.TaskInput{
			ProjectID: projID, OwnerID: owner, BusinessID: businessID,
			Title: title, ReporterID: owner,
		})
		if err != nil {
			t.Fatalf("CreateTask(%s): %v", title, err)
		}
		return id
	}
	taskID := mk("blocked")
	blockerID := mk("blocker")
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_task_dependencies WHERE task_id = ?", taskID)
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_tasks WHERE project_id = ?", projID)
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID)
	})

	if err := svc.AddTaskDependency(ctx, taskID, blockerID, businessID); err != nil {
		t.Fatalf("AddTaskDependency: %v", err)
	}

	// The whole point: what the service wrote must be readable with the same
	// business id the caller has.
	deps, err := repo.ListTaskDependencies(ctx, []int64{taskID}, businessID)
	if err != nil {
		t.Fatalf("ListTaskDependencies: %v", err)
	}
	if len(deps[taskID]) != 1 || deps[taskID][0] != blockerID {
		t.Fatalf("dependency written by the service is not readable back: got %v want [%d]",
			deps[taskID], blockerID)
	}

	// And removal, through the service, clears it.
	if err := svc.RemoveTaskDependency(ctx, taskID, blockerID, businessID); err != nil {
		t.Fatalf("RemoveTaskDependency: %v", err)
	}
	deps, err = repo.ListTaskDependencies(ctx, []int64{taskID}, businessID)
	if err != nil {
		t.Fatalf("ListTaskDependencies after remove: %v", err)
	}
	if len(deps[taskID]) != 0 {
		t.Fatalf("dependency survived removal: %v", deps[taskID])
	}
}

func TestGetTaskDetail_IncludesDependencies(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	svc := application.NewService(context.Background(), db)
	ctx := context.Background()
	const owner int64 = 970903

	projectID, err := repo.CreateProject(ctx, domain.ProjectInput{
		OwnerID: owner, BusinessID: testutil.BusinessID, Name: "Detail dependencies", CreatedBy: owner,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	createTask := func(title string) int64 {
		taskID, createErr := repo.CreateTask(ctx, domain.TaskInput{
			ProjectID: projectID, OwnerID: owner, BusinessID: testutil.BusinessID,
			Title: title, ReporterID: owner,
		})
		if createErr != nil {
			t.Fatalf("CreateTask(%s): %v", title, createErr)
		}
		return taskID
	}
	taskID := createTask("blocked")
	blockerID := createTask("blocker")
	t.Cleanup(func() {
		if _, cleanupErr := db.ExecContext(ctx, "DELETE FROM flow_task_dependencies WHERE task_id = ?", taskID); cleanupErr != nil {
			t.Errorf("cleanup dependencies: %v", cleanupErr)
		}
		if _, cleanupErr := db.ExecContext(ctx, "DELETE FROM flow_tasks WHERE project_id = ?", projectID); cleanupErr != nil {
			t.Errorf("cleanup tasks: %v", cleanupErr)
		}
		if _, cleanupErr := db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projectID); cleanupErr != nil {
			t.Errorf("cleanup project: %v", cleanupErr)
		}
	})

	if err := svc.AddTaskDependency(ctx, taskID, blockerID, testutil.BusinessID); err != nil {
		t.Fatalf("AddTaskDependency: %v", err)
	}

	detail, err := svc.GetTaskDetail(ctx, taskID)
	if err != nil {
		t.Fatalf("GetTaskDetail: %v", err)
	}
	if detail.Task.BusinessID != testutil.BusinessID {
		t.Fatalf("task business_id = %d, want %d", detail.Task.BusinessID, testutil.BusinessID)
	}
	if len(detail.Task.Dependencies) != 1 || detail.Task.Dependencies[0] != blockerID {
		t.Fatalf("task dependencies = %v, want [%d]", detail.Task.Dependencies, blockerID)
	}
}

// Ruling T2: dependencies are cross-project but within the business. The blocker
// is checked rather than trusted, so an id from another tenant is refused.
func TestAddTaskDependency_RefusesABlockerFromAnotherBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	svc := application.NewService(context.Background(), db)
	ctx := context.Background()

	const owner int64 = 970902
	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		OwnerID: owner, BusinessID: testutil.BusinessID, Name: "Dep Tenant", CreatedBy: owner,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	otherProjID, err := repo.CreateProject(ctx, domain.ProjectInput{
		OwnerID: testutil.OtherOwnerID, BusinessID: testutil.OtherBusinessID,
		Name: "Dep Other Tenant", CreatedBy: testutil.OtherOwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject(other): %v", err)
	}

	taskID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, OwnerID: owner, BusinessID: testutil.BusinessID,
		Title: "mine", ReporterID: owner,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	foreignID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: otherProjID, OwnerID: testutil.OtherOwnerID, BusinessID: testutil.OtherBusinessID,
		Title: "theirs", ReporterID: testutil.OtherOwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask(foreign): %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_task_dependencies WHERE task_id = ?", taskID)
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_tasks WHERE project_id IN (?, ?)", projID, otherProjID)
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id IN (?, ?)", projID, otherProjID)
	})

	err = svc.AddTaskDependency(ctx, taskID, foreignID, testutil.BusinessID)
	if err == nil {
		t.Fatal("a blocker from another business must be refused")
	}

	// A same-business blocker still works, so the refusal above is the tenant
	// check and not the call being broken outright.
	sibling, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, OwnerID: owner, BusinessID: testutil.BusinessID,
		Title: "sibling", ReporterID: owner,
	})
	if err != nil {
		t.Fatalf("CreateTask(sibling): %v", err)
	}
	if err := svc.AddTaskDependency(ctx, taskID, sibling, testutil.BusinessID); err != nil {
		t.Fatalf("same-business blocker should be accepted, got %v", err)
	}
}
