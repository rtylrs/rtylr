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
