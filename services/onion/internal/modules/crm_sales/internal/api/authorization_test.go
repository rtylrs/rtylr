package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/corezio/onekit/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

func crmUser(permissions ...string) *authPkg.AuthUser {
	return &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppCRM,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
			}},
		},
	}
}

func crmPermissionStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func TestListContactsRequiresFeaturePermission(t *testing.T) {
	h := &ListContactsHandler{}
	ctx := authPkg.WithAuthUser(context.Background(), crmUser())

	_, err := h.Serve(ctx, &ListContactsRequest{})
	if got := crmPermissionStatus(err); got != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", got, http.StatusForbidden)
	}
}

func TestCreateContactRequiresCreatePermission(t *testing.T) {
	h := &CreateContactHandler{}
	ctx := authPkg.WithAuthUser(context.Background(), crmUser("crm:customers:view"))

	_, err := h.Serve(ctx, &CreateContactRequest{})
	if got := crmPermissionStatus(err); got != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", got, http.StatusForbidden)
	}
}
