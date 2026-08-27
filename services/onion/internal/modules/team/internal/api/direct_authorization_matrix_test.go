package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/team/internal/application"

	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	hubteaminvitationsv1 "github.com/voxire/voxie/gen/go/hub/team/invitations/v1"
	hubteammembersv1 "github.com/voxire/voxie/gen/go/hub/team/members/v1"
	hubteampermissionsv1 "github.com/voxire/voxie/gen/go/hub/team/permissions/v1"
	hubuserv1 "github.com/voxire/voxie/gen/go/hub/user/v1"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	cases := []testutil.DirectAuthorizationCase{
		{Name: "hub team view", App: authorization.AppHub, Permission: "hub:team:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewMembersAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).ListTeam(ctx, nil)
			return response, err
		}},
		{Name: "hub team invite", App: authorization.AppHub, Permission: "hub:team:invite", Call: func(ctx context.Context) (any, error) {
			response, err := NewInvitationsAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).InviteUser(ctx, &hubteaminvitationsv1.InviteUserRequest{Email: "direct-invite@test.local", FullName: "Direct Invite"})
			return response, err
		}},
		{Name: "hub team edit", App: authorization.AppHub, Permission: "hub:team:edit", DelegationUserID: 1002, Call: func(ctx context.Context) (any, error) {
			fullName := "Direct Authorization Staff"
			role := "staff"
			response, err := NewPermissionsAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).UpdatePermissions(ctx, &hubteampermissionsv1.UpdatePermissionsRequest{
				UserId:   1002,
				Role:     &role,
				FullName: &fullName,
				AppAssignments: []*hubuserv1.AppRoleAssignment{{
					AppId:            "dash",
					RoleIds:          []string{"7305"},
					AssignmentSource: string(authorization.SourceExplicit),
					Scopes:           []*hubuserv1.RoleScope{{ScopeType: "business", ScopeId: "1001"}},
				}},
				GroupIds: []int64{},
			})
			return response, err
		}},
		{Name: "hub team deactivate", App: authorization.AppHub, Permission: "hub:team:deactivate", Call: func(ctx context.Context) (any, error) {
			response, err := NewMembersAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).RemoveMember(ctx, &hubteammembersv1.RemoveMemberRequest{UserId: 1003})
			return response, err
		}},
		{Name: "hub licenses manage", App: authorization.AppHub, Permission: "hub:licenses:manage", Extra: []string{"hub:team:edit"}, DelegationUserID: 1002, Call: func(ctx context.Context) (any, error) {
			response, err := NewPermissionsAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).UpdatePermissions(ctx, &hubteampermissionsv1.UpdatePermissionsRequest{UserId: 1002, Apps: []string{"pos"}})
			return response, err
		}},
		{Name: "hub groups manage", App: authorization.AppHub, Permission: "hub:groups:manage", Call: func(ctx context.Context) (any, error) {
			response, err := NewGroupsAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).ListGroups(ctx, nil)
			return response, err
		}},
		{Name: "hub hierarchy manage", App: authorization.AppHub, Permission: "hub:hierarchy:manage", Call: func(ctx context.Context) (any, error) {
			response, err := NewAccessAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).ListHierarchy(ctx, &ListHierarchyRequest{})
			return response, err
		}},
	}
	testutil.RunDirectAuthorizationCases(t, cases)
}
