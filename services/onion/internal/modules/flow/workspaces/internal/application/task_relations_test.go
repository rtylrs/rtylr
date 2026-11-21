package application_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"
	"github.com/voxire/onion/internal/testutil"
)

const (
	relationOwner    int64 = 975101
	relationAssignee int64 = 975102
)

type relationFixture struct {
	t    *testing.T
	ctx  context.Context
	db   *sql.DB
	repo *infrastructure.Repository
	svc  *application.Service
}

func newRelationFixture(t *testing.T) *relationFixture {
	t.Helper()
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	return &relationFixture{
		t: t, ctx: context.Background(), db: db, repo: repo,
		svc: application.NewService(context.Background(), repo),
	}
}

// project creates a project owned by relationOwner in businessID and removes
// it, its tasks, their relations and their notifications when the test ends.
func (f *relationFixture) project(businessID int64, name string) int64 {
	f.t.Helper()
	id, err := f.repo.CreateProject(f.ctx, domain.ProjectInput{
		OwnerID: relationOwner, BusinessID: businessID, Name: name, CreatedBy: relationOwner,
	})
	if err != nil {
		f.t.Fatalf("CreateProject(%s): %v", name, err)
	}
	f.t.Cleanup(func() {
		if _, cleanupErr := f.db.ExecContext(f.ctx,
			"DELETE FROM flow_task_dependencies WHERE task_id IN (SELECT id FROM flow_tasks WHERE project_id = ?) OR depends_on_task_id IN (SELECT id FROM flow_tasks WHERE project_id = ?)",
			id, id); cleanupErr != nil {
			f.t.Errorf("cleanup: %v", cleanupErr)
		}
		for _, stmt := range []string{
			"DELETE FROM flow_notifications WHERE task_id IN (SELECT id FROM flow_tasks WHERE project_id = ?)",
			"DELETE FROM flow_task_activity WHERE task_id IN (SELECT id FROM flow_tasks WHERE project_id = ?)",
			"DELETE FROM flow_tasks WHERE project_id = ?",
			"DELETE FROM flow_statuses WHERE project_id = ?",
			"DELETE FROM flow_projects WHERE id = ?",
		} {
			if _, cleanupErr := f.db.ExecContext(f.ctx, stmt, id); cleanupErr != nil {
				f.t.Errorf("cleanup: %v", cleanupErr)
			}
		}
	})
	return id
}

func (f *relationFixture) task(projectID, businessID, assignee int64, title string) int64 {
	f.t.Helper()
	input := domain.TaskInput{
		ProjectID: projectID, OwnerID: relationOwner, BusinessID: businessID,
		Title: title, ReporterID: relationOwner,
	}
	if assignee != 0 {
		input.AssigneeID = &assignee
	}
	id, err := f.repo.CreateTask(f.ctx, input)
	if err != nil {
		f.t.Fatalf("CreateTask(%s): %v", title, err)
	}
	return id
}

func (f *relationFixture) relate(taskID, otherID int64, relationType, direction string) int64 {
	f.t.Helper()
	change, err := f.svc.AddTaskRelation(f.ctx, taskID, otherID, relationOwner, testutil.BusinessID, relationType, direction)
	if err != nil {
		f.t.Fatalf("AddTaskRelation(%d %s/%s %d): %v", taskID, relationType, direction, otherID, err)
	}
	return change.RelationID
}

func (f *relationFixture) setStatus(taskID int64, status string) {
	f.t.Helper()
	if err := f.svc.UpdateTaskStatus(f.ctx, domain.UpdateTaskStatusInput{
		ID: taskID, OwnerID: relationOwner, BusinessID: testutil.BusinessID, Status: status,
	}, relationOwner); err != nil {
		f.t.Fatalf("UpdateTaskStatus(%d, %s): %v", taskID, status, err)
	}
}

// groups returns the listed task's relations as relation key -> other task ids.
func (f *relationFixture) groups(taskID int64) map[string][]int64 {
	f.t.Helper()
	groups, err := f.svc.ListTaskRelations(f.ctx, taskID, relationOwner, testutil.BusinessID)
	if err != nil {
		f.t.Fatalf("ListTaskRelations(%d): %v", taskID, err)
	}
	out := make(map[string][]int64, len(groups))
	for _, group := range groups {
		for _, relation := range group.Relations {
			out[group.Relation] = append(out[group.Relation], relation.Task.ID)
		}
	}
	return out
}

