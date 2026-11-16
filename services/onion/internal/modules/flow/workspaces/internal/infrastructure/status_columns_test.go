package infrastructure_test

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"
	"github.com/voxire/onion/internal/testutil"
)

// Board columns are project statuses. These run against MySQL because the
// whole point is that three columns on flow_tasks stay in step.

func columnsByName(t *testing.T, repo *infrastructure.Repository, projectID int64) map[string]int64 {
	t.Helper()
	cols, _, err := repo.GetColumnsByProject(context.Background(), projectID, 100, 0)
	if err != nil {
		t.Fatalf("GetColumnsByProject: %v", err)
	}
	out := make(map[string]int64, len(cols))
	for _, c := range cols {
		out[c.Name] = c.ID
	}
	return out
}

func taskState(t *testing.T, db *sql.DB, taskID int64) (status string, statusID, sectionID sql.NullInt64, completed bool) {
	t.Helper()
	var completedAt sql.NullTime
	if err := db.QueryRow("SELECT status, status_id, section_id, completed_at FROM flow_tasks WHERE id = ?", taskID).
		Scan(&status, &statusID, &sectionID, &completedAt); err != nil {
		t.Fatalf("read task %d: %v", taskID, err)
	}
	return status, statusID, sectionID, completedAt.Valid
}

func TestStatusColumns_BoardMovesChangeTheStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()
	const owner int64 = 991101
	projID := newStatusProbeProject(t, repo, owner, "Status Columns Probe")
	if err := repo.SeedDefaultColumns(ctx, projID, 0, owner); err != nil {
		t.Fatalf("SeedDefaultColumns: %v", err)
	}

	cols, _, err := repo.GetColumnsByProject(ctx, projID, 100, 0)
	if err != nil {
		t.Fatalf("GetColumnsByProject: %v", err)
	}
	var names []string
	for _, c := range cols {
		names = append(names, c.Name)
	}
	if got := strings.Join(names, ","); got != "Backlog,To Do,In Progress,In Review,Done,Cancelled" {
		t.Fatalf("columns = %s", got)
	}
	byName := columnsByName(t, repo, projID)

	taskID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, BusinessID: testutil.BusinessID, Title: "Columns probe",
		ReporterID: owner, OwnerID: owner, Status: domain.TaskStatusTodo,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	status, statusID, sectionID, _ := taskState(t, db, taskID)
	if status != "todo" || statusID.Int64 != byName["To Do"] || sectionID.Int64 != byName["To Do"] {
		t.Fatalf("new task: status=%s status_id=%v section_id=%v, want To Do", status, statusID, sectionID)
	}

	// Dragging to Done is a status change, with completed_at stamped.
	if err := repo.MoveTaskToColumn(ctx, taskID, byName["Done"], owner); err != nil {
		t.Fatalf("MoveTaskToColumn: %v", err)
	}
	status, statusID, sectionID, completed := taskState(t, db, taskID)
	if status != "done" || statusID.Int64 != byName["Done"] || sectionID.Int64 != byName["Done"] || !completed {
		t.Fatalf("after move: status=%s status_id=%v section_id=%v completed=%v", status, statusID, sectionID, completed)
	}

	// Setting the string status moves the card.
	if err := repo.UpdateTaskStatus(ctx, domain.UpdateTaskStatusInput{
		ID: taskID, BusinessID: testutil.BusinessID, Status: domain.TaskStatusInProgress,
	}); err != nil {
		t.Fatalf("UpdateTaskStatus: %v", err)
	}
	status, statusID, sectionID, completed = taskState(t, db, taskID)
	if status != "in_progress" || statusID.Int64 != byName["In Progress"] || sectionID.Int64 != byName["In Progress"] || completed {
		t.Fatalf("after status: status=%s status_id=%v section_id=%v completed=%v", status, statusID, sectionID, completed)
	}

	// A new column is a status, and a non-empty one cannot be deleted blind.
	designID, err := repo.CreateColumn(ctx, domain.ProjectColumnInput{ProjectID: projID, Name: "Design", SortOrder: 25, CreatedBy: owner})
	if err != nil {
		t.Fatalf("CreateColumn: %v", err)
	}
	if err := repo.MoveTaskToColumn(ctx, taskID, designID, owner); err != nil {
		t.Fatalf("move to Design: %v", err)
	}
	if err := repo.DeleteColumn(ctx, designID, owner); err == nil {
		t.Fatal("deleting a column that still has tasks should fail")
	}
}

