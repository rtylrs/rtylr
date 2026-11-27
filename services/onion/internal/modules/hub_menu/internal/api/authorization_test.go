package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	hubmenuv1 "github.com/voxire/voxie/gen/go/hub/menu/v1"
)

func menuUser(licensed bool, permissions ...string) *authPkg.AuthUser {
	return &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppMenu,
				Licensed:    licensed,
				HasAccess:   licensed,
				Permissions: permissions,
			}},
		},
	}
}

func menuPermissionStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func TestListMenuRequiresInternalMenuPermission(t *testing.T) {
	tests := []struct {
		name string
		user *authPkg.AuthUser
	}{
		{name: "missing authentication", user: nil},
		{name: "missing menu license", user: menuUser(false, "menu:menus:view")},
		{name: "missing menu permission", user: menuUser(true)},
		{name: "wrong app permission", user: &authPkg.AuthUser{
			BusinessID: 9,
			UserID:     7,
			EffectiveAccess: &authorization.EffectiveAccess{
				Apps: []authorization.EffectiveAppAccess{{
					AppID:       authorization.AppPOS,
					Licensed:    true,
					HasAccess:   true,
					Permissions: []string{"menu:menus:view"},
				}},
			},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.user != nil {
				ctx = authPkg.WithAuthUser(ctx, tt.user)
			}
			_, err := (&MenuAPI{}).ListMenu(ctx, &hubmenuv1.ListMenuRequest{})
			if got := menuPermissionStatus(err); got != http.StatusForbidden {
				t.Fatalf("status = %d, want %d", got, http.StatusForbidden)
			}
		})
	}
}

func TestUpdateMenuRequiresInternalMenuPermission(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), menuUser(true))
	_, err := (&MenuAPI{}).UpdateMenu(ctx, &hubmenuv1.UpdateMenuRequest{})
	if got := menuPermissionStatus(err); got != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", got, http.StatusForbidden)
	}
}

func TestInternalMenuPermissionAllowsGrantedPermission(t *testing.T) {
	user := menuUser(true, "menu:menus:view")
	if !authPkg.HasAppPermission(user, authorization.AppMenu, "menu:menus:view") {
		t.Fatal("expected the granted Menu permission to be recognized")
	}
}

func TestInternalMenuPermissionRequiresGrantedBrandingPermission(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), menuUser(true, "menu:menus:view"))
	if _, err := requireMenuPermission(ctx, "menu:branding:edit"); menuPermissionStatus(err) != http.StatusForbidden {
		t.Fatalf("expected branding permission denial, got %v", err)
	}
}

func TestMenuOrPOSPermissionAllowsScopedPOSStaffForOperations(t *testing.T) {
	posStaff := &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppPOS,
				Licensed:    true,
				HasAccess:   true,
				Permissions: []string{"pos:orders:view", "pos:orders:create"},
				PermissionScopes: map[string][]authorization.Scope{
					"pos:orders:view": {{Type: "location", ID: "4"}},
				},
			}},
		},
	}
	if !hasMenuOrPOSPermission(posStaff, "menu:orders:view", "pos:orders:view") {
		t.Fatal("expected scoped POS staff to pass operations permission check")
	}

	unscopedPOS := &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppPOS,
				Licensed:    true,
				HasAccess:   true,
				Permissions: []string{"pos:orders:view"},
			}},
		},
	}
	if hasMenuOrPOSPermission(unscopedPOS, "menu:orders:view", "pos:orders:view") {
		t.Fatal("expected unscoped POS user to be denied")
	}
	if _, err := requireMenuOrPOSPermission(authPkg.WithAuthUser(context.Background(), unscopedPOS), "menu:orders:view", "pos:orders:view"); menuPermissionStatus(err) != http.StatusForbidden {
		t.Fatalf("expected unscoped POS user to receive 403, got %v", err)
	}
}

func TestMenuReviewModerationRequiresBrandingEditPermission(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), menuUser(true, "menu:menus:view"))
	if _, err := (&MenuAPI{}).ListMenuReviews(ctx, &hubmenuv1.ListMenuReviewsRequest{}); menuPermissionStatus(err) != http.StatusForbidden {
		t.Fatalf("ListMenuReviews status = %d, want 403", menuPermissionStatus(err))
	}
	if _, err := (&MenuAPI{}).UpdateMenuReviewStatus(ctx, &hubmenuv1.UpdateMenuReviewStatusRequest{Id: 1, Status: "approved"}); menuPermissionStatus(err) != http.StatusForbidden {
		t.Fatalf("UpdateMenuReviewStatus status = %d, want 403", menuPermissionStatus(err))
	}
	if _, err := (&MenuAPI{}).DeleteMenuReview(ctx, &hubmenuv1.DeleteMenuReviewRequest{Id: 1}); menuPermissionStatus(err) != http.StatusForbidden {
		t.Fatalf("DeleteMenuReview status = %d, want 403", menuPermissionStatus(err))
	}
}
