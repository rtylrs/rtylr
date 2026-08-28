package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/roles/internal/application"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	hubpermissionv1 "github.com/voxire/voxie/gen/go/hub/permission/v1"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	cases := []testutil.DirectAuthorizationCase{
		{Name: "hub roles view", App: authorization.AppHub, Permission: "hub:roles:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewPermissionAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).GetPermissions(ctx, nil)
			return response, err
		}},
		{Name: "hub roles create", App: authorization.AppHub, Permission: "hub:roles:create", Role: "owner", DelegationUserID: 1002, Call: func(ctx context.Context) (any, error) {
			appID := "dash"
			response, err := NewPermissionAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).CreateRole(ctx, &hubpermissionv1.CreateRoleRequest{Name: "Authorization Probe", AppId: &appID, Permissions: []string{"hub:team:view"}})
			return response, err
		}},
		{Name: "hub roles edit", App: authorization.AppHub, Permission: "hub:roles:edit", Role: "owner", DelegationUserID: 1002, Call: func(ctx context.Context) (any, error) {
			response, err := NewPermissionAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).UpdateRole(ctx, &hubpermissionv1.UpdateRoleRequest{Id: 7305, Body: &hubpermissionv1.UpdateRoleBody{Name: "Authorization Probe", Permissions: []string{"hub:team:view"}}})
			return response, err
		}},
		{Name: "hub roles archive", App: authorization.AppHub, Permission: "hub:roles:archive", Call: func(ctx context.Context) (any, error) {
			response, err := NewPermissionAPI(application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))).ArchiveRole(ctx, &hubpermissionv1.ArchiveRoleRequest{Id: 7305, Reason: "authorization probe"})
			return response, err
		}},
	}
	testutil.RunDirectAuthorizationCases(t, cases)
}
