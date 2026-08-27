package api

import (
	"context"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	hubpermissionv1 "github.com/voxire/voxie/gen/go/hub/permission/v1"
)

func roleAuthorizationUser(permissions ...string) *authPkg.AuthUser {
	return &authPkg.AuthUser{
		UserID:     7,
		BusinessID: 9,
		Role:       "member",
		EffectiveAccess: &authorization.EffectiveAccess{
			DashRole:        "member",
			DashPermissions: permissions,
		},
	}
}

func TestPermissionAPIsRejectMissingCanonicalPermissionBeforeServiceAccess(t *testing.T) {
	tests := []struct {
		name       string
		permission string
		call       func(context.Context) error
	}{
		{name: "get permissions", permission: "hub:roles:view", call: func(ctx context.Context) error {
			_, err := (&PermissionAPI{}).GetPermissions(ctx, &hubpermissionv1.GetPermissionsRequest{})
			return err
		}},
		{name: "list roles", permission: "hub:roles:view", call: func(ctx context.Context) error {
			_, err := (&PermissionAPI{}).ListRoles(ctx, &hubpermissionv1.ListRolesRequest{})
			return err
		}},
		{name: "create role", permission: "hub:roles:create", call: func(ctx context.Context) error {
			_, err := (&PermissionAPI{}).CreateRole(ctx, &hubpermissionv1.CreateRoleRequest{})
			return err
		}},
		{name: "update permission matrix", permission: "hub:roles:edit", call: func(ctx context.Context) error {
			_, err := (&PermissionAPI{}).UpdatePermissionMatrix(ctx, &hubpermissionv1.UpdatePermissionMatrixRequest{})
			return err
		}},
		{name: "update role", permission: "hub:roles:edit", call: func(ctx context.Context) error {
			_, err := (&PermissionAPI{}).UpdateRole(ctx, &hubpermissionv1.UpdateRoleRequest{})
			return err
		}},
		{name: "delete role", permission: "hub:roles:archive", call: func(ctx context.Context) error {
			_, err := (&PermissionAPI{}).DeleteRole(ctx, &hubpermissionv1.DeleteRoleRequest{})
			return err
		}},
		{name: "get role", permission: "hub:roles:view", call: func(ctx context.Context) error {
			_, err := (&PermissionAPI{}).GetRole(ctx, &hubpermissionv1.GetRoleRequest{})
			return err
		}},
		{name: "clone role", permission: "hub:roles:create", call: func(ctx context.Context) error {
			_, err := (&PermissionAPI{}).CloneRole(ctx, &hubpermissionv1.CloneRoleRequest{})
			return err
		}},
		{name: "archive role", permission: "hub:roles:archive", call: func(ctx context.Context) error {
			_, err := (&PermissionAPI{}).ArchiveRole(ctx, &hubpermissionv1.ArchiveRoleRequest{})
			return err
		}},
		{name: "list assignments", permission: "hub:roles:view", call: func(ctx context.Context) error {
			_, err := (&PermissionAPI{}).ListRoleAssignments(ctx, &hubpermissionv1.ListRoleAssignmentsRequest{})
			return err
		}},
		{name: "assign role", permission: "hub:roles:edit", call: func(ctx context.Context) error {
			_, err := (&PermissionAPI{}).AssignRole(ctx, &hubpermissionv1.AssignRoleRequest{})
			return err
		}},
		{name: "remove assignment", permission: "hub:roles:edit", call: func(ctx context.Context) error {
			_, err := (&PermissionAPI{}).RemoveRoleAssignment(ctx, &hubpermissionv1.RemoveRoleAssignmentRequest{})
			return err
		}},
		{name: "preview impact", permission: "hub:roles:edit", call: func(ctx context.Context) error {
			_, err := (&PermissionAPI{}).PreviewPermissionImpact(ctx, &hubpermissionv1.PreviewPermissionImpactRequest{})
			return err
		}},
		{name: "list audit", permission: "hub:roles:view", call: func(ctx context.Context) error {
			_, err := (&PermissionAPI{}).ListRoleAudit(ctx, &hubpermissionv1.ListRoleAuditRequest{})
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := authPkg.WithAuthUser(context.Background(), roleAuthorizationUser())
			if err := tt.call(ctx); err == nil {
				t.Fatal("request without the canonical role permission reached the service")
			}
		})
	}
}

func TestRolePermissionGuardsUseExactDashPermissions(t *testing.T) {
	for _, permission := range []string{"hub:roles:view", "hub:roles:create", "hub:roles:edit", "hub:roles:archive"} {
		user := roleAuthorizationUser(permission)
		if permission == "hub:roles:view" && !canViewRoles(user) {
			t.Fatal("role view permission was not recognized")
		}
		if permission == "hub:roles:create" && !canEditRoles(user) {
			t.Fatal("role create permission was not recognized")
		}
		if permission == "hub:roles:edit" && !canEditRoles(user) {
			t.Fatal("role edit permission was not recognized")
		}
		if permission == "hub:roles:archive" && !canArchiveRoles(user) {
			t.Fatal("role archive permission was not recognized")
		}
	}
	if canViewRoles(roleAuthorizationUser("hub:team:view")) || canEditRoles(roleAuthorizationUser("hub:team:edit")) || canArchiveRoles(roleAuthorizationUser("hub:team:deactivate")) {
		t.Fatal("unrelated Dash permissions crossed the role administration boundary")
	}
}
