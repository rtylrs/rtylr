package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// Distinct from other integration tests' fixture ids so these rows cannot
// collide with data another test creates.
const (
	teamWSOwnerID    int64 = 920001 // owns the workspace; RequireWorkspaceRole admits owners with no membership row
	teamWSMemberID   int64 = 920002 // flow_workspace_members.role = member; has an hr_employees row and a work day
	teamWSStrangerID int64 = 920003 // no membership on the workspace at all
	teamOwnerID      int64 = 920100 // flow_work_days.owner_id, distinct from the HR tenant key
	teamBusinessID   int64 = 920200 // HR owner_id and business_id, and flow_work_days.business_id
)

// newTeamFixture builds a workspace, one ordinary member with an HR record and
// a recorded day, and returns the Service plus enough ids for a caller at
// every role level.
func newTeamFixture(t *testing.T) (svc *application.Service, repo *infrastructure.Repository, workspaceID int64) {
	t.Helper()

	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo = testsupport.NewRepository(db)
	svc = application.NewService(ctx, testsupport.NewRepository(db))

	workspaceID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: teamWSOwnerID, BusinessID: teamBusinessID,
		Name: "Team View WS", CreatedBy: teamWSOwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, workspaceID, teamWSOwnerID) })

	if err := repo.AddWorkspaceMember(ctx, domain.AddWorkspaceMemberInput{
		WorkspaceID: workspaceID, UserID: teamWSMemberID, Role: domain.FlowRoleMember, InvitedBy: teamWSOwnerID,
	}); err != nil {
		t.Fatalf("AddWorkspaceMember: %v", err)
	}

	if _, err := db.ExecContext(ctx,
		`INSERT INTO hr_employees (owner_id, business_id, user_id, first_name, last_name, email)
		 VALUES (?, ?, ?, 'Team', 'Member', 'team-member@example.test')`,
		teamBusinessID, teamBusinessID, teamWSMemberID,
	); err != nil {
		t.Fatalf("insert hr_employees: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(),
			`DELETE FROM hr_employees WHERE owner_id = ? AND business_id = ? AND user_id = ?`,
			teamBusinessID, teamBusinessID, teamWSMemberID)
	})

	var employeeID int64
	if err := db.QueryRowContext(ctx,
		`SELECT id FROM hr_employees WHERE owner_id = ? AND business_id = ? AND user_id = ?`,
		teamBusinessID, teamBusinessID, teamWSMemberID,
	).Scan(&employeeID); err != nil {
		t.Fatalf("read back hr_employees id: %v", err)
	}
	svc.SetWorkdayRosterSource(testWorkdayRosterSource{
		businessMembers: []int64{teamWSOwnerID, teamWSMemberID},
		employeeUserIDs: map[int64]int64{employeeID: teamWSMemberID},
	})

	if _, err := repo.UpsertWorkDay(ctx, domain.WorkDayInput{
		OwnerID: teamOwnerID, BusinessID: teamBusinessID, EmployeeID: employeeID,
		Date: "2026-08-16", Timezone: "UTC", WorkMode: domain.WorkModeOnsite, LoggedBy: teamWSMemberID,
	}); err != nil {
		t.Fatalf("UpsertWorkDay: %v", err)
	}

	return svc, repo, workspaceID
}

type testWorkdayRosterSource struct {
	businessMembers []int64
	employeeUserIDs map[int64]int64
}

func (s testWorkdayRosterSource) ResolveWorkdayRoster(_ context.Context, _ int64, _ []int64, includeAll bool) (application.WorkdayRoster, error) {
	active := []int64{}
	if includeAll {
		active = append(active, s.businessMembers...)
	}
	return application.WorkdayRoster{
		BusinessMemberUserIDs: append([]int64(nil), s.businessMembers...),
		ActiveMemberUserIDs:   active,
		EmployeeUserIDs:       s.employeeUserIDs,
	}, nil
}

