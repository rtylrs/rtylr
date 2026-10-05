package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/categories/internal/domain"
)

type mockRepo struct {
	createID   int64
	createErr  error
	updateErr  error
	deleteErr  error
	categories []domain.Category
	total      int64
	listGot    struct {
		businessID    int64
		limit, offset int32
	}
}

func (m *mockRepo) ListCategories(_ context.Context, _ int64) ([]domain.Category, error) {
	return nil, nil
}

func (m *mockRepo) ListCategoriesPaginated(_ context.Context, businessID int64, limit, offset int32) ([]domain.Category, error) {
	m.listGot.businessID = businessID
	m.listGot.limit = limit
	m.listGot.offset = offset
	return m.categories, nil
}

func (m *mockRepo) CountCategories(_ context.Context, _ int64) (int64, error) {
	return m.total, nil
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

func TestListCategoriesPage_HonorsBoundsAndReturnsTotal(t *testing.T) {
	repo := &mockRepo{
		categories: []domain.Category{{ID: 101, Name: "Beyond first page"}},
		total:      150,
	}
	s := &Service{repo: repo}

	page, err := s.ListCategoriesPage(context.Background(), 7, 25, 100)
	if err != nil {
		t.Fatalf("ListCategoriesPage: %v", err)
	}
	if repo.listGot.businessID != 7 || repo.listGot.limit != 25 || repo.listGot.offset != 100 {
		t.Fatalf("repository bounds = %+v", repo.listGot)
	}
	if page.Total != 150 || page.Limit != 25 || page.Offset != 100 || len(page.Categories) != 1 {
		t.Fatalf("page = %+v", page)
	}
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