func (f *relationFixture) unblockedNotifications(userID, taskID int64) int {
	f.t.Helper()
	notifications, _, err := f.svc.ListNotifications(f.ctx, userID, 500, 0)
	if err != nil {
		f.t.Fatalf("ListNotifications: %v", err)
	}
	want := fmt.Sprintf("T-%d is no longer blocked", taskID)
	count := 0
	for _, notification := range notifications {
		if notification.TaskID != nil && *notification.TaskID == taskID && notification.Title == want {
			count++
		}
	}
	return count
}

func sameIDs(got []int64, want ...int64) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestTaskRelations_ShowBothDirections(t *testing.T) {
	f := newRelationFixture(t)
	projectID := f.project(testutil.BusinessID, "Relations directions")
	otherProjectID := f.project(testutil.BusinessID, "Relations directions other project")
	a := f.task(projectID, testutil.BusinessID, relationAssignee, "A")
	b := f.task(otherProjectID, testutil.BusinessID, 0, "B")
	c := f.task(projectID, testutil.BusinessID, 0, "C")

	f.relate(a, b, domain.RelationBlocks, domain.RelationOutward)
	// Added from the object's side: C is duplicated by A, so A duplicates C.
	f.relate(c, a, domain.RelationDuplicates, domain.RelationInward)
	f.relate(c, b, domain.RelationClones, "")

	onA := f.groups(a)
	if !sameIDs(onA["blocks"], b) || !sameIDs(onA["duplicates"], c) || len(onA) != 2 {
		t.Fatalf("relations on A = %v, want blocks [%d] and duplicates [%d]", onA, b, c)
	}
	onB := f.groups(b)
	if !sameIDs(onB["blocked_by"], a) || !sameIDs(onB["cloned_by"], c) || len(onB) != 2 {
		t.Fatalf("relations on B = %v, want blocked_by [%d] and cloned_by [%d]", onB, a, c)
	}
	onC := f.groups(c)
	if !sameIDs(onC["duplicated_by"], a) || !sameIDs(onC["clones"], b) || len(onC) != 2 {
		t.Fatalf("relations on C = %v, want duplicated_by [%d] and clones [%d]", onC, a, c)
	}

	// The legacy dependency projection only ever means "blocked by".
	deps, err := f.repo.ListTaskDependencies(f.ctx, []int64{a, b, c}, testutil.BusinessID)
	if err != nil {
		t.Fatalf("ListTaskDependencies: %v", err)
	}
	if !sameIDs(deps[b], a) || len(deps[a]) != 0 || len(deps[c]) != 0 {
		t.Fatalf("dependencies = %v, want only %d blocked by %d", deps, b, a)
	}

	groups, err := f.svc.ListTaskRelations(f.ctx, b, relationOwner, testutil.BusinessID)
	if err != nil {
		t.Fatalf("ListTaskRelations: %v", err)
	}
	blocker := groups[0].Relations[0].Task
	if groups[0].Type != domain.RelationBlocks || groups[0].Direction != domain.RelationInward {
		t.Fatalf("first group on B = %s/%s, want blocks/inward", groups[0].Type, groups[0].Direction)
	}
	if blocker.Identifier != fmt.Sprintf("T-%d", a) || blocker.Title != "A" || blocker.ProjectID != projectID ||
		blocker.StatusBucket != domain.StatusBucketBacklog || blocker.AssigneeID == nil || *blocker.AssigneeID != relationAssignee {
		t.Fatalf("related task projection is incomplete: %+v", blocker)
	}
}

func TestTaskRelations_RelatesIsSymmetricAndStoredOnce(t *testing.T) {
	f := newRelationFixture(t)
	projectID := f.project(testutil.BusinessID, "Relations relates")
	low := f.task(projectID, testutil.BusinessID, 0, "low")
	high := f.task(projectID, testutil.BusinessID, 0, "high")

	relationID := f.relate(high, low, domain.RelationRelates, domain.RelationInward)

	var storedTask, storedOther int64
	if err := f.db.QueryRowContext(f.ctx,
		"SELECT task_id, depends_on_task_id FROM flow_task_dependencies WHERE id = ?", relationID,
	).Scan(&storedTask, &storedOther); err != nil {
		t.Fatalf("read stored relation: %v", err)
	}
	if storedTask != low || storedOther != high {
		t.Fatalf("relates stored as (%d, %d), want lower id first (%d, %d)", storedTask, storedOther, low, high)
	}
	if got := f.groups(low); !sameIDs(got["relates_to"], high) || len(got) != 1 {
		t.Fatalf("relations on low = %v, want relates_to [%d]", got, high)
	}
	if got := f.groups(high); !sameIDs(got["relates_to"], low) || len(got) != 1 {
		t.Fatalf("relations on high = %v, want relates_to [%d]", got, low)
	}

	// The same pair from the other end is the same relation, not a second one.
	_, err := f.svc.AddTaskRelation(f.ctx, low, high, relationOwner, testutil.BusinessID, domain.RelationRelates, "")
	if !errors.Is(err, domain.ErrRelationExists) {
		t.Fatalf("second relates = %v, want ErrRelationExists", err)
	}

	if _, err := f.svc.RemoveTaskRelation(f.ctx, high, relationID, testutil.BusinessID); err != nil {
		t.Fatalf("RemoveTaskRelation from the other end: %v", err)
	}
	if got := f.groups(low); len(got) != 0 {
		t.Fatalf("relations survived removal: %v", got)
	}
}

