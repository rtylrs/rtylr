package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/finance_bank_feeds/internal/domain"
)

type mockRepo struct {
	list    []domain.BankFeedTransaction
	total   int64
	err     error
	gotLim  int32
	gotOff  int32
	sync    int
	syncErr error
}

func (m *mockRepo) List(_ context.Context, _ int64, limit, offset int32) ([]domain.BankFeedTransaction, int64, error) {
	m.gotLim = limit
	m.gotOff = offset
	return m.list, m.total, m.err
}

func (m *mockRepo) Sync(_ context.Context, _ int64) (int, error) {
	return m.sync, m.syncErr
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

func TestSync_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{sync: 3}}
	n, err := s.Sync(context.Background(), 1)
	if err != nil || n != 3 {
		t.Errorf("expected (3, nil), got (%d, %v)", n, err)
	}
}

func TestSync_PropagatesError(t *testing.T) {
	want := errors.New("sync failed")
	s := &Service{repo: &mockRepo{syncErr: want}}
	if _, err := s.Sync(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestList_ReturnsTransactions(t *testing.T) {
	txns := []domain.BankFeedTransaction{{ID: 1}, {ID: 2}}
	s := &Service{repo: &mockRepo{list: txns, total: 2}}
	got, total, err := s.List(context.Background(), 1, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(got))
	}
	if total != 2 {
		t.Errorf("expected total=2, got %d", total)
	}
}

func TestList_EmptyRepoReturnsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.BankFeedTransaction{}, total: 0}}
	got, total, err := s.List(context.Background(), 1, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 || total != 0 {
		t.Errorf("expected empty result, got len=%d total=%d", len(got), total)
	}
}

func TestList_PassesOffsetToRepo(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	_, _, _ = s.List(context.Background(), 1, 10, 5)
	if m.gotOff != 5 {
		t.Errorf("expected offset 5, got %d", m.gotOff)
	}
}

func TestList_PassesLimitToRepo(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	_, _, _ = s.List(context.Background(), 1, 50, 0)
	if m.gotLim != 50 {
		t.Errorf("expected limit 50, got %d", m.gotLim)
	}
}

func TestList_NegativeLimitDefaultsTo20(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	_, _, _ = s.List(context.Background(), 1, -5, 0)
	if m.gotLim != 20 {
		t.Errorf("expected default limit 20 for negative input, got %d", m.gotLim)
	}
}
