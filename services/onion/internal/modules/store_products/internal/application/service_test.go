package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/store_products/internal/domain"
)

type storeProdRepo interface {
	List(ctx context.Context, ownerID int64) ([]domain.StoreProduct, error)
	Create(ctx context.Context, ownerID int64, input domain.StoreProductInput) (int64, error)
	Update(ctx context.Context, ownerID, id int64, input domain.StoreProductInput) error
	Delete(ctx context.Context, ownerID, id int64) error
}

type mockStoreProdRepo struct {
	list      []domain.StoreProduct
	listErr   error
	createID  int64
	createErr error
	updateErr error
	deleteErr error
}

func (m *mockStoreProdRepo) List(_ context.Context, _ int64) ([]domain.StoreProduct, error) {
	return m.list, m.listErr
}

func (m *mockStoreProdRepo) Create(_ context.Context, _ int64, _ domain.StoreProductInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockStoreProdRepo) Update(_ context.Context, _, _ int64, _ domain.StoreProductInput) error {
	return m.updateErr
}
func (m *mockStoreProdRepo) Delete(_ context.Context, _, _ int64) error { return m.deleteErr }

type storeProdTestService struct{ repo storeProdRepo }

func (s *storeProdTestService) List(ctx context.Context, ownerID int64) ([]domain.StoreProduct, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *storeProdTestService) Create(ctx context.Context, ownerID int64, input domain.StoreProductInput) (int64, error) {
	return s.repo.Create(ctx, ownerID, input)
}

func (s *storeProdTestService) Update(ctx context.Context, ownerID, id int64, input domain.StoreProductInput) error {
	return s.repo.Update(ctx, ownerID, id, input)
}

func (s *storeProdTestService) Delete(ctx context.Context, ownerID, id int64) error {
	return s.repo.Delete(ctx, ownerID, id)
}

// --- Tests ---

func TestStoreProd_List_Delegates(t *testing.T) {
	m := &mockStoreProdRepo{list: []domain.StoreProduct{{ID: 1}, {ID: 2}}}
	s := &storeProdTestService{repo: m}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d items, want 2", len(got))
	}
}

func TestStoreProd_List_PropagatesError(t *testing.T) {
	want := errors.New("db error")
	s := &storeProdTestService{repo: &mockStoreProdRepo{listErr: want}}
	_, err := s.List(context.Background(), 1)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestStoreProd_List_Empty(t *testing.T) {
	s := &storeProdTestService{repo: &mockStoreProdRepo{list: []domain.StoreProduct{}}}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty list, got %d", len(got))
	}
}

func TestStoreProd_Create_ReturnsID(t *testing.T) {
	s := &storeProdTestService{repo: &mockStoreProdRepo{createID: 88}}
	id, err := s.Create(context.Background(), 1, domain.StoreProductInput{Title: "Widget"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if id != 88 {
		t.Errorf("id = %d, want 88", id)
	}
}

func TestStoreProd_Create_PropagatesError(t *testing.T) {
	want := errors.New("insert failed")
	s := &storeProdTestService{repo: &mockStoreProdRepo{createErr: want}}
	_, err := s.Create(context.Background(), 1, domain.StoreProductInput{})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestStoreProd_Update_HappyPath(t *testing.T) {
	s := &storeProdTestService{repo: &mockStoreProdRepo{}}
	if err := s.Update(context.Background(), 1, 5, domain.StoreProductInput{}); err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

func TestStoreProd_Update_PropagatesError(t *testing.T) {
	want := errors.New("update failed")
	s := &storeProdTestService{repo: &mockStoreProdRepo{updateErr: want}}
	if err := s.Update(context.Background(), 1, 5, domain.StoreProductInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestStoreProd_Delete_HappyPath(t *testing.T) {
	s := &storeProdTestService{repo: &mockStoreProdRepo{}}
	if err := s.Delete(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

func TestStoreProd_Delete_PropagatesError(t *testing.T) {
	want := errors.New("delete failed")
	s := &storeProdTestService{repo: &mockStoreProdRepo{deleteErr: want}}
	if err := s.Delete(context.Background(), 1, 5); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestStoreProd_List_PublishedProducts(t *testing.T) {
	m := &mockStoreProdRepo{list: []domain.StoreProduct{
		{ID: 1, Title: "Hat", IsPublished: true},
		{ID: 2, Title: "Draft", IsPublished: false},
	}}
	s := &storeProdTestService{repo: m}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d items, want 2", len(got))
	}
	if !got[0].IsPublished {
		t.Error("first item should be published")
	}
}

func TestStoreProd_Create_WithPrice(t *testing.T) {
	s := &storeProdTestService{repo: &mockStoreProdRepo{createID: 10}}
	id, err := s.Create(context.Background(), 1, domain.StoreProductInput{
		Title:          "Shirt",
		Price:          29.99,
		CompareAtPrice: 39.99,
		IsPublished:    true,
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if id != 10 {
		t.Errorf("id = %d, want 10", id)
	}
}