func TestTaskRelations_RejectsSelfDuplicateAndCycle(t *testing.T) {
	f := newRelationFixture(t)
	projectID := f.project(testutil.BusinessID, "Relations validation")
	a := f.task(projectID, testutil.BusinessID, 0, "A")
	b := f.task(projectID, testutil.BusinessID, 0, "B")
	c := f.task(projectID, testutil.BusinessID, 0, "C")
	add := func(taskID, otherID int64, relationType, direction string) error {
		_, err := f.svc.AddTaskRelation(f.ctx, taskID, otherID, relationOwner, testutil.BusinessID, relationType, direction)
		return err
	}

	if err := add(a, a, domain.RelationBlocks, ""); !errors.Is(err, domain.ErrRelationSelf) {
		t.Fatalf("self link = %v, want ErrRelationSelf", err)
	}
	if err := add(a, b, "depends", ""); !errors.Is(err, domain.ErrRelationInvalidType) {
		t.Fatalf("unknown type = %v, want ErrRelationInvalidType", err)
	}
	if err := add(a, b, domain.RelationBlocks, "sideways"); !errors.Is(err, domain.ErrRelationInvalidDirection) {
		t.Fatalf("unknown direction = %v, want ErrRelationInvalidDirection", err)
	}

	f.relate(a, b, domain.RelationBlocks, "")
	f.relate(b, c, domain.RelationBlocks, "")

	if err := add(a, b, domain.RelationBlocks, ""); !errors.Is(err, domain.ErrRelationExists) {
		t.Fatalf("repeat blocks = %v, want ErrRelationExists", err)
	}
	// The same edge named from the other end is still the same edge.
	if err := add(b, a, domain.RelationBlocks, domain.RelationInward); !errors.Is(err, domain.ErrRelationExists) {
		t.Fatalf("repeat blocks from the other end = %v, want ErrRelationExists", err)
	}
	if err := add(b, a, domain.RelationBlocks, ""); !errors.Is(err, domain.ErrRelationCycle) {
		t.Fatalf("two-task cycle = %v, want ErrRelationCycle", err)
	}
	if err := add(c, a, domain.RelationBlocks, ""); !errors.Is(err, domain.ErrRelationCycle) {
		t.Fatalf("three-task cycle = %v, want ErrRelationCycle", err)
	}
	// The older dependency endpoint writes the same relation and must not be a
	// way around the check: "A depends on C" is "C blocks A".
	if err := f.svc.AddTaskDependency(f.ctx, a, c, testutil.BusinessID); !errors.Is(err, domain.ErrRelationCycle) {
		t.Fatalf("cycle through AddTaskDependency = %v, want ErrRelationCycle", err)
	}

	// Only blocks is acyclic. The other types say nothing about order.
	f.relate(c, a, domain.RelationRelates, "")
	f.relate(a, c, domain.RelationDuplicates, "")
	if err := add(c, a, domain.RelationDuplicates, ""); !errors.Is(err, domain.ErrRelationExists) {
		t.Fatalf("duplicates in both directions = %v, want ErrRelationExists", err)
	}
}

