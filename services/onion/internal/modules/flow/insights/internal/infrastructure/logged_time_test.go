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

// TestGetLoggedTime uses businesses no fixture or other test touches, so the
// sums it checks are exact.
func TestGetLoggedTime(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const biz, other, empty int64 = 956600001, 956600002, 956600003

	removeRows := func() {
		for _, q := range []string{
			`DELETE tl FROM flow_time_logs tl JOIN flow_tasks t ON t.id = tl.task_id WHERE t.business_id IN (?, ?)`,
			`DELETE FROM flow_tasks WHERE business_id IN (?, ?)`,
			`DELETE FROM flow_projects WHERE business_id IN (?, ?)`,
		} {
			if _, err := db.ExecContext(ctx, q, biz, other); err != nil {
				t.Fatalf("clear: %v", err)
			}
		}
	}
	removeRows()
	t.Cleanup(removeRows)

	project := func(businessID int64, name, visibility, status string) int64 {
		t.Helper()
		res, err := db.ExecContext(ctx, `INSERT INTO flow_projects (workspace_id, owner_id, business_id, name, status, visibility, created_by)
			VALUES (0, ?, ?, ?, ?, ?, ?)`, businessID, businessID, name, status, visibility, businessID)
		if err != nil {
			t.Fatalf("insert project %s: %v", name, err)
		}
		id, _ := res.LastInsertId()
		return id
	}
	task := func(businessID, projectID int64) int64 {
		t.Helper()
		res, err := db.ExecContext(ctx, `INSERT INTO flow_tasks (project_id, workspace_id, owner_id, business_id, title, reporter_id)
			VALUES (?, 0, ?, ?, 'Logged task', ?)`, projectID, businessID, businessID, businessID)
		if err != nil {
			t.Fatalf("insert task: %v", err)
		}
		id, _ := res.LastInsertId()
		return id
	}
	logTime := func(taskID, businessID int64, minutes int, at time.Time, timeEntryID sql.NullInt64) {
		t.Helper()
		if _, err := db.ExecContext(ctx, `INSERT INTO flow_time_logs (owner_id, business_id, task_id, time_entry_id, user_id, duration_minutes, logged_at)
			VALUES (?, ?, ?, ?, 1, ?, ?)`, businessID, businessID, taskID, timeEntryID, minutes, at); err != nil {
			t.Fatalf("insert time log: %v", err)
		}
	}
	none := sql.NullInt64{}
	from := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.October, 15, 12, 0, 0, 0, time.UTC)

	website := project(biz, "Website", "public", "active")
	audit := project(biz, "Payroll audit", "private", "active")
	legacy := project(biz, "Legacy", "public", "archived")
	websiteTask, auditTask, legacyTask := task(biz, website), task(biz, audit), task(biz, legacy)
	// A second task on Website, logged on both window edges.
	websiteTask2 := task(biz, website)

	logTime(websiteTask, biz, 90, time.Date(2026, time.October, 3, 9, 0, 0, 0, time.UTC), none)
	logTime(websiteTask2, biz, 30, from, none)
	logTime(websiteTask2, biz, 15, to, none)
	logTime(auditTask, biz, 120, time.Date(2026, time.October, 10, 14, 0, 0, 0, time.UTC), none)
	logTime(legacyTask, biz, 45, time.Date(2026, time.October, 6, 11, 0, 0, 0, time.UTC), none)
	// Outside the window by a second either side.
	logTime(websiteTask, biz, 1000, from.Add(-time.Second), none)
	logTime(websiteTask, biz, 2000, to.Add(time.Second), none)
	// Linked to an HR time entry: HR's time, not Flow's.
	logTime(websiteTask, biz, 3000, time.Date(2026, time.October, 4, 9, 0, 0, 0, time.UTC), sql.NullInt64{Int64: 77, Valid: true})
	// Another business's time stays out of biz's totals.
	otherProject := project(other, "Other website", "public", "active")
	logTime(task(other, otherProject), other, 500, time.Date(2026, time.October, 5, 9, 0, 0, 0, time.UTC), none)

	got, err := repo.GetLoggedTime(ctx, biz, from, to)
	if err != nil {
		t.Fatalf("GetLoggedTime: %v", err)
	}
	// 90 + 30 + 15 on Website, 120 on the private audit, 45 on the archived project.
	if got.TotalMinutes != 300 {
		t.Errorf("TotalMinutes = %d, want 300", got.TotalMinutes)
	}
	want := []domain.ProjectLoggedTime{
		{ProjectID: website, Name: "Website", Visibility: "public", Status: "active", Minutes: 135},
		{ProjectID: audit, Name: "Payroll audit", Visibility: "private", Status: "active", Minutes: 120},
		{ProjectID: legacy, Name: "Legacy", Visibility: "public", Status: "archived", Minutes: 45},
	}
	if len(got.Projects) != len(want) {
		t.Fatalf("Projects = %+v, want %+v", got.Projects, want)
	}
	for i := range want {
		if got.Projects[i] != want[i] {
			t.Errorf("Projects[%d] = %+v, want %+v", i, got.Projects[i], want[i])
		}
	}

	otherTime, err := repo.GetLoggedTime(ctx, other, from, to)
	if err != nil {
		t.Fatalf("GetLoggedTime(other): %v", err)
	}
	if otherTime.TotalMinutes != 500 || len(otherTime.Projects) != 1 || otherTime.Projects[0].ProjectID != otherProject {
		t.Errorf("other business = %+v, want 500 minutes on project %d only", otherTime, otherProject)
	}

	nothing, err := repo.GetLoggedTime(ctx, empty, from, to)
	if err != nil {
		t.Fatalf("GetLoggedTime(empty): %v", err)
	}
	if nothing.TotalMinutes != 0 || nothing.Projects == nil || len(nothing.Projects) != 0 {
		t.Errorf("business without time = %+v, want zero minutes and an empty project list", nothing)
	}
}
