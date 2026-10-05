package domain

import "testing"

func TestValidateOrderInput_RejectsEmpty(t *testing.T) {
	if err := ValidateOrderInput(OrderInput{}); err == nil {
		t.Fatal("expected error for empty items")
	}
}

func TestValidateOrderInput_RejectsBlankItemName(t *testing.T) {
	err := ValidateOrderInput(OrderInput{
		Items: []OrderItemInput{{Quantity: 1, UnitPrice: 5}},
	})
	if err == nil {
		t.Fatal("expected error for blank item name")
	}
}

func TestValidateOrderInput_RejectsZeroQuantity(t *testing.T) {
	err := ValidateOrderInput(OrderInput{
		Items: []OrderItemInput{{Name: "Burger", Quantity: 0, UnitPrice: 5}},
	})
	if err == nil {
		t.Fatal("expected error for zero quantity")
	}
}

func TestValidateOrderInput_RejectsNegativeQuantity(t *testing.T) {
	err := ValidateOrderInput(OrderInput{
		Items: []OrderItemInput{{Name: "Burger", Quantity: -1, UnitPrice: 5}},
	})
	if err == nil {
		t.Fatal("expected error for negative quantity")
	}
}

func TestValidateOrderInput_RejectsNegativeUnitPrice(t *testing.T) {
	err := ValidateOrderInput(OrderInput{
		Items: []OrderItemInput{{Name: "Burger", Quantity: 1, UnitPrice: -1}},
	})
	if err == nil {
		t.Fatal("expected error for negative unit price")
	}
}

func TestValidateOrderInput_AcceptsValidItem(t *testing.T) {
	err := ValidateOrderInput(OrderInput{
		Items: []OrderItemInput{{Name: "Fries", Quantity: 2, UnitPrice: 3.50}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateOrderInput_AcceptsZeroUnitPrice(t *testing.T) {
	err := ValidateOrderInput(OrderInput{
		Items: []OrderItemInput{{Name: "Free Item", Quantity: 1, UnitPrice: 0}},
	})
	if err != nil {
		t.Fatalf("unexpected error for zero unit price: %v", err)
	}
}

func TestValidateRefundInput_RejectsMissingOrderID(t *testing.T) {
	if err := ValidateRefundInput(RefundInput{Amount: 10}); err == nil {
		t.Fatal("expected error for missing order id")
	}
}

func TestValidateRefundInput_RejectsZeroAmount(t *testing.T) {
	if err := ValidateRefundInput(RefundInput{OrderID: 1, Amount: 0}); err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestValidateRefundInput_RejectsNegativeAmount(t *testing.T) {
	if err := ValidateRefundInput(RefundInput{OrderID: 1, Amount: -5}); err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestValidateRefundInput_AcceptsValid(t *testing.T) {
	if err := ValidateRefundInput(RefundInput{OrderID: 1, Amount: 25}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateDiscountInput_RejectsMissingOrderID(t *testing.T) {
	if err := ValidateDiscountInput(DiscountInput{DiscountType: "pct", DiscountValue: 10}); err == nil {
		t.Fatal("expected error for missing order id")
	}
}

func TestValidateDiscountInput_RejectsNegativeValue(t *testing.T) {
	if err := ValidateDiscountInput(DiscountInput{OrderID: 1, DiscountType: "pct", DiscountValue: -5}); err == nil {
		t.Fatal("expected error for negative discount value")
	}
}

func TestValidateDiscountInput_RejectsBlankType(t *testing.T) {
	if err := ValidateDiscountInput(DiscountInput{OrderID: 1, DiscountValue: 10}); err == nil {
		t.Fatal("expected error for blank discount type")
	}
}

func TestValidateDiscountInput_AcceptsValid(t *testing.T) {
	if err := ValidateDiscountInput(DiscountInput{OrderID: 1, DiscountType: "pct", DiscountValue: 10}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
