package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

func TestWorkspaces_CountWorkspacesByOwner_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
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
	repo := testsupport.NewRepository(db)
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
	repo := testsupport.NewRepository(db)
	ctx := context.Background()
	repo.SetWorkDayLineStore(newTestHRTimeStore(ctx, db))
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
	repo := testsupport.NewRepository(db)
	repo.SetWorkDayLineStore(newTestHRTimeStore(ctx, db))
	service := application.NewService(ctx, repo)
	service.SetWorkdayRosterSource(workdayRosterSourceForTest{employees: map[int64]int64{4101: 1001}})

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

type workdayRosterSourceForTest struct {
	employees map[int64]int64
}

func (s workdayRosterSourceForTest) ResolveWorkdayRoster(_ context.Context, _ int64, candidates []int64, includeAll bool) (application.WorkdayRoster, error) {
	users := make(map[int64]struct{}, len(candidates))
	for _, userID := range candidates {
		users[userID] = struct{}{}
	}
	roster := application.WorkdayRoster{EmployeeUserIDs: make(map[int64]int64)}
	for employeeID, userID := range s.employees {
		if includeAll {
			roster.EmployeeUserIDs[employeeID] = userID
			continue
		}
		if _, ok := users[userID]; ok {
			roster.EmployeeUserIDs[employeeID] = userID
		}
	}
	return roster, nil
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
	repo := testsupport.NewRepository(db)
	_, _, err := repo.GetWorkspacesByOwner(context.Background(), testutil.BusinessID, 100, 0)
	if err != nil {
		t.Fatalf("GetWorkspacesByOwner: %v", err)
	}
}

func TestWorkspaces_CreateWorkspace_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
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
			context.Background(), testutil.BusinessID, pageSize, offset,
		)
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
	repo := testsupport.NewRepository(db)
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
	repo := testsupport.NewRepository(db)
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
