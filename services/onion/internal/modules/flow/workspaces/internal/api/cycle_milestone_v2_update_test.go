package api_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/api"
	"github.com/voxire/onion/internal/testutil"
	flowcyclev2 "github.com/voxire/voxie/gen/go/flow/cycle/v2"
	flowmilestonev2 "github.com/voxire/voxie/gen/go/flow/milestone/v2"
)

// flow.cycle.v2 UpdateCycle / flow.milestone.v2 UpdateMilestone used to build
// their input from ID/Name/Status alone. ActorID and BusinessID stayed 0, the
// repository's membership predicate matched no row, and every update answered
// 500. The handlers also never read end_date / description / target_date, so
// those contract fields could not be changed even once the failure was fixed.

func TestCycleV2_UpdateAppliesAndPersistsEndDate(t *testing.T) {
	f := newRoleFixture(t)
	cycleV2 := api.NewCycleAPIV2(f.svc())
	db := testutil.OpenDB(t)
	ctx := context.Background()

	created, err := cycleV2.CreateCycle(ctxAs(roleMemberID), &flowcyclev2.CreateCycleRequest{
		ProjectId: testutil.Int64Ptr(f.projID), Name: "Sprint 1",
		StartDate: "2026-01-05", EndDate: "2026-01-19",
	})
	if err != nil {
		t.Fatalf("CreateCycle: %v", err)
	}
	cycleID := created.Cycle.Id
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_cycles WHERE id = ?", cycleID) })

	if _, err := cycleV2.UpdateCycle(ctxAs(roleMemberID), &flowcyclev2.UpdateCycleRequest{
		Id: cycleID, Name: "Sprint 1 renamed", Status: "completed", EndDate: "2026-02-02",
	}); err != nil {
		t.Fatalf("member UpdateCycle: expected success, got %v", err)
	}

	var name, status string
	var endDate sql.NullTime
	if err := db.QueryRowContext(ctx,
		"SELECT name, status, end_date FROM flow_cycles WHERE id = ?", cycleID,
	).Scan(&name, &status, &endDate); err != nil {
		t.Fatalf("read back cycle: %v", err)
	}
	if name != "Sprint 1 renamed" || status != "completed" {
		t.Errorf("cycle after update = (%q, %q), want (Sprint 1 renamed, completed)", name, status)
	}
	if !endDate.Valid || !endDate.Time.Equal(time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("cycle end_date = %v, want 2026-02-02", endDate)
	}
}

// The project owner holds no flow_project_members row, as CreateProject never
// writes one; the predicate must still admit them.
func TestCycleV2_ProjectOwnerCanUpdate(t *testing.T) {
	f := newRoleFixture(t)
	cycleV2 := api.NewCycleAPIV2(f.svc())
	db := testutil.OpenDB(t)
	ctx := context.Background()

	created, err := cycleV2.CreateCycle(ctxAs(roleMemberID), &flowcyclev2.CreateCycleRequest{
		ProjectId: testutil.Int64Ptr(f.projID), Name: "Owner update", StartDate: "2026-03-02", EndDate: "2026-03-16",
	})
	if err != nil {
		t.Fatalf("CreateCycle: %v", err)
	}
	cycleID := created.Cycle.Id
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_cycles WHERE id = ?", cycleID) })

	if _, err := cycleV2.UpdateCycle(ctxAs(roleOwnerID), &flowcyclev2.UpdateCycleRequest{
		Id: cycleID, Name: "Owner renamed",
	}); err != nil {
		t.Fatalf("project owner UpdateCycle: expected success, got %v", err)
	}
}

// A caller with no project or workspace membership gets a 403, not the 500 the
// handler used to answer for every repository error.
func TestCycleV2_UpdateDeniedForNonMember(t *testing.T) {
	f := newRoleFixture(t)
	cycleV2 := api.NewCycleAPIV2(f.svc())
	db := testutil.OpenDB(t)
	ctx := context.Background()

	created, err := cycleV2.CreateCycle(ctxAs(roleMemberID), &flowcyclev2.CreateCycleRequest{
		ProjectId: testutil.Int64Ptr(f.projID), Name: "Stranger update", StartDate: "2026-04-06", EndDate: "2026-04-20",
	})
	if err != nil {
		t.Fatalf("CreateCycle: %v", err)
	}
	cycleID := created.Cycle.Id
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_cycles WHERE id = ?", cycleID) })

	_, err = cycleV2.UpdateCycle(ctxAs(roleStrangerID), &flowcyclev2.UpdateCycleRequest{
		Id: cycleID, Name: "Stranger renamed",
	})
	assertPermissionDenied(t, err, "stranger UpdateCycle via v2")
}

