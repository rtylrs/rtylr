package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/categories/internal/domain"
)

type mockRepo struct {
	createID  int64
	createErr error
	updateErr error
	deleteErr error
}

func (m *mockRepo) ListCategories(_ context.Context, _ int64) ([]domain.Category, error) {
	return nil, nil
}

func (m *mockRepo) ListCategoriesPaginated(_ context.Context, _ int64, _, _ int32) ([]domain.Category, error) {
	return nil, nil
}

func (m *mockRepo) CreateCategory(_ context.Context, _ int64, _ domain.CategoryInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) UpdateCategory(_ context.Context, _, _ int64, _ domain.CategoryInput) error {
	return m.updateErr
}

func (m *mockRepo) DeleteCategory(_ context.Context, _, _ int64) error {
	return m.deleteErr
}

func TestCreateCategory_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.CreateCategory(context.Background(), 1, domain.CategoryInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateCategory_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{updateErr: want}}
	if err := s.UpdateCategory(context.Background(), 1, 5, domain.CategoryInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestDeleteCategory_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{deleteErr: want}}
	if err := s.DeleteCategory(context.Background(), 1, 5); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestCreateCategory_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 7}}
	id, err := s.CreateCategory(context.Background(), 1, domain.CategoryInput{})
	if err != nil || id != 7 {
		t.Errorf("id=%d err=%v", id, err)
	}
}
