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

// adminInBusiness makes the caller a business admin in one business and a
// plain member everywhere else, which is what lets these tests tell "an admin
// here" from "an admin somewhere".
type adminInBusiness struct{ businessID int64 }

func (r adminInBusiness) GetProfile(_ context.Context, _, businessID int64) (hubprofile.Profile, error) {
	access := authorization.EffectiveAccess{IsAdmin: businessID == r.businessID}
	return hubprofile.Profile{EffectiveAccess: &access}, nil
}

// adminProjectFixture is a private project v2UserB created, holding one task,
// one cycle and one milestone. v2UserA is not involved in it at all.
type adminProjectFixture struct {
	project domain.Project
	task    domain.Task
}

func newAdminProjectFixture(ctx context.Context, t *testing.T, svc *Service) adminProjectFixture {
	t.Helper()
	p := createV2Project(ctx, t, svc, "admin-opens", domain.VisibilityPrivate, v2UserB)
	task, err := svc.CreateTask(ctx, domain.TaskInput{
		ProjectID: p.ID, BusinessID: v2Business, OwnerID: v2UserB,
		Title: "zebrafinch quarterly plan", ReporterID: v2UserB,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if _, err := svc.CreateCycleV2(ctx, v2Business, p.ID, "Sprint", "2026-11-01", "2026-11-14"); err != nil {
		t.Fatalf("CreateCycleV2: %v", err)
	}
	if _, err := svc.CreateMilestoneV2(ctx, v2Business, p.ID, "Launch", "", "2026-12-01"); err != nil {
		t.Fatalf("CreateMilestoneV2: %v", err)
	}
	return adminProjectFixture{project: p, task: task}
}

func listProjectTasks(ctx context.Context, t *testing.T, svc *Service, callerID, businessID, projectID int64) int64 {
	t.Helper()
	_, total, err := svc.ListTasksV2(ctx, callerID, businessID, domain.ListTasksFilter{ProjectID: projectID, Limit: 50})
	if err != nil {
		t.Fatalf("ListTasksV2: %v", err)
	}
	return total
}

// Abed, 2026-11-18: business admins open private projects in Work again. The
// admin was never added to the project, and everything the open board reads
// has to answer for them.
func TestBusinessAdminOpensPrivateProjectAndItsWork(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	fx := newAdminProjectFixture(ctx, t, svc)
	svc.hubProfiles = adminInBusiness{businessID: v2Business}

	if _, err := svc.GetProjectV2(ctx, fx.project.ID, v2UserA, v2Business); err != nil {
		t.Fatalf("GetProjectV2 as admin: %v", err)
	}
	if !svc.IsProjectMemberOrBypass(ctx, v2UserA, v2Business, fx.project.ID) {
		t.Error("admin cannot read the project roster")
	}
	if got := listProjectTasks(ctx, t, svc, v2UserA, v2Business, fx.project.ID); got != 1 {
		t.Errorf("admin sees %d tasks on the board, want 1", got)
	}
	if _, total, err := svc.ListCyclesV2(ctx, fx.project.ID, v2Business, v2UserA, 50, 0); err != nil || total != 1 {
		t.Errorf("admin cycles = %d, %v; want 1", total, err)
	}
	pid := fx.project.ID
	if rows, _, err := svc.ListMilestonesV2(ctx, v2Business, &pid, v2UserA, 50, 0); err != nil || len(rows) != 1 {
		t.Errorf("admin milestones = %d, %v; want 1", len(rows), err)
	}
	if !svc.CanReadTask(ctx, v2UserA, v2Business, fx.task.ID) {
		t.Error("admin cannot read the subresources of a task on the board")
	}
	if !svc.CanSubscribeProject(ctx, v2UserA, fx.project.ID) {
		t.Error("admin cannot subscribe to the project topic")
	}
	if !svc.CanSubscribeTask(ctx, v2UserA, fx.task.ID) {
		t.Error("admin cannot subscribe to the task topic")
	}
}

// A plain member of the business who is not on the project still cannot open
// it or read anything inside it.
func TestPlainMemberStillCannotOpenPrivateProject(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	fx := newAdminProjectFixture(ctx, t, svc)
	svc.hubProfiles = adminInBusiness{businessID: 0}

	if _, err := svc.GetProjectV2(ctx, fx.project.ID, v2UserA, v2Business); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("GetProjectV2 as plain member: err = %v, want ErrUnauthorized", err)
	}
	if svc.IsProjectMemberOrBypass(ctx, v2UserA, v2Business, fx.project.ID) {
		t.Error("plain member can read the project roster")
	}
	if got := listProjectTasks(ctx, t, svc, v2UserA, v2Business, fx.project.ID); got != 0 {
		t.Errorf("plain member sees %d tasks, want 0", got)
	}
	if _, total, err := svc.ListCyclesV2(ctx, fx.project.ID, v2Business, v2UserA, 50, 0); err != nil || total != 0 {
		t.Errorf("plain member cycles = %d, %v; want 0", total, err)
	}
	pid := fx.project.ID
	if rows, _, err := svc.ListMilestonesV2(ctx, v2Business, &pid, v2UserA, 50, 0); err != nil || len(rows) != 0 {
		t.Errorf("plain member milestones = %d, %v; want 0", len(rows), err)
	}
	if svc.CanReadTask(ctx, v2UserA, v2Business, fx.task.ID) {
		t.Error("plain member can read a task's subresources")
	}
	if svc.CanSubscribeProject(ctx, v2UserA, fx.project.ID) || svc.CanSubscribeTask(ctx, v2UserA, fx.task.ID) {
		t.Error("plain member can subscribe to the project's topics")
	}
}

// Being an admin of one business opens nothing in another, whichever business
// the request claims to act in.
func TestAdminOfAnotherBusinessCannotOpenProject(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	fx := newAdminProjectFixture(ctx, t, svc)
	other := v2Business + 1
	svc.hubProfiles = adminInBusiness{businessID: other}

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
		if svc.CanReadTask(ctx, v2UserA, businessID, fx.task.ID) {
			t.Errorf("business %d: can read a task's subresources", businessID)
		}
	}
	if svc.CanSubscribeProject(ctx, v2UserA, fx.project.ID) || svc.CanSubscribeTask(ctx, v2UserA, fx.task.ID) {
		t.Error("admin of another business can subscribe to the project's topics")
	}
}

