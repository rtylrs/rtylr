package application

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// Abed, 2026-11-18: a department manager who opens a report's project sees its
// whole board, not an empty one. The manager (v2UserA) was never added to the
// project; they manage its creator (v2UserB).
func TestDepartmentManagerWorksInAReportsProject(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	fx := newAdminProjectFixture(ctx, t, svc)
	svc.hubProfiles = adminInBusiness{businessID: 0}
	svc.orgchart = &stubOrgchart{reports: []int64{v2UserB}}

	if _, err := svc.GetProjectV2(ctx, fx.project.ID, v2UserA, v2Business); err != nil {
		t.Fatalf("GetProjectV2 as manager: %v", err)
	}
	if got := listProjectTasks(ctx, t, svc, v2UserA, v2Business, fx.project.ID); got != 1 {
		t.Errorf("manager sees %d tasks on the board, want 1", got)
	}
	if _, total, err := svc.ListCyclesV2(ctx, fx.project.ID, v2Business, v2UserA, 50, 0); err != nil || total != 1 {
		t.Errorf("manager cycles = %d, %v; want 1", total, err)
	}
	pid := fx.project.ID
	if rows, _, err := svc.ListMilestonesV2(ctx, v2Business, &pid, v2UserA, 50, 0); err != nil || len(rows) != 1 {
		t.Errorf("manager milestones = %d, %v; want 1", len(rows), err)
	}
	if _, err := svc.ListProjectStatuses(ctx, fx.project.ID, v2UserA, v2Business); err != nil {
		t.Errorf("manager ListProjectStatuses: %v", err)
	}
	if err := svc.RequireTaskRole(ctx, fx.task.ID, v2UserA, v2Business, domain.FlowRoleViewer); err != nil {
		t.Errorf("manager cannot open the task drawer: %v", err)
	}
	if !svc.CanReadTask(ctx, v2UserA, v2Business, fx.task.ID) {
		t.Error("manager cannot read the subresources of a task on the board")
	}
	if !svc.CanSubscribeProject(ctx, v2UserA, fx.project.ID) {
		t.Error("manager cannot subscribe to the project topic")
	}
	if !svc.CanSubscribeTask(ctx, v2UserA, fx.task.ID) {
		t.Error("manager cannot subscribe to the task topic")
	}
}

// A manager whose reports are not involved in the project gets nothing: the
// proxy follows the reporting line, not the title.
func TestManagerOfNobodyInvolvedStillCannotOpenProject(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	fx := newAdminProjectFixture(ctx, t, svc)
	svc.hubProfiles = adminInBusiness{businessID: 0}
	svc.orgchart = &stubOrgchart{reports: []int64{v2UserB + 1000}}

	if _, err := svc.GetProjectV2(ctx, fx.project.ID, v2UserA, v2Business); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("GetProjectV2: err = %v, want ErrUnauthorized", err)
	}
	if got := listProjectTasks(ctx, t, svc, v2UserA, v2Business, fx.project.ID); got != 0 {
		t.Errorf("sees %d tasks, want 0", got)
	}
	if _, total, err := svc.ListCyclesV2(ctx, fx.project.ID, v2Business, v2UserA, 50, 0); err != nil || total != 0 {
		t.Errorf("cycles = %d, %v; want 0", total, err)
	}
	pid := fx.project.ID
	if rows, _, err := svc.ListMilestonesV2(ctx, v2Business, &pid, v2UserA, 50, 0); err != nil || len(rows) != 0 {
		t.Errorf("milestones = %d, %v; want 0", len(rows), err)
	}
	if err := svc.RequireProjectRole(ctx, fx.project.ID, v2UserA, v2Business, domain.FlowRoleViewer); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("RequireProjectRole viewer: err = %v, want ErrUnauthorized", err)
	}
	if svc.CanReadTask(ctx, v2UserA, v2Business, fx.task.ID) {
		t.Error("can read a task's subresources")
	}
	if svc.CanSubscribeProject(ctx, v2UserA, fx.project.ID) || svc.CanSubscribeTask(ctx, v2UserA, fx.task.ID) {
		t.Error("can subscribe to the project's topics")
	}
}

// Ruling P8 for managers: opening a report's project changes no list. The
// sidebar already groups the report's project under P9, and the business-wide
// task list, milestone list and search stay what they were.
func TestManagerListingsStayCurated(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	fx := newAdminProjectFixture(ctx, t, svc)
	svc.hubProfiles = adminInBusiness{businessID: 0}
	svc.orgchart = &stubOrgchart{reports: []int64{v2UserB}}

	projects, _, err := svc.ListProjectsV2(ctx, v2UserA, v2Business, "", "", "", 100, 0)
	if err != nil {
		t.Fatalf("ListProjectsV2: %v", err)
	}
	listed := false
	for _, p := range projects {
		if p.ID == fx.project.ID {
			listed = p.VisibleViaReport
		}
	}
	if !listed {
		t.Error("manager's sidebar no longer lists the report's project as visible via report")
	}
	if _, total, err := svc.ListTasksV2(ctx, v2UserA, v2Business, domain.ListTasksFilter{Limit: 50}); err != nil || total != 0 {
		t.Errorf("manager business-wide task list = %d, %v; want 0", total, err)
	}
	if rows, _, err := svc.ListMilestonesV2(ctx, v2Business, nil, v2UserA, 50, 0); err != nil || len(rows) != 0 {
		t.Errorf("manager business-wide milestones = %d, %v; want 0", len(rows), err)
	}
	if _, total, err := svc.SearchTasksV2(ctx, v2UserB, v2Business, "zebrafinch", 50, 0); err != nil || total != 1 {
		t.Fatalf("creator search = %d, %v; want 1", total, err)
	}
	if _, total, err := svc.SearchTasksV2(ctx, v2UserA, v2Business, "zebrafinch", 50, 0); err != nil || total != 0 {
		t.Errorf("manager search = %d, %v; want 0", total, err)
	}
}
