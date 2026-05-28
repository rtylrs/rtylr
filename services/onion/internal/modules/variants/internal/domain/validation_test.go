package domain

import "testing"

func TestVariantInput_Validate_EmptyName(t *testing.T) {
	err := (VariantInput{Name: "", ProductID: 1}).Validate()
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestVariantInput_Validate_ZeroProductID(t *testing.T) {
	err := (VariantInput{Name: "Red", ProductID: 0}).Validate()
	if err == nil {
		t.Fatal("expected error for zero ProductID, got nil")
	}
}

func TestVariantInput_Validate_NegativeProductID(t *testing.T) {
	err := (VariantInput{Name: "Red", ProductID: -1}).Validate()
	if err == nil {
		t.Fatal("expected error for negative ProductID, got nil")
	}
}

func TestVariantInput_Validate_NegativeSalePrice(t *testing.T) {
	err := (VariantInput{Name: "Red", ProductID: 1, SalePrice: -1}).Validate()
	if err == nil {
		t.Fatal("expected error for negative sale price, got nil")
	}
}

func TestVariantInput_Validate_NegativeCostPrice(t *testing.T) {
	err := (VariantInput{Name: "Red", ProductID: 1, CostPrice: -0.01}).Validate()
	if err == nil {
		t.Fatal("expected error for negative cost price, got nil")
	}
}

func TestVariantInput_Validate_NegativeStockQuantity(t *testing.T) {
	err := (VariantInput{Name: "Red", ProductID: 1, StockQuantity: -5}).Validate()
	if err == nil {
		t.Fatal("expected error for negative stock quantity, got nil")
	}
}

func TestVariantInput_Validate_ValidInput(t *testing.T) {
	err := (VariantInput{
		Name:          "Blue Large",
		ProductID:     10,
		CostPrice:     5.0,
		SalePrice:     12.0,
		StockQuantity: 100,
		IsActive:      true,
	}).Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVariantInput_Validate_ZeroPricesOK(t *testing.T) {
	err := (VariantInput{Name: "Freebie", ProductID: 1}).Validate()
	if err != nil {
		t.Fatalf("zero prices should be valid, got: %v", err)
	}
}
