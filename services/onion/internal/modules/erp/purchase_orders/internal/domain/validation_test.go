package domain

import "testing"

func TestPurchaseOrderInput_Validate_ZeroSupplier(t *testing.T) {
	err := (PurchaseOrderInput{SupplierID: 0, PONumber: "PO-1", Items: []PurchaseOrderItemInput{{ProductID: 1, QuantityOrdered: 1}}}).Validate()
	if err == nil {
		t.Fatal("expected error for zero supplier, got nil")
	}
}

func TestPurchaseOrderInput_Validate_EmptyPONumber(t *testing.T) {
	err := (PurchaseOrderInput{SupplierID: 1, PONumber: "", Items: []PurchaseOrderItemInput{{ProductID: 1, QuantityOrdered: 1}}}).Validate()
	if err == nil {
		t.Fatal("expected error for empty PO number, got nil")
	}
}

func TestPurchaseOrderInput_Validate_NoItems(t *testing.T) {
	err := (PurchaseOrderInput{SupplierID: 1, PONumber: "PO-1", Items: nil}).Validate()
	if err == nil {
		t.Fatal("expected error for no items, got nil")
	}
}

func TestPurchaseOrderInput_Validate_NegativeTax(t *testing.T) {
	err := (PurchaseOrderInput{
		SupplierID: 1,
		PONumber:   "PO-1",
		TaxAmount:  -5.0,
		Items:      []PurchaseOrderItemInput{{ProductID: 1, QuantityOrdered: 1}},
	}).Validate()
	if err == nil {
		t.Fatal("expected error for negative tax, got nil")
	}
}

func TestPurchaseOrderInput_Validate_ItemZeroProductID(t *testing.T) {
	err := (PurchaseOrderInput{
		SupplierID: 1,
		PONumber:   "PO-1",
		Items:      []PurchaseOrderItemInput{{ProductID: 0, QuantityOrdered: 1}},
	}).Validate()
	if err == nil {
		t.Fatal("expected error for item with zero product ID, got nil")
	}
}

func TestPurchaseOrderInput_Validate_ItemZeroQuantity(t *testing.T) {
	err := (PurchaseOrderInput{
		SupplierID: 1,
		PONumber:   "PO-1",
		Items:      []PurchaseOrderItemInput{{ProductID: 1, QuantityOrdered: 0}},
	}).Validate()
	if err == nil {
		t.Fatal("expected error for item with zero quantity, got nil")
	}
}

func TestPurchaseOrderInput_Validate_ItemNegativeUnitCost(t *testing.T) {
	err := (PurchaseOrderInput{
		SupplierID: 1,
		PONumber:   "PO-1",
		Items:      []PurchaseOrderItemInput{{ProductID: 1, QuantityOrdered: 1, UnitCost: -1}},
	}).Validate()
	if err == nil {
		t.Fatal("expected error for negative unit cost, got nil")
	}
}

func TestPurchaseOrderInput_Validate_ValidInput(t *testing.T) {
	err := (PurchaseOrderInput{
		SupplierID: 1,
		PONumber:   "PO-VALID-001",
		TaxAmount:  10.0,
		Items: []PurchaseOrderItemInput{
			{ProductID: 1, QuantityOrdered: 2, UnitCost: 5.0},
			{ProductID: 2, QuantityOrdered: 1, UnitCost: 0},
		},
	}).Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPurchaseOrderItemInput_Validate_ZeroProductID(t *testing.T) {
	err := (PurchaseOrderItemInput{ProductID: 0, QuantityOrdered: 1}).Validate()
	if err == nil {
		t.Fatal("expected error for zero product ID, got nil")
	}
}

func TestPurchaseOrderItemInput_Validate_ZeroQuantity(t *testing.T) {
	err := (PurchaseOrderItemInput{ProductID: 1, QuantityOrdered: 0}).Validate()
	if err == nil {
		t.Fatal("expected error for zero quantity, got nil")
	}
}

func TestPurchaseOrderItemInput_Validate_NegativeQuantity(t *testing.T) {
	err := (PurchaseOrderItemInput{ProductID: 1, QuantityOrdered: -1}).Validate()
	if err == nil {
		t.Fatal("expected error for negative quantity, got nil")
	}
}

func TestPurchaseOrderItemInput_Validate_ValidInput(t *testing.T) {
	err := (PurchaseOrderItemInput{ProductID: 5, QuantityOrdered: 3, UnitCost: 12.5}).Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
