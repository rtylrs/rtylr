package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onkithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	hubactivityv1 "github.com/voxire/voxie/gen/go/hub/activity/v1"
)

func activityUser(permissions ...string) *authPkg.AuthUser {
	return &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			DashPermissions: permissions,
		},
	}
}

func activityPermissionStatus(err error) int {
	var oneKitErr *onkithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func TestListActivityRequiresAuditPermissionBeforeServiceAccess(t *testing.T) {
	api := &ActivityAPI{}
	ctx := authPkg.WithAuthUser(context.Background(), activityUser())

	_, err := api.ListActivity(ctx, &hubactivityv1.ListActivityRequest{})
	if got := activityPermissionStatus(err); got != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", got, http.StatusForbidden)
	}
}

func TestListActivityAcceptsAuditPermission(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), activityUser("hub:audit:view"))
	if !authPkg.HasDashboardPermission(authPkg.DeviceCtx(ctx).User, "hub:audit:view") {
		t.Fatal("expected audit permission to be available to the handler")
	}
}
