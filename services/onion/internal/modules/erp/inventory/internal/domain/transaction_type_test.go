package domain

import "testing"

func TestInventoryTransactionType_Purchase_Value(t *testing.T) {
	if string(InventoryTypePurchase) != "purchase" {
		t.Errorf("expected 'purchase', got %q", InventoryTypePurchase)
	}
}

func TestInventoryTransactionType_Sale_Value(t *testing.T) {
	if string(InventoryTypeSale) != "sale" {
		t.Errorf("expected 'sale', got %q", InventoryTypeSale)
	}
}

func TestInventoryTransactionType_Adjustment_Value(t *testing.T) {
	if string(InventoryTypeAdjustment) != "adjustment" {
		t.Errorf("expected 'adjustment', got %q", InventoryTypeAdjustment)
	}
}

func TestInventoryTransactionType_Return_Value(t *testing.T) {
	if string(InventoryTypeReturn) != "return" {
		t.Errorf("expected 'return', got %q", InventoryTypeReturn)
	}
}

func TestInventoryTransactionType_Waste_Value(t *testing.T) {
	if string(InventoryTypeWaste) != "waste" {
		t.Errorf("expected 'waste', got %q", InventoryTypeWaste)
	}
}

func TestInventoryTransactionType_Distinct(t *testing.T) {
	types := []InventoryTransactionType{
		InventoryTypePurchase,
		InventoryTypeSale,
		InventoryTypeAdjustment,
		InventoryTypeReturn,
		InventoryTypeWaste,
	}
	seen := make(map[InventoryTransactionType]bool)
	for _, txType := range types {
		if seen[txType] {
			t.Errorf("duplicate transaction type: %q", txType)
		}
		seen[txType] = true
	}
}
