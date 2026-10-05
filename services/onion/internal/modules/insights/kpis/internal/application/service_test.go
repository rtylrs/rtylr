package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/insights/kpis/internal/domain"
)

type mockRepo struct {
	result domain.KPISummary
	err    error
}

func (m *mockRepo) GetKPIs(_ context.Context, _ int64) (domain.KPISummary, error) {
	return m.result, m.err
}

func TestGetKPIs_Delegates(t *testing.T) {
	want := domain.KPISummary{}
	s := &Service{repo: &mockRepo{result: want}}
	if _, err := s.GetKPIs(context.Background(), 1); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestGetKPIs_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{err: want}}
	if _, err := s.GetKPIs(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
