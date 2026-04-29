package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/tax/internal/domain"
)

type mockRepo struct {
	cfg      domain.TaxConfig
	pl       domain.PLReport
	cfgErr   error
	upErr    error
	plErr    error
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

func TestUpdateConfig_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{upErr: want}}
	if err := s.UpdateConfig(context.Background(), domain.TaxConfigInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetPLReport_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{plErr: want}}
	if _, err := s.GetPLReport(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
