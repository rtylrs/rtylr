package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"
	"github.com/voxire/onion/internal/testutil"
)

// A task's number (T-n) used to be its global row id, so a new business's
// first task showed how many tasks every other business had made.
func TestTaskNumbersCountWithinTheBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	fresh := 7_000_000_000 + time.Now().UnixNano()%1_000_000_000
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_tasks WHERE business_id = ?", fresh)
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_task_sequences WHERE business_id = ?", fresh)
	})

	mkProject := func(businessID int64, name string) int64 {
		wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
			OwnerID: testutil.OwnerID, BusinessID: businessID, Name: name, CreatedBy: testutil.OwnerID,
		}, false)
		if err != nil {
			t.Fatalf("CreateWorkspace: %v", err)
		}
		t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_workspaces WHERE id = ?", wsID) })
		id, err := repo.CreateProject(ctx, domain.ProjectInput{
			WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: businessID,
			Name: name, CreatedBy: testutil.OwnerID,
		})
		if err != nil {
			t.Fatalf("CreateProject: %v", err)
		}
		t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", id) })
		return id
	}
	mkTask := func(businessID, projectID int64, title string) int64 {
		id, err := repo.CreateTask(ctx, domain.TaskInput{
			ProjectID: projectID, OwnerID: testutil.OwnerID, BusinessID: businessID,
			Title: title, ReporterID: testutil.OwnerID,
		})
		if err != nil {
			t.Fatalf("CreateTask %s: %v", title, err)
		}
		t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_tasks WHERE id = ?", id) })
		return id
	}

	freshProject := mkProject(fresh, "task-numbers-fresh")
	otherProject := mkProject(testutil.BusinessID, "task-numbers-other")

	first := mkTask(fresh, freshProject, "first")
	mkTask(testutil.BusinessID, otherProject, "elsewhere")
	second := mkTask(fresh, freshProject, "second")

	for id, want := range map[int64]string{first: "T-1", second: "T-2"} {
		task, err := repo.GetTaskByID(ctx, id)
		if err != nil {
			t.Fatalf("GetTaskByID: %v", err)
		}
		if task.Identifier != want {
			t.Errorf("task %d identifier = %q, want %q", id, task.Identifier, want)
		}
	}

	listed, _, err := repo.GetTasksByBusinessV2(ctx, domain.ListTasksFilter{
		BusinessID: fresh, ProjectID: freshProject, Limit: 10,
	}, testutil.OwnerID, true)
	if err != nil {
		t.Fatalf("GetTasksByBusinessV2: %v", err)
	}
	got := []string{}
	for _, task := range listed {
		got = append(got, task.Identifier)
	}
	if len(got) != 2 || got[0] != "T-1" || got[1] != "T-2" {
		t.Errorf("listed identifiers = %v, want [T-1 T-2]", got)
	}
}

// Creates in one business run at the same time all the time (and in this suite,
// across packages). Reserving a number used to take a shared lock and then
// upgrade it, so two creates could deadlock on each other.
func TestTaskNumbersSurviveConcurrentCreates(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	business := 7_100_000_000 + time.Now().UnixNano()%1_000_000_000
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_tasks WHERE business_id = ?", business)
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_task_sequences WHERE business_id = ?", business)
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE business_id = ?", business)
	})
	project, err := repo.CreateProject(ctx, domain.ProjectInput{
		OwnerID: testutil.OwnerID, BusinessID: business, Name: "concurrent-numbers", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	const creates = 20
	errs := make(chan error, creates)
	ids := make(chan int64, creates)
	for i := 0; i < creates; i++ {
		go func(i int) {
			id, err := repo.CreateTask(ctx, domain.TaskInput{
				ProjectID: project, OwnerID: testutil.OwnerID, BusinessID: business,
				Title: "concurrent", ReporterID: testutil.OwnerID,
			})
			errs <- err
			ids <- id
		}(i)
	}
	seen := map[string]bool{}
	for i := 0; i < creates; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("CreateTask: %v", err)
		}
		task, err := repo.GetTaskByID(ctx, <-ids)
		if err != nil {
			t.Fatalf("GetTaskByID: %v", err)
		}
		if seen[task.Identifier] {
			t.Fatalf("number %s given out twice", task.Identifier)
		}
		seen[task.Identifier] = true
	}
	for n := 1; n <= creates; n++ {
		if !seen[fmt.Sprintf("T-%d", n)] {
			t.Errorf("T-%d missing; got %v", n, seen)
		}
	}
}

// While a release rolls out, an older server still inserts tasks without a
// number, and those show T-<id>. A new number must never repeat such a label.
func TestTaskNumbersStayAboveUnnumberedTasks(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	business := 7_200_000_000 + time.Now().UnixNano()%1_000_000_000
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_tasks WHERE business_id = ?", business)
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_task_sequences WHERE business_id = ?", business)
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE business_id = ?", business)
	})
	project, err := repo.CreateProject(ctx, domain.ProjectInput{
		OwnerID: testutil.OwnerID, BusinessID: business, Name: "unnumbered", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	create := func() int64 {
		id, err := repo.CreateTask(ctx, domain.TaskInput{
			ProjectID: project, OwnerID: testutil.OwnerID, BusinessID: business,
			Title: "task", ReporterID: testutil.OwnerID,
		})
		if err != nil {
			t.Fatalf("CreateTask: %v", err)
		}
		return id
	}

	legacy := create()
	if _, err := db.ExecContext(ctx, "UPDATE flow_tasks SET number = NULL WHERE id = ?", legacy); err != nil {
		t.Fatalf("clear number: %v", err)
	}
	next, err := repo.GetTaskByID(ctx, create())
	if err != nil {
		t.Fatalf("GetTaskByID: %v", err)
	}
	if want := fmt.Sprintf("T-%d", legacy+1); next.Identifier != want {
		t.Errorf("identifier after an unnumbered task = %q, want %q", next.Identifier, want)
	}
}
