package domain

import "testing"

func TestInventoryAdjustInput_Validate_ZeroProductID(t *testing.T) {
	err := (InventoryAdjustInput{ProductID: 0, Change: 1, Type: InventoryTypeAdjustment}).Validate()
	if err == nil {
		t.Fatal("expected error for zero ProductID, got nil")
	}
}

func TestInventoryAdjustInput_Validate_NegativeProductID(t *testing.T) {
	err := (InventoryAdjustInput{ProductID: -1, Change: 1, Type: InventoryTypeAdjustment}).Validate()
	if err == nil {
		t.Fatal("expected error for negative ProductID, got nil")
	}
}

func TestInventoryAdjustInput_Validate_ZeroChange(t *testing.T) {
	err := (InventoryAdjustInput{ProductID: 1, Change: 0, Type: InventoryTypeAdjustment}).Validate()
	if err == nil {
		t.Fatal("expected error for zero change, got nil")
	}
}

func TestInventoryAdjustInput_Validate_UnknownType(t *testing.T) {
	err := (InventoryAdjustInput{ProductID: 1, Change: 5, Type: "unknown"}).Validate()
	if err == nil {
		t.Fatal("expected error for unknown type, got nil")
	}
}

func TestInventoryAdjustInput_Validate_PurchaseType(t *testing.T) {
	err := (InventoryAdjustInput{ProductID: 1, Change: 10, Type: InventoryTypePurchase}).Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInventoryAdjustInput_Validate_SaleType(t *testing.T) {
	err := (InventoryAdjustInput{ProductID: 1, Change: -3, Type: InventoryTypeSale}).Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInventoryAdjustInput_Validate_AdjustmentType(t *testing.T) {
	err := (InventoryAdjustInput{ProductID: 1, Change: 1, Type: InventoryTypeAdjustment}).Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInventoryAdjustInput_Validate_ReturnType(t *testing.T) {
	err := (InventoryAdjustInput{ProductID: 1, Change: 2, Type: InventoryTypeReturn}).Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInventoryAdjustInput_Validate_WasteType(t *testing.T) {
	err := (InventoryAdjustInput{ProductID: 1, Change: -1, Type: InventoryTypeWaste}).Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInventoryAdjustInput_Validate_NegativeChangeValid(t *testing.T) {
	err := (InventoryAdjustInput{ProductID: 5, Change: -100, Type: InventoryTypeSale}).Validate()
	if err != nil {
		t.Fatalf("negative change (sale) should be valid, got: %v", err)
	}
}
