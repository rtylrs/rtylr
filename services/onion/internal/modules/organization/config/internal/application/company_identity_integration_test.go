package application

import (
	"context"
	"strconv"
	"testing"

	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	organizationbusiness "github.com/voxire/onion/internal/modules/organization/business"
	"github.com/voxire/onion/internal/modules/organization/config/internal/domain"
	moduleinfra "github.com/voxire/onion/internal/modules/organization/config/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/testutil"
)

type noopTestEmailQueue struct{}

func (noopTestEmailQueue) Enqueue(context.Context, TestEmail) error { return nil }

func TestCompanyIdentityRenameRefreshesHRWithoutCrossingBusinesses(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	if err := cacher.InitializeLocalCache(ctx); err != nil {
		t.Fatal(err)
	}
	service := NewService(moduleinfra.NewRepository(db), noopTestEmailQueue{})
	hr := hrcore.NewModule(ctx, db).WithBusinessIdentityResolver(organizationbusiness.NewModule(ctx, db))
	businesses := []int64{97101, 97102}
	for _, id := range businesses {
		testutil.GrantBusinessAccess(t, db, id, testutil.OwnerID)
		t.Cleanup(func() {
			for _, table := range []string{"hr_company_config", "config"} {
				if _, err := db.ExecContext(ctx, "DELETE FROM "+table+" WHERE business_id = ?", id); err != nil {
					t.Error(err)
				}
			}
			cacher.InvalidateScope(ctx, cacheKeys.BusinessPrefix(id))
		})
		if err := service.CreateConfig(ctx, domain.ConfigInput{OwnerId: strconv.FormatInt(id, 10), SiteName: "Shared " + strconv.FormatInt(id, 10)}); err != nil {
			t.Fatal(err)
		}
		if _, err := hr.UpsertConfig(ctx, id, hrcore.ConfigInput{CompanyName: "Stale HR name", DefaultCurrency: "EUR", DefaultTimezone: "Europe/Berlin", FiscalYearStartMonth: 4, OvertimeMultiplier: 2}); err != nil {
			t.Fatal(err)
		}
		cfg, err := hr.GetConfig(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.CompanyName != "Shared "+strconv.FormatInt(id, 10) {
			t.Fatalf("HR failed to inherit shared identity: %+v", cfg)
		}
	}
	if err := service.UpdateConfig(ctx, "97101", domain.ConfigInput{SiteName: "Renamed company"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := hr.GetConfig(ctx, 97101)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.CompanyName != "Renamed company" {
		t.Fatalf("warm HR cache kept old identity: %+v", cfg)
	}
	if cfg.DefaultCurrency != "EUR" || cfg.DefaultTimezone != "Europe/Berlin" || cfg.FiscalYearStartMonth != 4 || cfg.OvertimeMultiplier != 2 {
		t.Fatalf("rename changed HR-specific settings: %+v", cfg)
	}
	other, err := hr.GetConfig(ctx, 97102)
	if err != nil {
		t.Fatal(err)
	}
	if other.CompanyName != "Shared 97102" {
		t.Fatalf("rename crossed business boundary: %+v", other)
	}
	if err := service.UpdateConfig(ctx, "97101", domain.ConfigInput{SiteName: ""}); err != nil {
		t.Fatal(err)
	}
	cfg, err = hr.GetConfig(ctx, 97101)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.CompanyName != "Integration Test Business 97101" {
		t.Fatalf("missing business name fallback: %+v", cfg)
	}
}
