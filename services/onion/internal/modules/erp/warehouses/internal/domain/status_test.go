package domain

import "testing"

func TestWarehouse_Status_Active(t *testing.T) {
	w := Warehouse{IsActive: true}
	if got := w.Status(); got != WarehouseStatusActive {
		t.Errorf("expected %q, got %q", WarehouseStatusActive, got)
	}
}

func TestWarehouse_Status_Inactive(t *testing.T) {
	w := Warehouse{IsActive: false}
	if got := w.Status(); got != WarehouseStatusInactive {
		t.Errorf("expected %q, got %q", WarehouseStatusInactive, got)
	}
}

func TestWarehouseStatusActive_StringValue(t *testing.T) {
	if string(WarehouseStatusActive) != "active" {
		t.Errorf("expected 'active', got %q", WarehouseStatusActive)
	}
}

func TestWarehouseStatusInactive_StringValue(t *testing.T) {
	if string(WarehouseStatusInactive) != "inactive" {
		t.Errorf("expected 'inactive', got %q", WarehouseStatusInactive)
	}
}

func TestWarehouseStatus_Distinct(t *testing.T) {
	if WarehouseStatusActive == WarehouseStatusInactive {
		t.Error("active and inactive statuses must be distinct")
	}
}