func TestTaskRelations_BlockedFollowsBlockerStatus(t *testing.T) {
	f := newRelationFixture(t)
	projectID := f.project(testutil.BusinessID, "Relations blocked flag")
	blockerOne := f.task(projectID, testutil.BusinessID, 0, "blocker one")
	blockerTwo := f.task(projectID, testutil.BusinessID, 0, "blocker two")
	blocked := f.task(projectID, testutil.BusinessID, 0, "blocked")
	duplicate := f.task(projectID, testutil.BusinessID, 0, "duplicate")

	f.relate(blockerOne, blocked, domain.RelationBlocks, "")
	f.relate(blocked, blockerTwo, domain.RelationBlocks, domain.RelationInward)
	// Not a blocker: marking a duplicate neither blocks nor closes anything.
	f.relate(duplicate, blocked, domain.RelationDuplicates, "")

	list := func(blockedFilter *bool) map[int64]domain.Task {
		t.Helper()
		tasks, total, err := f.svc.ListTasksV2(f.ctx, relationOwner, testutil.BusinessID, domain.ListTasksFilter{
			ProjectID: projectID, Blocked: blockedFilter, Limit: 100,
		})
		if err != nil {
			t.Fatalf("ListTasksV2: %v", err)
		}
		if int(total) != len(tasks) {
			t.Fatalf("total = %d but %d tasks returned: the count ignores the blocked filter", total, len(tasks))
		}
		out := make(map[int64]domain.Task, len(tasks))
		for _, task := range tasks {
			out[task.ID] = task
		}
		return out
	}
	assertBlocked := func(step string, wantCount int32) {
		t.Helper()
		all := list(nil)
		if got := all[blocked]; got.BlockerCount != wantCount || got.Blocked != (wantCount > 0) {
			t.Fatalf("%s: list shows blocked=%v blocker_count=%d, want count %d", step, got.Blocked, got.BlockerCount, wantCount)
		}
		for _, id := range []int64{blockerOne, blockerTwo, duplicate} {
			if all[id].Blocked || all[id].BlockerCount != 0 {
				t.Fatalf("%s: task %d is reported blocked", step, id)
			}
		}
		detail, err := f.svc.GetTaskDetail(f.ctx, blocked)
		if err != nil {
			t.Fatalf("GetTaskDetail: %v", err)
		}
		if detail.Task.BlockerCount != wantCount || detail.Task.Blocked != (wantCount > 0) {
			t.Fatalf("%s: detail shows blocked=%v blocker_count=%d, want count %d",
				step, detail.Task.Blocked, detail.Task.BlockerCount, wantCount)
		}
		yes, no := true, false
		_, inBlocked := list(&yes)[blocked]
		_, inUnblocked := list(&no)[blocked]
		if inBlocked != (wantCount > 0) || inUnblocked == inBlocked {
			t.Fatalf("%s: blocked=true lists the task: %v, blocked=false lists it: %v", step, inBlocked, inUnblocked)
		}
		if len(list(&yes))+len(list(&no)) != len(all) {
			t.Fatalf("%s: blocked=true and blocked=false do not partition the project", step)
		}
	}

	assertBlocked("two open blockers", 2)
	f.setStatus(blockerOne, domain.TaskStatusDone)
	assertBlocked("one blocker done", 1)
	f.setStatus(blockerTwo, domain.TaskStatusCancelled)
	assertBlocked("other blocker cancelled", 0)
	f.setStatus(blockerOne, domain.TaskStatusInProgress)
	assertBlocked("blocker reopened", 1)

	if detail, err := f.svc.GetTaskDetail(f.ctx, duplicate); err != nil || detail.Task.Status == domain.TaskStatusDone {
		t.Fatalf("marking a duplicate changed its status: %v %v", detail.Task.Status, err)
	}
}