func TestListWorkDaysForWorkspace_OwnerSeesMemberDay(t *testing.T) {
	svc, _, workspaceID := newTeamFixture(t)

	days, total, err := svc.ListWorkDaysForWorkspace(
		context.Background(), teamWSOwnerID, workspaceID, teamOwnerID, teamBusinessID,
		"2026-08-01", "2026-08-31", 50, 0,
	)
	if err != nil {
		t.Fatalf("ListWorkDaysForWorkspace as owner: %v", err)
	}
	if total != 1 || len(days) != 1 {
		t.Fatalf("want exactly the one seeded day, got total=%d len=%d", total, len(days))
	}
	if days[0].Date.Format("2006-01-02") != "2026-08-16" {
		t.Fatalf("want the seeded date, got %s", days[0].Date)
	}
	if days[0].UserID != teamWSMemberID {
		t.Fatalf("want UserID resolved back to the member's auth id (%d), got %d", teamWSMemberID, days[0].UserID)
	}
}

func TestListWorkDaysForWorkspace_StrangerDenied(t *testing.T) {
	svc, _, workspaceID := newTeamFixture(t)

	_, _, err := svc.ListWorkDaysForWorkspace(
		context.Background(), teamWSStrangerID, workspaceID, teamOwnerID, teamBusinessID,
		"2026-08-01", "2026-08-31", 50, 0,
	)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("want ErrUnauthorized for a caller with no role on the workspace, got %v", err)
	}
}

func TestListWorkDaysForWorkspace_OrdinaryMemberDenied(t *testing.T) {
	svc, _, workspaceID := newTeamFixture(t)

	// The member can see their own day elsewhere, but the team view itself
	// needs admin: an ordinary member should not see the whole roster.
	_, _, err := svc.ListWorkDaysForWorkspace(
		context.Background(), teamWSMemberID, workspaceID, teamOwnerID, teamBusinessID,
		"2026-08-01", "2026-08-31", 50, 0,
	)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("want ErrUnauthorized for a member-role caller, got %v", err)
	}
}

func TestWorkDaysForDate_ResolvesByEmployeeID(t *testing.T) {
	svc, _, _ := newTeamFixture(t)

	db := testutil.OpenDB(t)
	var employeeID int64
	if err := db.QueryRowContext(context.Background(),
		`SELECT id FROM hr_employees WHERE owner_id = ? AND business_id = ? AND user_id = ?`,
		teamBusinessID, teamBusinessID, teamWSMemberID,
	).Scan(&employeeID); err != nil {
		t.Fatalf("read back hr_employees id: %v", err)
	}

	byEmployeeID, err := svc.WorkDaysForDate(
		context.Background(), teamOwnerID, teamBusinessID, []int64{employeeID, 999999999}, "2026-08-16",
	)
	if err != nil {
		t.Fatalf("WorkDaysForDate: %v", err)
	}
	day, ok := byEmployeeID[employeeID]
	if !ok {
		t.Fatalf("want a work day for employee %d, got none: %+v", employeeID, byEmployeeID)
	}
	if day.WorkMode != domain.WorkModeOnsite {
		t.Fatalf("WorkMode = %q, want %q", day.WorkMode, domain.WorkModeOnsite)
	}
	if _, ok := byEmployeeID[999999999]; ok {
		t.Fatal("an employee id with no row must be absent from the map, not a zero-value entry")
	}
}

func TestWorkDaysForDate_EmptyEmployeeIDsReturnsEmptyMap(t *testing.T) {
	svc, _, _ := newTeamFixture(t)

	byEmployeeID, err := svc.WorkDaysForDate(context.Background(), teamOwnerID, teamBusinessID, nil, "2026-08-16")
	if err != nil {
		t.Fatalf("WorkDaysForDate: %v", err)
	}
	if len(byEmployeeID) != 0 {
		t.Fatalf("want an empty map for no employee ids, got %+v", byEmployeeID)
	}
}
