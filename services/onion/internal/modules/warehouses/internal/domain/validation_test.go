package domain

import "testing"

func TestWarehouseInput_Validate_EmptyName(t *testing.T) {
	err := (WarehouseInput{Name: ""}).Validate()
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestWarehouseInput_Validate_ValidName(t *testing.T) {
	err := (WarehouseInput{Name: "Main Warehouse"}).Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWarehouseInput_Validate_NegativeCapacity(t *testing.T) {
	err := (WarehouseInput{Name: "WH", Capacity: -1}).Validate()
	if err == nil {
		t.Fatal("expected error for negative capacity, got nil")
	}
}

func TestWarehouseInput_Validate_ZeroCapacityOK(t *testing.T) {
	err := (WarehouseInput{Name: "WH", Capacity: 0}).Validate()
	if err != nil {
		t.Fatalf("zero capacity should be valid, got: %v", err)
	}
}

func TestWarehouseInput_Validate_FullFields(t *testing.T) {
	err := (WarehouseInput{
		Name:        "East Wing",
		Address:     "100 Industrial Rd",
		City:        "Testville",
		Country:     "US",
		ManagerName: "Jane",
		Capacity:    500,
		IsActive:    true,
	}).Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