func TestTaskRelations_UnblockNotifiesAssigneeOnceForLastBlocker(t *testing.T) {
	f := newRelationFixture(t)
	projectID := f.project(testutil.BusinessID, "Relations unblock notification")
	blockerOne := f.task(projectID, testutil.BusinessID, 0, "blocker one")
	blockerTwo := f.task(projectID, testutil.BusinessID, 0, "blocker two")
	blocked := f.task(projectID, testutil.BusinessID, relationAssignee, "waiting")
	// Assigned to the person who closes its blocker: they already know.
	ownBlocked := f.task(projectID, testutil.BusinessID, relationOwner, "mine")
	// Already finished: nobody is waiting to start it.
	finished := f.task(projectID, testutil.BusinessID, relationAssignee, "finished")
	f.setStatus(finished, domain.TaskStatusDone)

	f.relate(blockerOne, blocked, domain.RelationBlocks, "")
	f.relate(blockerTwo, blocked, domain.RelationBlocks, "")
	f.relate(blockerTwo, ownBlocked, domain.RelationBlocks, "")
	f.relate(blockerTwo, finished, domain.RelationBlocks, "")

	f.setStatus(blockerOne, domain.TaskStatusDone)
	if got := f.unblockedNotifications(relationAssignee, blocked); got != 0 {
		t.Fatalf("%d notifications after the first of two blockers closed, want 0", got)
	}

	// The last blocker closes through the general update path this time.
	if err := f.svc.UpdateTask(f.ctx, domain.UpdateTaskInput{
		ID: blockerTwo, OwnerID: relationOwner, BusinessID: testutil.BusinessID, Status: domain.TaskStatusCancelled,
	}, relationOwner); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	if got := f.unblockedNotifications(relationAssignee, blocked); got != 1 {
		t.Fatalf("%d notifications after the last blocker closed, want exactly 1", got)
	}
	if got := f.unblockedNotifications(relationOwner, ownBlocked); got != 0 {
		t.Fatalf("the actor was notified about their own task %d times", got)
	}
	if got := f.unblockedNotifications(relationAssignee, finished); got != 0 {
		t.Fatalf("a finished task produced %d unblock notifications", got)
	}

	// Closed to closed is not a transition, whichever path writes it.
	f.setStatus(blockerTwo, domain.TaskStatusDone)
	f.setStatus(blockerOne, domain.TaskStatusDone)
	if got := f.unblockedNotifications(relationAssignee, blocked); got != 1 {
		t.Fatalf("%d notifications after re-closing closed blockers, want still 1", got)
	}

	// Blocked again and unblocked again is a new event.
	f.setStatus(blockerOne, domain.TaskStatusTodo)
	f.setStatus(blockerOne, domain.TaskStatusDone)
	if got := f.unblockedNotifications(relationAssignee, blocked); got != 2 {
		t.Fatalf("%d notifications after a second unblock, want 2", got)
	}
}

func TestTaskRelations_UnblockNotifiesOnBoardMoveAndStatusDelete(t *testing.T) {
	f := newRelationFixture(t)
	projectID := f.project(testutil.BusinessID, "Relations unblock board paths")
	if err := f.repo.SeedDefaultColumns(f.ctx, projectID, 0, relationOwner); err != nil {
		t.Fatalf("SeedDefaultColumns: %v", err)
	}
	statuses, err := f.repo.ListStatusesByProject(f.ctx, testutil.BusinessID, projectID)
	if err != nil {
		t.Fatalf("ListStatusesByProject: %v", err)
	}
	var done domain.Status
	for _, status := range statuses {
		if status.Bucket == domain.StatusBucketDone {
			done = status
		}
	}
	if done.ID == 0 {
		t.Fatal("project has no done status")
	}
	review, err := f.repo.CreateStatus(f.ctx, testutil.BusinessID, relationOwner, domain.StatusInput{
		ProjectID: projectID, Name: "Relations review", Bucket: domain.StatusBucketActive,
		LegacyStatus: domain.TaskStatusInReview, Color: domain.DefaultStatusColor,
	})
	if err != nil {
		t.Fatalf("CreateStatus: %v", err)
	}

	dragged := f.task(projectID, testutil.BusinessID, 0, "dragged blocker")
	afterDrag := f.task(projectID, testutil.BusinessID, relationAssignee, "after drag")
	f.relate(dragged, afterDrag, domain.RelationBlocks, "")

	if err := f.svc.MoveTaskToColumn(f.ctx, dragged, done.ID, relationOwner); err != nil {
		t.Fatalf("MoveTaskToColumn: %v", err)
	}
	if got := f.unblockedNotifications(relationAssignee, afterDrag); got != 1 {
		t.Fatalf("%d notifications after dragging the blocker to done, want 1", got)
	}

	// Two blockers of one task sit on a status that is deleted with its tasks
	// moved to done: both close in one statement, and that is one unblock.
	bulkOne := f.task(projectID, testutil.BusinessID, 0, "bulk blocker one")
	bulkTwo := f.task(projectID, testutil.BusinessID, 0, "bulk blocker two")
	afterBulk := f.task(projectID, testutil.BusinessID, relationAssignee, "after bulk")
	for _, blocker := range []int64{bulkOne, bulkTwo} {
		if _, err := f.svc.SetTaskStatus(f.ctx, blocker, review.ID, relationOwner, testutil.BusinessID); err != nil {
			t.Fatalf("SetTaskStatus: %v", err)
		}
		f.relate(blocker, afterBulk, domain.RelationBlocks, "")
	}
	if detail, err := f.svc.GetTaskDetail(f.ctx, afterBulk); err != nil || detail.Task.BlockerCount != 2 {
		t.Fatalf("before the delete blocker_count = %d (%v), want 2", detail.Task.BlockerCount, err)
	}
	moved, err := f.svc.DeleteProjectStatus(f.ctx, review.ID, done.ID, relationOwner, testutil.BusinessID)
	if err != nil || moved != 2 {
		t.Fatalf("DeleteProjectStatus moved %d (%v), want 2", moved, err)
	}
	if got := f.unblockedNotifications(relationAssignee, afterBulk); got != 1 {
		t.Fatalf("%d notifications after the bulk move closed both blockers, want 1", got)
	}
}

