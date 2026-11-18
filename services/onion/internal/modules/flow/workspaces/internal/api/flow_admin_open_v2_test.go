package api_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/api"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/hubprofile"
	"github.com/voxire/onion/internal/testutil"
	flowprojectv2 "github.com/voxire/voxie/gen/go/flow/project/v2"
	flowstatusv2 "github.com/voxire/voxie/gen/go/flow/status/v2"
	flowtaskv2 "github.com/voxire/voxie/gen/go/flow/task/v2"
)

// roleFlowAdminID holds flow:team:manage in the fixture business and is on no
// workspace or project. It is neither the business owner nor an admin.
const roleFlowAdminID int64 = 900006

// flowAdminProfiles answers the central access lookup the way the hub does for
// a Flow Admin: flow:team:manage for roleFlowAdminID and nothing for anyone
// else, so the fixture's other callers keep the access they had.
type flowAdminProfiles struct{}

func (flowAdminProfiles) GetProfile(_ context.Context, userID, businessID int64) (hubprofile.Profile, error) {
	access := authorization.EffectiveAccess{BusinessID: businessID}
	if userID == roleFlowAdminID && businessID == testutil.BusinessID {
		access.Apps = []authorization.EffectiveAppAccess{{
			AppID: authorization.AppFlow, Licensed: true, HasAccess: true,
			Permissions: []string{"flow:team:manage"},
		}}
	}
	return hubprofile.Profile{EffectiveAccess: &access}, nil
}

// Abed, 2026-11-18: a Flow Admin opens a project they are not on and reads its
// board, statuses and task drawer through the same handlers the portal calls,
// and every write on it is still refused with a 403.
func TestFlowAdminOpensProjectReadOnlyV2(t *testing.T) {
	f := newRoleFixture(t)
	testutil.GrantBusinessAccess(t, testutil.OpenDB(t), testutil.BusinessID, roleFlowAdminID)
	f.service.SetHubProfileResolver(flowAdminProfiles{})
	taskID := f.newTask(t)
	ctx := ctxAs(roleFlowAdminID)
	projID := f.projID

	if _, err := f.projAPI.GetProject(ctx, &flowprojectv2.GetProjectRequest{Id: projID}); err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if _, err := f.projAPI.ListProjectMembers(ctx, &flowprojectv2.ListProjectMembersRequest{ProjectId: &projID}); err != nil {
		t.Errorf("ListProjectMembers: %v", err)
	}
	if _, err := f.projAPI.ListProjectColumns(ctx, &flowprojectv2.ListProjectColumnsRequest{ProjectId: &projID}); err != nil {
		t.Errorf("ListProjectColumns: %v", err)
	}
	tasks := api.NewTaskAPIV2(f.service)
	listed, err := tasks.ListTasks(ctx, &flowtaskv2.ListTasksRequest{ProjectId: &projID})
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if got := len(listed.GetTasks()); got != 1 {
		t.Errorf("ListTasks returned %d tasks, want 1", got)
	}
	if _, err := api.NewStatusAPIV2(f.service).ListStatuses(ctx, &flowstatusv2.ListStatusesRequest{ProjectId: &projID}); err != nil {
		t.Errorf("ListStatuses: %v", err)
	}
	if _, err := tasks.GetTaskDetail(ctx, &flowtaskv2.GetTaskDetailRequest{Id: taskID}); err != nil {
		t.Errorf("GetTaskDetail: %v", err)
	}
	if _, err := tasks.GetTaskActivity(ctx, &flowtaskv2.GetTaskActivityRequest{TaskId: taskID}); err != nil {
		t.Errorf("GetTaskActivity: %v", err)
	}

	_, err = tasks.UpdateTask(ctx, &flowtaskv2.UpdateTaskRequest{
		Id: taskID, Body: &flowtaskv2.UpdateTaskBody{Title: "Retitled by a Flow Admin"},
	})
	assertPermissionDenied(t, err, "Flow Admin UpdateTask")
	_, err = tasks.AddComment(ctx, &flowtaskv2.AddCommentRequest{TaskId: taskID, Content: "hello"})
	assertPermissionDenied(t, err, "Flow Admin AddComment")
	_, err = f.projAPI.GetProjectLabourCost(ctx, &flowprojectv2.GetProjectLabourCostRequest{Id: projID})
	assertPermissionDenied(t, err, "Flow Admin GetProjectLabourCost")
}

// The same handlers still refuse a plain member of the business who is not on
// the project, so the Flow Admin pass above is the Flow Admin's and nobody
// else's.
func TestStrangerStillCannotOpenProjectV2(t *testing.T) {
	f := newRoleFixture(t)
	f.service.SetHubProfileResolver(flowAdminProfiles{})
	taskID := f.newTask(t)
	ctx := ctxAs(roleStrangerID)
	projID := f.projID

	if _, err := f.projAPI.GetProject(ctx, &flowprojectv2.GetProjectRequest{Id: projID}); err == nil {
		t.Error("GetProject: a stranger opened the project")
	}
	tasks := api.NewTaskAPIV2(f.service)
	listed, err := tasks.ListTasks(ctx, &flowtaskv2.ListTasksRequest{ProjectId: &projID})
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if got := len(listed.GetTasks()); got != 0 {
		t.Errorf("ListTasks returned %d tasks to a stranger, want 0", got)
	}
	_, err = api.NewStatusAPIV2(f.service).ListStatuses(ctx, &flowstatusv2.ListStatusesRequest{ProjectId: &projID})
	assertPermissionDenied(t, err, "stranger ListStatuses")
	_, err = tasks.GetTaskDetail(ctx, &flowtaskv2.GetTaskDetailRequest{Id: taskID})
	assertPermissionDenied(t, err, "stranger GetTaskDetail")
}
