package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/erp_settings/internal/domain"
	"github.com/voxire/onion/internal/modules/erp_settings/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestErpSettingsRepositoryRoundTripPreservesFalseFlags(t *testing.T) {
	db := testutil.OpenDB(t)
	const businessID int64 = 987654321

	_, _ = db.ExecContext(context.Background(), "DELETE FROM erp_business_settings WHERE business_id = ?", businessID)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM erp_business_settings WHERE business_id = ?", businessID)
	})

	repo := infrastructure.NewRepository(db)
	input := domain.SettingsInput{
		BusinessID:         businessID,
		Currency:           "EUR",
		TaxRate:            20,
		LowStockThreshold:  4,
		DefaultWarehouseID: 7,
		NotifyOnLowStock:   false,
		NotifyOnExpiry:     false,
	}
	if err := repo.Upsert(context.Background(), input); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, err := repo.Get(context.Background(), businessID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.BusinessID != businessID || got.Currency != input.Currency || got.TaxRate != input.TaxRate || got.LowStockThreshold != input.LowStockThreshold || got.DefaultWarehouseID != input.DefaultWarehouseID {
		t.Fatalf("settings did not round trip: got %+v, want %+v", got, input)
	}
	if got.NotifyOnLowStock || got.NotifyOnExpiry {
		t.Fatalf("false notification flags were not persisted: %+v", got)
	}
}
