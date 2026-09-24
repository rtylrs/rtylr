package http

import (
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/platform/authorization"
)

// featurePermissionForRequest maps the HTTP surface to the canonical feature
// permission that protects it. App access remains a separate check in the
// authentication middleware. A route is left unmapped when the matrix does
// not define a feature permission for that surface, rather than guessing a
// broader permission.
func featurePermissionForRequest(method, path string) (authorization.AppID, string, bool) {
	path = strings.ToLower(strings.TrimRight(path, "/"))
	method = strings.ToUpper(method)

	if permission, ok := hubFeaturePermission(method, path); ok {
		return authorization.AppHub, permission, true
	}
	if permission, ok := financeFeaturePermission(method, path); ok {
		return authorization.AppFinance, permission, true
	}
	if permission, ok := crmFeaturePermission(method, path); ok {
		return authorization.AppCRM, permission, true
	}
	if permission, ok := erpFeaturePermission(method, path); ok {
		return authorization.AppERP, permission, true
	}
	if permission, ok := hrFeaturePermission(method, path); ok {
		return authorization.AppHR, permission, true
	}
	if permission, ok := insightsFeaturePermission(method, path); ok {
		return authorization.AppInsights, permission, true
	}
	if permission, ok := posFeaturePermission(method, path); ok {
		return authorization.AppPOS, permission, true
	}
	if permission, ok := flowFeaturePermission(method, path); ok {
		return authorization.AppFlow, permission, true
	}
	if permission, ok := menuFeaturePermission(method, path); ok {
		return authorization.AppMenu, permission, true
	}
	if permission, ok := recruitFeaturePermission(method, path); ok {
		return authorization.AppRecruit, permission, true
	}
	return "", "", false
}

func hubFeaturePermission(method, path string) (string, bool) {
	switch {
	case strings.HasPrefix(path, "/hub/dashboard"):
		return "hub:dashboard:view", true
	case strings.HasPrefix(path, "/hub/team"):
		if strings.Contains(path, "/invite") {
			return "hub:team:invite", true
		}
		if strings.Contains(path, "/deactivate") {
			return "hub:team:deactivate", true
		}
		if method == http.MethodGet {
			return "hub:team:view", true
		}
		return "hub:team:edit", true
	case strings.HasPrefix(path, "/hub/roles"), strings.HasPrefix(path, "/hub/permission"):
		if method == http.MethodGet {
			return "hub:roles:view", true
		}
		if method == http.MethodPost {
			return "hub:roles:create", true
		}
		if method == http.MethodDelete {
			return "hub:roles:archive", true
		}
		return "hub:roles:edit", true
	case strings.HasPrefix(path, "/hub/licenses"):
		return "hub:licenses:manage", true
	case strings.HasPrefix(path, "/hub/groups"):
		return "hub:groups:manage", true
	case strings.HasPrefix(path, "/hub/hierarchy"):
		return "hub:hierarchy:manage", true
	case strings.HasPrefix(path, "/hub/audit"):
		return "hub:audit:view", true
	case strings.HasPrefix(path, "/hub/security"):
		return "hub:security:manage", true
	case strings.HasPrefix(path, "/hub/api-keys"), strings.HasPrefix(path, "/hub/api_keys"):
		return "hub:api_keys:manage", true
	}
	return "", false
}

func financeFeaturePermission(method, path string) (string, bool) {
	switch {
	case strings.HasPrefix(path, "/finance/invoices"):
		if strings.HasSuffix(path, "/send") {
			return "finance:invoices:send", true
		}
		if method == http.MethodGet {
			return "finance:invoices:view", true
		}
		if method == http.MethodPost {
			return "finance:invoices:create", true
		}
	case strings.HasPrefix(path, "/finance/expenses"):
		if strings.Contains(path, "/approve") {
			return "finance:expenses:approve", true
		}
		if method == http.MethodGet {
			return "finance:expenses:view", true
		}
		if method == http.MethodPost {
			return "finance:expenses:create", true
		}
	case strings.HasPrefix(path, "/finance/bank"):
		if strings.Contains(path, "/sensitive") {
			return "finance:bank_feeds:sensitive_view", true
		}
		if method == http.MethodGet {
			return "finance:bank_feeds:view", true
		}
	case strings.HasPrefix(path, "/finance/reconciliation"):
		return "finance:reconciliation:manage", true
	case strings.HasPrefix(path, "/finance/report"), strings.HasPrefix(path, "/finance/dashboard"):
		if strings.HasSuffix(path, "/export") {
			return "finance:reports:export", true
		}
		if method == http.MethodGet {
			return "finance:reports:view", true
		}
	case strings.HasPrefix(path, "/finance/tax"):
		return "finance:tax:configure", true
	}
	return "", false
}

