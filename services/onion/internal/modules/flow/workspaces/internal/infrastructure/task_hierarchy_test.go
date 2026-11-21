package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"
	"github.com/voxire/onion/internal/testutil"
)

type hierarchyFixture struct {
	t          *testing.T
	db         *sql.DB
	repo       *infrastructure.Repository
	businessID int64
	projectID  int64
}

func newHierarchyFixture(t *testing.T, name string, businessID int64) *hierarchyFixture {
	t.Helper()
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	projectID, err := repo.CreateProject(ctx, domain.ProjectInput{
		OwnerID: testutil.OwnerID, BusinessID: businessID, Name: name, CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE tl FROM flow_time_logs tl JOIN flow_tasks t ON t.id = tl.task_id WHERE t.project_id = ?", projectID)
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_tasks WHERE project_id = ?", projectID)
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_statuses WHERE project_id = ?", projectID)
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projectID)
	})
	return &hierarchyFixture{t: t, db: db, repo: repo, businessID: businessID, projectID: projectID}
}

func (f *hierarchyFixture) task(title string, mutate func(*domain.TaskInput)) int64 {
	f.t.Helper()
	input := domain.TaskInput{
		ProjectID: f.projectID, OwnerID: testutil.OwnerID, BusinessID: f.businessID,
		Title: title, ReporterID: testutil.OwnerID,
	}
	if mutate != nil {
		mutate(&input)
	}
	id, err := f.repo.CreateTask(context.Background(), input)
	if err != nil {
		f.t.Fatalf("CreateTask(%s): %v", title, err)
	}
	return id
}

func (f *hierarchyFixture) epic(title, color string) int64 {
	return f.task(title, func(in *domain.TaskInput) {
		in.TaskType = domain.TaskTypeEpic
		in.EpicColor = color
	})
}

func (f *hierarchyFixture) under(epicID int64, title string, mutate func(*domain.TaskInput)) int64 {
	return f.task(title, func(in *domain.TaskInput) {
		in.EpicID = &epicID
		if mutate != nil {
			mutate(in)
		}
	})
}

func (f *hierarchyFixture) subtask(parentID int64, title string, mutate func(*domain.TaskInput)) int64 {
	return f.task(title, func(in *domain.TaskInput) {
		in.ParentTaskID = &parentID
		if mutate != nil {
			mutate(in)
		}
	})
}

func (f *hierarchyFixture) exec(query string, args ...interface{}) {
	f.t.Helper()
	if _, err := f.db.ExecContext(context.Background(), query, args...); err != nil {
		f.t.Fatalf("%s: %v", query, err)
	}
}

func (f *hierarchyFixture) logTime(taskID int64, minutes int) {
	f.exec("INSERT INTO flow_time_logs (owner_id, business_id, task_id, user_id, duration_minutes) VALUES (?, ?, ?, ?, ?)",
		testutil.OwnerID, f.businessID, taskID, testutil.OwnerID, minutes)
}

func (f *hierarchyFixture) placement(taskID int64) (parentID, epicID sql.NullInt64) {
	f.t.Helper()
	err := f.db.QueryRowContext(context.Background(),
		"SELECT parent_task_id, epic_id FROM flow_tasks WHERE id = ?", taskID).Scan(&parentID, &epicID)
	if err != nil {
		f.t.Fatalf("read placement of %d: %v", taskID, err)
	}
	return parentID, epicID
}

func (f *hierarchyFixture) exists(taskID int64) bool {
	f.t.Helper()
	var n int
	if err := f.db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM flow_tasks WHERE id = ?", taskID).Scan(&n); err != nil {
		f.t.Fatalf("count task %d: %v", taskID, err)
	}
	return n == 1
}

func placeByRules(child domain.HierarchyNode, parent *domain.HierarchyNode) (domain.TaskPlacement, error) {
	if parent == nil {
		return domain.TaskPlacement{}, domain.ErrTaskParentNotFound
	}
	return domain.PlaceUnderParent(child, *parent)
}

func effort(minutes int32) func(*domain.TaskInput) {
	return func(in *domain.TaskInput) { in.EffortMinutes = &minutes }
}

