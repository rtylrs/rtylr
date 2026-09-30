package http

import (
	"net/http"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

func TestCustomerExportUsesCanonicalPermission(t *testing.T) {
	appID, permission, ok := featurePermissionForRequest(http.MethodGet, "/crm/customers/export")
	if !ok || appID != "crm" || permission != "crm:customers:export" {
		t.Fatalf("customer export mapping = %s:%s,%t", appID, permission, ok)
	}
}

func TestERPFeatureRoutesUseCanonicalPermissions(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		permission string
	}{
		{name: "settings view", method: http.MethodGet, path: "/erp/settings/v1/settings", permission: "erp:settings:view"},
		{name: "settings manage", method: http.MethodPut, path: "/erp/settings/v1/settings", permission: "erp:settings:manage"},
		{name: "supplier view", method: http.MethodGet, path: "/erp/supplier/v1/suppliers", permission: "erp:suppliers:view"},
		{name: "supplier create", method: http.MethodPost, path: "/erp/supplier/v1/suppliers", permission: "erp:suppliers:create"},
		{name: "supplier edit", method: http.MethodPut, path: "/erp/supplier/v1/suppliers/7", permission: "erp:suppliers:edit"},
		{name: "supplier delete", method: http.MethodDelete, path: "/erp/supplier/v1/suppliers/7", permission: "erp:suppliers:edit"},
		{name: "variant view", method: http.MethodGet, path: "/erp/variant/v1/variants", permission: "erp:products:view"},
		{name: "variant create", method: http.MethodPost, path: "/erp/variant/v1/variants", permission: "erp:products:create"},
		{name: "stock take create", method: http.MethodPost, path: "/erp/stock-take/v1/stock-takes", permission: "erp:inventory:adjust"},
		{name: "stock take complete", method: http.MethodPut, path: "/erp/stock-take/v1/stock-takes/7", permission: "erp:inventory:approve"},
		{name: "stock transfer receive", method: http.MethodPatch, path: "/erp/stock-transfer/v1/stock-transfers/7/status", permission: "erp:inventory:transfer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appID, permission, ok := featurePermissionForRequest(tt.method, tt.path)
			if !ok || appID != "erp" || permission != tt.permission {
				t.Fatalf("mapping = %s:%s,%t, want erp:%s,true", appID, permission, ok, tt.permission)
			}
		})
	}
}

func TestAttendanceClockInRoutesAreSelfServiceExempt(t *testing.T) {
	for _, path := range []string{
		"/hr/attendance/clock-in",
		"/hr/attendance/v1/attendance/clock-in",
	} {
		t.Run(path, func(t *testing.T) {
			if appID, permission, ok := featurePermissionForRequest(http.MethodPost, path); ok {
				t.Fatalf("self-service attendance route %q mapped to %q:%q", path, appID, permission)
			}
		})
	}
}

func TestAttendanceClockOutRoutesAreSelfServiceExempt(t *testing.T) {
	for _, path := range []string{
		"/hr/attendance/clock-out",
		"/hr/attendance/v1/attendance/clock-out",
	} {
		t.Run(path, func(t *testing.T) {
			if appID, permission, ok := featurePermissionForRequest(http.MethodPost, path); ok {
				t.Fatalf("self-service attendance route %q mapped to %q:%q", path, appID, permission)
			}
		})
	}
}

func TestAISurfaceRoutesUseTheHubDashboardPermission(t *testing.T) {
	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/caret/tools"},
		{http.MethodPost, "/caret/tools/decision"},
		{http.MethodGet, "/ai/status"},
		{http.MethodPost, "/ai/generate-text-stream"},
		{http.MethodPost, "/hub/caret/v1/tools"},
		{http.MethodGet, "/hub/ai/v1/usage"},
		{http.MethodPut, "/hub/ai/v1/quota-settings"},
	} {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			appID, permission, ok := featurePermissionForRequest(route.method, route.path)
			if !ok || appID != authorization.AppHub || permission != "hub:dashboard:view" {
				t.Fatalf("mapping = %s:%s,%t, want hub:hub:dashboard:view,true", appID, permission, ok)
			}
		})
	}
}

// An AI sub-path with no matrix row must still reach the middleware's
// permission check. Falling through to the unmapped catch-all is what let any
// authenticated tenant member invoke the tool-decision write path.
func TestUnmappedAISurfaceSubPathsStillRequireAuthorization(t *testing.T) {
	t.Setenv("AUTHORIZATION_V2_MODE", "v2")

	member := &authPkg.AuthUser{
		UserID:     7,
		BusinessID: 9,
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID: 9,
			Apps: []authorization.EffectiveAppAccess{{
				AppID:     authorization.AppPOS,
				Licensed:  true,
				HasAccess: true,
			}},
		},
	}

	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/caret"},
		{http.MethodPost, "/caret/tools/decision"},
		{http.MethodPost, "/caret/tools/decision/approve-all"},
		{http.MethodDelete, "/caret/internal/exec"},
		{http.MethodGet, "/ai"},
		{http.MethodPost, "/ai/anything"},
		{http.MethodPost, "/hub/caret/v1/unknown"},
		{http.MethodGet, "/hub/ai/v1/unknown"},
	} {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			appID, permission, ok := featurePermissionForRequest(route.method, route.path)
			if !ok {
				t.Fatalf("%q reached the unmapped catch-all, so the middleware skips authorization", route.path)
			}
			if hasFeaturePermission(member, appID, permission) {
				t.Fatalf("%q authorized a tenant member holding no hub permission", route.path)
			}
		})
	}
}

func TestAISurfaceRoutesAllowAMemberWithTheHubDashboardPermission(t *testing.T) {
	t.Setenv("AUTHORIZATION_V2_MODE", "v2")

	member := &authPkg.AuthUser{
		UserID:     7,
		BusinessID: 9,
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID:      9,
			DashPermissions: []string{"hub:dashboard:view"},
		},
	}

	for _, path := range []string{"/caret/tools/decision", "/ai/generate-text-stream", "/hub/ai/v1/usage"} {
		t.Run(path, func(t *testing.T) {
			appID, permission, ok := featurePermissionForRequest(http.MethodPost, path)
			if !ok || !hasFeaturePermission(member, appID, permission) {
				t.Fatalf("%q denied a member holding hub:dashboard:view", path)
			}
		})
	}
}

func TestAISurfaceMatcherDoesNotClaimNeighbouringPrefixes(t *testing.T) {
	for _, path := range []string{"/aircraft/v1/parts", "/cares/v1/intake", "/hub/air/v1/usage", "/a"} {
		if appID, permission, ok := featurePermissionForRequest(http.MethodGet, path); ok {
			t.Fatalf("%q was claimed by the AI namespace as %s:%s", path, appID, permission)
		}
	}
}
