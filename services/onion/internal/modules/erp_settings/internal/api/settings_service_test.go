package api

import (
	"testing"

	"github.com/voxire/onion/internal/modules/erp_settings/internal/domain"
)

func TestSettingsToProtoPreservesDisabledNotificationFlags(t *testing.T) {
	settings := settingsToProto(domain.Settings{
		ID:                 3,
		BusinessID:         42,
		Currency:           "EUR",
		TaxRate:            20,
		LowStockThreshold:  4,
		DefaultWarehouseID: 7,
		NotifyOnLowStock:   false,
		NotifyOnExpiry:     false,
	})

	if settings.NotifyOnLowStock || settings.NotifyOnExpiry {
		t.Fatalf("disabled notification flags were changed: %+v", settings)
	}
	if settings.TaxRate == nil || settings.LowStockThreshold == nil || settings.DefaultWarehouseId == nil {
		t.Fatalf("required settings pointers were not populated: %+v", settings)
	}
	if settings.DefaultWarehouseId == nil || *settings.DefaultWarehouseId != 7 {
		t.Fatalf("default warehouse = %v, want 7", settings.DefaultWarehouseId)
	}
	if err := settings.Validate(); err != nil {
		t.Fatalf("generated contract validation failed: %v", err)
	}
}
