package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	flowprojectv1 "github.com/voxire/voxie/gen/go/flow/project/v1"
)

func flowPermissionStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func flowUser(permissions ...string) *authPkg.AuthUser {
	return &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
			AppID:       authorization.AppFlow,
			Licensed:    true,
			HasAccess:   true,
			Permissions: permissions,
		}}},
	}
}

func TestListProjectsRequiresFlowPermissionBeforeServiceAccess(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), flowUser())
	_, err := (&ProjectAPI{}).ListProjects(ctx, &flowprojectv1.ListProjectsRequest{})
	if got := flowPermissionStatus(err); got != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", got, http.StatusForbidden)
	}
}

func TestRequireFlowPermissionRequiresLicensedFlowApp(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), flowUser("flow:projects:view"))
	if err := requireFlowPermission(ctx, "flow:projects:view"); err != nil {
		t.Fatalf("expected granted flow permission to pass, got %v", err)
	}

	wrongApp := &authPkg.AuthUser{EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
		AppID:       authorization.AppHR,
		Licensed:    true,
		HasAccess:   true,
		Permissions: []string{"flow:projects:view"},
	}}}}
	if err := requireFlowPermission(authPkg.WithAuthUser(context.Background(), wrongApp), "flow:projects:view"); flowPermissionStatus(err) != http.StatusForbidden {
		t.Fatalf("expected wrong app to be denied, got %v", err)
	}
}
