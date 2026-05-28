package domain

import (
	"math"
	"testing"
)

const floatEps = 0.001

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < floatEps
}

func TestCalculateTotal_SimpleLineNoDiscount(t *testing.T) {
	item := InvoiceItemInput{Quantity: 2, UnitPrice: 50}
	got := item.CalculateTotal()
	if !almostEqual(got, 100) {
		t.Errorf("CalculateTotal() = %f, want 100", got)
	}
}

func TestCalculateTotal_WithDiscountOnly(t *testing.T) {
	// 10% discount on 100 = 90
	item := InvoiceItemInput{Quantity: 1, UnitPrice: 100, DiscountPct: 10}
	got := item.CalculateTotal()
	if !almostEqual(got, 90) {
		t.Errorf("CalculateTotal() = %f, want 90", got)
	}
}

func TestCalculateTotal_WithTaxOnly(t *testing.T) {
	// 10% tax on 100 = 110
	item := InvoiceItemInput{Quantity: 1, UnitPrice: 100, TaxRate: 10}
	got := item.CalculateTotal()
	if !almostEqual(got, 110) {
		t.Errorf("CalculateTotal() = %f, want 110", got)
	}
}

func TestCalculateTotal_DiscountThenTax(t *testing.T) {
	// base=200, discount 10%=20, after=180, tax 5%=9, total=189
	item := InvoiceItemInput{Quantity: 2, UnitPrice: 100, DiscountPct: 10, TaxRate: 5}
	got := item.CalculateTotal()
	if !almostEqual(got, 189) {
		t.Errorf("CalculateTotal() = %f, want 189", got)
	}
}

func TestCalculateTotal_ZeroQuantity(t *testing.T) {
	item := InvoiceItemInput{Quantity: 0, UnitPrice: 100}
	got := item.CalculateTotal()
	if got != 0 {
		t.Errorf("CalculateTotal() = %f, want 0", got)
	}
}

func TestCalculateTotal_ZeroUnitPrice(t *testing.T) {
	item := InvoiceItemInput{Quantity: 5, UnitPrice: 0}
	got := item.CalculateTotal()
	if got != 0 {
		t.Errorf("CalculateTotal() = %f, want 0", got)
	}
}

func TestCalculateTotal_FullDiscount(t *testing.T) {
	// 100% discount => no charge even before tax
	item := InvoiceItemInput{Quantity: 1, UnitPrice: 999, DiscountPct: 100}
	got := item.CalculateTotal()
	if !almostEqual(got, 0) {
		t.Errorf("CalculateTotal() = %f, want 0 for 100%% discount", got)
	}
}

func TestCalculateTotal_FractionalQuantity(t *testing.T) {
	// 1.5 units at 200 = 300
	item := InvoiceItemInput{Quantity: 1.5, UnitPrice: 200}
	got := item.CalculateTotal()
	if !almostEqual(got, 300) {
		t.Errorf("CalculateTotal() = %f, want 300", got)
	}
}
