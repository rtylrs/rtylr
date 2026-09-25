package application

import (
	"context"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/testutil"
)

// The fixture tenant: two users in one business, matching scripts/fixtures.
// userA edits, userB reads, which is the shape the my-tasks staleness was
// reported in.
const (
	cacheTestUserA    = int64(1001)
	cacheTestUserB    = int64(1002)
	cacheTestBusiness = int64(1)
	cacheTestLimit    = int64(50)
)

// initCache wires the process caches the way an API process does, minus Redis.
// The local cache is real; DisableRedisForProcess installs the no-op remote
// tier, whose DeleteByPattern still clears the local one. That is the
// configuration the staleness was observed in, so the tests reproduce it
// without needing a Redis container.
func initCache(t *testing.T) context.Context {
	t.Helper()

	ctx := context.Background()
	if err := cacher.InitializeLocalCache(ctx); err != nil {
		t.Fatalf("InitializeLocalCache: %v", err)
	}
	cacher.DisableRedisForProcess(ctx)

	return ctx
}

// myTasksKey rebuilds the key GetMyTasks writes under, so the assertions break
// if that key shape ever changes without the invalidation following it.
func myTasksKey(userID, page, limit int64) string {
	return cacheKeys.Scoped("user", userID, fmt.Sprintf("workspace_tasks:page=%d:limit=%d", page, limit))
}

func myTasksCached(ctx context.Context, userID, page, limit int64) bool {
	_, found := cacher.Of[myTasksPage]().Get(ctx, myTasksKey(userID, page, limit))

	return found
}

func putMyTasksPage(ctx context.Context, userID, page, limit int64) {
	cacher.Of[myTasksPage]().Put(ctx, myTasksKey(userID, page, limit),
		myTasksPage{Total: page},
		cacher.WithExpiry(ctx, workspacesTTL),
	)
}

// primeMyTasks fills a user's page-1 cache through the real read path, so the
// tests assert against an entry the service itself wrote.
func primeMyTasks(ctx context.Context, t *testing.T, svc *Service, userID int64) {
	t.Helper()

	if _, _, err := svc.GetMyTasks(ctx, userID, 1, cacheTestLimit); err != nil {
		t.Fatalf("GetMyTasks(%d): %v", userID, err)
	}
	if !myTasksCached(ctx, userID, 1, cacheTestLimit) {
		t.Fatalf("my-tasks cache for user %d was not primed", userID)
	}
}

// seedProject creates a throwaway workspace and project for one test.
func seedProject(ctx context.Context, t *testing.T, svc *Service) (workspaceID, projectID int64) {
	t.Helper()

	workspace, err := svc.CreateWorkspace(ctx, domain.WorkspaceInput{
		BusinessID: cacheTestBusiness,
		OwnerID:    cacheTestUserA,
		Name:       t.Name(),
		CreatedBy:  cacheTestUserA,
	})
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	project, err := svc.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: workspace.ID,
		OwnerID:     cacheTestUserA,
		BusinessID:  cacheTestBusiness,
		Name:        t.Name(),
		CreatedBy:   cacheTestUserA,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	return workspace.ID, project.ID
}