// Someone's My Tasks bucket is theirs alone: an admin of the same business
// opens it no more than anyone else does.
func TestAdminCannotOpenSomeoneElsesMyTasks(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	svc.hubProfiles = adminInBusiness{businessID: v2Business}

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

	if _, err := svc.GetProjectV2(ctx, bucketID, v2UserA, v2Business); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("GetProjectV2 on someone's bucket: err = %v, want ErrUnauthorized", err)
	}
	if got := listProjectTasks(ctx, t, svc, v2UserA, v2Business, bucketID); got != 0 {
		t.Errorf("admin sees %d tasks in someone's bucket, want 0", got)
	}
	if svc.CanReadTask(ctx, v2UserA, v2Business, task.ID) {
		t.Error("admin can read the subresources of someone's My Tasks task")
	}
	if svc.CanSubscribeProject(ctx, v2UserA, bucketID) || svc.CanSubscribeTask(ctx, v2UserA, task.ID) {
		t.Error("admin can subscribe to someone's My Tasks topics")
	}

	// The owner still opens their own bucket.
	if _, err := svc.GetProjectV2(ctx, bucketID, v2UserB, v2Business); err != nil {
		t.Errorf("owner GetProjectV2 on their own bucket: %v", err)
	}
}

// Ruling P8: opening is not listing. The sidebar, the business-wide task list,
// the business-wide milestone list and search stay what they were for an admin.
func TestAdminListingsStayCurated(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	fx := newAdminProjectFixture(ctx, t, svc)
	svc.hubProfiles = adminInBusiness{businessID: v2Business}

	projects, _, err := svc.ListProjectsV2(ctx, v2UserA, v2Business, "", "", "", 100, 0)
	if err != nil {
		t.Fatalf("ListProjectsV2: %v", err)
	}
	for _, p := range projects {
		if p.ID == fx.project.ID {
			t.Error("admin's sidebar lists a private project they are not on")
		}
	}

	if _, total, err := svc.ListTasksV2(ctx, v2UserA, v2Business, domain.ListTasksFilter{Limit: 50}); err != nil || total != 0 {
		t.Errorf("admin business-wide task list = %d, %v; want 0", total, err)
	}
	if rows, _, err := svc.ListMilestonesV2(ctx, v2Business, nil, v2UserA, 50, 0); err != nil || len(rows) != 0 {
		t.Errorf("admin business-wide milestones = %d, %v; want 0", len(rows), err)
	}

	// The creator finding the task proves the search can see it at all, so the
	// admin's empty result is the visibility rule and not an empty index.
	if _, total, err := svc.SearchTasksV2(ctx, v2UserB, v2Business, "zebrafinch", 50, 0); err != nil || total != 1 {
		t.Fatalf("creator search = %d, %v; want 1", total, err)
	}
	if _, total, err := svc.SearchTasksV2(ctx, v2UserA, v2Business, "zebrafinch", 50, 0); err != nil || total != 0 {
		t.Errorf("admin search = %d, %v; want 0", total, err)
	}
}
