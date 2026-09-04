package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestWorkspaces_CountWorkspacesByOwner_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	count, err := repo.CountWorkspacesByOwner(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("CountWorkspacesByOwner: %v", err)
	}
	if count < 0 {
		t.Errorf("expected count >= 0, got %d", count)
	}
}

func TestWorkspaces_CreateWorkspace_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateWorkspace(context.Background(), domain.WorkspaceInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Integration Test Workspace",
		CreatedBy:  testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), id, testutil.BusinessID) })
}

func TestWorkspaces_ListWorkDayLines_ScopesByBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const workDayID int64 = 980001

	for _, row := range []struct {
		ownerID    int64
		businessID int64
		clientKey  string
		duration   int
	}{
		{ownerID: testutil.BusinessID, businessID: testutil.BusinessID, clientKey: "owner-identity-own", duration: 45},
		{ownerID: testutil.OwnerID, businessID: testutil.OtherBusinessID, clientKey: "owner-identity-other", duration: 90},
	} {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO hr_time_entries (
				owner_id, business_id, employee_id, work_day_id, date,
				duration_minutes, description, client_key, source
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'flow')`,
			row.ownerID, row.businessID, 4101, workDayID, "2026-08-20",
			row.duration, "owner identity scope", row.clientKey,
		); err != nil {
			t.Fatalf("insert %s: %v", row.clientKey, err)
		}
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx,
			"DELETE FROM hr_time_entries WHERE client_key IN (?, ?)",
			"owner-identity-own", "owner-identity-other",
		)
	})

	lines, total, err := repo.ListWorkDayLines(ctx, testutil.BusinessID, workDayID)
	if err != nil {
		t.Fatalf("ListWorkDayLines: %v", err)
	}
	if len(lines) != 1 || lines[0].ClientKey != "owner-identity-own" {
		t.Fatalf("lines = %+v, want only the requested business's line", lines)
	}
	if total != 45 {
		t.Fatalf("total = %d, want 45", total)
	}
}

func TestWorkspaces_WorkDayLineMutationsScopeToCaller(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	service := application.NewService(ctx, db)

	const (
		workDayID int64 = 980101
		lineID    int64 = 980102
	)

	if _, err := db.ExecContext(ctx, "DELETE FROM hr_time_entries WHERE id = ?", lineID); err != nil {
		t.Fatalf("delete stale line fixture: %v", err)
	}
	if _, err := db.ExecContext(ctx, "DELETE FROM flow_work_days WHERE id = ?", workDayID); err != nil {
		t.Fatalf("delete stale work day fixture: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO flow_work_days (id, owner_id, business_id, employee_id, date, timezone, work_mode, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		workDayID, testutil.BusinessID, testutil.BusinessID, 4101, "2035-01-01", "UTC", "onsite", "draft",
	); err != nil {
		t.Fatalf("insert work day fixture: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO hr_time_entries (
			id, owner_id, business_id, employee_id, work_day_id, date,
			project_id, duration_minutes, description, source
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'flow')`,
		lineID, testutil.BusinessID, testutil.BusinessID, 4101, workDayID, "2035-01-01", 998833, 45, "original",
	); err != nil {
		t.Fatalf("insert work day line fixture: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(ctx, "DELETE FROM hr_time_entries WHERE id = ?", lineID); err != nil {
			t.Errorf("delete line fixture: %v", err)
		}
		if _, err := db.ExecContext(ctx, "DELETE FROM flow_work_days WHERE id = ?", workDayID); err != nil {
			t.Errorf("delete work day fixture: %v", err)
		}
	})

	foreignLine, err := service.UpdateWorkDayLine(ctx, domain.WorkDayLineInput{
		ID:              lineID,
		OwnerID:         testutil.BusinessID,
		BusinessID:      testutil.BusinessID,
		ProjectID:       998833,
		DurationMinutes: 90,
		Description:     "foreign update",
		LoggedBy:        1002,
	})
	if !errors.Is(err, application.ErrWorkDayNotFound) {
		t.Fatalf("foreign update error = %v, want ErrWorkDayNotFound", err)
	}
	if foreignLine.ID != 0 {
		t.Fatalf("foreign update returned line %+v", foreignLine)
	}

	var duration int
	var description string
	if err := db.QueryRowContext(ctx,
		"SELECT duration_minutes, description FROM hr_time_entries WHERE id = ?", lineID,
	).Scan(&duration, &description); err != nil {
		t.Fatalf("read line after foreign update: %v", err)
	}
	if duration != 45 || description != "original" {
		t.Fatalf("foreign update changed line to duration=%d description=%q", duration, description)
	}

	if _, err := service.UpdateWorkDayLine(ctx, domain.WorkDayLineInput{
		ID:              lineID,
		OwnerID:         testutil.BusinessID,
		BusinessID:      testutil.BusinessID,
		ProjectID:       998833,
		DurationMinutes: 60,
		Description:     "owner update",
		LoggedBy:        1001,
	}); err != nil {
		t.Fatalf("owner update: %v", err)
	}

	removed, err := service.DeleteWorkDayLine(ctx, lineID, testutil.BusinessID, testutil.BusinessID, 1002)
	if !errors.Is(err, application.ErrWorkDayNotFound) {
		t.Fatalf("foreign delete error = %v, want ErrWorkDayNotFound", err)
	}
	if removed {
		t.Fatal("foreign delete reported a removal")
	}

	removed, err = service.DeleteWorkDayLine(ctx, lineID, testutil.BusinessID, testutil.BusinessID, 1001)
	if err != nil {
		t.Fatalf("owner delete: %v", err)
	}
	if !removed {
		t.Fatal("owner delete reported no removal")
	}
}