func crmFeaturePermission(method, path string) (string, bool) {
	switch {
	case strings.Contains(path, "/crm/campaign"):
		if strings.HasSuffix(path, "/send") {
			return "crm:campaigns:send", true
		}
		if method == http.MethodGet {
			return "crm:campaigns:view", true
		}
		if method == http.MethodPost {
			return "crm:campaigns:create", true
		}
	case strings.Contains(path, "/crm/loyalty"):
		return "crm:loyalty:manage", true
	case strings.Contains(path, "/crm/automation"):
		return "crm:automations:manage", true
	case strings.HasSuffix(path, "/crm/customers/export"):
		return "crm:customers:export", true
	case strings.Contains(path, "/crm/segment"), strings.Contains(path, "/crm/customer"), strings.Contains(path, "/crm/sales"):
		if method == http.MethodGet {
			return "crm:customers:view", true
		}
		if method == http.MethodPost {
			return "crm:customers:create", true
		}
		if method == http.MethodPut || method == http.MethodPatch {
			return "crm:customers:edit", true
		}
		if method == http.MethodDelete {
			return "crm:customers:delete", true
		}
	}
	return "", false
}

func erpFeaturePermission(method, path string) (string, bool) {
	switch {
	case strings.Contains(path, "/erp/product"):
		if method == http.MethodGet {
			return "erp:products:view", true
		}
		if method == http.MethodPost {
			return "erp:products:create", true
		}
		if method == http.MethodPut || method == http.MethodPatch {
			return "erp:products:edit", true
		}
		if method == http.MethodDelete {
			return "erp:products:delete", true
		}
	case strings.Contains(path, "/erp/inventory"):
		if strings.Contains(path, "/transfer") {
			return "erp:inventory:transfer", true
		}
		if strings.Contains(path, "/approve") {
			return "erp:inventory:approve", true
		}
		if method != http.MethodGet {
			return "erp:inventory:adjust", true
		}
		return "erp:inventory:view", true
	case strings.Contains(path, "/erp/warehouse"):
		return "erp:warehouses:manage", true
	case strings.Contains(path, "/erp/purchas"):
		if strings.Contains(path, "/approve") {
			return "erp:purchasing:approve", true
		}
		if method == http.MethodGet {
			return "erp:purchasing:view", true
		}
		return "erp:purchasing:create", true
	}
	return "", false
}

func isHRPolicyExemptPath(method, path string) bool {
	// Attendance clock-in/clock-out and availability are self-scoped by their
	// handlers. Keep these routes unmapped here rather than requiring a broad
	// employee/schedule permission before the handler can resolve the caller.
	if method == http.MethodPost && strings.HasPrefix(path, "/hr/attendance/") &&
		(strings.HasSuffix(path, "/clock-in") || strings.HasSuffix(path, "/clock-out")) {
		return true
	}
	if path == "/hr/shifts/availability" || strings.HasPrefix(path, "/hr/shifts/availability/") ||
		path == "/hr/shift/v1/availability" || strings.HasPrefix(path, "/hr/shift/v1/availability/") {
		return true
	}
	return method == http.MethodGet &&
		(path == "/hr/payroll/my-runs" || strings.HasPrefix(path, "/hr/payroll/my-runs/"))
}

func isPayrollExecutePath(path string) bool {
	return strings.HasSuffix(path, "/run") ||
		strings.HasSuffix(path, "/execute") ||
		strings.HasSuffix(path, "/runs/trigger")
}

func hrFeaturePermission(method, path string) (string, bool) {
	switch {
	case isHRPolicyExemptPath(method, path):
		return "", false
	case strings.Contains(path, "/hr/payroll"):
		if isPayrollExecutePath(path) {
			return "hr:payroll:execute", true
		}
		if method == http.MethodGet {
			return "hr:payroll:view", true
		}
		return "hr:payroll:manage", true
	case strings.Contains(path, "/hr/time-off"), strings.Contains(path, "/hr/time_off"), strings.Contains(path, "/hr/leave"):
		if strings.Contains(path, "/approve") {
			return "hr:time_off:approve", true
		}
		return "hr:time_off:request", true
	case strings.Contains(path, "/hr/schedule"), strings.Contains(path, "/hr/shift"):
		return "hr:schedules:manage", true
	case strings.Contains(path, "/hr/role"):
		return "hr:roles:manage", true
	case strings.Contains(path, "/hr/employee"), strings.Contains(path, "/hr/staff"), strings.Contains(path, "/hr/attendance"):
		if strings.Contains(path, "/sensitive") {
			return "hr:employees:sensitive_view", true
		}
		if method == http.MethodGet {
			return "hr:employees:view", true
		}
		return "hr:employees:manage", true
	}
	return "", false
}

func insightsFeaturePermission(method, path string) (string, bool) {
	switch {
	case strings.Contains(path, "/insights/report"):
		if strings.HasSuffix(path, "/export") {
			return "insights:reports:export", true
		}
		if strings.HasSuffix(path, "/share") {
			return "insights:reports:share", true
		}
		if method == http.MethodGet {
			return "insights:dashboard:view", true
		}
		return "insights:reports:create", true
	case strings.Contains(path, "/insights/financial"):
		return "insights:financial:view", true
	case strings.Contains(path, "/insights/customers"):
		return "insights:customers:view", true
	case strings.Contains(path, "/insights/dashboard"):
		return "insights:dashboard:view", true
	}
	return "", false
}

