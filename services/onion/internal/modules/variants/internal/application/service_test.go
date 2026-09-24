package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/variants/internal/domain"
)

type mockRepo struct {
	pageItems   []domain.Variant
	pageHasMore bool
	pageTotal   int64
	pageErr     error
	createID    int64
	createErr   error
	updateErr   error
	deleteErr   error
}

func (m *mockRepo) ListVariantsPage(_ context.Context, _, _ int64, _, _ int) ([]domain.Variant, bool, int64, error) {
	return m.pageItems, m.pageHasMore, m.pageTotal, m.pageErr
}
func (m *mockRepo) CreateVariant(_ context.Context, _ int64, _ domain.VariantInput) (int64, error) {
	return m.createID, m.createErr
}
func (m *mockRepo) UpdateVariant(_ context.Context, _, _ int64, _ domain.VariantInput) error {
	return m.updateErr
}
func (m *mockRepo) DeleteVariant(_ context.Context, _, _ int64) error {
	return m.deleteErr
}

func newServiceWithRepo(repo repository) *Service { return &Service{repo: repo} }

func TestVariantService_CreateVariant_HappyPath(t *testing.T) {
	s := newServiceWithRepo(&mockRepo{createID: 42})
	id, err := s.CreateVariant(context.Background(), 1, domain.VariantInput{ProductID: 10, Name: "Red Small"})
	if err != nil || id != 42 {
		t.Fatalf("id=%d err=%v", id, err)
	}
}

func TestVariantService_CreateVariant_PropagatesError(t *testing.T) {
	want := errors.New("db boom")
	s := newServiceWithRepo(&mockRepo{createErr: want})
	id, err := s.CreateVariant(context.Background(), 1, domain.VariantInput{})
	if id != 0 || !errors.Is(err, want) {
		t.Fatalf("id=%d err=%v, want 0 and %v", id, err, want)
	}
}

func TestVariantService_UpdateVariant_HappyPath(t *testing.T) {
	s := newServiceWithRepo(&mockRepo{})
	if err := s.UpdateVariant(context.Background(), 1, 5, domain.VariantInput{ProductID: 10, Name: "Updated"}); err != nil {
		t.Fatalf("UpdateVariant: %v", err)
	}
}

func TestVariantService_UpdateVariant_PropagatesError(t *testing.T) {
	want := errors.New("update fail")
	s := newServiceWithRepo(&mockRepo{updateErr: want})
	if err := s.UpdateVariant(context.Background(), 1, 5, domain.VariantInput{}); !errors.Is(err, want) {
		t.Fatalf("err=%v, want %v", err, want)
	}
}

func TestVariantService_DeleteVariant_HappyPath(t *testing.T) {
	s := newServiceWithRepo(&mockRepo{})
	if err := s.DeleteVariant(context.Background(), 1, 5); err != nil {
		t.Fatalf("DeleteVariant: %v", err)
	}
}

func TestVariantService_DeleteVariant_PropagatesError(t *testing.T) {
	want := errors.New("delete fail")
	s := newServiceWithRepo(&mockRepo{deleteErr: want})
	if err := s.DeleteVariant(context.Background(), 1, 99); !errors.Is(err, want) {
		t.Fatalf("err=%v, want %v", err, want)
	}
}

func TestVariantService_ListVariantsPage_PreservesPagination(t *testing.T) {
	items := []domain.Variant{
		{ID: 1, Name: "Small Red", BusinessID: 1},
		{ID: 2, Name: "Large Blue", BusinessID: 1},
	}
	s := newServiceWithRepo(&mockRepo{pageItems: items, pageHasMore: true, pageTotal: 101})
	page, err := s.ListVariantsPage(context.Background(), 1, 0, 2, 20)
	if err != nil {
		t.Fatalf("ListVariantsPage: %v", err)
	}
	if len(page.Items) != 2 || !page.HasMore || page.Total != 101 || page.Limit != 2 || page.Offset != 20 {
		t.Fatalf("unexpected page: %+v", page)
	}
}

func TestVariantService_ListVariants_ReturnsEmpty(t *testing.T) {
	s := newServiceWithRepo(&mockRepo{pageItems: []domain.Variant{}})
	items, err := s.ListVariants(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("ListVariants: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty slice, got %d items", len(items))
	}
}

func TestVariantService_ListVariants_PreservesFields(t *testing.T) {
	items := []domain.Variant{{ID: 7, Name: "Green Medium", OwnerID: 1, ProductID: 10, SalePrice: 29.99}}
	s := newServiceWithRepo(&mockRepo{pageItems: items})
	got, err := s.ListVariants(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("ListVariants: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Green Medium" || got[0].SalePrice != 29.99 {
		t.Fatalf("fields not preserved: %+v", got)
	}
}

func TestVariantService_ListVariantsPage_PropagatesError(t *testing.T) {
	want := errors.New("list fail")
	s := newServiceWithRepo(&mockRepo{pageErr: want})
	if _, err := s.ListVariantsPage(context.Background(), 1, 10, 10, 0); !errors.Is(err, want) {
		t.Fatalf("err=%v, want %v", err, want)
	}
}

func TestPaginateVariants_ClampsBounds(t *testing.T) {
	limit, offset := paginateVariants(-1, 0)
	if limit != 50 || offset != 0 {
		t.Fatalf("defaults limit=%d offset=%d", limit, offset)
	}
	limit, offset = paginateVariants(10, 1000)
	if limit != 500 || offset != 10 {
		t.Fatalf("bounds limit=%d offset=%d", limit, offset)
	}
}
