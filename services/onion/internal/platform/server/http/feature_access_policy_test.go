package http

import (
	"net/http"
	"testing"
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
