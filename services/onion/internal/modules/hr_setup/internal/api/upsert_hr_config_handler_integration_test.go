package api

import (
	"context"
	"encoding/json"
	"testing"

	hrcore "github.com/voxire/onion/internal/modules/hr_core"
	"github.com/voxire/onion/internal/modules/hr_setup/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

// TestUpsertHRConfigPartialCorePayloadKeepsOmittedFields drives the real
// handler end to end: a configured hr_company_config must survive a PUT that
// only carries company_name + directory_enabled. Before the merge fix every
// column was written from the request struct, so the omitted currency,
// timezone, fiscal-year month, work schedule and overtime settings were wiped.
func TestUpsertHRConfigPartialCorePayloadKeepsOmittedFields(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	coreModule := hrcore.NewModule(ctx, db)
	if _, err := coreModule.UpsertConfig(ctx, testutil.BusinessID, hrcore.ConfigInput{
		CompanyName:           "Acme",
		FiscalYearStartMonth:  4,
		DefaultCurrency:       "EUR",
		DefaultTimezone:       "Europe/Berlin",
		DefaultWorkScheduleID: 0,
		OvertimeEnabled:       true,
		OvertimeMultiplier:    2,
		DirectoryEnabled:      false,
	}); err != nil {
		t.Fatalf("seed core config: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), "DELETE FROM hr_company_config WHERE business_id = ?", testutil.BusinessID); err != nil {
			t.Errorf("cleanup hr_company_config: %v", err)
		}
		if _, err := db.ExecContext(context.Background(), "DELETE FROM hr_configs WHERE owner_id = ?", testutil.OwnerID); err != nil {
			t.Errorf("cleanup hr_configs: %v", err)
		}
	})

	handler := &UpsertHRConfigHandler{
		Service:     application.NewService(ctx, db),
		CoreService: coreModule,
	}
	var req UpsertHRConfigRequest
	if err := json.Unmarshal([]byte(`{"company_name":"Acme Renamed","directory_enabled":true}`), &req); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if err := handler.Validate(ctx, &req); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	apiCtx := authPkg.WithAuthUser(ctx, &authPkg.AuthUser{
		UserID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		EffectiveAccess: &authorization.EffectiveAccess{
			IsOwner: true,
			Apps:    []authorization.EffectiveAppAccess{{AppID: authorization.AppHR, Licensed: true, HasAccess: true}},
		},
	})
	resp, err := handler.Serve(apiCtx, &req)
	if err != nil {
		t.Fatalf("Serve: %v", err)
	}
	if resp.Config == nil {
		t.Fatal("Serve returned no config")
	}
	if resp.Config.CompanyName != "Acme Renamed" || !resp.Config.DirectoryEnabled {
		t.Fatalf("provided fields not applied: %+v", resp.Config)
	}
	if resp.Config.DefaultCurrency != "EUR" || resp.Config.DefaultTimezone != "Europe/Berlin" {
		t.Fatalf("omitted currency/timezone were overwritten: %+v", resp.Config)
	}
	if resp.Config.FiscalYearStartMonth != 4 {
		t.Fatalf("omitted fiscal_year_start_month was overwritten: %+v", resp.Config)
	}
	if !resp.Config.OvertimeEnabled || resp.Config.OvertimeMultiplier != 2 {
		t.Fatalf("omitted overtime settings were overwritten: %+v", resp.Config)
	}

	stored, err := coreModule.GetConfig(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("GetConfig after Serve: %v", err)
	}
	if stored.DefaultCurrency != "EUR" || stored.DefaultTimezone != "Europe/Berlin" || stored.FiscalYearStartMonth != 4 {
		t.Fatalf("stored config lost omitted fields: %+v", stored)
	}
}
