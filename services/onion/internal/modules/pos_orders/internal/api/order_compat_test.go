package api

import (
	"context"
	"testing"
)

func TestValidateOrderCompatItemRequiresExactlyOneProductKind(t *testing.T) {
	if err := validateOrderCompatItem(OrderCompatItemInput{ComboID: 7, Quantity: 1}); err != nil {
		t.Fatalf("combo-only item rejected: %v", err)
	}
	if err := validateOrderCompatItem(OrderCompatItemInput{MenuItemID: 7, ComboID: 8, Quantity: 1}); err == nil {
		t.Fatal("item referencing both menu and combo was accepted")
	}
	if err := validateOrderCompatItem(OrderCompatItemInput{MenuItemID: 7, Quantity: 0}); err == nil {
		t.Fatal("zero-quantity item was accepted")
	}
}

func TestCreateOrderCompatHandlerRequiresSession(t *testing.T) {
	handler := &CreateOrderCompatHandler{}
	if err := handler.Validate(context.Background(), &CreateOrderCompatRequest{
		Items: []OrderCompatItemInput{{MenuItemID: 7, Quantity: 1}},
	}); err == nil {
		t.Fatal("compat create accepted a request without an open session")
	}
}