func seedTask(ctx context.Context, t *testing.T, svc *Service, title string, assignee int64) domain.Task {
	t.Helper()

	workspaceID, projectID := seedProject(ctx, t, svc)

	task, err := svc.CreateTask(ctx, domain.TaskInput{
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
		OwnerID:     cacheTestUserA,
		BusinessID:  cacheTestBusiness,
		Title:       title,
		AssigneeID:  &assignee,
		ReporterID:  cacheTestUserA,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	return task
}

func titleOfTask(tasks []domain.Task, taskID int64) (string, bool) {
	for _, task := range tasks {
		if task.ID == taskID {
			return task.Title, true
		}
	}

	return "", false
}

// The page and limit are baked into every my-tasks key, so an exact-key delete
// would only ever clear the page that was read last. This is the test that
// pins the invalidation to a prefix delete.
func TestInvalidateUserScopeClearsEveryCachedPage(t *testing.T) {
	ctx := initCache(t)

	pages := []int64{1, 2, 3}
	for _, page := range pages {
		putMyTasksPage(ctx, cacheTestUserA, page, cacheTestLimit)
	}
	putMyTasksPage(ctx, cacheTestUserB, 1, cacheTestLimit)

	invalidateUserScope(ctx, cacheTestUserA)

	for _, page := range pages {
		if myTasksCached(ctx, cacheTestUserA, page, cacheTestLimit) {
			t.Errorf("page %d of the mutating user's my-tasks survived invalidation", page)
		}
	}
	if !myTasksCached(ctx, cacheTestUserB, 1, cacheTestLimit) {
		t.Error("another user's my-tasks page was evicted; the prefix is too wide")
	}
}

// The original bug: the my-tasks entry is written under the user scope, and the
// only invalidation the task write paths had deleted under the workspaces
// scope. The two patterns share no segment, so the entry was TTL-only.
func TestInvalidateWorkspacesScopeDoesNotReachMyTasks(t *testing.T) {
	ctx := initCache(t)

	putMyTasksPage(ctx, cacheTestUserA, 1, cacheTestLimit)

	invalidateWorkspacesScope(ctx, cacheTestBusiness)

	if !myTasksCached(ctx, cacheTestUserA, 1, cacheTestLimit) {
		t.Fatal("workspaces-scope invalidation now clears my-tasks; the user-scope eviction may be redundant")
	}

	invalidateUserScope(ctx, cacheTestUserA)

	if myTasksCached(ctx, cacheTestUserA, 1, cacheTestLimit) {
		t.Error("user-scope invalidation left the my-tasks page in place")
	}
}

func TestUpdateTaskReassignmentInvalidatesBothAssignees(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := initCache(t)
	svc := NewService(ctx, db)

	task := seedTask(ctx, t, svc, "before reassignment", cacheTestUserA)

	primeMyTasks(ctx, t, svc, cacheTestUserA)
	primeMyTasks(ctx, t, svc, cacheTestUserB)

	newAssignee := cacheTestUserB
	if err := svc.UpdateTask(ctx, domain.UpdateTaskInput{
		ID:         task.ID,
		OwnerID:    cacheTestUserA,
		BusinessID: cacheTestBusiness,
		Title:      "after reassignment",
		AssigneeID: &newAssignee,
	}, cacheTestUserA); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	if myTasksCached(ctx, cacheTestUserA, 1, cacheTestLimit) {
		t.Error("the previous assignee's my-tasks cache survived the reassignment; the task lingers in their list")
	}
	if myTasksCached(ctx, cacheTestUserB, 1, cacheTestLimit) {
		t.Error("the new assignee's my-tasks cache survived the reassignment; the task is missing from their list")
	}

	// The reported symptom, end to end: the new assignee's next read has to
	// show the edit now, not two minutes from now.
	tasks, _, err := svc.GetMyTasks(ctx, cacheTestUserB, 1, cacheTestLimit)
	if err != nil {
		t.Fatalf("GetMyTasks after reassignment: %v", err)
	}
	title, found := titleOfTask(tasks, task.ID)
	if !found {
		t.Fatal("the reassigned task is missing from the new assignee's my-tasks")
	}
	if title != "after reassignment" {
		t.Errorf("title = %q, want %q", title, "after reassignment")
	}
}

func TestUpdateTaskWithoutReassignmentInvalidatesOnlyTheAssignee(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := initCache(t)
	svc := NewService(ctx, db)

	task := seedTask(ctx, t, svc, "before rename", cacheTestUserA)

	primeMyTasks(ctx, t, svc, cacheTestUserA)
	primeMyTasks(ctx, t, svc, cacheTestUserB)

	if err := svc.UpdateTask(ctx, domain.UpdateTaskInput{
		ID:         task.ID,
		OwnerID:    cacheTestUserA,
		BusinessID: cacheTestBusiness,
		Title:      "after rename",
	}, cacheTestUserA); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	if myTasksCached(ctx, cacheTestUserA, 1, cacheTestLimit) {
		t.Error("the assignee's my-tasks cache survived a title change")
	}
	if !myTasksCached(ctx, cacheTestUserB, 1, cacheTestLimit) {
		t.Error("an uninvolved user's my-tasks cache was evicted by a title change")
	}
}

func TestUpdateTaskStatusInvalidatesOnlyTheAssignee(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := initCache(t)
	svc := NewService(ctx, db)

	task := seedTask(ctx, t, svc, "status change", cacheTestUserA)

	primeMyTasks(ctx, t, svc, cacheTestUserA)
	primeMyTasks(ctx, t, svc, cacheTestUserB)

	if err := svc.UpdateTaskStatus(ctx, domain.UpdateTaskStatusInput{
		ID:         task.ID,
		OwnerID:    cacheTestUserA,
		BusinessID: cacheTestBusiness,
		Status:     "in_progress",
	}, cacheTestUserA); err != nil {
		t.Fatalf("UpdateTaskStatus: %v", err)
	}

	if myTasksCached(ctx, cacheTestUserA, 1, cacheTestLimit) {
		t.Error("the assignee's my-tasks cache survived a status change")
	}
	if !myTasksCached(ctx, cacheTestUserB, 1, cacheTestLimit) {
		t.Error("an uninvolved user's my-tasks cache was evicted by a status change")
	}
}

func TestCreateTaskInvalidatesTheAssignee(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := initCache(t)
	svc := NewService(ctx, db)

	workspaceID, projectID := seedProject(ctx, t, svc)

	primeMyTasks(ctx, t, svc, cacheTestUserA)
	primeMyTasks(ctx, t, svc, cacheTestUserB)

	assignee := cacheTestUserA
	if _, err := svc.CreateTask(ctx, domain.TaskInput{
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
		OwnerID:     cacheTestUserA,
		BusinessID:  cacheTestBusiness,
		Title:       "freshly assigned",
		AssigneeID:  &assignee,
		ReporterID:  cacheTestUserA,
	}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	if myTasksCached(ctx, cacheTestUserA, 1, cacheTestLimit) {
		t.Error("the assignee's my-tasks cache survived a task being assigned to them")
	}
	if !myTasksCached(ctx, cacheTestUserB, 1, cacheTestLimit) {
		t.Error("an uninvolved user's my-tasks cache was evicted by someone else's new task")
	}
}

func TestDeleteTaskInvalidatesTheAssignee(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := initCache(t)
	svc := NewService(ctx, db)

	task := seedTask(ctx, t, svc, "to be deleted", cacheTestUserA)

	primeMyTasks(ctx, t, svc, cacheTestUserA)

	if err := svc.DeleteTask(ctx, task.ID, cacheTestBusiness); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}

	if myTasksCached(ctx, cacheTestUserA, 1, cacheTestLimit) {
		t.Fatal("the assignee's my-tasks cache survived the delete")
	}

	tasks, _, err := svc.GetMyTasks(ctx, cacheTestUserA, 1, cacheTestLimit)
	if err != nil {
		t.Fatalf("GetMyTasks after delete: %v", err)
	}
	if _, found := titleOfTask(tasks, task.ID); found {
		t.Error("the deleted task is still listed in my-tasks")
	}
}

func TestMoveTaskToColumnInvalidatesTheAssignee(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := initCache(t)
	svc := NewService(ctx, db)

	workspaceID, projectID := seedProject(ctx, t, svc)

	assignee := cacheTestUserA
	task, err := svc.CreateTask(ctx, domain.TaskInput{
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
		OwnerID:     cacheTestUserA,
		BusinessID:  cacheTestBusiness,
		Title:       "board card",
		AssigneeID:  &assignee,
		ReporterID:  cacheTestUserA,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	column, err := svc.CreateColumn(ctx, domain.ProjectColumnInput{
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
		Name:        "Blocked",
		Color:       "#ef4444",
		CreatedBy:   cacheTestUserA,
	})
	if err != nil {
		t.Fatalf("CreateColumn: %v", err)
	}

	primeMyTasks(ctx, t, svc, cacheTestUserA)
	primeMyTasks(ctx, t, svc, cacheTestUserB)

	if err := svc.MoveTaskToColumn(ctx, task.ID, column.ID, cacheTestUserA); err != nil {
		t.Fatalf("MoveTaskToColumn: %v", err)
	}

	if myTasksCached(ctx, cacheTestUserA, 1, cacheTestLimit) {
		t.Error("the assignee's my-tasks cache survived the task moving column")
	}
	if !myTasksCached(ctx, cacheTestUserB, 1, cacheTestLimit) {
		t.Error("an uninvolved user's my-tasks cache was evicted by someone else's board move")
	}
}

// Labels are part of the cached my-tasks payload, so relabelling a task is a
// task-visible write like any other and has to evict the assignee's pages.
func TestSetTaskLabelsInvalidatesTheAssignee(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := initCache(t)
	svc := NewService(ctx, db)

	workspaceID, projectID := seedProject(ctx, t, svc)

	assignee := cacheTestUserA
	task, err := svc.CreateTask(ctx, domain.TaskInput{
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
		OwnerID:     cacheTestUserA,
		BusinessID:  cacheTestBusiness,
		Title:       "relabelled card",
		AssigneeID:  &assignee,
		ReporterID:  cacheTestUserA,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	primeMyTasks(ctx, t, svc, cacheTestUserA)
	primeMyTasks(ctx, t, svc, cacheTestUserB)

	if err := svc.SetTaskLabels(ctx, task.ID, cacheTestUserA, []int64{101, 102}); err != nil {
		t.Fatalf("SetTaskLabels: %v", err)
	}

	if myTasksCached(ctx, cacheTestUserA, 1, cacheTestLimit) {
		t.Error("the assignee's my-tasks cache survived a label change")
	}
	if !myTasksCached(ctx, cacheTestUserB, 1, cacheTestLimit) {
		t.Error("an uninvolved user's my-tasks cache was evicted by someone else's label change")
	}

	// The reported symptom, end to end: the assignee's next read has to show
	// the new labels now, not two minutes from now.
	tasks, _, err := svc.GetMyTasks(ctx, cacheTestUserA, 1, cacheTestLimit)
	if err != nil {
		t.Fatalf("GetMyTasks after label change: %v", err)
	}
	if !labelsOfTaskContain(tasks, task.ID, 101) {
		t.Error("the label change is missing from the assignee's my-tasks row")
	}
}

func labelsOfTaskContain(tasks []domain.Task, taskID, labelID int64) bool {
	for _, task := range tasks {
		if task.ID != taskID {
			continue
		}
		for _, id := range task.Labels {
			if id == labelID {
				return true
			}
		}
	}

	return false
}
