package domain

import (
	"math"
	"testing"
)

func TestTotalStockValue_Empty(t *testing.T) {
	if got := TotalStockValue(nil); got != 0 {
		t.Errorf("expected 0 for nil slice, got %.2f", got)
	}
}

func TestTotalStockValue_SingleItem(t *testing.T) {
	items := []InventoryReportItem{{StockValue: 100.0}}
	if got := TotalStockValue(items); math.Abs(got-100.0) > 0.001 {
		t.Errorf("expected 100.0, got %.2f", got)
	}
}

func TestTotalStockValue_MultipleItems(t *testing.T) {
	items := []InventoryReportItem{
		{StockValue: 50.0},
		{StockValue: 75.5},
		{StockValue: 24.5},
	}
	want := 150.0
	if got := TotalStockValue(items); math.Abs(got-want) > 0.001 {
		t.Errorf("expected %.2f, got %.2f", want, got)
	}
}

func TestLowStockCount_Empty(t *testing.T) {
	if got := LowStockCount(nil); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestLowStockCount_NoneIsLow(t *testing.T) {
	items := []InventoryReportItem{
		{IsLowStock: false},
		{IsLowStock: false},
	}
	if got := LowStockCount(items); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestLowStockCount_AllLow(t *testing.T) {
	items := []InventoryReportItem{
		{IsLowStock: true},
		{IsLowStock: true},
		{IsLowStock: true},
	}
	if got := LowStockCount(items); got != 3 {
		t.Errorf("expected 3, got %d", got)
	}
}

func TestLowStockCount_Mixed(t *testing.T) {
	items := []InventoryReportItem{
		{IsLowStock: true},
		{IsLowStock: false},
		{IsLowStock: true},
	}
	if got := LowStockCount(items); got != 2 {
		t.Errorf("expected 2, got %d", got)
	}
}

func TestFilterLowStock_Empty(t *testing.T) {
	if got := FilterLowStock(nil); got != nil {
		t.Errorf("expected nil for nil input, got %v", got)
	}
}

func TestFilterLowStock_NoneIsLow(t *testing.T) {
	items := []InventoryReportItem{
		{ProductName: "A", IsLowStock: false},
		{ProductName: "B", IsLowStock: false},
	}
	got := FilterLowStock(items)
	if len(got) != 0 {
		t.Errorf("expected 0 items, got %d", len(got))
	}
}

func TestFilterLowStock_SomeLow(t *testing.T) {
	items := []InventoryReportItem{
		{ProductName: "A", IsLowStock: true},
		{ProductName: "B", IsLowStock: false},
		{ProductName: "C", IsLowStock: true},
	}
	got := FilterLowStock(items)
	if len(got) != 2 {
		t.Errorf("expected 2 low-stock items, got %d", len(got))
	}
	for _, item := range got {
		if !item.IsLowStock {
			t.Errorf("FilterLowStock returned non-low-stock item: %q", item.ProductName)
		}
	}
}

func TestFilterLowStock_AllLow(t *testing.T) {
	items := []InventoryReportItem{
		{ProductName: "X", IsLowStock: true},
		{ProductName: "Y", IsLowStock: true},
	}
	got := FilterLowStock(items)
	if len(got) != 2 {
		t.Errorf("expected 2, got %d", len(got))
	}
}
