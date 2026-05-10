package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/support_inbox/internal/domain"
)

type mockRepo struct {
	list   []domain.InboxItem
	total  int
	count  int
	listEr error
	cntErr error
}

func (m *mockRepo) ListInbox(_ context.Context, _ int64, _, _ int) ([]domain.InboxItem, int, error) {
	return m.list, m.total, m.listEr
}

func (m *mockRepo) UnreadCount(_ context.Context, _ int64) (int, error) {
	return m.count, m.cntErr
}

func TestListInbox_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.InboxItem{{}}, total: 5}}
	got, total, err := s.ListInbox(context.Background(), 1, 1, 10)
	if err != nil || len(got) != 1 || total != 5 {
		t.Errorf("got %v total %d err %v", got, total, err)
	}
}

func TestUnreadCount_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{cntErr: want}}
	if _, err := s.UnreadCount(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
