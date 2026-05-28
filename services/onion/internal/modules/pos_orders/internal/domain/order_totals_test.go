package domain

import (
	"math"
	"testing"
)

const floatEps = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < floatEps
}

func TestComputeTotals_EmptyItems(t *testing.T) {
	totals := ComputeTotals(nil)
	if totals.Subtotal != 0 || totals.TaxAmount != 0 || totals.Total != 0 {
		t.Errorf("expected all zeros for empty items, got %+v", totals)
	}
}

func TestComputeTotals_SingleItemNoTax(t *testing.T) {
	items := []OrderItemInput{{Name: "Burger", Quantity: 2, UnitPrice: 10}}
	totals := ComputeTotals(items)
	if !almostEqual(totals.Subtotal, 20) {
		t.Errorf("subtotal: got %g, want 20", totals.Subtotal)
	}
	if !almostEqual(totals.TaxAmount, 0) {
		t.Errorf("tax: got %g, want 0", totals.TaxAmount)
	}
	if !almostEqual(totals.Total, 20) {
		t.Errorf("total: got %g, want 20", totals.Total)
	}
}

func TestComputeTotals_SingleItemWithTax(t *testing.T) {
	items := []OrderItemInput{{Name: "Burger", Quantity: 1, UnitPrice: 100, TaxRate: 10}}
	totals := ComputeTotals(items)
	if !almostEqual(totals.Subtotal, 100) {
		t.Errorf("subtotal: got %g, want 100", totals.Subtotal)
	}
	if !almostEqual(totals.TaxAmount, 10) {
		t.Errorf("tax: got %g, want 10", totals.TaxAmount)
	}
	if !almostEqual(totals.Total, 110) {
		t.Errorf("total: got %g, want 110", totals.Total)
	}
}

func TestComputeTotals_DiscountReducesSubtotal(t *testing.T) {
	items := []OrderItemInput{{Name: "Pizza", Quantity: 1, UnitPrice: 50, DiscountPct: 10}}
	totals := ComputeTotals(items)
	if !almostEqual(totals.Subtotal, 45) {
		t.Errorf("subtotal after 10%% discount: got %g, want 45", totals.Subtotal)
	}
}

func TestComputeTotals_DiscountThenTax(t *testing.T) {
	items := []OrderItemInput{{Name: "Steak", Quantity: 1, UnitPrice: 100, DiscountPct: 20, TaxRate: 10}}
	totals := ComputeTotals(items)
	// After 20% discount: 80; tax 10% of 80 = 8; total = 88.
	if !almostEqual(totals.Subtotal, 80) {
		t.Errorf("subtotal: got %g, want 80", totals.Subtotal)
	}
	if !almostEqual(totals.TaxAmount, 8) {
		t.Errorf("tax: got %g, want 8", totals.TaxAmount)
	}
	if !almostEqual(totals.Total, 88) {
		t.Errorf("total: got %g, want 88", totals.Total)
	}
}

func TestComputeTotals_MultipleItems(t *testing.T) {
	items := []OrderItemInput{
		{Name: "A", Quantity: 2, UnitPrice: 10},
		{Name: "B", Quantity: 1, UnitPrice: 5},
	}
	totals := ComputeTotals(items)
	if !almostEqual(totals.Subtotal, 25) {
		t.Errorf("subtotal: got %g, want 25", totals.Subtotal)
	}
}

func TestComputeTotals_TotalEqualsSubtotalPlusTax(t *testing.T) {
	items := []OrderItemInput{{Name: "Item", Quantity: 3, UnitPrice: 20, TaxRate: 5}}
	totals := ComputeTotals(items)
	if !almostEqual(totals.Total, totals.Subtotal+totals.TaxAmount) {
		t.Errorf("total mismatch: %g != %g + %g", totals.Total, totals.Subtotal, totals.TaxAmount)
	}
}
