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
