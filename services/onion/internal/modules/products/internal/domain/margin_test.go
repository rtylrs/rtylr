package domain

import (
	"math"
	"testing"
)

func TestProduct_Margin_ZeroSalePrice(t *testing.T) {
	p := Product{CostPrice: 5.0, SalePrice: 0}
	if got := p.Margin(); got != 0 {
		t.Errorf("expected 0 margin for zero sale price, got %.2f", got)
	}
}

func TestProduct_Margin_FullMargin(t *testing.T) {
	p := Product{CostPrice: 0, SalePrice: 10}
	got := p.Margin()
	if got != 100 {
		t.Errorf("expected 100%% margin, got %.2f", got)
	}
}

func TestProduct_Margin_HalfMargin(t *testing.T) {
	p := Product{CostPrice: 5, SalePrice: 10}
	got := p.Margin()
	if math.Abs(got-50) > 0.001 {
		t.Errorf("expected 50%% margin, got %.2f", got)
	}
}

func TestProduct_Margin_NegativeMargin(t *testing.T) {
	p := Product{CostPrice: 15, SalePrice: 10}
	got := p.Margin()
	if got >= 0 {
		t.Errorf("expected negative margin when cost > sale, got %.2f", got)
	}
}

func TestProduct_GrossProfit_Positive(t *testing.T) {
	p := Product{CostPrice: 3.0, SalePrice: 8.0}
	got := p.GrossProfit()
	if math.Abs(got-5.0) > 0.001 {
		t.Errorf("expected gross profit=5.0, got %.2f", got)
	}
}

func TestProduct_GrossProfit_Zero(t *testing.T) {
	p := Product{CostPrice: 5.0, SalePrice: 5.0}
	if got := p.GrossProfit(); got != 0 {
		t.Errorf("expected gross profit=0, got %.2f", got)
	}
}

func TestProduct_GrossProfit_Negative(t *testing.T) {
	p := Product{CostPrice: 10.0, SalePrice: 7.0}
	got := p.GrossProfit()
	if got >= 0 {
		t.Errorf("expected negative gross profit, got %.2f", got)
	}
}

func TestProduct_StockValue_Basic(t *testing.T) {
	p := Product{StockQuantity: 100, CostPrice: 2.5}
	want := 250.0
	if got := p.StockValue(); math.Abs(got-want) > 0.001 {
		t.Errorf("expected stock value=%.2f, got %.2f", want, got)
	}
}

func TestProduct_StockValue_ZeroQuantity(t *testing.T) {
	p := Product{StockQuantity: 0, CostPrice: 5.0}
	if got := p.StockValue(); got != 0 {
		t.Errorf("expected 0 stock value, got %.2f", got)
	}
}

func TestProduct_IsLowStock_ZeroThreshold(t *testing.T) {
	p := Product{StockQuantity: 0, LowStockThreshold: 0}
	if p.IsLowStock() {
		t.Error("product with zero threshold should never be considered low stock")
	}
}

func TestProduct_IsLowStock_AtThreshold(t *testing.T) {
	p := Product{StockQuantity: 5, LowStockThreshold: 5}
	if !p.IsLowStock() {
		t.Error("stock at threshold should be considered low stock")
	}
}

func TestProduct_IsLowStock_BelowThreshold(t *testing.T) {
	p := Product{StockQuantity: 3, LowStockThreshold: 10}
	if !p.IsLowStock() {
		t.Error("stock below threshold should be considered low stock")
	}
}

func TestProduct_IsLowStock_AboveThreshold(t *testing.T) {
	p := Product{StockQuantity: 50, LowStockThreshold: 10}
	if p.IsLowStock() {
		t.Error("stock above threshold should not be considered low stock")
	}
}