func TestCycleV2_UpdateRejectsUnparseableEndDate(t *testing.T) {
	f := newRoleFixture(t)
	cycleV2 := api.NewCycleAPIV2(f.svc())

	_, err := cycleV2.UpdateCycle(ctxAs(roleMemberID), &flowcyclev2.UpdateCycleRequest{
		Id: 1, Name: "x", EndDate: "not-a-date",
	})
	if err == nil || !strings.Contains(err.Error(), "Invalid end_date") {
		t.Fatalf("UpdateCycle with bad end_date: expected Invalid end_date, got %v", err)
	}
}

func TestMilestoneV2_UpdateAppliesAndPersistsDescriptionAndTargetDate(t *testing.T) {
	f := newRoleFixture(t)
	milestoneV2 := api.NewMilestoneAPIV2(f.svc())
	db := testutil.OpenDB(t)
	ctx := context.Background()

	created, err := milestoneV2.CreateMilestone(ctxAs(roleMemberID), &flowmilestonev2.CreateMilestoneRequest{
		ProjectId: testutil.Int64Ptr(f.projID), Name: "M1", Description: "first", TargetDate: "2026-05-01",
	})
	if err != nil {
		t.Fatalf("CreateMilestone: %v", err)
	}
	milestoneID := created.Milestone.Id
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_milestones WHERE id = ?", milestoneID) })

	if _, err := milestoneV2.UpdateMilestone(ctxAs(roleMemberID), &flowmilestonev2.UpdateMilestoneRequest{
		Id: milestoneID, Name: "M1 renamed", Description: "second", Status: "completed", TargetDate: "2026-06-15",
	}); err != nil {
		t.Fatalf("member UpdateMilestone: expected success, got %v", err)
	}

	var name, description, status string
	var targetDate sql.NullTime
	if err := db.QueryRowContext(ctx,
		"SELECT name, description, status, target_date FROM flow_milestones WHERE id = ?", milestoneID,
	).Scan(&name, &description, &status, &targetDate); err != nil {
		t.Fatalf("read back milestone: %v", err)
	}
	if name != "M1 renamed" || description != "second" || status != "completed" {
		t.Errorf("milestone after update = (%q, %q, %q), want (M1 renamed, second, completed)", name, description, status)
	}
	if !targetDate.Valid || !targetDate.Time.Equal(time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("milestone target_date = %v, want 2026-06-15", targetDate)
	}
}

func TestMilestoneV2_UpdateDeniedForNonMember(t *testing.T) {
	f := newRoleFixture(t)
	milestoneV2 := api.NewMilestoneAPIV2(f.svc())
	db := testutil.OpenDB(t)
	ctx := context.Background()

	created, err := milestoneV2.CreateMilestone(ctxAs(roleMemberID), &flowmilestonev2.CreateMilestoneRequest{
		ProjectId: testutil.Int64Ptr(f.projID), Name: "Stranger target",
	})
	if err != nil {
		t.Fatalf("CreateMilestone: %v", err)
	}
	milestoneID := created.Milestone.Id
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_milestones WHERE id = ?", milestoneID) })

	_, err = milestoneV2.UpdateMilestone(ctxAs(roleStrangerID), &flowmilestonev2.UpdateMilestoneRequest{
		Id: milestoneID, Name: "Stranger renamed",
	})
	assertPermissionDenied(t, err, "stranger UpdateMilestone via v2")
}

// The list used to attach an empty MilestoneProject to every row, so the
// Milestones page rendered a blank project line.
func TestMilestoneV2_ListCarriesProjectNameAndColor(t *testing.T) {
	f := newRoleFixture(t)
	milestoneV2 := api.NewMilestoneAPIV2(f.svc())
	db := testutil.OpenDB(t)
	ctx := context.Background()

	created, err := milestoneV2.CreateMilestone(ctxAs(roleMemberID), &flowmilestonev2.CreateMilestoneRequest{
		ProjectId: testutil.Int64Ptr(f.projID), Name: "With project",
	})
	if err != nil {
		t.Fatalf("CreateMilestone: %v", err)
	}
	milestoneID := created.Milestone.Id
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_milestones WHERE id = ?", milestoneID) })

	list, err := milestoneV2.ListMilestones(ctxAs(roleMemberID), &flowmilestonev2.ListMilestonesRequest{ProjectId: testutil.Int64Ptr(f.projID)})
	if err != nil {
		t.Fatalf("ListMilestones: %v", err)
	}

	var found *flowmilestonev2.Milestone
	for _, m := range list.Milestones {
		if m.Id == milestoneID {
			found = m
			break
		}
	}
	if found == nil {
		t.Fatal("created milestone missing from ListMilestones")
	}
	if found.Project == nil || found.Project.Name != "Role Gate Project" {
		t.Errorf("milestone project = %+v, want name Role Gate Project", found.Project)
	}
}
