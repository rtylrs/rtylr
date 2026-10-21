package domain

import "testing"

func TestWarehouse_HasLocation_CityOnly(t *testing.T) {
	w := Warehouse{City: "London"}
	if !w.HasLocation() {
		t.Error("expected true when city is set")
	}
}

func TestWarehouse_HasLocation_CountryOnly(t *testing.T) {
	w := Warehouse{Country: "UK"}
	if !w.HasLocation() {
		t.Error("expected true when country is set")
	}
}

func TestWarehouse_HasLocation_Neither(t *testing.T) {
	w := Warehouse{}
	if w.HasLocation() {
		t.Error("expected false when neither city nor country is set")
	}
}

func TestWarehouse_HasManager_Set(t *testing.T) {
	w := Warehouse{ManagerName: "Alice"}
	if !w.HasManager() {
		t.Error("expected true when manager name is set")
	}
}

func TestWarehouse_HasManager_Empty(t *testing.T) {
	w := Warehouse{}
	if w.HasManager() {
		t.Error("expected false when manager name is empty")
	}
}

func TestWarehouse_HasManager_Whitespace(t *testing.T) {
	w := Warehouse{ManagerName: "   "}
	if w.HasManager() {
		t.Error("expected false for whitespace-only manager name")
	}
}

func TestWarehouse_HasCapacity_Positive(t *testing.T) {
	w := Warehouse{Capacity: 1000}
	if !w.HasCapacity() {
		t.Error("expected true for positive capacity")
	}
}

func TestWarehouse_HasCapacity_Zero(t *testing.T) {
	w := Warehouse{Capacity: 0}
	if w.HasCapacity() {
		t.Error("expected false for zero capacity")
	}
}

func TestWarehouse_DisplayName_Trimmed(t *testing.T) {
	w := Warehouse{Name: "  Main Warehouse  "}
	if got := w.DisplayName(); got != "Main Warehouse" {
		t.Errorf("expected 'Main Warehouse', got %q", got)
	}
}

func TestWarehouse_DisplayName_Clean(t *testing.T) {
	w := Warehouse{Name: "East Wing"}
	if got := w.DisplayName(); got != "East Wing" {
		t.Errorf("expected 'East Wing', got %q", got)
	}
}
