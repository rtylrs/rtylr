package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/notifications/internal/domain"
)

type mockRepo struct {
	list  []domain.Notification
	mrErr error
	maErr error
}

func (m *mockRepo) List(_ context.Context, _, _ int64) ([]domain.Notification, error) {
	return m.list, nil
}

func (m *mockRepo) MarkRead(_ context.Context, _, _ int64) error {
	return m.mrErr
}

func (m *mockRepo) MarkAllRead(_ context.Context, _ int64) error {
	return m.maErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Notification{{}}}}
	got, _ := s.List(context.Background(), 1, 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestMarkRead_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{mrErr: want}}
	if err := s.MarkRead(context.Background(), 1, 1, 5); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestMarkAllRead_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.MarkAllRead(context.Background(), 1, 1); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}
