package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	flowprojectv1 "github.com/voxire/voxie/gen/go/flow/project/v1"
	flowtaskv1 "github.com/voxire/voxie/gen/go/flow/task/v1"
	flowworkspacev1 "github.com/voxire/voxie/gen/go/flow/workspace/v1"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	assignedUserID := int64(8)
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "flow projects view", App: authorization.AppFlow, Permission: "flow:projects:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewProjectAPI(application.NewService(ctx, testsupport.NewRepository(testutil.DirectAuthorizationDB(ctx)))).ListProjects(ctx, nil)
			return response, err
		}},
		{Name: "flow projects create", App: authorization.AppFlow, Permission: "flow:projects:create", Call: func(ctx context.Context) (any, error) {
			response, err := NewProjectAPI(application.NewService(ctx, testsupport.NewRepository(testutil.DirectAuthorizationDB(ctx)))).CreateProject(ctx, nil)
			return response, err
		}},
		{Name: "flow projects edit", App: authorization.AppFlow, Permission: "flow:projects:edit", Call: func(ctx context.Context) (any, error) {
			response, err := NewProjectAPI(application.NewService(ctx, testsupport.NewRepository(testutil.DirectAuthorizationDB(ctx)))).UpdateProject(ctx, &flowprojectv1.UpdateProjectRequest{Id: 1})
			return response, err
		}},
		{Name: "flow projects assign", App: authorization.AppFlow, Permission: "flow:projects:assign", Call: func(ctx context.Context) (any, error) {
			response, err := NewProjectAPI(application.NewService(ctx, testsupport.NewRepository(testutil.DirectAuthorizationDB(ctx)))).AddProjectMember(ctx, &flowprojectv1.AddProjectMemberRequest{ProjectId: testutil.Int64Ptr(1), UserId: testutil.Int64Ptr(assignedUserID)})
			return response, err
		}},
		{Name: "flow tasks view", App: authorization.AppFlow, Permission: "flow:tasks:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewTaskAPI(application.NewService(ctx, testsupport.NewRepository(testutil.DirectAuthorizationDB(ctx)))).ListTasks(ctx, &flowtaskv1.ListTasksRequest{})
			return response, err
		}},
		{Name: "flow tasks create", App: authorization.AppFlow, Permission: "flow:tasks:create", Call: func(ctx context.Context) (any, error) {
			response, err := NewTaskAPI(application.NewService(ctx, testsupport.NewRepository(testutil.DirectAuthorizationDB(ctx)))).CreateTask(ctx, &flowtaskv1.CreateTaskRequest{})
			return response, err
		}},
		{Name: "flow tasks edit", App: authorization.AppFlow, Permission: "flow:tasks:edit", Call: func(ctx context.Context) (any, error) {
			response, err := NewTaskAPI(application.NewService(ctx, testsupport.NewRepository(testutil.DirectAuthorizationDB(ctx)))).UpdateTask(ctx, &flowtaskv1.UpdateTaskRequest{Id: 1, Body: &flowtaskv1.UpdateTaskBody{Title: "Direct Authorization Task"}})
			return response, err
		}},
		{Name: "flow tasks assign", App: authorization.AppFlow, Permission: "flow:tasks:assign", Extra: []string{"flow:tasks:create"}, Call: func(ctx context.Context) (any, error) {
			response, err := NewTaskAPI(application.NewService(ctx, testsupport.NewRepository(testutil.DirectAuthorizationDB(ctx)))).CreateTask(ctx, &flowtaskv1.CreateTaskRequest{AssigneeId: &assignedUserID})
			return response, err
		}},
		{Name: "flow team manage", App: authorization.AppFlow, Permission: "flow:team:manage", Call: func(ctx context.Context) (any, error) {
			response, err := NewWorkspaceAPI(application.NewService(ctx, testsupport.NewRepository(testutil.DirectAuthorizationDB(ctx)))).AddWorkspaceMember(ctx, &flowworkspacev1.AddWorkspaceMemberRequest{WorkspaceId: testutil.Int64Ptr(1), UserId: testutil.Int64Ptr(assignedUserID), Role: "member"})
			return response, err
		}},
		{Name: "flow managers manage", App: authorization.AppFlow, Permission: "flow:managers:manage", Call: func(ctx context.Context) (any, error) {
			response, err := NewWorkspaceAPI(application.NewService(ctx, testsupport.NewRepository(testutil.DirectAuthorizationDB(ctx)))).UpdateWorkspaceMemberRole(ctx, &flowworkspacev1.UpdateWorkspaceMemberRoleRequest{WorkspaceId: testutil.Int64Ptr(1), UserId: testutil.Int64Ptr(assignedUserID), Role: "admin"})
			return response, err
		}},
	})
}
