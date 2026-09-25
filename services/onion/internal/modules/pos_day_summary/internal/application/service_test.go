package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_day_summary/internal/domain"
)

type mockRepo struct {
	summary domain.DaySummary
	err     error
}

func (m *mockRepo) GetDaySummary(_ context.Context, _ int64, _ string) (domain.DaySummary, error) {
	return m.summary, m.err
}

func TestGetDaySummary_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if _, err := s.GetDaySummary(context.Background(), 1, "2026-04-28"); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestGetDaySummary_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{err: want}}
	if _, err := s.GetDaySummary(context.Background(), 1, ""); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
