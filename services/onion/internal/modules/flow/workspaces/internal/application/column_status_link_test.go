package application

import (
	"context"
	"database/sql"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// Board columns and task status were stored apart: dropping a card on Done
// left the task To Do, and marking it done left the card where it was. The
// four default columns are now tied to a status both ways; a column someone
// adds is not.

type linkFixture struct {
	svc     *Service
	ctx     context.Context
	db      *sql.DB
	project domain.Project
	columns map[string]int64 // by name
}

func newLinkFixture(t *testing.T) *linkFixture {
	t.Helper()
	svc, ctx, db := v2Svc(t)
	p := createV2Project(ctx, t, svc, "column-status-link", domain.VisibilityPrivate, v2UserA)
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM flow_project_columns WHERE project_id = ?", p.ID) })
	cols, _, err := svc.GetColumnsByProject(ctx, p.ID, 50, 0)
	if err != nil {
		t.Fatalf("GetColumnsByProject: %v", err)
	}
	f := &linkFixture{svc: svc, ctx: ctx, db: db, project: p, columns: map[string]int64{}}
	for _, c := range cols {
		f.columns[c.Name] = c.ID
	}
	custom, err := svc.CreateColumn(ctx, domain.ProjectColumnInput{ProjectID: p.ID, Name: "Blocked", CreatedBy: v2UserA})
	if err != nil {
		t.Fatalf("CreateColumn: %v", err)
	}
	f.columns["Blocked"] = custom.ID
	return f
}

func (f *linkFixture) task(t *testing.T, section *int64, status string) domain.Task {
	t.Helper()
	task, err := f.svc.CreateTask(f.ctx, domain.TaskInput{
		ProjectID: f.project.ID, OwnerID: v2UserA, BusinessID: v2Business,
		Title: "linked", ReporterID: v2UserA, SectionID: section, Status: status,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	return f.reload(t, task.ID)
}

func (f *linkFixture) reload(t *testing.T, id int64) domain.Task {
	t.Helper()
	task, err := f.svc.repo.GetTaskByID(f.ctx, id)
	if err != nil {
		t.Fatalf("GetTaskByID: %v", err)
	}
	return task
}

func ptr(v int64) *int64 { return &v }

func sectionOf(task domain.Task) int64 {
	if task.SectionID == nil {
		return 0
	}
	return *task.SectionID
}

func TestDroppingACardOnALinkedColumnSetsItsStatus(t *testing.T) {
	f := newLinkFixture(t)
	task := f.task(t, ptr(f.columns["To Do"]), "")

	if err := f.svc.MoveTaskToColumnOrdered(f.ctx, task.ID, f.columns["Done"], v2UserA, []int64{task.ID}); err != nil {
		t.Fatalf("move to Done: %v", err)
	}
	got := f.reload(t, task.ID)
	if got.Status != domain.TaskStatusDone || got.CompletedAt == nil {
		t.Fatalf("after dropping on Done: status %q completed_at %v, want done and stamped", got.Status, got.CompletedAt)
	}

	if err := f.svc.MoveTaskToColumn(f.ctx, task.ID, f.columns["In Progress"], v2UserA); err != nil {
		t.Fatalf("move to In Progress: %v", err)
	}
	if got := f.reload(t, task.ID); got.Status != "in_progress" || got.CompletedAt != nil {
		t.Fatalf("after dropping on In Progress: status %q completed_at %v", got.Status, got.CompletedAt)
	}

	// A column someone added has no status, so a drop there leaves it alone.
	if err := f.svc.MoveTaskToColumn(f.ctx, task.ID, f.columns["Blocked"], v2UserA); err != nil {
		t.Fatalf("move to Blocked: %v", err)
	}
	if got := f.reload(t, task.ID); got.Status != "in_progress" {
		t.Fatalf("after dropping on a custom column: status %q, want it unchanged", got.Status)
	}
}

func TestChangingStatusMovesTheCardToItsColumn(t *testing.T) {
	f := newLinkFixture(t)
	task := f.task(t, ptr(f.columns["To Do"]), "")

	if err := f.svc.UpdateTaskStatus(f.ctx, domain.UpdateTaskStatusInput{
		ID: task.ID, OwnerID: v2UserA, BusinessID: v2Business, Status: domain.TaskStatusDone,
	}, v2UserA); err != nil {
		t.Fatalf("UpdateTaskStatus: %v", err)
	}
	if got := f.reload(t, task.ID); sectionOf(got) != f.columns["Done"] {
		t.Fatalf("after marking done the card is in column %d, want Done (%d)", sectionOf(got), f.columns["Done"])
	}

	if err := f.svc.UpdateTask(f.ctx, domain.UpdateTaskInput{
		ID: task.ID, BusinessID: v2Business, Status: "in_review",
	}, v2UserA); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	if got := f.reload(t, task.ID); sectionOf(got) != f.columns["In Review"] {
		t.Fatalf("after a partial update to in_review the card is in column %d, want In Review", sectionOf(got))
	}

	// Backlog has no column, so the card stays where it is.
	if err := f.svc.UpdateTaskStatus(f.ctx, domain.UpdateTaskStatusInput{
		ID: task.ID, OwnerID: v2UserA, BusinessID: v2Business, Status: "backlog",
	}, v2UserA); err != nil {
		t.Fatalf("UpdateTaskStatus backlog: %v", err)
	}
	if got := f.reload(t, task.ID); sectionOf(got) != f.columns["In Review"] {
		t.Fatalf("after moving to backlog the card is in column %d, want it left in In Review", sectionOf(got))
	}
}

func TestNewTasksTakeTheColumnAndStatusTheyBelongTo(t *testing.T) {
	f := newLinkFixture(t)

	// The header's New task sends no column; the task used to have none.
	if got := f.task(t, nil, ""); sectionOf(got) != f.columns["To Do"] {
		t.Fatalf("a new task with no column is in %d, want To Do (%d)", sectionOf(got), f.columns["To Do"])
	}
	if got := f.task(t, nil, "in_progress"); sectionOf(got) != f.columns["In Progress"] {
		t.Fatalf("a new in-progress task is in %d, want In Progress", sectionOf(got))
	}
	// Backlog has no column of its own, so it starts in the first column.
	if got := f.task(t, nil, "backlog"); sectionOf(got) != f.columns["To Do"] {
		t.Fatalf("a new backlog task is in %d, want the first column", sectionOf(got))
	}
	// Add task inside a column gives the task that column's status.
	if got := f.task(t, ptr(f.columns["Done"]), ""); got.Status != domain.TaskStatusDone {
		t.Fatalf("a task added in Done has status %q, want done", got.Status)
	}
	if got := f.task(t, ptr(f.columns["Blocked"]), ""); got.Status != domain.TaskStatusTodo || sectionOf(got) != f.columns["Blocked"] {
		t.Fatalf("a task added in a custom column: status %q column %d", got.Status, sectionOf(got))
	}
}
