package domain

import "testing"

func TestProductInput_Validate_EmptyName(t *testing.T) {
	err := (ProductInput{Name: ""}).Validate()
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestProductInput_Validate_ValidName(t *testing.T) {
	err := (ProductInput{Name: "Burger", SalePrice: 5.0}).Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProductInput_Validate_NegativeSalePrice(t *testing.T) {
	err := (ProductInput{Name: "Widget", SalePrice: -1.0}).Validate()
	if err == nil {
		t.Fatal("expected error for negative sale price, got nil")
	}
}

func TestProductInput_Validate_NegativeCostPrice(t *testing.T) {
	err := (ProductInput{Name: "Widget", CostPrice: -0.01}).Validate()
	if err == nil {
		t.Fatal("expected error for negative cost price, got nil")
	}
}

func TestProductInput_Validate_ZeroPricesOK(t *testing.T) {
	err := (ProductInput{Name: "Free Sample", SalePrice: 0, CostPrice: 0}).Validate()
	if err != nil {
		t.Fatalf("zero prices should be valid, got: %v", err)
	}
}

func TestProductInput_Validate_NegativeLowStockThreshold(t *testing.T) {
	err := (ProductInput{Name: "Widget", LowStockThreshold: -5}).Validate()
	if err == nil {
		t.Fatal("expected error for negative low stock threshold, got nil")
	}
}

func TestProductInput_Validate_NegativeReorderQuantity(t *testing.T) {
	err := (ProductInput{Name: "Widget", ReorderQuantity: -1}).Validate()
	if err == nil {
		t.Fatal("expected error for negative reorder quantity, got nil")
	}
}

func TestProductInput_Validate_AllFieldsValid(t *testing.T) {
	err := (ProductInput{
		Name:              "Premium Burger",
		CostPrice:         3.0,
		SalePrice:         9.99,
		LowStockThreshold: 10,
		ReorderQuantity:   50,
		IsActive:          true,
	}).Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