func posFeaturePermission(method, path string) (string, bool) {
	// The POS matcher is intentionally anchored to the POS namespace. Broad
	// substrings such as "/employees" and "/reports" otherwise classify
	// unmapped HR/Finance routes as POS features and deny legitimate users.
	if path != "/pos" && !strings.HasPrefix(path, "/pos/") {
		return "", false
	}

	switch {
	case strings.Contains(path, "/void"):
		if strings.Contains(path, "/approve") {
			return "pos:orders:void:approve", true
		}
		if strings.Contains(path, "/override") {
			return "pos:orders:void:override", true
		}
		return "pos:orders:void", true
	case strings.Contains(path, "/refund"):
		if strings.Contains(path, "/approve") {
			return "pos:orders:refund:approve", true
		}
		if strings.Contains(path, "/override") {
			return "pos:orders:refund:override", true
		}
		return "pos:orders:refund", true
	case strings.Contains(path, "/discount"):
		return "pos:orders:discount", true
	case strings.Contains(path, "/orders"):
		if method == http.MethodGet {
			return "pos:orders:view", true
		}
		return "pos:orders:create", true
	case strings.Contains(path, "/cash"):
		if strings.Contains(path, "/drop") {
			return "pos:cash:drop", true
		}
		if strings.Contains(path, "/pickup") {
			return "pos:cash:pickup", true
		}
		if strings.Contains(path, "/approve") {
			return "pos:cash:approve", true
		}
		return "pos:cash:view", true
	case strings.Contains(path, "/employees"):
		return "pos:employees:manage", true
	case strings.Contains(path, "/terminals"):
		return "pos:terminals:manage", true
	case strings.Contains(path, "/reports"), strings.Contains(path, "/audit"):
		return "pos:reports:view", true
	}
	return "", false
}

// Flow's only /settings routes are GET and PUT /flow/setting/v1/settings, which
// read and write one row keyed on the caller's own user id: notification
// preferences and the sidebar layout. Nothing there is business-wide, so they
// are deliberately absent from this table. Gating them on
// flow:settings:manage stopped an ordinary member from turning off their own
// email notifications, and the handlers' own GetAuthUser scoping is what keeps
// a caller to their own row. Business-wide Flow settings, if any are ever
// added, need their own path rather than this substring.
func flowFeaturePermission(method, path string) (string, bool) {
	switch {
	case strings.Contains(path, "/managers"):
		return "flow:managers:manage", true
	case strings.Contains(path, "/team"):
		return "flow:team:manage", true
	case strings.Contains(path, "/tasks"):
		if strings.Contains(path, "/assign") {
			return "flow:tasks:assign", true
		}
		if method == http.MethodGet {
			return "flow:tasks:view", true
		}
		if method == http.MethodPost {
			return "flow:tasks:create", true
		}
		return "flow:tasks:edit", true
	case strings.Contains(path, "/projects"):
		if strings.Contains(path, "/assign") || strings.Contains(path, "/members") {
			return "flow:projects:assign", true
		}
		if method == http.MethodGet {
			return "flow:projects:view", true
		}
		if method == http.MethodPost {
			return "flow:projects:create", true
		}
		return "flow:projects:edit", true
	}
	return "", false
}

func menuFeaturePermission(method, path string) (string, bool) {
	if !strings.HasPrefix(path, "/menu/") && !strings.HasPrefix(path, "/hub/menu/") {
		return "", false
	}
	if !strings.HasPrefix(path, "/hub/menu/") &&
		!strings.Contains(path, "/menus") &&
		!strings.Contains(path, "/branding") &&
		!strings.Contains(path, "/orders") &&
		!strings.Contains(path, "/analytics") {
		return "", false
	}
	switch {
	case strings.Contains(path, "/branding"):
		return "menu:branding:edit", true
	case strings.Contains(path, "/orders"):
		return "menu:orders:view", true
	case strings.Contains(path, "/analytics"):
		return "menu:analytics:view", true
	case method == http.MethodGet:
		return "menu:menus:view", true
	default:
		return "menu:menus:edit", true
	}
}

func recruitFeaturePermission(method, path string) (string, bool) {
	switch {
	case strings.Contains(path, "/submissions/resume-url") || strings.Contains(path, "/candidates/sensitive"):
		return "recruit:candidates:sensitive_view", true
	case strings.Contains(path, "/submissions") || strings.Contains(path, "/candidates"):
		if strings.Contains(path, "/export") {
			return "recruit:candidates:export", true
		}
		if method == http.MethodGet {
			return "recruit:candidates:view", true
		}
		return "recruit:candidates:edit", true
	case strings.Contains(path, "/careers") || strings.Contains(path, "/jobs"):
		if strings.Contains(path, "/publish") {
			return "recruit:jobs:publish", true
		}
		if method == http.MethodGet {
			return "recruit:jobs:view", true
		}
		return "recruit:jobs:manage", true
	}
	return "", false
}
