package domain

import (
	"math"
	"testing"
)

func TestPurchaseOrder_ComputedSubtotal_NoItems(t *testing.T) {
	po := PurchaseOrder{}
	if got := po.ComputedSubtotal(); got != 0 {
		t.Errorf("expected 0, got %.2f", got)
	}
}

func TestPurchaseOrder_ComputedSubtotal_MultipleItems(t *testing.T) {
	po := PurchaseOrder{
		Items: []PurchaseOrderItem{
			{TotalCost: 20.0},
			{TotalCost: 30.0},
			{TotalCost: 50.0},
		},
	}
	want := 100.0
	if got := po.ComputedSubtotal(); math.Abs(got-want) > 0.001 {
		t.Errorf("expected %.2f, got %.2f", want, got)
	}
}

func TestPurchaseOrder_ComputedTotal_WithTax(t *testing.T) {
	po := PurchaseOrder{
		TaxAmount: 10.0,
		Items:     []PurchaseOrderItem{{TotalCost: 90.0}},
	}
	if got := po.ComputedTotal(); math.Abs(got-100.0) > 0.001 {
		t.Errorf("expected 100.0, got %.2f", got)
	}
}

func TestPurchaseOrder_ComputedTotal_ZeroTax(t *testing.T) {
	po := PurchaseOrder{
		TaxAmount: 0,
		Items:     []PurchaseOrderItem{{TotalCost: 50.0}},
	}
	if got := po.ComputedTotal(); math.Abs(got-50.0) > 0.001 {
		t.Errorf("expected 50.0, got %.2f", got)
	}
}

func TestPurchaseOrder_IsCancellable_Draft(t *testing.T) {
	po := PurchaseOrder{Status: POStatusDraft}
	if !po.IsCancellable() {
		t.Error("draft order should be cancellable")
	}
}

func TestPurchaseOrder_IsCancellable_Sent(t *testing.T) {
	po := PurchaseOrder{Status: POStatusSent}
	if !po.IsCancellable() {
		t.Error("sent order should be cancellable")
	}
}

func TestPurchaseOrder_IsCancellable_Approved(t *testing.T) {
	po := PurchaseOrder{Status: POStatusApproved}
	if po.IsCancellable() {
		t.Error("approved order should not be cancellable")
	}
}

func TestPurchaseOrder_IsCancellable_Received(t *testing.T) {
	po := PurchaseOrder{Status: POStatusReceived}
	if po.IsCancellable() {
		t.Error("received order should not be cancellable")
	}
}

func TestPurchaseOrder_IsEditable_Draft(t *testing.T) {
	po := PurchaseOrder{Status: POStatusDraft}
	if !po.IsEditable() {
		t.Error("draft order should be editable")
	}
}

func TestPurchaseOrder_IsEditable_Sent(t *testing.T) {
	po := PurchaseOrder{Status: POStatusSent}
	if po.IsEditable() {
		t.Error("sent order should not be editable")
	}
}

func TestPurchaseOrder_IsEditable_Approved(t *testing.T) {
	po := PurchaseOrder{Status: POStatusApproved}
	if po.IsEditable() {
		t.Error("approved order should not be editable")
	}
}

func TestPurchaseOrder_ItemCount_Empty(t *testing.T) {
	po := PurchaseOrder{}
	if po.ItemCount() != 0 {
		t.Errorf("expected 0, got %d", po.ItemCount())
	}
}

func TestPurchaseOrder_ItemCount_Two(t *testing.T) {
	po := PurchaseOrder{
		Items: []PurchaseOrderItem{{ID: 1}, {ID: 2}},
	}
	if po.ItemCount() != 2 {
		t.Errorf("expected 2, got %d", po.ItemCount())
	}
}

func TestPurchaseOrder_TotalQuantityOrdered_Single(t *testing.T) {
	po := PurchaseOrder{
		Items: []PurchaseOrderItem{{QuantityOrdered: 5}},
	}
	if got := po.TotalQuantityOrdered(); got != 5 {
		t.Errorf("expected 5, got %.2f", got)
	}
}

func TestPurchaseOrder_TotalQuantityOrdered_Multiple(t *testing.T) {
	po := PurchaseOrder{
		Items: []PurchaseOrderItem{
			{QuantityOrdered: 10},
			{QuantityOrdered: 3},
			{QuantityOrdered: 7},
		},
	}
	if got := po.TotalQuantityOrdered(); math.Abs(got-20) > 0.001 {
		t.Errorf("expected 20, got %.2f", got)
	}
}
