package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/inventory/internal/domain"
)

type mockRepo struct {
	reportItems  []domain.InventoryReportItem
	valuations   []domain.ValuationItem
	adjustErr    error
	reportErr    error
	lowStockErr  error
	valuationErr error
	reportBounds struct {
		limit, offset int
	}
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

func (m *mockRepo) GetInventoryReportPage(_ context.Context, _ int64, limit, offset int) (domain.InventoryReportPage, error) {
	m.reportBounds.limit = limit
	m.reportBounds.offset = offset
	return domain.InventoryReportPage{
		Items:      m.reportItems,
		Total:      int64(len(m.reportItems)),
		TotalValue: domain.TotalStockValue(m.reportItems),
		HasMore:    int64(len(m.reportItems))+int64(offset) > int64(limit),
		Limit:      int64(limit),
		Offset:     int64(offset),
	}, m.reportErr
}

func (m *mockRepo) GetLowStockProductsPage(_ context.Context, _ int64, limit, offset int) (domain.LowStockPage, error) {
	return domain.LowStockPage{
		Items:   m.reportItems,
		Total:   int64(len(m.reportItems)),
		HasMore: int64(len(m.reportItems))+int64(offset) > int64(limit),
		Limit:   int64(limit),
		Offset:  int64(offset),
	}, m.lowStockErr
}

func (m *mockRepo) GetValuationPage(_ context.Context, _ int64, limit, offset int) (domain.ValuationPage, error) {
	totalValue := 0.0
	for _, valuation := range m.valuations {
		totalValue += valuation.TotalValue
	}
	return domain.ValuationPage{
		Items:      m.valuations,
		Total:      int64(len(m.valuations)),
		TotalValue: totalValue,
		HasMore:    int64(len(m.valuations))+int64(offset) > int64(limit),
		Limit:      int64(limit),
		Offset:     int64(offset),
	}, m.valuationErr
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

func TestGetInventoryReportPage_HonorsOffsetAndReturnsMetadata(t *testing.T) {
	repo := &mockRepo{reportItems: []domain.InventoryReportItem{{ProductID: 101}}}
	s := &Service{repo: repo}

	page, err := s.GetInventoryReportPage(context.Background(), 7, 25, 10_100)
	if err != nil {
		t.Fatalf("GetInventoryReportPage: %v", err)
	}
	if repo.reportBounds.limit != 25 || repo.reportBounds.offset != 10_100 {
		t.Fatalf("repository bounds = %+v", repo.reportBounds)
	}
	if page.Limit != 25 || page.Offset != 10_100 || page.Total != 1 || page.HasMore {
		t.Fatalf("page metadata = %+v", page)
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
