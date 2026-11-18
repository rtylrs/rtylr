package application

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/hubprofile"
)

// flowAdminInBusiness makes the caller a Flow Admin (flow:team:manage, not a
// business owner or admin) in one business and a plain member everywhere else.
type flowAdminInBusiness struct{ businessID int64 }

func (r flowAdminInBusiness) GetProfile(_ context.Context, _, businessID int64) (hubprofile.Profile, error) {
	access := authorization.EffectiveAccess{}
	if businessID == r.businessID {
		access.Apps = []authorization.EffectiveAppAccess{{
			AppID: "flow", Licensed: true, HasAccess: true, Permissions: []string{"flow:team:manage"},
		}}
	}
	return hubprofile.Profile{EffectiveAccess: &access}, nil
}

// Abed, 2026-11-18: a Flow Admin opens a private project in Work as Insights
// already let them open its dashboard. v2UserA is not on the project.
func TestFlowAdminOpensPrivateProjectAndReadsItsWork(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	fx := newAdminProjectFixture(ctx, t, svc)
	svc.hubProfiles = flowAdminInBusiness{businessID: v2Business}

	if _, err := svc.GetProjectV2(ctx, fx.project.ID, v2UserA, v2Business); err != nil {
		t.Fatalf("GetProjectV2 as Flow Admin: %v", err)
	}
	if !svc.IsProjectMemberOrBypass(ctx, v2UserA, v2Business, fx.project.ID) {
		t.Error("Flow Admin cannot read the project roster")
	}
	if got := listProjectTasks(ctx, t, svc, v2UserA, v2Business, fx.project.ID); got != 1 {
		t.Errorf("Flow Admin sees %d tasks on the board, want 1", got)
	}
	if _, total, err := svc.ListCyclesV2(ctx, fx.project.ID, v2Business, v2UserA, 50, 0); err != nil || total != 1 {
		t.Errorf("Flow Admin cycles = %d, %v; want 1", total, err)
	}
	pid := fx.project.ID
	if rows, _, err := svc.ListMilestonesV2(ctx, v2Business, &pid, v2UserA, 50, 0); err != nil || len(rows) != 1 {
		t.Errorf("Flow Admin milestones = %d, %v; want 1", len(rows), err)
	}
	if statuses, err := svc.ListProjectStatuses(ctx, fx.project.ID, v2UserA, v2Business); err != nil || len(statuses) == 0 {
		t.Errorf("Flow Admin statuses = %d, %v; want the project's set", len(statuses), err)
	}
	if err := svc.RequireTaskRole(ctx, fx.task.ID, v2UserA, v2Business, domain.FlowRoleViewer); err != nil {
		t.Errorf("Flow Admin cannot open the task drawer: %v", err)
	}
	if !svc.CanReadTask(ctx, v2UserA, v2Business, fx.task.ID) {
		t.Error("Flow Admin cannot read the subresources of a task on the board")
	}
	if !svc.CanSubscribeProject(ctx, v2UserA, fx.project.ID) || !svc.CanSubscribeTask(ctx, v2UserA, fx.task.ID) {
		t.Error("Flow Admin cannot subscribe to the project's topics")
	}
}

// Opening is reading. Every member or admin gate still refuses a Flow Admin
// who is not on the project, and so does the comment gate.
func TestFlowAdminGainsNoWriteOnAProjectTheyOpen(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	fx := newAdminProjectFixture(ctx, t, svc)
	svc.hubProfiles = flowAdminInBusiness{businessID: v2Business}

	for _, role := range []string{domain.FlowRoleMember, domain.FlowRoleAdmin} {
		if err := svc.RequireProjectRole(ctx, fx.project.ID, v2UserA, v2Business, role); !errors.Is(err, domain.ErrUnauthorized) {
			t.Errorf("RequireProjectRole %s: err = %v, want ErrUnauthorized", role, err)
		}
		if err := svc.RequireTaskRole(ctx, fx.task.ID, v2UserA, v2Business, role); !errors.Is(err, domain.ErrUnauthorized) {
			t.Errorf("RequireTaskRole %s: err = %v, want ErrUnauthorized", role, err)
		}
	}
	statuses, err := svc.ListProjectStatuses(ctx, fx.project.ID, v2UserB, v2Business)
	if err != nil || len(statuses) == 0 {
		t.Fatalf("creator ListProjectStatuses = %d, %v", len(statuses), err)
	}
	if _, err := svc.SetTaskStatus(ctx, fx.task.ID, statuses[len(statuses)-1].ID, v2UserA, v2Business); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("SetTaskStatus: err = %v, want ErrUnauthorized", err)
	}
	if svc.IsTaskMember(ctx, v2UserA, fx.task.ID) {
		t.Error("Flow Admin passes the comment gate on a project they are not on")
	}
}

