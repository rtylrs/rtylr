package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/activity/internal/domain"
)

type mockRepo struct {
	list  []domain.ActivityLog
	total int64
	err   error
}

func (m *mockRepo) ListActivity(_ context.Context, _ int64, _ string, _, _ int) ([]domain.ActivityLog, int64, error) {
	return m.list, m.total, m.err
}

func TestListActivity_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.ActivityLog{{}}, total: 1}}
	got, err := s.ListActivity(context.Background(), 1, "", 10, 0)
	if err != nil {
		t.Errorf("unexpected: %v", err)
	}
	if len(got.Entries) != 1 {
		t.Errorf("got %d entries", len(got.Entries))
	}
	if got.Total != 1 {
		t.Errorf("got total %d, want 1", got.Total)
	}
}
