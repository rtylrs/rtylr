package domain

import (
	"math"
	"testing"
)

func TestPurchaseOrderItem_LineTotal_Basic(t *testing.T) {
	item := PurchaseOrderItem{QuantityOrdered: 5, UnitCost: 10.0}
	want := 50.0
	if got := item.LineTotal(); math.Abs(got-want) > 0.001 {
		t.Errorf("expected %.2f, got %.2f", want, got)
	}
}

func TestPurchaseOrderItem_LineTotal_ZeroQuantity(t *testing.T) {
	item := PurchaseOrderItem{QuantityOrdered: 0, UnitCost: 10.0}
	if got := item.LineTotal(); got != 0 {
		t.Errorf("expected 0, got %.2f", got)
	}
}

func TestPurchaseOrderItem_LineTotal_ZeroCost(t *testing.T) {
	item := PurchaseOrderItem{QuantityOrdered: 3, UnitCost: 0}
	if got := item.LineTotal(); got != 0 {
		t.Errorf("expected 0, got %.2f", got)
	}
}

func TestPurchaseOrderItem_RemainingQuantity_Full(t *testing.T) {
	item := PurchaseOrderItem{QuantityOrdered: 10, QuantityReceived: 10}
	if got := item.RemainingQuantity(); got != 0 {
		t.Errorf("expected 0, got %.2f", got)
	}
}

func TestPurchaseOrderItem_RemainingQuantity_None(t *testing.T) {
	item := PurchaseOrderItem{QuantityOrdered: 10, QuantityReceived: 0}
	if got := item.RemainingQuantity(); got != 10 {
		t.Errorf("expected 10, got %.2f", got)
	}
}

func TestPurchaseOrderItem_RemainingQuantity_Partial(t *testing.T) {
	item := PurchaseOrderItem{QuantityOrdered: 10, QuantityReceived: 3}
	if got := item.RemainingQuantity(); math.Abs(got-7) > 0.001 {
		t.Errorf("expected 7, got %.2f", got)
	}
}

func TestPurchaseOrderItem_RemainingQuantity_OverReceived(t *testing.T) {
	item := PurchaseOrderItem{QuantityOrdered: 5, QuantityReceived: 10}
	if got := item.RemainingQuantity(); got != 0 {
		t.Errorf("expected 0 for over-received item, got %.2f", got)
	}
}

func TestPurchaseOrderItem_IsFullyReceived_True(t *testing.T) {
	item := PurchaseOrderItem{QuantityOrdered: 5, QuantityReceived: 5}
	if !item.IsFullyReceived() {
		t.Error("expected true when received == ordered")
	}
}

func TestPurchaseOrderItem_IsFullyReceived_OverReceived(t *testing.T) {
	item := PurchaseOrderItem{QuantityOrdered: 5, QuantityReceived: 7}
	if !item.IsFullyReceived() {
		t.Error("expected true when received > ordered")
	}
}

func TestPurchaseOrderItem_IsFullyReceived_NotReceived(t *testing.T) {
	item := PurchaseOrderItem{QuantityOrdered: 5, QuantityReceived: 0}
	if item.IsFullyReceived() {
		t.Error("expected false when received == 0")
	}
}

func TestPurchaseOrderItem_IsFullyReceived_Partial(t *testing.T) {
	item := PurchaseOrderItem{QuantityOrdered: 10, QuantityReceived: 4}
	if item.IsFullyReceived() {
		t.Error("expected false for partial receipt")
	}
}