func TestWorkspaces_EpicRollup_SeededFixture(t *testing.T) {
	f := newHierarchyFixture(t, "EpicRollup", testutil.BusinessID)
	ctx := context.Background()

	epicID := f.epic("Checkout rebuild", "#ef4444")
	emptyEpicID := f.epic("Nothing yet", "")

	backlog := f.under(epicID, "Backlog task", effort(60))
	hours := 1.5
	active := f.under(epicID, "Active task", func(in *domain.TaskInput) {
		in.Status = domain.TaskStatusInProgress
		in.EstimatedHours = &hours
	})
	done := f.under(epicID, "Done task", func(in *domain.TaskInput) {
		in.Status = domain.TaskStatusDone
		effort(30)(in)
	})
	// A custom status decides the bucket even when the built-in string says
	// otherwise: this one reads in_review but its status is a done status.
	customDone := f.under(epicID, "Shipped task", func(in *domain.TaskInput) {
		in.Status = domain.TaskStatusInReview
		effort(45)(in)
	})
	res, err := f.db.ExecContext(ctx,
		`INSERT INTO flow_statuses (business_id, project_id, name, bucket, legacy_status, created_by)
		 VALUES (?, ?, 'Shipped', 'done', 'in_review', ?)`, f.businessID, f.projectID, testutil.OwnerID)
	if err != nil {
		t.Fatalf("insert status: %v", err)
	}
	statusID, _ := res.LastInsertId()
	f.exec("UPDATE flow_tasks SET status_id = ? WHERE id = ?", statusID, customDone)

	f.under(epicID, "Cancelled task", func(in *domain.TaskInput) { in.Status = domain.TaskStatusCancelled })
	archived := f.under(epicID, "Archived task", effort(500))
	f.exec("UPDATE flow_tasks SET archived = TRUE WHERE id = ?", archived)

	// Subtask estimates and counts stay out; subtask time comes in.
	sub := f.subtask(backlog, "Subtask", effort(999))
	outside := f.task("Not under any epic", effort(77))

	f.logTime(backlog, 15)
	f.logTime(active, 10)
	f.logTime(active, 5)
	f.logTime(sub, 20)
	f.logTime(archived, 100)
	f.logTime(outside, 77)
	_ = done

	rollups, err := f.repo.ListEpicRollups(ctx, f.projectID, f.businessID)
	if err != nil {
		t.Fatalf("ListEpicRollups: %v", err)
	}
	if len(rollups) != 2 {
		t.Fatalf("got %d epics, want 2: %+v", len(rollups), rollups)
	}
	byID := map[int64]domain.EpicRollup{}
	for _, r := range rollups {
		byID[r.EpicID] = r
	}

	got := byID[epicID]
	want := domain.EpicRollup{
		EpicID: epicID, Identifier: got.Identifier, Title: "Checkout rebuild", Color: "#ef4444", Status: domain.TaskStatusTodo,
		TotalTasks: 5, BacklogTasks: 1, ActiveTasks: 1, DoneTasks: 2, CancelledTasks: 1,
		EstimateMinutes: 60 + 90 + 30 + 45, DoneEstimateMinutes: 30 + 45,
		LoggedMinutes: 15 + 10 + 5 + 20,
	}
	if got != want {
		t.Fatalf("rollup = %+v\n   want = %+v", got, want)
	}
	if got.Identifier == "" {
		t.Fatal("rollup carries no identifier")
	}

	empty := byID[emptyEpicID]
	if empty.Title != "Nothing yet" || empty.TotalTasks != 0 || empty.EstimateMinutes != 0 || empty.LoggedMinutes != 0 || empty.Color != "" {
		t.Fatalf("an epic with no tasks should report zeros: %+v", empty)
	}

	// Another business asking about this project sees none of it.
	foreign, err := f.repo.ListEpicRollups(ctx, f.projectID, testutil.OtherBusinessID)
	if err != nil {
		t.Fatalf("ListEpicRollups as other business: %v", err)
	}
	if len(foreign) != 0 {
		t.Fatalf("cross-tenant leak: %+v", foreign)
	}
}

