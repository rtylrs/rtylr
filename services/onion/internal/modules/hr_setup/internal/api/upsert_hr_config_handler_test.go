package api

import (
	"context"
	"encoding/json"
	"testing"
)

func TestUpsertHRConfigRequestAcceptsFrontendSetupPayload(t *testing.T) {
	var req UpsertHRConfigRequest
	if err := json.Unmarshal([]byte(`{"setupCompletedAt":"1755439200"}`), &req); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if req.SetupCompletedAt == nil || *req.SetupCompletedAt != 1755439200 {
		t.Fatalf("SetupCompletedAt = %v, want 1755439200", req.SetupCompletedAt)
	}
	if req.hasCoreConfigFields {
		t.Fatal("setup-only payload was classified as an HR-core config update")
	}
	if err := (&UpsertHRConfigHandler{}).Validate(context.Background(), &req); err != nil {
		t.Fatalf("Validate setup-only payload: %v", err)
	}
}

func TestUpsertHRConfigRequestAcceptsSnakeCaseCorePayload(t *testing.T) {
	var req UpsertHRConfigRequest
	if err := json.Unmarshal([]byte(`{"company_name":"Acme","default_currency":"USD","directory_enabled":true}`), &req); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !req.hasCoreConfigFields || req.CompanyName != "Acme" || req.DefaultCurrency != "USD" || !req.DirectoryEnabled {
		t.Fatalf("decoded request = %+v", req)
	}
}

func TestUpsertHRConfigRequestRejectsCoreUpdateWithoutCompanyName(t *testing.T) {
	var req UpsertHRConfigRequest
	if err := json.Unmarshal([]byte(`{"default_currency":"USD"}`), &req); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if err := (&UpsertHRConfigHandler{}).Validate(context.Background(), &req); err == nil {
		t.Fatal("Validate accepted a core config update without company_name")
	}
}

func TestSetupConfigRoutesStayOnCompatibilityPath(t *testing.T) {
	routes := GetHRConfigRoutes()
	if len(routes) != 1 || routes[0] != "/hr/config" {
		t.Fatalf("setup config routes = %v, want [/hr/config]", routes)
	}
}

// Kept from the pre-existing suite: older clients still send the snake_case
// field as a JSON number rather than the setup wizard's camelCase string.
func TestUpsertHRConfigRequestAcceptsLegacyPayload(t *testing.T) {
	var req UpsertHRConfigRequest
	if err := json.Unmarshal([]byte(`{"setup_completed_at":1755439200}`), &req); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if req.SetupCompletedAt == nil || *req.SetupCompletedAt != 1755439200 {
		t.Fatalf("SetupCompletedAt = %v, want 1755439200", req.SetupCompletedAt)
	}
	if req.hasCoreConfigFields {
		t.Fatal("setup-only payload was classified as an HR-core config update")
	}
}

func TestUpsertHRConfigRequestTreatsExplicitNullAsAbsent(t *testing.T) {
	var req UpsertHRConfigRequest
	if err := json.Unmarshal([]byte(`{"setup_completed_at":null}`), &req); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if req.SetupCompletedAt != nil {
		t.Fatalf("SetupCompletedAt = %v, want nil", req.SetupCompletedAt)
	}
}
