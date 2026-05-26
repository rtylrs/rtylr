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

func TestListInbox_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{listEr: want}}
	if _, _, err := s.ListInbox(context.Background(), 1, 1, 10); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestListInbox_ReturnsEmptySlice(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.InboxItem{}, total: 0}}
	got, total, err := s.ListInbox(context.Background(), 1, 1, 10)
	if err != nil || len(got) != 0 || total != 0 {
		t.Errorf("got %v total %d err %v", got, total, err)
	}
}

func TestListInbox_MultipleItems(t *testing.T) {
	items := []domain.InboxItem{{TicketID: 1}, {TicketID: 2}, {TicketID: 3}}
	s := &Service{repo: &mockRepo{list: items, total: 3}}
	got, total, err := s.ListInbox(context.Background(), 1, 1, 10)
	if err != nil || len(got) != 3 || total != 3 {
		t.Errorf("got %d items total=%d err=%v", len(got), total, err)
	}
}

func TestListInbox_FieldsPreserved(t *testing.T) {
	items := []domain.InboxItem{{TicketID: 7, Subject: "Help!"}}
	s := &Service{repo: &mockRepo{list: items, total: 1}}
	got, _, err := s.ListInbox(context.Background(), 1, 1, 10)
	if err != nil || got[0].TicketID != 7 || got[0].Subject != "Help!" {
		t.Errorf("fields not preserved: %v err %v", got, err)
	}
}

func TestUnreadCount_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{cntErr: want}}
	if _, err := s.UnreadCount(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUnreadCount_ReturnsZero(t *testing.T) {
	s := &Service{repo: &mockRepo{count: 0}}
	n, err := s.UnreadCount(context.Background(), 1)
	if err != nil || n != 0 {
		t.Errorf("count=%d err=%v", n, err)
	}
}

func TestUnreadCount_ReturnsValue(t *testing.T) {
	s := &Service{repo: &mockRepo{count: 7}}
	n, err := s.UnreadCount(context.Background(), 1)
	if err != nil || n != 7 {
		t.Errorf("expected 7, got %d err %v", n, err)
	}
}

func TestListInbox_TotalMismatch(t *testing.T) {
	// total can be larger than page size (total items vs page items)
	items := []domain.InboxItem{{TicketID: 1}}
	s := &Service{repo: &mockRepo{list: items, total: 100}}
	_, total, err := s.ListInbox(context.Background(), 1, 1, 10)
	if err != nil || total != 100 {
		t.Errorf("expected total=100 err=nil, got total=%d err=%v", total, err)
	}
}

func TestListInbox_ErrorZerosTotal(t *testing.T) {
	s := &Service{repo: &mockRepo{listEr: errors.New("fail"), total: 10}}
	_, total, err := s.ListInbox(context.Background(), 1, 1, 10)
	if err == nil {
		t.Fatal("expected error")
	}
	if total != 0 {
		t.Errorf("expected total=0 on error, got %d", total)
	}
}

func TestUnreadCount_LargeValue(t *testing.T) {
	s := &Service{repo: &mockRepo{count: 999}}
	n, err := s.UnreadCount(context.Background(), 1)
	if err != nil || n != 999 {
		t.Errorf("expected 999, got %d err %v", n, err)
	}
}

func TestListInbox_PageTwoWorks(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.InboxItem{{TicketID: 11}}, total: 20}}
	got, total, err := s.ListInbox(context.Background(), 1, 2, 10)
	if err != nil || len(got) != 1 || total != 20 {
		t.Errorf("page2: got %d items total=%d err=%v", len(got), total, err)
	}
}