// Flow Admin in one business opens nothing in another, whichever business the
// request claims to act in.
func TestFlowAdminOfAnotherBusinessCannotOpenProject(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	fx := newAdminProjectFixture(ctx, t, svc)
	other := v2Business + 1
	svc.hubProfiles = flowAdminInBusiness{businessID: other}

	if _, err := svc.GetProjectV2(ctx, fx.project.ID, v2UserA, v2Business); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("acting in the project's business: err = %v, want ErrUnauthorized", err)
	}
	if _, err := svc.GetProjectV2(ctx, fx.project.ID, v2UserA, other); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("acting in their own business: err = %v, want sql.ErrNoRows", err)
	}
	for _, businessID := range []int64{v2Business, other} {
		if got := listProjectTasks(ctx, t, svc, v2UserA, businessID, fx.project.ID); got != 0 {
			t.Errorf("business %d: sees %d tasks, want 0", businessID, got)
		}
		if err := svc.RequireProjectRole(ctx, fx.project.ID, v2UserA, businessID, domain.FlowRoleViewer); !errors.Is(err, domain.ErrUnauthorized) {
			t.Errorf("business %d: RequireProjectRole viewer = %v, want ErrUnauthorized", businessID, err)
		}
		if svc.CanReadTask(ctx, v2UserA, businessID, fx.task.ID) {
			t.Errorf("business %d: can read a task's subresources", businessID)
		}
	}
	if svc.CanSubscribeProject(ctx, v2UserA, fx.project.ID) || svc.CanSubscribeTask(ctx, v2UserA, fx.task.ID) {
		t.Error("Flow Admin of another business can subscribe to the project's topics")
	}
}

// Someone's My Tasks bucket stays theirs: managing its owner opens it no more
// than holding business-wide Flow access does.
func TestManagerAndFlowAdminCannotOpenSomeoneElsesMyTasks(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	task, err := svc.CreateTask(ctx, domain.TaskInput{
		BusinessID: v2Business, OwnerID: v2UserB, Title: "private errand", ReporterID: v2UserB,
	})
	if err != nil {
		t.Fatalf("CreateTask into My Tasks: %v", err)
	}
	bucketID := task.ProjectID
	if bucketID == 0 {
		t.Fatal("task did not land in a My Tasks bucket")
	}

	for _, tc := range []struct {
		name  string
		setup func()
	}{
		{"manager of the owner", func() {
			svc.hubProfiles = adminInBusiness{businessID: 0}
			svc.orgchart = &stubOrgchart{reports: []int64{v2UserB}}
		}},
		{"flow admin", func() {
			svc.hubProfiles = flowAdminInBusiness{businessID: v2Business}
			svc.orgchart = nil
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()
			if _, err := svc.GetProjectV2(ctx, bucketID, v2UserA, v2Business); !errors.Is(err, domain.ErrUnauthorized) {
				t.Errorf("GetProjectV2: err = %v, want ErrUnauthorized", err)
			}
			if got := listProjectTasks(ctx, t, svc, v2UserA, v2Business, bucketID); got != 0 {
				t.Errorf("sees %d tasks in someone's bucket, want 0", got)
			}
			if err := svc.RequireTaskRole(ctx, task.ID, v2UserA, v2Business, domain.FlowRoleViewer); !errors.Is(err, domain.ErrUnauthorized) {
				t.Errorf("RequireTaskRole viewer: err = %v, want ErrUnauthorized", err)
			}
			if svc.CanReadTask(ctx, v2UserA, v2Business, task.ID) {
				t.Error("can read the subresources of someone's My Tasks task")
			}
			if svc.CanSubscribeProject(ctx, v2UserA, bucketID) || svc.CanSubscribeTask(ctx, v2UserA, task.ID) {
				t.Error("can subscribe to someone's My Tasks topics")
			}
		})
	}
}

// Ruling P8 for Flow Admins: the sidebar, the business-wide task list, the
// business-wide milestone list and search stay curated.
func TestFlowAdminListingsStayCurated(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	fx := newAdminProjectFixture(ctx, t, svc)
	svc.hubProfiles = flowAdminInBusiness{businessID: v2Business}

	projects, _, err := svc.ListProjectsV2(ctx, v2UserA, v2Business, "", "", "", 100, 0)
	if err != nil {
		t.Fatalf("ListProjectsV2: %v", err)
	}
	for _, p := range projects {
		if p.ID == fx.project.ID {
			t.Error("Flow Admin's sidebar lists a private project they are not on")
		}
	}
	if _, total, err := svc.ListTasksV2(ctx, v2UserA, v2Business, domain.ListTasksFilter{Limit: 50}); err != nil || total != 0 {
		t.Errorf("Flow Admin business-wide task list = %d, %v; want 0", total, err)
	}
	if rows, _, err := svc.ListMilestonesV2(ctx, v2Business, nil, v2UserA, 50, 0); err != nil || len(rows) != 0 {
		t.Errorf("Flow Admin business-wide milestones = %d, %v; want 0", len(rows), err)
	}
	if _, total, err := svc.SearchTasksV2(ctx, v2UserB, v2Business, "zebrafinch", 50, 0); err != nil || total != 1 {
		t.Fatalf("creator search = %d, %v; want 1", total, err)
	}
	if _, total, err := svc.SearchTasksV2(ctx, v2UserA, v2Business, "zebrafinch", 50, 0); err != nil || total != 0 {
		t.Errorf("Flow Admin search = %d, %v; want 0", total, err)
	}
}
