package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_dashboard/internal/domain"
)

type mockRepo struct {
	widgets   []*domain.DashboardWidget
	getErr    error
	upsertErr error
}

func (m *mockRepo) GetWidgets(_ context.Context, _ int64) ([]*domain.DashboardWidget, error) {
	return m.widgets, m.getErr
}

func (m *mockRepo) UpsertWidgets(_ context.Context, _ int64, _ []*domain.DashboardWidget) error {
	return m.upsertErr
}

func TestGetWidgets_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{widgets: []*domain.DashboardWidget{{}}}}
	got, _ := s.GetWidgets(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestUpsertWidgets_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{upsertErr: want}}
	if err := s.UpsertWidgets(context.Background(), 1, nil); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
