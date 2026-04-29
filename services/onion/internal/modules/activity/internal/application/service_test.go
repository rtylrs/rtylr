package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/activity/internal/domain"
)

type mockRepo struct {
	list []domain.ActivityLog
	err  error
}

func (m *mockRepo) ListActivity(_ context.Context, _ int64, _ int) ([]domain.ActivityLog, error) {
	return m.list, m.err
}

func TestListActivity_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.ActivityLog{{}}}}
	got, err := s.ListActivity(context.Background(), 1, 10)
	if err != nil {
		t.Errorf("unexpected: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}
