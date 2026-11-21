package api

import (
	"context"
	"encoding/json"
	"testing"

	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
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

// A single-setting update (the directory toggle) must not have to resend the
// company name, which this table no longer stores; mergeCoreConfigInput keeps
// everything the payload omits.
func TestUpsertHRConfigRequestAcceptsPartialUpdateWithoutCompanyName(t *testing.T) {
	var req UpsertHRConfigRequest
	if err := json.Unmarshal([]byte(`{"directory_enabled":false}`), &req); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if err := (&UpsertHRConfigHandler{}).Validate(context.Background(), &req); err != nil {
		t.Fatalf("Validate rejected a partial update: %v", err)
	}
}

func TestUpsertHRConfigRequestRejectsBlankCompanyName(t *testing.T) {
	var req UpsertHRConfigRequest
	if err := json.Unmarshal([]byte(`{"company_name":"  ","default_currency":"USD"}`), &req); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if err := (&UpsertHRConfigHandler{}).Validate(context.Background(), &req); err == nil {
		t.Fatal("Validate accepted a blank company_name")
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

func TestUpsertHRConfigRequestTracksPresentCoreFields(t *testing.T) {
	var req UpsertHRConfigRequest
	if err := json.Unmarshal([]byte(`{"companyName":"Acme","directoryEnabled":true}`), &req); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !req.coreFieldPresent("company_name") || !req.coreFieldPresent("directory_enabled") {
		t.Fatalf("present fields = %v, want company_name and directory_enabled", req.presentCoreFields)
	}
	for _, field := range []string{
		"fiscal_year_start_month", "default_currency", "default_timezone",
		"default_work_schedule_id", "overtime_enabled", "overtime_multiplier",
	} {
		if req.coreFieldPresent(field) {
			t.Errorf("field %q reported present but was omitted from the payload", field)
		}
	}
}

// A partial HR-core payload must not zero the settings the caller omitted:
// previously every column was written from the request struct, so
// {"company_name":...,"directory_enabled":true} wiped currency, timezone,
// fiscal-year month, work schedule and overtime settings.
func TestMergeCoreConfigInputKeepsOmittedFields(t *testing.T) {
	current := &hrcore.Config{
		CompanyName:           "Acme",
		FiscalYearStartMonth:  4,
		DefaultCurrency:       "EUR",
		DefaultTimezone:       "Europe/Berlin",
		DefaultWorkScheduleID: 77,
		OvertimeEnabled:       true,
		OvertimeMultiplier:    2,
		DirectoryEnabled:      false,
	}
	var req UpsertHRConfigRequest
	if err := json.Unmarshal([]byte(`{"company_name":"Acme Renamed","directory_enabled":true}`), &req); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	got := mergeCoreConfigInput(current, &req)

	if got.CompanyName != "Acme Renamed" {
		t.Errorf("CompanyName = %q, want the provided value", got.CompanyName)
	}
	if !got.DirectoryEnabled {
		t.Error("DirectoryEnabled = false, want the provided true")
	}
	if got.DefaultCurrency != "EUR" || got.DefaultTimezone != "Europe/Berlin" {
		t.Errorf("currency/timezone = %q/%q, want the stored values", got.DefaultCurrency, got.DefaultTimezone)
	}
	if got.FiscalYearStartMonth != 4 {
		t.Errorf("FiscalYearStartMonth = %d, want the stored 4", got.FiscalYearStartMonth)
	}
	if got.DefaultWorkScheduleID != 77 {
		t.Errorf("DefaultWorkScheduleID = %d, want the stored 77", got.DefaultWorkScheduleID)
	}
	if !got.OvertimeEnabled || got.OvertimeMultiplier != 2 {
		t.Errorf("overtime = %v/%v, want the stored true/2", got.OvertimeEnabled, got.OvertimeMultiplier)
	}
}

// A first write with no stored config falls back to the column defaults rather
// than zero values.
func TestMergeCoreConfigInputSeedsSchemaDefaults(t *testing.T) {
	var req UpsertHRConfigRequest
	if err := json.Unmarshal([]byte(`{"company_name":"Acme"}`), &req); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	got := mergeCoreConfigInput(nil, &req)

	if got.DefaultCurrency != "USD" || got.DefaultTimezone != "UTC" || got.FiscalYearStartMonth != 1 || got.OvertimeMultiplier != 1.5 {
		t.Fatalf("defaults not applied: %+v", got)
	}
}

// The HR Settings toggle sends only directory_enabled; turning it off must
// stick rather than merge back to the stored true.
func TestMergeCoreConfigInputTurnsDirectoryOff(t *testing.T) {
	var req UpsertHRConfigRequest
	if err := json.Unmarshal([]byte(`{"directory_enabled":false}`), &req); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	got := mergeCoreConfigInput(&hrcore.Config{CompanyName: "Acme", DirectoryEnabled: true}, &req)

	if got.DirectoryEnabled {
		t.Fatal("an explicit directory_enabled=false must turn the directory off")
	}
}

// The HR setup wizard's Work Schedule step receives schedule.id as a string
// from the OneKit client and PUTs {defaultWorkScheduleId: "9"} to /hr/config.
func TestUpsertHRConfigRequestParsesStringWorkScheduleID(t *testing.T) {
	for _, raw := range []string{
		`{"defaultWorkScheduleId":"9"}`,
		`{"default_work_schedule_id":"14"}`,
		`{"defaultWorkScheduleId":21}`,
	} {
		var req UpsertHRConfigRequest
		if err := json.Unmarshal([]byte(raw), &req); err != nil {
			t.Fatalf("Unmarshal(%s): %v", raw, err)
		}
		if !req.coreFieldPresent("default_work_schedule_id") || req.DefaultWorkScheduleID <= 0 {
			t.Fatalf("Unmarshal(%s) got id=%d present=%v", raw, req.DefaultWorkScheduleID, req.coreFieldPresent("default_work_schedule_id"))
		}
	}
}