func TestStatusTemplate_NewProjectsCopyIt(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()
	const business, owner int64 = 5302, 991201
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM flow_status_templates WHERE business_id = ?", business)
		_, _ = db.Exec("DELETE FROM flow_statuses WHERE business_id = ?", business)
		_, _ = db.Exec("DELETE FROM flow_projects WHERE business_id = ?", business)
	})

	items, err := domain.NormalizeStatusTemplate([]domain.StatusTemplateItem{
		{Name: "Inbox", Bucket: domain.StatusBucketBacklog, LegacyStatus: domain.TaskStatusTodo},
		{Name: "Doing", Bucket: domain.StatusBucketActive},
		{Name: "Shipped", Bucket: domain.StatusBucketDone},
	})
	if err != nil {
		t.Fatalf("NormalizeStatusTemplate: %v", err)
	}
	if err := repo.SaveStatusTemplate(ctx, business, items); err != nil {
		t.Fatalf("SaveStatusTemplate: %v", err)
	}

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{OwnerID: owner, BusinessID: business, Name: "Template probe", CreatedBy: owner})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := repo.SeedDefaultColumns(ctx, projID, 0, owner); err != nil {
		t.Fatalf("SeedDefaultColumns: %v", err)
	}
	statuses, err := repo.ListStatusesByProject(ctx, business, projID)
	if err != nil {
		t.Fatalf("ListStatusesByProject: %v", err)
	}
	var names []string
	defaultName := ""
	for _, s := range statuses {
		names = append(names, s.Name)
		if s.IsDefault {
			defaultName = s.Name
		}
	}
	if strings.Join(names, ",") != "Inbox,Doing,Shipped" || defaultName != "Inbox" {
		t.Fatalf("project statuses = %v default = %q, want the template with Inbox as default", names, defaultName)
	}
}

// TestStatusColumnsMigration runs v12 against a project with old-style columns:
// a custom "Design" column, and tasks whose column and status disagree.
func TestStatusColumnsMigration(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()
	const owner int64 = 991301
	projID := newStatusProbeProject(t, repo, owner, "Migration probe")

	colIDs := map[string]int64{}
	for i, name := range []string{"To Do", "In Progress", "Design", "Done"} {
		res, err := db.Exec(`INSERT INTO flow_project_columns (project_id, workspace_id, name, color, sort_order, created_by)
VALUES (?, 0, ?, '#123456', ?, ?)`, projID, name, i, owner)
		if err != nil {
			t.Fatalf("insert column: %v", err)
		}
		colIDs[name], _ = res.LastInsertId()
	}
	newTask := func(title, status, column string) int64 {
		res, err := db.Exec(`INSERT INTO flow_tasks (project_id, workspace_id, owner_id, business_id, title, status, section_id, reporter_id)
VALUES (?, 0, ?, ?, ?, ?, ?, ?)`, projID, owner, testutil.BusinessID, title, status, sql.NullInt64{Int64: colIDs[column], Valid: column != ""}, owner)
		if err != nil {
			t.Fatalf("insert task: %v", err)
		}
		id, _ := res.LastInsertId()
		return id
	}
	inDesign := newTask("todo in Design", "todo", "Design")
	doneInTodo := newTask("done in To Do", "done", "To Do")
	progressInTodo := newTask("in_progress in To Do", "in_progress", "To Do")
	backlogNoColumn := newTask("backlog nowhere", "backlog", "")

	raw, err := os.ReadFile("../../../sql/schema/flow/v12_status_columns.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	var body strings.Builder
	for _, line := range strings.Split(string(raw), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "--") {
			body.WriteString(line + "\n")
		}
	}
	// Twice: the migration must be safe to re-run after a stop.
	for run := 0; run < 2; run++ {
		for _, stmt := range strings.Split(body.String(), ";\n") {
			if strings.TrimSpace(stmt) == "" {
				continue
			}
			if _, err := db.Exec(stmt); err != nil {
				t.Fatalf("run %d: %v\n%s", run, err, stmt)
			}
		}
	}

	statuses, err := repo.ListStatusesByProject(ctx, testutil.BusinessID, projID)
	if err != nil {
		t.Fatalf("ListStatusesByProject: %v", err)
	}
	byName := map[string]domain.Status{}
	for _, s := range statuses {
		byName[s.Name] = s
	}
	for _, name := range []string{"Backlog", "To Do", "In Progress", "Design", "In Review", "Done", "Cancelled"} {
		if _, ok := byName[name]; !ok {
			t.Errorf("missing status %q (have %v)", name, statuses)
		}
	}
	if len(statuses) != 7 {
		t.Errorf("got %d statuses, want 7 (no duplicates on re-run)", len(statuses))
	}
	if !byName["To Do"].IsDefault {
		t.Error("To Do should be the default status")
	}

	check := func(taskID int64, wantStatus, wantColumn string) {
		t.Helper()
		status, statusID, sectionID, _ := taskState(t, db, taskID)
		want := byName[wantColumn].ID
		if status != wantStatus || statusID.Int64 != want || sectionID.Int64 != want {
			t.Errorf("task %d: status=%s status_id=%v section_id=%v, want %s on %s(%d)",
				taskID, status, statusID, sectionID, wantStatus, wantColumn, want)
		}
	}
	check(inDesign, "in_progress", "Design")
	check(doneInTodo, "done", "Done")
	check(progressInTodo, "in_progress", "In Progress")
	check(backlogNoColumn, "backlog", "Backlog")
}
