package api

import (
	"context"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	hubteamaccessv1 "github.com/voxire/voxie/gen/go/hub/team/access/v1"
	hubteamgroupsv1 "github.com/voxire/voxie/gen/go/hub/team/groups/v1"
	hubteaminvitationsv1 "github.com/voxire/voxie/gen/go/hub/team/invitations/v1"
	hubteammembersv1 "github.com/voxire/voxie/gen/go/hub/team/members/v1"
)

func teamAuthorizationUser(permission string) *authPkg.AuthUser {
	return &authPkg.AuthUser{
		UserID:     7,
		BusinessID: 9,
		Role:       "member",
		EffectiveAccess: &authorization.EffectiveAccess{
			DashRole:        "member",
			DashPermissions: []string{permission},
		},
	}
}

func TestTeamAPIsRequireTheirCanonicalDashPermissionBeforeServiceAccess(t *testing.T) {
	tests := []struct {
		name       string
		permission string
		call       func(context.Context) error
	}{
		{name: "list team", permission: "hub:team:view", call: func(ctx context.Context) error {
			_, err := (&MembersAPI{}).ListTeam(ctx, &hubteammembersv1.ListTeamRequest{})
			return err
		}},
		{name: "get member", permission: "hub:team:view", call: func(ctx context.Context) error {
			_, err := (&MembersAPI{}).GetMember(ctx, &hubteammembersv1.GetMemberRequest{UserId: 1})
			return err
		}},
		{name: "remove member", permission: "hub:team:deactivate", call: func(ctx context.Context) error {
			_, err := (&MembersAPI{}).RemoveMember(ctx, &hubteammembersv1.RemoveMemberRequest{UserId: 1})
			return err
		}},
		{name: "invite user", permission: "hub:team:invite", call: func(ctx context.Context) error {
			_, err := (&InvitationsAPI{}).InviteUser(ctx, &hubteaminvitationsv1.InviteUserRequest{})
			return err
		}},
		{name: "bulk invite", permission: "hub:team:invite", call: func(ctx context.Context) error {
			_, err := (&InvitationsAPI{}).BulkInvite(ctx, &hubteaminvitationsv1.BulkInviteRequest{})
			return err
		}},
		{name: "list groups", permission: "hub:groups:manage", call: func(ctx context.Context) error {
			_, err := (&GroupsAPI{}).ListGroups(ctx, &hubteamgroupsv1.ListGroupsRequest{})
			return err
		}},
		{name: "create group", permission: "hub:groups:manage", call: func(ctx context.Context) error {
			_, err := (&GroupsAPI{}).CreateGroup(ctx, &hubteamgroupsv1.CreateGroupRequest{})
			return err
		}},
		{name: "get group", permission: "hub:groups:manage", call: func(ctx context.Context) error {
			_, err := (&GroupsAPI{}).GetGroupDetail(ctx, &hubteamgroupsv1.GetGroupDetailRequest{Id: 1})
			return err
		}},
		{name: "update group", permission: "hub:groups:manage", call: func(ctx context.Context) error {
			_, err := (&GroupsAPI{}).UpdateGroup(ctx, &hubteamgroupsv1.UpdateGroupRequest{Id: 1})
			return err
		}},
		{name: "delete group", permission: "hub:groups:manage", call: func(ctx context.Context) error {
			_, err := (&GroupsAPI{}).DeleteGroup(ctx, &hubteamgroupsv1.DeleteGroupRequest{Id: 1})
			return err
		}},
		{name: "add group members", permission: "hub:groups:manage", call: func(ctx context.Context) error {
			_, err := (&GroupsAPI{}).AddGroupMembers(ctx, &hubteamgroupsv1.AddGroupMembersRequest{Id: 1, UserIds: []int64{2}})
			return err
		}},
		{name: "remove group member", permission: "hub:groups:manage", call: func(ctx context.Context) error {
			_, err := (&GroupsAPI{}).RemoveGroupMember(ctx, &hubteamgroupsv1.RemoveGroupMemberRequest{Id: 1, UserId: 2})
			return err
		}},
		{name: "get user groups", permission: "hub:groups:manage", call: func(ctx context.Context) error {
			_, err := (&GroupsAPI{}).GetUserGroups(ctx, &hubteamgroupsv1.GetUserGroupsRequest{UserId: 2})
			return err
		}},
		{name: "list access requests", permission: "hub:team:view", call: func(ctx context.Context) error {
			_, err := (&AccessAPI{}).ListAccessRequests(ctx, &hubteamaccessv1.ListAccessRequestsRequest{})
			return err
		}},
		{name: "approve access request", permission: "hub:team:edit", call: func(ctx context.Context) error {
			_, err := (&AccessAPI{}).ApproveAccessRequest(ctx, &hubteamaccessv1.ApproveAccessRequestRequest{Id: 1, UserId: testutil.Int64Ptr(2), App: "crm"})
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := authPkg.WithAuthUser(context.Background(), teamAuthorizationUser(""))
			if err := tt.call(ctx); err == nil {
				t.Fatal("request without the canonical permission reached the service")
			}
		})
	}
}

func TestTeamPermissionGuardAcceptsOnlyTheRequestedDashPermission(t *testing.T) {
	for _, permission := range []string{"hub:team:view", "hub:team:invite", "hub:team:edit", "hub:team:deactivate", "hub:groups:manage"} {
		ctx := authPkg.WithAuthUser(context.Background(), teamAuthorizationUser(permission))
		if _, err := requireHubPermission(ctx, permission); err != nil {
			t.Fatalf("permission %q was denied: %v", permission, err)
		}
		if _, err := requireHubPermission(ctx, "hub:roles:archive"); err == nil {
			t.Fatalf("permission %q incorrectly granted hub:roles:archive", permission)
		}
	}
}