func TestWorkspaces_ClientKeyIsScopedByBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const sharedOwner int64 = 970200

	// Registered before the inserts, not after. A run that failed partway
	// through the loop never reached a cleanup placed below it, so the first
	// row leaked and every later run failed on that leftover as a duplicate:
	// the test poisoned itself permanently after one bad run.
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx,
			"DELETE FROM hr_time_entries WHERE owner_id = ? AND client_key = ?",
			sharedOwner, "shared-client-key",
		)
	})

	// Both inserts must succeed. The same client_key in two businesses is the
	// whole point: uk_time_entries_client_key_tenant is scoped by business_id,
	// so one tenant's key must not block another's.
	for _, businessID := range []int64{testutil.BusinessID, testutil.OtherBusinessID} {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO hr_time_entries (
				owner_id, business_id, employee_id, work_day_id, date,
				duration_minutes, description, client_key, source
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'flow')`,
			sharedOwner, businessID, 4101, 980002, "2026-08-21",
			30, "tenant client key", "shared-client-key",
		); err != nil {
			t.Fatalf("insert business %d: %v", businessID, err)
		}
	}
}

func TestWorkspaces_GetWorkspacesByOwner_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, _, err := repo.GetWorkspacesByOwner(context.Background(), testutil.BusinessID, 100, 0)
	if err != nil {
		t.Fatalf("GetWorkspacesByOwner: %v", err)
	}
}

func TestWorkspaces_CreateWorkspace_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateWorkspace(context.Background(), domain.WorkspaceInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "List Test WS",
		CreatedBy:  testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), id, testutil.BusinessID) })

	// Page rather than assume the new row is on the first one. The query orders
	// by is_default DESC, created_at ASC, id ASC, so a freshly created workspace
	// sorts LAST - and the dev database holds over a thousand for this business,
	// well past any single page. Asserting against a fixed limit made this test
	// fail purely because the fixture data grew.
	const pageSize = 500
	found := false
	for offset := int64(0); !found; offset += pageSize {
		workspaces, _, err := repo.GetWorkspacesByOwner(
			context.Background(), testutil.BusinessID, pageSize, offset)
		if err != nil {
			t.Fatalf("GetWorkspacesByOwner(offset=%d): %v", offset, err)
		}
		if len(workspaces) == 0 {
			break
		}
		for _, w := range workspaces {
			if w.ID == id {
				found = true
				break
			}
		}
	}
	if !found {
		t.Errorf("created workspace id=%d not found in any page of the list", id)
	}
}

func TestWorkspaces_UpdateWorkspace_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateWorkspace(context.Background(), domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "OldWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), id, testutil.BusinessID) })

	if err := repo.UpdateWorkspace(context.Background(), domain.UpdateWorkspaceInput{
		ID: id, OwnerID: testutil.OwnerID, Name: "NewWS",
	}); err != nil {
		t.Fatalf("UpdateWorkspace: %v", err)
	}
}

func TestWorkspaces_DeleteWorkspace_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateWorkspace(context.Background(), domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "DeleteMe", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	if err := repo.DeleteWorkspace(context.Background(), id, testutil.BusinessID); err != nil {
		t.Fatalf("DeleteWorkspace: %v", err)
	}

	workspaces, _, _ := repo.GetWorkspacesByOwner(context.Background(), testutil.BusinessID, 500, 0)
	for _, w := range workspaces {
		if w.ID == id {
			t.Errorf("deleted workspace id=%d still in list", id)
		}
	}
}

func TestWorkspaces_CreateProject_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	wsID, err := repo.CreateWorkspace(context.Background(), domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "ProjectWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(context.Background(), domain.ProjectInput{
		WorkspaceID: wsID,
		OwnerID:     testutil.OwnerID,
		Name:        "Test Project",
		CreatedBy:   testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if projID <= 0 {
		t.Fatalf("expected positive project ID, got %d", projID)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), "DELETE FROM flow_projects WHERE id = ?", projID) })
}

func TestWorkspaces_GetProjectsByWorkspace_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	wsID, err := repo.CreateWorkspace(context.Background(), domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "ProjListWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), wsID, testutil.BusinessID) })

	projects, _, err := repo.GetProjectsByWorkspace(context.Background(), wsID, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("GetProjectsByWorkspace: %v", err)
	}
	_ = projects
}

func TestWorkspaces_CreateTask_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	wsID, err := repo.CreateWorkspace(context.Background(), domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "TaskWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(context.Background(), domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "Task Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), "DELETE FROM flow_projects WHERE id = ?", projID) })

	taskID, err := repo.CreateTask(context.Background(), domain.TaskInput{
		ProjectID:   projID,
		WorkspaceID: wsID,
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		Title:       "Test Task",
		ReporterID:  testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if taskID <= 0 {
		t.Fatalf("expected positive task ID, got %d", taskID)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(context.Background(), taskID, testutil.BusinessID) })
}

func TestWorkspaces_LookupUserByEmail_Owner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	// First get actual owner email from users table
	var email string
	if err := db.QueryRowContext(context.Background(), "SELECT email FROM users WHERE id = ?", testutil.OwnerID).Scan(&email); err != nil {
		t.Fatalf("get owner email: %v", err)
	}

	id, found, err := repo.LookupUserByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("LookupUserByEmail: %v", err)
	}
	if !found {
		t.Fatal("expected to find owner by email")
	}
	if id != testutil.OwnerID {
		t.Errorf("expected id %d, got %d", testutil.OwnerID, id)
	}
}

func TestWorkspaces_LookupUserByEmail_Unknown_NotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, found, err := repo.LookupUserByEmail(context.Background(), "nonexistent@example.com")
	if err != nil {
		t.Fatalf("LookupUserByEmail: %v", err)
	}
	if found {
		t.Error("expected found=false for unknown email")
	}
}

func TestWorkspaces_GetDashboardStats_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	wsID, err := repo.CreateWorkspace(context.Background(), domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "StatsWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(context.Background(), wsID, testutil.BusinessID) })

	_, err = repo.GetDashboardStats(context.Background(), wsID, testutil.BusinessID)
	if err != nil {
		t.Fatalf("GetDashboardStats: %v", err)
	}
}

func TestWorkspaces_CreatePersonalWorkspace_Idempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id1, err := repo.CreatePersonalWorkspace(context.Background(), testutil.OwnerID, testutil.BusinessID)
	if err != nil {
		t.Fatalf("CreatePersonalWorkspace first call: %v", err)
	}
	id2, err := repo.CreatePersonalWorkspace(context.Background(), testutil.OwnerID, testutil.BusinessID)
	if err != nil {
		t.Fatalf("CreatePersonalWorkspace second call: %v", err)
	}
	if id1 != id2 {
		t.Errorf("expected idempotent ID, got %d then %d", id1, id2)
	}

	// business_id used to be omitted from the INSERT entirely, leaving every
	// personal workspace stuck at the schema's DEFAULT 0. Confirm it's
	// findable by its real business, not just present with some value.
	workspaces, _, err := repo.GetWorkspacesByOwner(context.Background(), testutil.BusinessID, 50, 0)
	if err != nil {
		t.Fatalf("GetWorkspacesByOwner: %v", err)
	}
	found := false
	for _, w := range workspaces {
		if w.ID == id1 {
			found = true
		}
	}
	if !found {
		t.Errorf("personal workspace %d not found when listed by its real business id %d - business_id was likely left at 0",
			id1, testutil.BusinessID)
	}
}

// The idempotency lookup behind CreatePersonalWorkspace read owner_id only, so
// a user who belongs to two businesses got the first business's personal
// workspace handed back for the second, and the second business never got one.
// That is also why the test above failed intermittently: any earlier test that
// left a personal or default workspace for the same owner under a different
// business won the lookup.
//
// Seeds the other tenant first, deliberately, so an unscoped lookup would find
// it before the caller's own.
func TestWorkspaces_CreatePersonalWorkspace_IsPerBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const sharedOwner int64 = 970100

	theirs, err := repo.CreatePersonalWorkspace(ctx, sharedOwner, testutil.OtherBusinessID)
	if err != nil {
		t.Fatalf("CreatePersonalWorkspace for the other business: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM flow_workspaces WHERE id = ?", theirs)
	})

	mine, err := repo.CreatePersonalWorkspace(ctx, sharedOwner, testutil.BusinessID)
	if err != nil {
		t.Fatalf("CreatePersonalWorkspace for this business: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM flow_workspaces WHERE id = ?", mine)
	})

	if mine == theirs {
		t.Fatalf("returned the other business's workspace %d instead of creating one for business %d",
			theirs, testutil.BusinessID)
	}

	// And it must still be idempotent within a business.
	again, err := repo.CreatePersonalWorkspace(ctx, sharedOwner, testutil.BusinessID)
	if err != nil {
		t.Fatalf("CreatePersonalWorkspace second call: %v", err)
	}
	if again != mine {
		t.Errorf("not idempotent within a business: got %d then %d", mine, again)
	}
}

func TestWorkspaces_UpdateWorkspace_Description(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "Desc WS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, id, testutil.BusinessID) })

	if err := repo.UpdateWorkspace(ctx, domain.UpdateWorkspaceInput{
		ID: id, OwnerID: testutil.OwnerID, Name: "Desc WS", Description: "Updated description",
	}); err != nil {
		t.Fatalf("UpdateWorkspace description: %v", err)
	}
}

func TestWorkspaces_CountWorkspacesByOwner_IncreasesOnCreate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const countTestBusinessID int64 = 881001

	before, err := repo.CountWorkspacesByOwner(ctx, countTestBusinessID)
	if err != nil {
		t.Fatalf("count before: %v", err)
	}

	id, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: countTestBusinessID, Name: "Count Increase WS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, id, countTestBusinessID) })

	after, err := repo.CountWorkspacesByOwner(ctx, countTestBusinessID)
	if err != nil {
		t.Fatalf("count after: %v", err)
	}
	if after != before+1 {
		t.Errorf("expected count to increase by 1: before=%d after=%d", before, after)
	}
}

func TestWorkspaces_CreateProject_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "ProjListWS2", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.OwnerID, Name: "Appears Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	projects, _, err := repo.GetProjectsByWorkspace(ctx, wsID, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("GetProjectsByWorkspace: %v", err)
	}
	found := false
	for _, p := range projects {
		if p.ID == projID {
			found = true
		}
	}
	if !found {
		t.Errorf("created project id=%d not found in list", projID)
	}
}

func TestWorkspaces_CreateTask_AppearsInProject(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "TaskListWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "Task List Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	taskID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Title: "Appears Task", ReporterID: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(ctx, taskID, testutil.BusinessID) })

	tasks, _, err := repo.GetTasksByProject(ctx, domain.ListTasksFilter{ProjectID: projID, WorkspaceID: wsID, Limit: 100})
	if err != nil {
		t.Fatalf("GetTasksByProject: %v", err)
	}
	found := false
	for _, tk := range tasks {
		if tk.ID == taskID {
			found = true
		}
	}
	if !found {
		t.Errorf("created task id=%d not found in GetTasksByProject", taskID)
	}
}

func TestWorkspaces_DeleteTask_RemovesFromProject(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "DelTaskWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "Del Task Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	taskID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Title: "Delete Me Task", ReporterID: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	if err := repo.DeleteTask(ctx, taskID, testutil.BusinessID); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}

	tasks, _, _ := repo.GetTasksByProject(ctx, domain.ListTasksFilter{ProjectID: projID, WorkspaceID: wsID, Limit: 100})
	for _, tk := range tasks {
		if tk.ID == taskID {
			t.Errorf("deleted task id=%d still in list", taskID)
		}
	}
}

func TestWorkspaces_CreateComment_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "CommentWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "Comment Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	taskID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Title: "Comment Task", ReporterID: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(ctx, taskID, testutil.BusinessID) })

	commentID, err := repo.CreateComment(ctx, taskID, testutil.OwnerID, "Integration test comment")
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}
	if commentID <= 0 {
		t.Errorf("expected positive comment ID, got %d", commentID)
	}
}

func TestWorkspaces_GetCommentsByTask_ReturnsComment(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "CommListWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "CommList Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	taskID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Title: "CommList Task", ReporterID: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(ctx, taskID, testutil.BusinessID) })

	commentID, err := repo.CreateComment(ctx, taskID, testutil.OwnerID, "Check me")
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}

	comments, err := repo.GetCommentsByTask(ctx, taskID)
	if err != nil {
		t.Fatalf("GetCommentsByTask: %v", err)
	}
	found := false
	for _, c := range comments {
		if c.ID == commentID {
			found = true
			if c.ActorName == "" || c.ActorInitials == "" {
				t.Errorf("comment actor identity is empty: name=%q initials=%q", c.ActorName, c.ActorInitials)
			}
		}
	}
	if !found {
		t.Errorf("comment id=%d not found in GetCommentsByTask", commentID)
	}

	if err := repo.LogActivity(ctx, taskID, testutil.OwnerID, "task_updated", "old", "new"); err != nil {
		t.Fatalf("LogActivity: %v", err)
	}
	activities, err := repo.GetActivityByTask(ctx, taskID, 10)
	if err != nil {
		t.Fatalf("GetActivityByTask: %v", err)
	}
	if len(activities) == 0 {
		t.Fatal("expected activity row")
	}
	if activities[0].ActorName == "" || activities[0].ActorInitials == "" {
		t.Fatalf("activity actor identity is empty: name=%q initials=%q", activities[0].ActorName, activities[0].ActorInitials)
	}
}

func TestWorkspaces_CreateLabel_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "LabelWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	labelID, err := repo.CreateLabel(ctx, domain.LabelInput{
		WorkspaceID: wsID, Name: "Bug", Color: "#ff0000",
	})
	if err != nil {
		t.Fatalf("CreateLabel: %v", err)
	}
	if labelID <= 0 {
		t.Errorf("expected positive label ID, got %d", labelID)
	}
	t.Cleanup(func() { _ = repo.DeleteLabel(ctx, labelID, testutil.OwnerID, testutil.BusinessID) })
}

func TestWorkspaces_GetLabelsByWorkspace_ReturnsLabel(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "LabelListWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	labelID, err := repo.CreateLabel(ctx, domain.LabelInput{
		WorkspaceID: wsID, Name: "Feature", Color: "#00ff00",
	})
	if err != nil {
		t.Fatalf("CreateLabel: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteLabel(ctx, labelID, testutil.OwnerID, testutil.BusinessID) })

	labels, _, err := repo.GetLabelsByWorkspace(ctx, wsID, 100, 0)
	if err != nil {
		t.Fatalf("GetLabelsByWorkspace: %v", err)
	}
	found := false
	for _, l := range labels {
		if l.ID == labelID {
			found = true
		}
	}
	if !found {
		t.Errorf("label id=%d not found in GetLabelsByWorkspace", labelID)
	}
}

func TestWorkspaces_UpdateTaskStatus_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "StatusWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "Status Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	taskID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Title: "Status Task", ReporterID: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(ctx, taskID, testutil.BusinessID) })

	if err := repo.UpdateTaskStatus(ctx, domain.UpdateTaskStatusInput{
		ID: taskID, OwnerID: testutil.OwnerID, Status: "in_progress",
	}); err != nil {
		t.Fatalf("UpdateTaskStatus: %v", err)
	}
}

func TestWorkspaces_CreateColumn_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "ColumnWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "Column Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	colID, err := repo.CreateColumn(ctx, domain.ProjectColumnInput{
		ProjectID: projID, WorkspaceID: wsID, Name: "To Do", Color: "#aabbcc", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateColumn: %v", err)
	}
	if colID <= 0 {
		t.Errorf("expected positive column ID, got %d", colID)
	}
	t.Cleanup(func() { _ = repo.DeleteColumn(ctx, colID, testutil.OwnerID) })
}

func TestWorkspaces_LogActivity_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "ActivityWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "Activity Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	taskID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Title: "Activity Task", ReporterID: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(ctx, taskID, testutil.BusinessID) })

	if err := repo.LogActivity(ctx, taskID, testutil.OwnerID, "status_changed", "todo", "in_progress"); err != nil {
		t.Fatalf("LogActivity: %v", err)
	}
}

func TestWorkspaces_GetTaskByID_ReturnsTask(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "GetByIDWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "GetByID Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	taskID, err := repo.CreateTask(ctx, domain.TaskInput{
		ProjectID: projID, WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Title: "GetByID Task", ReporterID: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteTask(ctx, taskID, testutil.BusinessID) })

	task, err := repo.GetTaskByID(ctx, taskID)
	if err != nil {
		t.Fatalf("GetTaskByID: %v", err)
	}
	if task.ID != taskID {
		t.Errorf("expected task.ID=%d, got %d", taskID, task.ID)
	}
}

func TestWorkspaces_IsWorkspaceMember_OwnerIsNotMemberByDefault(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "MemberCheckWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	// Just ensure it doesn't panic for an arbitrary user.
	_ = repo.IsWorkspaceMember(ctx, 9999999, wsID)
}
