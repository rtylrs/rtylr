package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// F-A's repository, against a real database.
//
// The unit tests cover the bucket mapping, which is pure. Everything that can
// only go wrong against MySQL is here: the single-default invariant that has no
// schema constraint behind it, the duplicate-name key, and the two columns that
// must be written together.

func newStatusProbeProject(t *testing.T, repo *infrastructure.Repository, owner int64, name string) int64 {
	t.Helper()
	ctx := context.Background()
	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		OwnerID: owner, BusinessID: testutil.BusinessID,
		Name: name, CreatedBy: owner,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	return projID
}

func TestCreateStatus_AppendsAndListsInOrder(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const owner int64 = 990101
	projID := newStatusProbeProject(t, repo, owner, "Status Order Probe")

	// No sort_order supplied, so each one appends. The first must land at 0,
	// which is the case COALESCE over MAX exists for.
	for _, name := range []string{"Icebox", "Building", "Shipped"} {
		in, err := domain.NormalizeStatusInput(domain.StatusInput{
			ProjectID: projID, Name: name, Bucket: domain.StatusBucketBacklog,
		})
		if err != nil {
			t.Fatalf("normalize %s: %v", name, err)
		}
		if _, err := repo.CreateStatus(ctx, testutil.BusinessID, owner, in); err != nil {
			t.Fatalf("CreateStatus %s: %v", name, err)
		}
	}

	statuses, err := repo.ListStatusesByProject(ctx, testutil.BusinessID, projID)
	if err != nil {
		t.Fatalf("ListStatusesByProject: %v", err)
	}
	if len(statuses) != 3 {
		t.Fatalf("got %d statuses, want 3", len(statuses))
	}
	wantOrder := []struct {
		name  string
		order int32
	}{{"Icebox", 0}, {"Building", 1}, {"Shipped", 2}}
	for i, w := range wantOrder {
		if statuses[i].Name != w.name || statuses[i].SortOrder != w.order {
			t.Errorf("position %d = %q at %d, want %q at %d",
				i, statuses[i].Name, statuses[i].SortOrder, w.name, w.order)
		}
	}
}

func TestListStatuses_EmptyProjectIsNotAnError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	projID := newStatusProbeProject(t, repo, 990201, "Status Empty Probe")

	statuses, err := repo.ListStatusesByProject(ctx, testutil.BusinessID, projID)
	if err != nil {
		t.Fatalf("an unconfigured project must not error: %v", err)
	}
	if len(statuses) != 0 {
		t.Fatalf("got %d statuses for a fresh project, want 0", len(statuses))
	}
	if statuses == nil {
		t.Error("want an empty slice rather than nil, so callers need no null check")
	}
}

func TestCreateStatus_DuplicateNameInSameProjectIsRejected(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const owner int64 = 990301
	projID := newStatusProbeProject(t, repo, owner, "Status Dupe Probe")

	in, _ := domain.NormalizeStatusInput(domain.StatusInput{
		ProjectID: projID, Name: "Building", Bucket: domain.StatusBucketActive,
	})
	if _, err := repo.CreateStatus(ctx, testutil.BusinessID, owner, in); err != nil {
		t.Fatalf("first CreateStatus: %v", err)
	}
	_, err := repo.CreateStatus(ctx, testutil.BusinessID, owner, in)
	if !errors.Is(err, domain.ErrStatusNameTaken) {
		t.Fatalf("second CreateStatus err = %v, want ErrStatusNameTaken", err)
	}
}

