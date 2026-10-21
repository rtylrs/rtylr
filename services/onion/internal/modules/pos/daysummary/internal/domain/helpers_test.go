package domain

import "testing"

func TestDaySummary_NetRevenue_NoRefunds(t *testing.T) {
	d := DaySummary{TotalSales: 1000, TotalRefunds: 0}
	if d.NetRevenue() != 1000 {
		t.Errorf("NetRevenue: got %g, want 1000", d.NetRevenue())
	}
}

func TestDaySummary_NetRevenue_WithRefunds(t *testing.T) {
	d := DaySummary{TotalSales: 1000, TotalRefunds: 150}
	if d.NetRevenue() != 850 {
		t.Errorf("NetRevenue: got %g, want 850", d.NetRevenue())
	}
}

func TestDaySummary_NetSalesAfterTax(t *testing.T) {
	d := DaySummary{TotalSales: 1000, TotalTax: 100}
	if d.NetSalesAfterTax() != 900 {
		t.Errorf("NetSalesAfterTax: got %g, want 900", d.NetSalesAfterTax())
	}
}

func TestDaySummary_AverageOrderValue_NoOrders(t *testing.T) {
	d := DaySummary{TotalSales: 0, TotalOrders: 0}
	if d.AverageOrderValue() != 0 {
		t.Errorf("AverageOrderValue with no orders: got %g, want 0", d.AverageOrderValue())
	}
}

func TestDaySummary_AverageOrderValue_WithOrders(t *testing.T) {
	d := DaySummary{TotalSales: 600, TotalOrders: 12}
	if d.AverageOrderValue() != 50 {
		t.Errorf("AverageOrderValue: got %g, want 50", d.AverageOrderValue())
	}
}

func TestDaySummary_HasRefunds_False(t *testing.T) {
	d := DaySummary{TotalRefunds: 0}
	if d.HasRefunds() {
		t.Error("expected HasRefunds=false when no refunds")
	}
}

func TestDaySummary_HasRefunds_True(t *testing.T) {
	d := DaySummary{TotalRefunds: 50}
	if !d.HasRefunds() {
		t.Error("expected HasRefunds=true when refunds > 0")
	}
}

func TestDaySummary_CashSales_NilMap(t *testing.T) {
	d := DaySummary{}
	if d.CashSales() != 0 {
		t.Errorf("CashSales with nil map: got %g, want 0", d.CashSales())
	}
}

func TestDaySummary_CashSales_WithBreakdown(t *testing.T) {
	d := DaySummary{PaymentBreakdown: map[string]float64{"cash": 300, "card": 700}}
	if d.CashSales() != 300 {
		t.Errorf("CashSales: got %g, want 300", d.CashSales())
	}
}

func TestDaySummary_CardSales_WithBreakdown(t *testing.T) {
	d := DaySummary{PaymentBreakdown: map[string]float64{"cash": 300, "card": 700}}
	if d.CardSales() != 700 {
		t.Errorf("CardSales: got %g, want 700", d.CardSales())
	}
}
