package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/settings/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestSettingsPreserveStoredFeatureOverrides(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const businessID int64 = 948231
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM pos_business_settings WHERE business_id = ?", businessID) })
	if _, err := db.ExecContext(ctx, `INSERT INTO pos_business_settings
		(business_id, business_profile, theme, features, receipt_footer)
		VALUES (?, 'restaurant', 'dark', ?, 'Saved footer')`, businessID,
		`{"kitchen_display":false,"gift_cards":false,"cash_drawer":false}`); err != nil {
		t.Fatal(err)
	}
	repo := infrastructure.NewRepository(db)
	settings, err := repo.GetSettings(ctx, businessID)
	if err != nil {
		t.Fatal(err)
	}
	if settings.Features.KitchenDisplay || settings.Features.GiftCards || settings.Features.CashDrawer || !settings.Features.Tables || !settings.Features.Inventory {
		t.Fatalf("stored overrides or omitted defaults were lost: %+v", settings.Features)
	}
	settings.Features.KitchenDisplay = true
	settings.Features.Tables = false
	if err := repo.UpsertSettings(ctx, settings); err != nil {
		t.Fatal(err)
	}
	loaded, err := repo.GetSettings(ctx, businessID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Features != settings.Features || loaded.ReceiptFooter != settings.ReceiptFooter {
		t.Fatalf("feature settings did not survive persistence: %+v", loaded)
	}
	var kitchenDisplay, tables bool
	if err := db.QueryRowContext(ctx, `SELECT JSON_EXTRACT(features, '$.kitchen_display'), JSON_EXTRACT(features, '$.tables')
		FROM pos_business_settings WHERE business_id = ?`, businessID).Scan(&kitchenDisplay, &tables); err != nil {
		t.Fatal(err)
	}
	if !kitchenDisplay || tables {
		t.Fatal("stored JSON no longer uses the existing feature keys")
	}
}
