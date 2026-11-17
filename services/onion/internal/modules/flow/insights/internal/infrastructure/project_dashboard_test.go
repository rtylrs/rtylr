package infrastructure_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/flow/insights/internal/domain"
	"github.com/voxire/onion/internal/modules/flow/insights/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// projectDashboardFixture seeds two businesses no fixture or other test
// touches, so every count the project dashboard tests check is exact.
type projectDashboardFixture struct {
	t   *testing.T
	db  *sql.DB
	ctx context.Context
}

const (
	dashBiz, dashOther int64 = 956700001, 956700002
	// Users: the creator of every project, a project member, an outsider.
	dashCreator, dashMember, dashOutsider int64 = 956700011, 956700012, 956700013
)

func newProjectDashboardFixture(t *testing.T) projectDashboardFixture {
	t.Helper()
	f := projectDashboardFixture{t: t, db: testutil.OpenDB(t), ctx: context.Background()}
	remove := func() {
		for _, q := range []string{
			`DELETE tl FROM flow_time_logs tl JOIN flow_tasks t ON t.id = tl.task_id WHERE t.business_id IN (?, ?)`,
			`DELETE FROM flow_tasks WHERE business_id IN (?, ?)`,
			`DELETE pm FROM flow_project_members pm JOIN flow_projects p ON p.id = pm.project_id WHERE p.business_id IN (?, ?)`,
			`DELETE FROM flow_projects WHERE business_id IN (?, ?)`,
		} {
			if _, err := f.db.ExecContext(f.ctx, q, dashBiz, dashOther); err != nil {
				t.Fatalf("clear: %v", err)
			}
		}
	}
	remove()
	t.Cleanup(remove)
	return f
}

func (f projectDashboardFixture) project(businessID int64, name, visibility, status string, personal bool) int64 {
	f.t.Helper()
	res, err := f.db.ExecContext(f.ctx, `INSERT INTO flow_projects (workspace_id, owner_id, business_id, name, status, visibility, is_personal, created_by)
		VALUES (0, ?, ?, ?, ?, ?, ?, ?)`, dashCreator, businessID, name, status, visibility, personal, dashCreator)
	if err != nil {
		f.t.Fatalf("insert project %s: %v", name, err)
	}
	id, _ := res.LastInsertId()
	return id
}

func (f projectDashboardFixture) member(projectID, userID int64, role string) {
	f.t.Helper()
	if _, err := f.db.ExecContext(f.ctx, `INSERT INTO flow_project_members (project_id, user_id, role) VALUES (?, ?, ?)`, projectID, userID, role); err != nil {
		f.t.Fatalf("insert member: %v", err)
	}
}

// task inserts a task. due is a YYYY-MM-DD date or "" for none; parent is 0
// for a top-level task.
func (f projectDashboardFixture) task(businessID, projectID int64, title, status, due string, parent, assignee int64, archived bool) int64 {
	f.t.Helper()
	res, err := f.db.ExecContext(f.ctx, `INSERT INTO flow_tasks (project_id, workspace_id, owner_id, business_id, parent_task_id, title, status, due_date, assignee_id, reporter_id, archived)
		VALUES (?, 0, ?, ?, NULLIF(?, 0), ?, ?, NULLIF(?, ''), NULLIF(?, 0), ?, ?)`,
		projectID, dashCreator, businessID, parent, title, status, due, assignee, dashCreator, archived)
	if err != nil {
		f.t.Fatalf("insert task %s: %v", title, err)
	}
	id, _ := res.LastInsertId()
	return id
}

