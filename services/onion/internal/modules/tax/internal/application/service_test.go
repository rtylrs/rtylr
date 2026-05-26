package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/tax/internal/domain"
)

type mockRepo struct {
	cfg    domain.TaxConfig
	pl     domain.PLReport
	cfgErr error
	upErr  error
	plErr  error
}

func (m *mockRepo) GetConfig(_ context.Context, _ int64) (domain.TaxConfig, error) {
	return m.cfg, m.cfgErr
}

func (m *mockRepo) UpsertConfig(_ context.Context, _ domain.TaxConfigInput) error {
	return m.upErr
}

func (m *mockRepo) GetPLReport(_ context.Context, _ int64) (domain.PLReport, error) {
	return m.pl, m.plErr
}

func TestGetConfig_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if _, err := s.GetConfig(context.Background(), 1); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestGetConfig_ReturnsConfig(t *testing.T) {
	cfg := domain.TaxConfig{
		OwnerID: 5,
		TaxName: "VAT",
		TaxRate: 0.20,
		IsActive: true,
	}
	s := &Service{repo: &mockRepo{cfg: cfg}}
	got, err := s.GetConfig(context.Background(), 5)
	if err != nil || got.TaxName != "VAT" || got.TaxRate != 0.20 {
		t.Errorf("fields not preserved: %+v err=%v", got, err)
	}
}

func TestGetConfig_PropagatesError(t *testing.T) {
	want := errors.New("db fail")
	s := &Service{repo: &mockRepo{cfgErr: want}}
	if _, err := s.GetConfig(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateConfig_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.UpdateConfig(context.Background(), domain.TaxConfigInput{
		OwnerID: 1,
		TaxName: "GST",
		TaxRate: 0.10,
	})
	if err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestUpdateConfig_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{upErr: want}}
	if err := s.UpdateConfig(context.Background(), domain.TaxConfigInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetPLReport_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{pl: domain.PLReport{TotalRevenue: 1000}}}
	got, err := s.GetPLReport(context.Background(), 1)
	if err != nil || got.TotalRevenue != 1000 {
		t.Errorf("got %+v err %v", got, err)
	}
}

func TestGetPLReport_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{plErr: want}}
	if _, err := s.GetPLReport(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetPLReport_FieldsPreserved(t *testing.T) {
	pl := domain.PLReport{
		TotalRevenue:  5000,
		TotalExpenses: 2000,
		GrossProfit:   3000,
		NetProfit:     2500,
	}
	s := &Service{repo: &mockRepo{pl: pl}}
	got, err := s.GetPLReport(context.Background(), 1)
	if err != nil || got.NetProfit != 2500 || got.GrossProfit != 3000 {
		t.Errorf("fields not preserved: %+v err=%v", got, err)
	}
}

func TestGetConfig_IsInclusivePreserved(t *testing.T) {
	cfg := domain.TaxConfig{IsInclusive: true, TaxRate: 0.15}
	s := &Service{repo: &mockRepo{cfg: cfg}}
	got, err := s.GetConfig(context.Background(), 1)
	if err != nil || !got.IsInclusive {
		t.Errorf("IsInclusive not preserved: %+v err=%v", got, err)
	}
}

func TestUpdateConfig_InactiveConfig(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.UpdateConfig(context.Background(), domain.TaxConfigInput{
		OwnerID:  1,
		IsActive: false,
	})
	if err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestGetPLReport_ZeroValues(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	got, err := s.GetPLReport(context.Background(), 1)
	if err != nil || got.TotalRevenue != 0 || got.NetProfit != 0 {
		t.Errorf("expected zero values: %+v err=%v", got, err)
	}
}

func TestGetConfig_ZeroOwnerID(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if _, err := s.GetConfig(context.Background(), 0); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}
