package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/documents/internal/domain"
)

type mockRepo struct {
	list      []domain.Document
	createID  int64
	createErr error
	delErr    error
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.Document, error) {
	return m.list, nil
}

func (m *mockRepo) ListPaginated(_ context.Context, _ int64, _, _ int32) ([]domain.Document, error) {
	return m.list, nil
}

func (m *mockRepo) Create(_ context.Context, _ int64, _ domain.DocumentInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) Delete(_ context.Context, _, _ int64) error {
	return m.delErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Document{{ID: 1, Name: "x"}}}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), 1, domain.DocumentInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestDelete_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Delete(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestDelete_PropagatesNotFound(t *testing.T) {
	s := &Service{repo: &mockRepo{delErr: domain.ErrDocumentNotFound}}
	if err := s.Delete(context.Background(), 1, 5); !errors.Is(err, domain.ErrDocumentNotFound) {
		t.Errorf("Delete error = %v, want ErrDocumentNotFound", err)
	}
}
