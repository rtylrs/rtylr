package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_sessions/internal/domain"
)

type mockRepo struct {
	list      []domain.Session
	sess      *domain.Session
	openID    int64
	listErr   error
	getErr    error
	openErr   error
	closeErr  error
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.Session, error) {
	return m.list, m.listErr
}
func (m *mockRepo) Get(_ context.Context, _, _ int64) (*domain.Session, error) {
	return m.sess, m.getErr
}
func (m *mockRepo) Open(_ context.Context, _ domain.OpenSessionInput) (int64, error) {
	return m.openID, m.openErr
}
func (m *mockRepo) Close(_ context.Context, _ domain.CloseSessionInput) error {
	return m.closeErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Session{{}}}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestOpen_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{openErr: want}}
	if _, err := s.Open(context.Background(), domain.OpenSessionInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestClose_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{closeErr: want}}
	if err := s.Close(context.Background(), domain.CloseSessionInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGet_HappyPath(t *testing.T) {
	want := &domain.Session{ID: 7}
	s := &Service{repo: &mockRepo{sess: want}}
	got, _ := s.Get(context.Background(), 1, 7)
	if got != want {
		t.Errorf("got %+v", got)
	}
}
