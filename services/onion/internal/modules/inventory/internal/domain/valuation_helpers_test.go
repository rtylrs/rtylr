package domain

import (
	"math"
	"testing"
)

func TestTotalValuationValue_Empty(t *testing.T) {
	if got := TotalValuationValue(nil); got != 0 {
		t.Errorf("expected 0 for nil slice, got %.2f", got)
	}
}

func TestTotalValuationValue_Single(t *testing.T) {
	items := []ValuationItem{{TotalValue: 500.0}}
	if got := TotalValuationValue(items); math.Abs(got-500.0) > 0.001 {
		t.Errorf("expected 500.0, got %.2f", got)
	}
}

func TestTotalValuationValue_Multiple(t *testing.T) {
	items := []ValuationItem{
		{TotalValue: 100.0},
		{TotalValue: 200.0},
		{TotalValue: 300.0},
	}
	want := 600.0
	if got := TotalValuationValue(items); math.Abs(got-want) > 0.001 {
		t.Errorf("expected %.2f, got %.2f", want, got)
	}
}

func TestTotalValuationCost_Empty(t *testing.T) {
	if got := TotalValuationCost(nil); got != 0 {
		t.Errorf("expected 0 for nil slice, got %.2f", got)
	}
}

func TestTotalValuationCost_Single(t *testing.T) {
	items := []ValuationItem{{TotalCost: 250.0}}
	if got := TotalValuationCost(items); math.Abs(got-250.0) > 0.001 {
		t.Errorf("expected 250.0, got %.2f", got)
	}
}

func TestTotalValuationCost_Multiple(t *testing.T) {
	items := []ValuationItem{
		{TotalCost: 50.0},
		{TotalCost: 75.0},
	}
	want := 125.0
	if got := TotalValuationCost(items); math.Abs(got-want) > 0.001 {
		t.Errorf("expected %.2f, got %.2f", want, got)
	}
}

func TestTotalItemCount_Empty(t *testing.T) {
	if got := TotalItemCount(nil); got != 0 {
		t.Errorf("expected 0 for nil slice, got %d", got)
	}
}

func TestTotalItemCount_Single(t *testing.T) {
	items := []ValuationItem{{TotalItems: 100}}
	if got := TotalItemCount(items); got != 100 {
		t.Errorf("expected 100, got %d", got)
	}
}

func TestTotalItemCount_Multiple(t *testing.T) {
	items := []ValuationItem{
		{TotalItems: 10},
		{TotalItems: 20},
		{TotalItems: 5},
	}
	if got := TotalItemCount(items); got != 35 {
		t.Errorf("expected 35, got %d", got)
	}
}
