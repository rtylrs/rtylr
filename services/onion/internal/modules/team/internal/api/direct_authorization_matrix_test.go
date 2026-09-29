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
					// No scope: dash is not a resource scoped app, and
					// chk_auth_assignment_scope_type cannot store a dash scope.
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

	// A scope the assignment table cannot store must be rejected before any
	// write. The normalized team writer revokes the member's licenses and role
	// assignments in the same transaction, so an unvalidated scope used to roll
	// the whole authorization back and return an opaque 500.
	t.Run("invalid assignment scope is rejected without a partial rewrite", func(t *testing.T) {
		caseCtx := context.Background()
		db := testutil.DirectAuthorizationDB(caseCtx)
		role := "staff"
		fullName := "Direct Authorization Staff"
		_, err := NewPermissionsAPI(application.NewService(caseCtx, db)).UpdatePermissions(caseCtx, &hubteampermissionsv1.UpdatePermissionsRequest{
			UserId:   1002,
			Role:     &role,
			FullName: &fullName,
			AppAssignments: []*hubuserv1.AppRoleAssignment{{
				AppId:            "pos",
				RoleIds:          []string{"7305"},
				AssignmentSource: string(authorization.SourceExplicit),
				Scopes:           []*hubuserv1.RoleScope{{ScopeType: "business", ScopeId: "1001"}},
			}},
			GroupIds: []int64{},
		})
		if err == nil {
			t.Fatal("expected a business scope to be rejected")
		}
		var rowCount int
		if scanErr := db.QueryRowContext(caseCtx, `
			SELECT COUNT(*) FROM authorization_role_assignment_scopes s
			JOIN authorization_role_assignments a ON a.id = s.assignment_id
			WHERE a.user_id = 1002 AND s.scope_type = 'business'`).Scan(&rowCount); scanErr != nil {
			t.Fatalf("count scopes: %v", scanErr)
		}
		if rowCount != 0 {
			t.Fatalf("a rejected scope left %d rows behind", rowCount)
		}
	})
}
