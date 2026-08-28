package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/corezio/onekit/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	posuserv1 "github.com/voxire/voxie/gen/go/pos/user/v1"
)

func posUsersStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func TestPOSUserManagementRequiresEmployeesPermissionBeforeServiceAccess(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:     authorization.AppPOS,
				Licensed:  true,
				HasAccess: true,
			}},
		},
	})
	api := &UserAPI{}
	cases := []func() error{
		func() error { _, err := api.ListPOSUsers(ctx, &posuserv1.ListPOSUsersRequest{}); return err },
		func() error { _, err := api.CreatePOSUser(ctx, &posuserv1.CreatePOSUserRequest{}); return err },
		func() error { _, err := api.UpdatePOSUser(ctx, &posuserv1.UpdatePOSUserRequest{}); return err },
		func() error { _, err := api.DeactivatePOSUser(ctx, &posuserv1.DeactivatePOSUserRequest{}); return err },
	}
	for i, call := range cases {
		if err := call(); posUsersStatus(err) != http.StatusForbidden {
			t.Fatalf("case %d returned %v, want 403 before service access", i, err)
		}
	}
}

func TestPOSUserOrderVerificationRequiresOrderViewPermission(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:     authorization.AppPOS,
				Licensed:  true,
				HasAccess: true,
			}},
		},
	})
	_, err := (&UserAPI{}).VerifyPOSUserPIN(ctx, &posuserv1.VerifyPOSUserPINRequest{})
	if posUsersStatus(err) != http.StatusForbidden {
		t.Fatalf("returned %v, want 403 before service access", err)
	}
}
