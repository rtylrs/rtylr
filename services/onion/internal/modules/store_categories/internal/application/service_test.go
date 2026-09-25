package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/store_categories/internal/domain"
)

type storeCatRepo interface {
	List(ctx context.Context, ownerID int64) ([]domain.StoreCategory, error)
	Create(ctx context.Context, ownerID int64, input domain.StoreCategoryInput) (int64, error)
	Update(ctx context.Context, ownerID, id int64, input domain.StoreCategoryInput) error
	Delete(ctx context.Context, ownerID, id int64) error
}

type mockCatRepo struct {
	list      []domain.StoreCategory
	listErr   error
	createID  int64
	createErr error
	updateErr error
	deleteErr error
}

func (m *mockCatRepo) List(_ context.Context, _ int64) ([]domain.StoreCategory, error) {
	return m.list, m.listErr
}

func (m *mockCatRepo) Create(_ context.Context, _ int64, _ domain.StoreCategoryInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockCatRepo) Update(_ context.Context, _, _ int64, _ domain.StoreCategoryInput) error {
	return m.updateErr
}
func (m *mockCatRepo) Delete(_ context.Context, _, _ int64) error { return m.deleteErr }

type catTestService struct{ repo storeCatRepo }

func (s *catTestService) List(ctx context.Context, ownerID int64) ([]domain.StoreCategory, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *catTestService) Create(ctx context.Context, ownerID int64, input domain.StoreCategoryInput) (int64, error) {
	return s.repo.Create(ctx, ownerID, input)
}

func (s *catTestService) Update(ctx context.Context, ownerID, id int64, input domain.StoreCategoryInput) error {
	return s.repo.Update(ctx, ownerID, id, input)
}

func (s *catTestService) Delete(ctx context.Context, ownerID, id int64) error {
	return s.repo.Delete(ctx, ownerID, id)
}

func TestStoreCat_List_Delegates(t *testing.T) {
	m := &mockCatRepo{list: []domain.StoreCategory{{ID: 1, Name: "Apparel"}, {ID: 2}}}
	s := &catTestService{repo: m}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d items, want 2", len(got))
	}
}

func TestStoreCat_List_PropagatesError(t *testing.T) {
	want := errors.New("db error")
	s := &catTestService{repo: &mockCatRepo{listErr: want}}
	_, err := s.List(context.Background(), 1)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestStoreCat_List_Empty(t *testing.T) {
	s := &catTestService{repo: &mockCatRepo{list: []domain.StoreCategory{}}}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty list, got %d", len(got))
	}
}

func TestStoreCat_Create_ReturnsID(t *testing.T) {
	s := &catTestService{repo: &mockCatRepo{createID: 99}}
	id, err := s.Create(context.Background(), 1, domain.StoreCategoryInput{Name: "Shoes"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if id != 99 {
		t.Errorf("id = %d, want 99", id)
	}
}

func TestStoreCat_Create_PropagatesError(t *testing.T) {
	want := errors.New("insert failed")
	s := &catTestService{repo: &mockCatRepo{createErr: want}}
	_, err := s.Create(context.Background(), 1, domain.StoreCategoryInput{})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestStoreCat_Update_HappyPath(t *testing.T) {
	s := &catTestService{repo: &mockCatRepo{}}
	if err := s.Update(context.Background(), 1, 5, domain.StoreCategoryInput{}); err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

func TestStoreCat_Update_PropagatesError(t *testing.T) {
	want := errors.New("update failed")
	s := &catTestService{repo: &mockCatRepo{updateErr: want}}
	if err := s.Update(context.Background(), 1, 5, domain.StoreCategoryInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestStoreCat_Delete_HappyPath(t *testing.T) {
	s := &catTestService{repo: &mockCatRepo{}}
	if err := s.Delete(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

func TestStoreCat_Delete_PropagatesError(t *testing.T) {
	want := errors.New("delete failed")
	s := &catTestService{repo: &mockCatRepo{deleteErr: want}}
	if err := s.Delete(context.Background(), 1, 5); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestStoreCat_Create_ZeroOwner(t *testing.T) {
	s := &catTestService{repo: &mockCatRepo{createID: 1}}
	id, err := s.Create(context.Background(), 0, domain.StoreCategoryInput{Name: "root"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if id != 1 {
		t.Errorf("id = %d, want 1", id)
	}
}