func (f projectDashboardFixture) logTime(taskID int64, minutes int, at time.Time, timeEntryID sql.NullInt64) {
	f.t.Helper()
	if _, err := f.db.ExecContext(f.ctx, `INSERT INTO flow_time_logs (owner_id, business_id, task_id, time_entry_id, user_id, duration_minutes, logged_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, dashCreator, dashBiz, taskID, timeEntryID, dashCreator, minutes, at); err != nil {
		f.t.Fatalf("insert time log: %v", err)
	}
}

func projectIDs(projects []domain.DashboardProject) []int64 {
	ids := make([]int64, 0, len(projects))
	for _, p := range projects {
		ids = append(ids, p.ID)
	}
	return ids
}

func sameIDs(got, want []int64) bool {
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

// A private project must look exactly like a missing one to anyone who is not
// its owner, creator or a project member, in the list and by id.
func TestDashboardProjectVisibility(t *testing.T) {
	f := newProjectDashboardFixture(t)
	repo := infrastructure.NewRepository(f.db)

	// Names sort the list: A, B, C, ...
	public := f.project(dashBiz, "A public site", "public", "active", false)
	private := f.project(dashBiz, "B private audit", "private", "active", false)
	viewerOnly := f.project(dashBiz, "C private viewer", "private", "completed", false)
	archived := f.project(dashBiz, "D archived public", "public", "archived", false)
	personal := f.project(dashBiz, "E My Tasks", "private", "active", true)
	foreign := f.project(dashOther, "F other business", "public", "active", false)
	f.member(private, dashMember, "member")
	f.member(viewerOnly, dashMember, "viewer")

	cases := []struct {
		name string
		user int64
		want []int64
	}{
		{"creator sees public and the private projects they created", dashCreator, []int64{public, private, viewerOnly}},
		{"member sees public and the private projects they belong to", dashMember, []int64{public, private, viewerOnly}},
		{"outsider sees only the public project", dashOutsider, []int64{public}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.ListDashboardProjects(f.ctx, dashBiz, tc.user)
			if err != nil {
				t.Fatalf("ListDashboardProjects: %v", err)
			}
			if !sameIDs(projectIDs(got), tc.want) {
				t.Errorf("projects = %+v, want ids %v", got, tc.want)
			}
		})
	}

	listed, err := repo.ListDashboardProjects(f.ctx, dashBiz, dashOutsider)
	if err != nil || len(listed) != 1 {
		t.Fatalf("outsider list = %+v, %v", listed, err)
	}
	if want := (domain.DashboardProject{ID: public, Name: "A public site", Status: "active", Visibility: "public"}); listed[0] != want {
		t.Errorf("public project = %+v, want %+v", listed[0], want)
	}

	lookups := []struct {
		name      string
		user      int64
		projectID int64
		found     bool
	}{
		{"outsider opens a public project", dashOutsider, public, true},
		{"outsider cannot open a private project", dashOutsider, private, false},
		{"member opens their private project", dashMember, private, true},
		{"viewer member opens their private project", dashMember, viewerOnly, true},
		{"creator opens their private project", dashCreator, private, true},
		{"an archived public project still opens by id", dashOutsider, archived, true},
		{"a personal My Tasks project never opens", dashCreator, personal, false},
		{"another business's project never opens", dashCreator, foreign, false},
		{"a missing project never opens", dashCreator, foreign + 1000, false},
	}
	for _, tc := range lookups {
		t.Run(tc.name, func(t *testing.T) {
			got, found, err := repo.GetDashboardProject(f.ctx, dashBiz, tc.user, tc.projectID)
			if err != nil {
				t.Fatalf("GetDashboardProject: %v", err)
			}
			if found != tc.found {
				t.Fatalf("found = %v, want %v (project %+v)", found, tc.found, got)
			}
			if found && got.ID != tc.projectID {
				t.Errorf("ID = %d, want %d", got.ID, tc.projectID)
			}
		})
	}
}

func TestProjectTaskProgress(t *testing.T) {
	f := newProjectDashboardFixture(t)
	repo := infrastructure.NewRepository(f.db)
	p := f.project(dashBiz, "Website", "public", "active", false)
	other := f.project(dashBiz, "Other", "public", "active", false)

	today := time.Now()
	day := func(offset int) string { return today.AddDate(0, 0, offset).Format(time.DateOnly) }

	f.task(dashBiz, p, "Done", "done", day(-30), 0, 0, false)
	f.task(dashBiz, p, "Done late is not overdue", "done", day(-3), 0, 0, false)
	f.task(dashBiz, p, "Cancelled is not overdue", "cancelled", day(-3), 0, 0, false)
	oldest := f.task(dashBiz, p, "Oldest overdue", "todo", day(-20), 0, dashMember, false)
	f.task(dashBiz, p, "Due today is not overdue", "todo", day(0), 0, 0, false)
	f.task(dashBiz, p, "No due date", "in_progress", "", 0, 0, false)
	parent := f.task(dashBiz, p, "Overdue parent", "in_progress", day(-10), 0, 0, false)
	sub := f.task(dashBiz, p, "Overdue subtask", "todo", day(-15), parent, 0, false)
	f.task(dashBiz, p, "Archived overdue", "todo", day(-40), 0, 0, true)
	for i := 1; i <= 4; i++ {
		f.task(dashBiz, p, "Recent overdue", "in_review", day(-i), 0, 0, false)
	}
	f.task(dashBiz, other, "Other project overdue", "todo", day(-50), 0, 0, false)

	got, err := repo.GetProjectTaskProgress(f.ctx, dashBiz, p)
	if err != nil {
		t.Fatalf("GetProjectTaskProgress: %v", err)
	}
	// Top-level, not archived: done x2, cancelled, oldest, today, no date,
	// parent and the four recent ones.
	if got.TotalTasks != 11 || got.CompletedTasks != 2 {
		t.Errorf("progress = %d/%d, want 2/11", got.CompletedTasks, got.TotalTasks)
	}
	// Oldest, parent, subtask and the four recent ones.
	if got.OverdueTasks != 7 {
		t.Errorf("OverdueTasks = %d, want 7", got.OverdueTasks)
	}
	if len(got.Overdue) != 5 {
		t.Fatalf("Overdue = %+v, want the five oldest", got.Overdue)
	}
	if got.Overdue[0].ID != oldest || got.Overdue[0].DueDate != day(-20) || got.Overdue[0].AssigneeID != dashMember || got.Overdue[0].Title != "Oldest overdue" {
		t.Errorf("Overdue[0] = %+v, want the oldest task assigned to the member", got.Overdue[0])
	}
	if got.Overdue[1].ID != sub || got.Overdue[2].ID != parent || got.Overdue[1].AssigneeID != 0 {
		t.Errorf("Overdue[1:3] = %+v, want the subtask then its parent", got.Overdue[1:3])
	}

	// Another business sees nothing of this project.
	foreign, err := repo.GetProjectTaskProgress(f.ctx, dashOther, p)
	if err != nil {
		t.Fatalf("GetProjectTaskProgress(other): %v", err)
	}
	if foreign.TotalTasks != 0 || foreign.OverdueTasks != 0 || foreign.Overdue == nil || len(foreign.Overdue) != 0 {
		t.Errorf("other business = %+v, want zeros and an empty list", foreign)
	}
}

func TestProjectTimeTotals(t *testing.T) {
	f := newProjectDashboardFixture(t)
	repo := infrastructure.NewRepository(f.db)
	p := f.project(dashBiz, "Website", "public", "active", false)
	other := f.project(dashBiz, "Other", "public", "active", false)
	taskA := f.task(dashBiz, p, "A", "todo", "", 0, 0, false)
	sub := f.task(dashBiz, p, "A.1", "todo", "", taskA, 0, false)
	otherTask := f.task(dashBiz, other, "Other", "todo", "", 0, 0, false)
	none := sql.NullInt64{}

	thisFrom := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
	thisTo := time.Date(2026, time.October, 15, 12, 0, 0, 0, time.UTC)
	priorFrom := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	priorTo := time.Date(2026, time.September, 15, 12, 0, 0, 0, time.UTC)

	f.logTime(taskA, 90, time.Date(2026, time.October, 3, 9, 0, 0, 0, time.UTC), none)
	f.logTime(sub, 30, thisTo, none)
	f.logTime(taskA, 60, priorFrom, none)
	// Between the windows and before them: all time only.
	f.logTime(taskA, 200, time.Date(2026, time.September, 20, 9, 0, 0, 0, time.UTC), none)
	f.logTime(taskA, 100, time.Date(2026, time.March, 2, 9, 0, 0, 0, time.UTC), none)
	// HR's time and another project's time never count.
	f.logTime(taskA, 5000, time.Date(2026, time.October, 4, 9, 0, 0, 0, time.UTC), sql.NullInt64{Int64: 77, Valid: true})
	f.logTime(otherTask, 700, time.Date(2026, time.October, 5, 9, 0, 0, 0, time.UTC), none)

	got, err := repo.GetProjectTimeTotals(f.ctx, dashBiz, p, thisFrom, thisTo, priorFrom, priorTo)
	if err != nil {
		t.Fatalf("GetProjectTimeTotals: %v", err)
	}
	if want := (domain.ProjectTimeTotals{AllTimeMinutes: 480, ThisMinutes: 120, PriorMinutes: 60}); got != want {
		t.Errorf("totals = %+v, want %+v", got, want)
	}

	foreign, err := repo.GetProjectTimeTotals(f.ctx, dashOther, p, thisFrom, thisTo, priorFrom, priorTo)
	if err != nil {
		t.Fatalf("GetProjectTimeTotals(other): %v", err)
	}
	if foreign != (domain.ProjectTimeTotals{}) {
		t.Errorf("other business = %+v, want zeros", foreign)
	}
}
