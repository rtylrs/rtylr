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
		OwnerID:            5,
		TaxName:            "VAT",
		TaxRateBasisPoints: 2000,
		IsActive:           true,
	}
	s := &Service{repo: &mockRepo{cfg: cfg}}
	got, err := s.GetConfig(context.Background(), 5)
	if err != nil || got.TaxName != "VAT" || got.TaxRateBasisPoints != 2000 {
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
		OwnerID:            1,
		TaxName:            "GST",
		TaxRateBasisPoints: 1000,
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
	s := &Service{repo: &mockRepo{pl: domain.PLReport{TotalRevenueMinorUnits: 100000}}}
	got, err := s.GetPLReport(context.Background(), 1)
	if err != nil || got.TotalRevenueMinorUnits != 100000 {
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
		TotalRevenueMinorUnits:  500000,
		TotalExpensesMinorUnits: 200000,
		GrossProfitMinorUnits:   300000,
		NetProfitMinorUnits:     250000,
	}
	s := &Service{repo: &mockRepo{pl: pl}}
	got, err := s.GetPLReport(context.Background(), 1)
	if err != nil || got.NetProfitMinorUnits != 250000 || got.GrossProfitMinorUnits != 300000 {
		t.Errorf("fields not preserved: %+v err=%v", got, err)
	}
}

func TestGetConfig_IsInclusivePreserved(t *testing.T) {
	cfg := domain.TaxConfig{IsInclusive: true, TaxRateBasisPoints: 1500}
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
	if err != nil || got.TotalRevenueMinorUnits != 0 || got.NetProfitMinorUnits != 0 {
		t.Errorf("expected zero values: %+v err=%v", got, err)
	}
}

func TestGetConfig_ZeroOwnerID(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if _, err := s.GetConfig(context.Background(), 0); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}
