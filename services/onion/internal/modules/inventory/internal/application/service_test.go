package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/inventory/internal/domain"
)

type mockRepo struct {
	reportItems  []domain.InventoryReportItem
	valuations   []domain.ValuationItem
	adjustErr    error
	reportErr    error
	lowStockErr  error
	valuationErr error
}

func (m *mockRepo) AdjustInventory(_ context.Context, _, _ int64, _ float64, _ domain.InventoryTransactionType, _ string) error {
	return m.adjustErr
}

func (m *mockRepo) GetInventoryReport(_ context.Context, _ int64) ([]domain.InventoryReportItem, error) {
	return m.reportItems, m.reportErr
}

func (m *mockRepo) GetLowStockProducts(_ context.Context, _ int64) ([]domain.InventoryReportItem, error) {
	return m.reportItems, m.lowStockErr
}

func (m *mockRepo) GetValuation(_ context.Context, _ int64) ([]domain.ValuationItem, error) {
	return m.valuations, m.valuationErr
}

func TestAdjustInventory_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.AdjustInventory(context.Background(), 1, domain.InventoryAdjustInput{
		ProductID: 10,
		Change:    5,
		Type:      domain.InventoryTypeAdjustment,
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAdjustInventory_PropagatesRepoError(t *testing.T) {
	want := errors.New("db failure")
	s := &Service{repo: &mockRepo{adjustErr: want}}
	err := s.AdjustInventory(context.Background(), 1, domain.InventoryAdjustInput{ProductID: 1})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestAdjustInventory_LowStockCheckDoesNotBlockSuccess(t *testing.T) {
	// Even if GetLowStockProducts would fail, AdjustInventory still succeeds
	// because the low-stock check is advisory only.
	s := &Service{repo: &mockRepo{lowStockErr: errors.New("low stock check failed")}}
	err := s.AdjustInventory(context.Background(), 1, domain.InventoryAdjustInput{ProductID: 1, Change: -1})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetInventoryReport_ReturnsDelegatedItems(t *testing.T) {
	items := []domain.InventoryReportItem{
		{ProductID: 1, ProductName: "Widget", SKU: "W-001"},
	}
	s := &Service{repo: &mockRepo{reportItems: items}}
	got, err := s.GetInventoryReport(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ProductID != 1 {
		t.Errorf("expected 1 item with ID 1, got %+v", got)
	}
}

func TestGetInventoryReport_ReturnsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{reportItems: nil}}
	got, err := s.GetInventoryReport(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %+v", got)
	}
}

func TestGetInventoryReport_PropagatesRepoError(t *testing.T) {
	want := errors.New("report error")
	s := &Service{repo: &mockRepo{reportErr: want}}
	_, err := s.GetInventoryReport(context.Background(), 1)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetLowStockProducts_ReturnsDelegatedItems(t *testing.T) {
	items := []domain.InventoryReportItem{
		{ProductID: 2, IsLowStock: true},
	}
	s := &Service{repo: &mockRepo{reportItems: items}}
	got, err := s.GetLowStockProducts(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || !got[0].IsLowStock {
		t.Errorf("expected low-stock item, got %+v", got)
	}
}

func TestGetLowStockProducts_PropagatesRepoError(t *testing.T) {
	want := errors.New("low stock error")
	s := &Service{repo: &mockRepo{lowStockErr: want}}
	_, err := s.GetLowStockProducts(context.Background(), 1)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetValuation_ReturnsDelegatedItems(t *testing.T) {
	vals := []domain.ValuationItem{
		{CategoryName: "Electronics", TotalValue: 1000},
	}
	s := &Service{repo: &mockRepo{valuations: vals}}
	got, err := s.GetValuation(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].TotalValue != 1000 {
		t.Errorf("expected valuation item, got %+v", got)
	}
}

func TestGetValuation_ReturnsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{valuations: nil}}
	got, err := s.GetValuation(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %+v", got)
	}
}

func TestGetValuation_PropagatesRepoError(t *testing.T) {
	want := errors.New("valuation error")
	s := &Service{repo: &mockRepo{valuationErr: want}}
	_, err := s.GetValuation(context.Background(), 1)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