func TestTaskRelations_DoNotCrossBusinesses(t *testing.T) {
	f := newRelationFixture(t)
	projectID := f.project(testutil.BusinessID, "Relations tenancy")
	foreignProjectID := f.project(testutil.OtherBusinessID, "Relations tenancy foreign")
	mine := f.task(projectID, testutil.BusinessID, 0, "mine")
	alsoMine := f.task(projectID, testutil.BusinessID, 0, "also mine")
	foreign := f.task(foreignProjectID, testutil.OtherBusinessID, 0, "foreign")
	alsoForeign := f.task(foreignProjectID, testutil.OtherBusinessID, 0, "also foreign")

	// Writes: neither end may sit in another business.
	for _, pair := range [][2]int64{{mine, foreign}, {foreign, mine}, {foreign, alsoForeign}} {
		_, err := f.svc.AddTaskRelation(f.ctx, pair[0], pair[1], relationOwner, testutil.BusinessID, domain.RelationBlocks, "")
		if !errors.Is(err, domain.ErrUnauthorized) {
			t.Fatalf("AddTaskRelation(%d, %d) across businesses = %v, want ErrUnauthorized", pair[0], pair[1], err)
		}
	}
	var rows int
	if err := f.db.QueryRowContext(f.ctx,
		"SELECT COUNT(*) FROM flow_task_dependencies WHERE task_id IN (?, ?) OR depends_on_task_id IN (?, ?)",
		foreign, alsoForeign, foreign, alsoForeign).Scan(&rows); err != nil || rows != 0 {
		t.Fatalf("refused relations left %d rows behind (%v)", rows, err)
	}

	relationID := f.relate(mine, alsoMine, domain.RelationBlocks, "")

	// Reads and deletes from the other business see nothing of it.
	if _, err := f.svc.ListTaskRelations(f.ctx, mine, relationOwner, testutil.OtherBusinessID); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("ListTaskRelations from another business = %v, want ErrUnauthorized", err)
	}
	if _, err := f.svc.RemoveTaskRelation(f.ctx, mine, relationID, testutil.OtherBusinessID); !errors.Is(err, domain.ErrRelationNotFound) {
		t.Fatalf("RemoveTaskRelation from another business = %v, want ErrRelationNotFound", err)
	}
	if counts, err := f.repo.CountOpenBlockers(f.ctx, []int64{alsoMine}, testutil.OtherBusinessID); err != nil || len(counts) != 0 {
		t.Fatalf("CountOpenBlockers from another business = %v (%v), want empty", counts, err)
	}
	if relations, err := f.repo.ListTaskRelations(f.ctx, mine, testutil.OtherBusinessID); err != nil || len(relations) != 0 {
		t.Fatalf("repository listed %d relations for another business (%v)", len(relations), err)
	}
	if got := f.groups(mine); !sameIDs(got["blocks"], alsoMine) {
		t.Fatalf("the relation was disturbed by the refused calls: %v", got)
	}

	// A row that claims this business but points at a foreign task, as a bug
	// or a hand edit could leave behind, must not surface that task or count
	// it as a blocker.
	if _, err := f.db.ExecContext(f.ctx,
		"INSERT INTO flow_task_dependencies (task_id, depends_on_task_id, link_type, business_id) VALUES (?, ?, 'blocks', ?)",
		mine, foreign, testutil.BusinessID); err != nil {
		t.Fatalf("plant cross-business row: %v", err)
	}
	if got := f.groups(mine); len(got["blocked_by"]) != 0 {
		t.Fatalf("a foreign task surfaced through a planted row: %v", got)
	}
	detail, err := f.svc.GetTaskDetail(f.ctx, mine)
	if err != nil || detail.Task.Blocked {
		t.Fatalf("a foreign task counted as a blocker: blocked=%v (%v)", detail.Task.Blocked, err)
	}
}
