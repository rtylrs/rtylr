package domain

import (
	"math"
	"testing"
)

func TestVariant_Margin_ZeroSalePrice(t *testing.T) {
	v := Variant{CostPrice: 5, SalePrice: 0}
	if got := v.Margin(); got != 0 {
		t.Errorf("expected 0 for zero sale price, got %.2f", got)
	}
}

func TestVariant_Margin_FullMargin(t *testing.T) {
	v := Variant{CostPrice: 0, SalePrice: 20}
	if got := v.Margin(); got != 100 {
		t.Errorf("expected 100%%, got %.2f", got)
	}
}

func TestVariant_Margin_HalfMargin(t *testing.T) {
	v := Variant{CostPrice: 10, SalePrice: 20}
	got := v.Margin()
	if math.Abs(got-50) > 0.001 {
		t.Errorf("expected 50%%, got %.2f", got)
	}
}

func TestVariant_GrossProfit_Positive(t *testing.T) {
	v := Variant{CostPrice: 4, SalePrice: 10}
	if got := v.GrossProfit(); math.Abs(got-6) > 0.001 {
		t.Errorf("expected 6.0, got %.2f", got)
	}
}

func TestVariant_GrossProfit_Zero(t *testing.T) {
	v := Variant{CostPrice: 5, SalePrice: 5}
	if got := v.GrossProfit(); got != 0 {
		t.Errorf("expected 0, got %.2f", got)
	}
}

func TestVariant_StockValue_Basic(t *testing.T) {
	v := Variant{StockQuantity: 50, CostPrice: 3.0}
	want := 150.0
	if got := v.StockValue(); math.Abs(got-want) > 0.001 {
		t.Errorf("expected %.2f, got %.2f", want, got)
	}
}

func TestVariant_StockValue_ZeroQuantity(t *testing.T) {
	v := Variant{StockQuantity: 0, CostPrice: 5.0}
	if got := v.StockValue(); got != 0 {
		t.Errorf("expected 0, got %.2f", got)
	}
}