func TestCreateStatus_OnlyOneDefaultSurvivesPerProject(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const owner int64 = 990401
	projID := newStatusProbeProject(t, repo, owner, "Status Default Probe")

	// There is NO schema constraint behind this. MySQL cannot express a partial
	// unique index on is_default, so if the transaction in CreateStatus stopped
	// clearing the flag, a project would carry two defaults and a new task would
	// land on whichever the read happened to find first.
	for _, name := range []string{"First", "Second", "Third"} {
		in, _ := domain.NormalizeStatusInput(domain.StatusInput{
			ProjectID: projID, Name: name, Bucket: domain.StatusBucketBacklog, IsDefault: true,
		})
		if _, err := repo.CreateStatus(ctx, testutil.BusinessID, owner, in); err != nil {
			t.Fatalf("CreateStatus %s: %v", name, err)
		}
	}

	statuses, err := repo.ListStatusesByProject(ctx, testutil.BusinessID, projID)
	if err != nil {
		t.Fatalf("ListStatusesByProject: %v", err)
	}
	defaults := make([]string, 0, len(statuses))
	for _, s := range statuses {
		if s.IsDefault {
			defaults = append(defaults, s.Name)
		}
	}
	if len(defaults) != 1 {
		t.Fatalf("defaults = %v, want exactly one", defaults)
	}
	if defaults[0] != "Third" {
		t.Errorf("default = %q, want the most recently set (Third)", defaults[0])
	}
}

func TestGetStatusByID_IsScopedToTheBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const owner int64 = 990501
	projID := newStatusProbeProject(t, repo, owner, "Status Tenancy Probe")

	in, _ := domain.NormalizeStatusInput(domain.StatusInput{
		ProjectID: projID, Name: "Building", Bucket: domain.StatusBucketActive,
	})
	created, err := repo.CreateStatus(ctx, testutil.BusinessID, owner, in)
	if err != nil {
		t.Fatalf("CreateStatus: %v", err)
	}

	// Same id, a different business. This is the read side of the mistake
	// onion#929 made on the write side, where business_id was omitted entirely.
	if _, err := repo.GetStatusByID(ctx, created.ID, testutil.BusinessID+9999); !errors.Is(err, domain.ErrStatusNotFound) {
		t.Fatalf("cross-business read err = %v, want ErrStatusNotFound", err)
	}
	if _, err := repo.GetStatusByID(ctx, created.ID, testutil.BusinessID); err != nil {
		t.Fatalf("same-business read failed: %v", err)
	}
}

func TestDeleteStatus_MovesItsTasksBackToTheStringStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const owner int64 = 990601
	projID := newStatusProbeProject(t, repo, owner, "Status Delete Probe")

	in, _ := domain.NormalizeStatusInput(domain.StatusInput{
		ProjectID: projID, Name: "In legal", Bucket: domain.StatusBucketActive,
		LegacyStatus: domain.TaskStatusInReview,
	})
	status, err := repo.CreateStatus(ctx, testutil.BusinessID, owner, in)
	if err != nil {
		t.Fatalf("CreateStatus: %v", err)
	}

	taskID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, BusinessID: testutil.BusinessID,
		Title: "Status delete probe task", ReporterID: owner, OwnerID: owner,
		// TaskInput has no Status field: a new task takes the column default,
		// which is 'todo'. That is the starting point this test wants anyway.
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if err := repo.SetTaskStatusID(ctx, taskID, testutil.BusinessID, status.ID, status.LegacyStatus); err != nil {
		t.Fatalf("SetTaskStatusID: %v", err)
	}

	// Both columns move together. If only status_id were written the task would
	// show one status on the board and another in every report that reads the
	// string column.
	task, err := repo.GetTaskByID(ctx, taskID)
	if err != nil {
		t.Fatalf("GetTaskByID: %v", err)
	}
	if task.Status != domain.TaskStatusInReview {
		t.Fatalf("string status = %q, want in_review from the custom status", task.Status)
	}

	moved, err := repo.DeleteStatus(ctx, status.ID, testutil.BusinessID)
	if err != nil {
		t.Fatalf("DeleteStatus: %v", err)
	}
	if moved != 1 {
		t.Errorf("moved = %d, want 1", moved)
	}

	// The task survives, and keeps the status the user last saw.
	after, err := repo.GetTaskByID(ctx, taskID)
	if err != nil {
		t.Fatalf("deleting a status must not delete its tasks: %v", err)
	}
	if after.Status != domain.TaskStatusInReview {
		t.Errorf("after delete, string status = %q, want it unchanged at in_review", after.Status)
	}
	if _, err := repo.GetStatusByID(ctx, status.ID, testutil.BusinessID); !errors.Is(err, domain.ErrStatusNotFound) {
		t.Errorf("status should be gone, got err = %v", err)
	}
}
