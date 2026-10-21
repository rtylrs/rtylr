package http

import (
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

func TestExtractToken(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"bare token", "abcdef", "abcdef"},
		{"bearer token", "Bearer abcdef", "abcdef"},
		{"bearer token with extra spaces preserved", "Bearer  abc", " abc"},
		{"lowercase bearer is treated as bare", "bearer abcdef", "bearer abcdef"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractToken(tc.in); got != tc.want {
				t.Fatalf("extractToken(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestAuthSuccessLogsDisabledByDefault(t *testing.T) {
	t.Setenv("AUTH_SUCCESS_LOGS", "")
	t.Setenv("AUTH_REQUEST_LOGS", "")

	if authSuccessLogsEnabled() {
		t.Fatal("expected auth success logs to be disabled")
	}
}

func TestAuthSuccessLogsCanBeEnabled(t *testing.T) {
	t.Setenv("AUTH_SUCCESS_LOGS", "true")
	t.Setenv("AUTH_REQUEST_LOGS", "")

	if !authSuccessLogsEnabled() {
		t.Fatal("expected auth success logs to be enabled")
	}
}

func TestAppForPathCoversLicensedApplicationPrefixes(t *testing.T) {
	cases := []struct {
		path string
		want authorization.AppID
	}{
		{path: "/pos/orders/v1/orders", want: authorization.AppPOS},
		{path: "/pos/menu-item/v1/catalog", want: authorization.AppPOS},
		{path: "/pos/menu-item/v1/menu-items", want: authorization.AppMenu},
		{path: "/pos/menu-item/v1/menu-items/42", want: authorization.AppMenu},
		{path: "/erp/inventory/v1/items", want: authorization.AppERP},
		{path: "/hr/employees/v1/staff", want: authorization.AppHR},
		{path: "/crm/customers/v1/customers", want: authorization.AppCRM},
		{path: "/finance/invoices/v1/invoices", want: authorization.AppFinance},
		{path: "/flow/project/v1/projects", want: authorization.AppFlow},
		{path: "/insights/revenue/v1/revenue", want: authorization.AppInsights},
		{path: "/menu/catalog/v1/items", want: authorization.AppMenu},
		{path: "/hub/menu/v1/menu", want: authorization.AppMenu},
		{path: "/recruit/careers/v1/jobs", want: authorization.AppRecruit},
	}
	for _, tc := range cases {
		t.Run(string(tc.want), func(t *testing.T) {
			got, ok := appForPath(tc.path)
			if !ok || got != tc.want {
				t.Fatalf("appForPath(%q) = (%q, %v), want (%q, true)", tc.path, got, ok, tc.want)
			}
		})
	}
}

func TestAppForPathDoesNotTreatHubOrCommonRoutesAsAppRoutes(t *testing.T) {
	for _, path := range []string{"/hub/user/v1/users/me", "/common/auth/v1/verify-email", "/health"} {
		if got, ok := appForPath(path); ok {
			t.Fatalf("appForPath(%q) = (%q, true), want no app", path, got)
		}
	}
}

func TestFeaturePermissionForRequestUsesCanonicalMatrixPermissions(t *testing.T) {
	cases := []struct {
		method     string
		path       string
		app        authorization.AppID
		permission string
	}{
		{method: "GET", path: "/finance/invoices/v1/invoices", app: authorization.AppFinance, permission: "finance:invoices:view"},
		{method: "POST", path: "/finance/invoices/v1/invoices/12/send", app: authorization.AppFinance, permission: "finance:invoices:send"},
		{method: "GET", path: "/finance/quote/v1/quotes", app: authorization.AppFinance, permission: "finance:quotes:view"},
		{method: "POST", path: "/finance/quote/v1/quotes/12/convert", app: authorization.AppFinance, permission: "finance:quotes:manage"},
		{method: "GET", path: "/finance/recurring-invoice/v1/recurring-invoices", app: authorization.AppFinance, permission: "finance:recurring_invoices:view"},
		{method: "POST", path: "/finance/recurring-invoice/v1/recurring-invoices/12/pause", app: authorization.AppFinance, permission: "finance:recurring_invoices:manage"},
		{method: "GET", path: "/crm/sales/contacts", app: authorization.AppCRM, permission: "crm:customers:view"},
		{method: "DELETE", path: "/erp/product/v1/products/12", app: authorization.AppERP, permission: "erp:products:delete"},
		{method: "GET", path: "/hr/attendance/v1/attendance", app: authorization.AppHR, permission: "hr:employees:view"},
		{method: "POST", path: "/insights/report/v1/reports", app: authorization.AppInsights, permission: "insights:reports:create"},
		{method: "GET", path: "/hub/team/v1/members", app: authorization.AppHub, permission: "hub:team:view"},
		{method: "POST", path: "/pos/orders/v1/orders/12/void/approve", app: authorization.AppPOS, permission: "pos:orders:void:approve"},
		{method: "POST", path: "/flow/task/v2/tasks", app: authorization.AppFlow, permission: "flow:tasks:create"},
		{method: "GET", path: "/hub/menu/v1/menu", app: authorization.AppMenu, permission: "menu:menus:view"},
		{method: "GET", path: "/recruit/submissions/resume-url", app: authorization.AppRecruit, permission: "recruit:candidates:sensitive_view"},
	}
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			app, permission, ok := featurePermissionForRequest(tc.method, tc.path)
			if !ok || app != tc.app || permission != tc.permission {
				t.Fatalf("featurePermissionForRequest(%q, %q) = (%q, %q, %v), want (%q, %q, true)", tc.method, tc.path, app, permission, ok, tc.app, tc.permission)
			}
		})
	}
}

func TestFeaturePermissionForRequestDoesNotGuessUnmappedRoutes(t *testing.T) {
	if app, permission, ok := featurePermissionForRequest("GET", "/menu/public/bistro"); ok {
		t.Fatalf("unmapped route returned (%q, %q, true)", app, permission)
	}
}

func TestHubFeaturePermissionHonorsResolvedOwnerAndAdminAccess(t *testing.T) {
	cases := []struct {
		name string
		user *authPkg.AuthUser
		want bool
	}{
		{
			name: "business owner",
			user: &authPkg.AuthUser{UserID: 1, BusinessID: 1, EffectiveAccess: &authorization.EffectiveAccess{IsOwner: true}},
			want: true,
		},
		{
			name: "business admin",
			user: &authPkg.AuthUser{UserID: 2, BusinessID: 1, EffectiveAccess: &authorization.EffectiveAccess{IsAdmin: true}},
			want: true,
		},
		{
			name: "member without permission",
			user: &authPkg.AuthUser{UserID: 3, BusinessID: 1, EffectiveAccess: &authorization.EffectiveAccess{}},
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasFeaturePermission(tc.user, authorization.AppHub, "hub:team:invite"); got != tc.want {
				t.Fatalf("hasFeaturePermission() = %v, want %v", got, tc.want)
			}
		})
	}
}