func TestWorkspaces_TaskReadsCarryTheirEpic(t *testing.T) {
	f := newHierarchyFixture(t, "EpicProjection", testutil.BusinessID)
	ctx := context.Background()

	epicID := f.epic("Billing", "#22c55e")
	childID := f.under(epicID, "Invoices", func(in *domain.TaskInput) { in.TaskType = domain.TaskTypeFeature })
	subID := f.subtask(childID, "PDF layout", nil)
	looseID := f.task("Loose bug", func(in *domain.TaskInput) { in.TaskType = domain.TaskTypeBug })

	epic, err := f.repo.GetTaskByID(ctx, epicID)
	if err != nil {
		t.Fatal(err)
	}
	if epic.TaskType != domain.TaskTypeEpic || epic.EpicColor != "#22c55e" || epic.EpicID != nil || epic.EpicTitle != "" {
		t.Fatalf("epic read wrong: type=%q color=%q epic=%v title=%q", epic.TaskType, epic.EpicColor, epic.EpicID, epic.EpicTitle)
	}
	for name, id := range map[string]int64{"task under the epic": childID, "its subtask": subID} {
		task, err := f.repo.GetTaskByID(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if task.EpicID == nil || *task.EpicID != epicID || task.EpicTitle != "Billing" || task.EpicColor != "#22c55e" {
			t.Fatalf("%s lost its epic: epic=%v title=%q color=%q", name, task.EpicID, task.EpicTitle, task.EpicColor)
		}
	}
	child, _ := f.repo.GetTaskByID(ctx, childID)
	if child.ParentTaskID != nil || child.TaskType != domain.TaskTypeFeature {
		t.Fatalf("a task under an epic must stay top-level: parent=%v type=%q", child.ParentTaskID, child.TaskType)
	}
	loose, _ := f.repo.GetTaskByID(ctx, looseID)
	if loose.EpicID != nil || loose.EpicTitle != "" || loose.EpicColor != "" || loose.TaskType != domain.TaskTypeBug {
		t.Fatalf("a task with no epic reported one: %+v", loose)
	}

	// The epic's "subtasks" are its direct tasks; the task's are unchanged.
	ofEpic, total, err := f.repo.GetSubtasks(ctx, epicID, 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(ofEpic) != 1 || ofEpic[0].ID != childID {
		t.Fatalf("GetSubtasks(epic) = %d rows (total %d), want only task %d", len(ofEpic), total, childID)
	}
	ofChild, _, err := f.repo.GetSubtasks(ctx, childID, 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(ofChild) != 1 || ofChild[0].ID != subID {
		t.Fatalf("GetSubtasks(task) = %+v, want only subtask %d", ofChild, subID)
	}

	list := func(filter domain.ListTasksFilter) []int64 {
		t.Helper()
		filter.ProjectID, filter.BusinessID, filter.Limit = f.projectID, f.businessID, 50
		tasks, _, err := f.repo.GetTasksByBusinessV2(ctx, filter, testutil.OwnerID, true)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		ids := make([]int64, 0, len(tasks))
		for _, task := range tasks {
			ids = append(ids, task.ID)
		}
		return ids
	}
	assertIDs := func(what string, got []int64, want ...int64) {
		t.Helper()
		if len(got) != len(want) {
			t.Fatalf("%s = %v, want %v", what, got, want)
		}
		seen := map[int64]bool{}
		for _, id := range got {
			seen[id] = true
		}
		for _, id := range want {
			if !seen[id] {
				t.Fatalf("%s = %v, want %v", what, got, want)
			}
		}
	}
	assertIDs("unfiltered list", list(domain.ListTasksFilter{}), epicID, childID, looseID)
	assertIDs("epic filter", list(domain.ListTasksFilter{EpicID: &epicID}), childID)
	assertIDs("no-epic filter", list(domain.ListTasksFilter{NoEpic: true}), looseID)
	assertIDs("type filter", list(domain.ListTasksFilter{TaskTypes: []string{domain.TaskTypeEpic}}), epicID)
	assertIDs("multi-type filter", list(domain.ListTasksFilter{TaskTypes: []string{domain.TaskTypeBug, domain.TaskTypeFeature}}), childID, looseID)
}

func TestWorkspaces_SetTaskParent_MovesAndRefusesLoops(t *testing.T) {
	f := newHierarchyFixture(t, "SetParent", testutil.BusinessID)
	ctx := context.Background()

	epicID := f.epic("Epic", "")
	parentID := f.task("Parent", nil)
	subID := f.subtask(parentID, "Sub", nil)
	looseID := f.task("Loose", nil)

	// A nil target clears the placement, as the service's decide does; a
	// named target that is missing reaches placeByRules as a nil parent.
	move := func(taskID int64, to *int64) error {
		decide := placeByRules
		if to == nil {
			decide = func(domain.HierarchyNode, *domain.HierarchyNode) (domain.TaskPlacement, error) {
				return domain.TaskPlacement{}, nil
			}
		}
		return f.repo.SetTaskParent(ctx, taskID, f.businessID, to, decide)
	}

	// A task that has subtasks may still go under an epic.
	if err := move(parentID, &epicID); err != nil {
		t.Fatalf("task with subtasks under an epic: %v", err)
	}
	if p, e := f.placement(parentID); p.Valid || e.Int64 != epicID {
		t.Fatalf("after move: parent=%v epic=%v", p, e)
	}
	if p, e := f.placement(subID); p.Int64 != parentID || e.Valid {
		t.Fatalf("the subtask must be untouched: parent=%v epic=%v", p, e)
	}

	// Becoming a subtask drops the task's own epic link.
	if err := move(looseID, &epicID); err != nil {
		t.Fatal(err)
	}
	other := f.task("Other", nil)
	if err := move(looseID, &other); err != nil {
		t.Fatalf("task under a task: %v", err)
	}
	if p, e := f.placement(looseID); p.Int64 != other || e.Valid {
		t.Fatalf("after becoming a subtask: parent=%v epic=%v", p, e)
	}

	refused := []struct {
		name   string
		taskID int64
		to     int64
		want   error
	}{
		{"under itself", parentID, parentID, domain.ErrTaskParentCycle},
		{"under its own subtask", parentID, subID, domain.ErrTaskParentCycle},
		{"under another task's subtask", other, subID, domain.ErrTaskParentTooDeep},
		{"with subtasks, under a task", parentID, f.task("Target", nil), domain.ErrTaskParentTooDeep},
		{"an epic under a task", epicID, other, domain.ErrEpicCannotHaveParent},
	}
	for _, tc := range refused {
		beforeParent, beforeEpic := f.placement(tc.taskID)
		to := tc.to
		if err := move(tc.taskID, &to); !errors.Is(err, tc.want) {
			t.Fatalf("%s: error = %v, want %v", tc.name, err, tc.want)
		}
		if p, e := f.placement(tc.taskID); p != beforeParent || e != beforeEpic {
			t.Fatalf("%s: a refused move changed the task: parent=%v epic=%v", tc.name, p, e)
		}
	}

	// A parent in another project of the same business is refused too.
	elsewhere := newHierarchyFixture(t, "SetParentElsewhere", testutil.BusinessID)
	foreignProjectEpic := elsewhere.epic("Elsewhere", "")
	if err := move(other, &foreignProjectEpic); !errors.Is(err, domain.ErrTaskParentOtherProject) {
		t.Fatalf("parent in another project: %v", err)
	}

	// Clearing returns a task to the top level with no epic.
	if err := move(parentID, nil); err != nil {
		t.Fatal(err)
	}
	if p, e := f.placement(parentID); p.Valid || e.Valid {
		t.Fatalf("after clear: parent=%v epic=%v", p, e)
	}
}

func TestWorkspaces_TaskHierarchy_IsolatedByBusiness(t *testing.T) {
	mine := newHierarchyFixture(t, "HierarchyMine", testutil.BusinessID)
	theirs := newHierarchyFixture(t, "HierarchyTheirs", testutil.OtherBusinessID)
	ctx := context.Background()

	myTask := mine.task("Mine", nil)
	theirEpic := theirs.epic("Their epic", "#000000")
	theirTask := theirs.under(theirEpic, "Their task", nil)

	// Their epic cannot be named as my task's parent.
	if err := mine.repo.SetTaskParent(ctx, myTask, mine.businessID, &theirEpic, placeByRules); !errors.Is(err, domain.ErrTaskParentNotFound) {
		t.Fatalf("foreign parent: error = %v, want not found", err)
	}
	if p, e := mine.placement(myTask); p.Valid || e.Valid {
		t.Fatalf("cross-tenant write: parent=%v epic=%v", p, e)
	}

	// Their task cannot be moved, or detached, through my business.
	if err := mine.repo.SetTaskParent(ctx, theirTask, mine.businessID, nil, placeByRules); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("moving a foreign task: error = %v, want no rows", err)
	}
	if _, e := theirs.placement(theirTask); e.Int64 != theirEpic {
		t.Fatalf("cross-tenant write detached their task: epic=%v", e)
	}

	// Nor can their epic be read or deleted through it.
	if _, found, err := mine.repo.GetTaskHierarchyNode(ctx, theirEpic, mine.businessID); err != nil || found {
		t.Fatalf("foreign node visible: found=%v err=%v", found, err)
	}
	detached, err := mine.repo.DeleteTaskDetachingEpicTasks(ctx, theirEpic, mine.businessID)
	if err != nil || detached != 0 {
		t.Fatalf("foreign delete: detached=%d err=%v", detached, err)
	}
	if !theirs.exists(theirEpic) {
		t.Fatal("cross-tenant delete removed their epic")
	}
	if _, e := theirs.placement(theirTask); e.Int64 != theirEpic {
		t.Fatalf("cross-tenant delete detached their task: epic=%v", e)
	}

	// A stray epic_id pointing across businesses resolves to nothing on read.
	mine.exec("UPDATE flow_tasks SET epic_id = ? WHERE id = ?", theirEpic, myTask)
	task, err := mine.repo.GetTaskByID(ctx, myTask)
	if err != nil {
		t.Fatal(err)
	}
	if task.EpicID != nil || task.EpicTitle != "" || task.EpicColor != "" {
		t.Fatalf("cross-tenant leak through the epic join: %+v", task)
	}
}

func TestWorkspaces_DeleteEpic_DetachesItsTasksAtomically(t *testing.T) {
	f := newHierarchyFixture(t, "DeleteEpic", testutil.BusinessID)
	ctx := context.Background()

	epicID := f.epic("Doomed epic", "#111111")
	first := f.under(epicID, "First", nil)
	second := f.under(epicID, "HierarchyDetachBlocked", nil)
	sub := f.subtask(first, "Sub of first", nil)

	// Make the detach fail after the epic row is already deleted inside the
	// transaction. If the two were not one unit the epic would be gone and its
	// tasks left pointing at nothing.
	f.exec("DROP TRIGGER IF EXISTS trg_test_hierarchy_detach_blocked")
	f.exec(`CREATE TRIGGER trg_test_hierarchy_detach_blocked BEFORE UPDATE ON flow_tasks FOR EACH ROW
		BEGIN
			IF OLD.title = 'HierarchyDetachBlocked' AND OLD.epic_id IS NOT NULL AND NEW.epic_id IS NULL THEN
				SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'detach blocked by test';
			END IF;
		END`)
	dropTrigger := func() { f.exec("DROP TRIGGER IF EXISTS trg_test_hierarchy_detach_blocked") }
	t.Cleanup(dropTrigger)

	if _, err := f.repo.DeleteTaskDetachingEpicTasks(ctx, epicID, f.businessID); err == nil {
		t.Fatal("expected the blocked detach to fail the delete")
	}
	if !f.exists(epicID) {
		t.Fatal("the epic was deleted although its tasks could not be detached")
	}
	for _, id := range []int64{first, second} {
		if _, e := f.placement(id); e.Int64 != epicID {
			t.Fatalf("task %d was detached by a delete that rolled back: epic=%v", id, e)
		}
	}

	dropTrigger()
	detached, err := f.repo.DeleteTaskDetachingEpicTasks(ctx, epicID, f.businessID)
	if err != nil {
		t.Fatalf("delete epic: %v", err)
	}
	if detached != 2 {
		t.Fatalf("detached = %d, want 2", detached)
	}
	if f.exists(epicID) {
		t.Fatal("the epic survived its delete")
	}
	for _, id := range []int64{first, second} {
		if !f.exists(id) {
			t.Fatalf("task %d was deleted with its epic", id)
		}
		if p, e := f.placement(id); p.Valid || e.Valid {
			t.Fatalf("task %d still attached: parent=%v epic=%v", id, p, e)
		}
	}
	if p, _ := f.placement(sub); p.Int64 != first {
		t.Fatalf("the subtask lost its parent: %v", p)
	}

	// Deleting an ordinary task reports nothing detached and, as before,
	// leaves its subtasks in place.
	detached, err = f.repo.DeleteTaskDetachingEpicTasks(ctx, first, f.businessID)
	if err != nil || detached != 0 {
		t.Fatalf("delete task: detached=%d err=%v", detached, err)
	}
	if !f.exists(sub) {
		t.Fatal("deleting a task removed its subtask, which it never did before")
	}
}
