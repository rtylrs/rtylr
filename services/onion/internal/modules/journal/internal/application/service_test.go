package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/journal/internal/domain"
)

type mockRepo struct {
	list      []domain.JournalEntry
	createID  int64
	createErr error
	postErr   error
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.JournalEntry, error) {
	return m.list, nil
}
func (m *mockRepo) Create(_ context.Context, _ int64, _ domain.JournalEntryInput) (int64, error) {
	return m.createID, m.createErr
}
func (m *mockRepo) Post(_ context.Context, _, _ int64) error {
	return m.postErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.JournalEntry{{ID: 1}}}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), 1, domain.JournalEntryInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestPost_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{postErr: want}}
	if err := s.Post(context.Background(), 1, 5); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
