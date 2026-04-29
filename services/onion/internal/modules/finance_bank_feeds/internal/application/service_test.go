package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/finance_bank_feeds/internal/domain"
)

type mockRepo struct {
	list   []domain.BankFeedTransaction
	total  int64
	err    error
	gotLim int32
	gotOff int32
}

func (m *mockRepo) List(_ context.Context, _ int64, limit, offset int32) ([]domain.BankFeedTransaction, int64, error) {
	m.gotLim = limit
	m.gotOff = offset
	return m.list, m.total, m.err
}

func TestList_AppliesPaginationDefaults(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	if _, _, err := s.List(context.Background(), 1, 0, -1); err != nil {
		t.Fatalf("err: %v", err)
	}
	if m.gotLim != 20 {
		t.Errorf("expected default limit 20, got %d", m.gotLim)
	}
	if m.gotOff != 0 {
		t.Errorf("expected normalized offset 0, got %d", m.gotOff)
	}
}

func TestList_CapsLimit(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	_, _, _ = s.List(context.Background(), 1, 500, 0)
	if m.gotLim != 100 {
		t.Errorf("expected limit capped at 100, got %d", m.gotLim)
	}
}

func TestList_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{err: want}}
	if _, _, err := s.List(context.Background(), 1, 10, 0); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestSync_StubReturnsZero(t *testing.T) {
	s := &Service{}
	n, err := s.Sync(context.Background(), 1)
	if err != nil || n != 0 {
		t.Errorf("expected (0, nil), got (%d, %v)", n, err)
	}
}
