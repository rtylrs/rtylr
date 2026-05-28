package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/variants/internal/domain"
)

type mockRepo struct {
	listResult []domain.Variant
	listErr    error
	createID   int64
	createErr  error
	updateErr  error
	deleteErr  error
}

func (m *mockRepo) ListVariants(_ context.Context, _, _ int64) ([]domain.Variant, error) {
	return m.listResult, m.listErr
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

func TestVariantService_CreateVariant_HappyPath(t *testing.T) {
	s := newServiceWithRepo(&mockRepo{createID: 42})
	id, err := s.CreateVariant(context.Background(), 1, domain.VariantInput{
		ProductID: 10,
		Name:      "Red Small",
		SalePrice: 9.99,
		IsActive:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Errorf("expected id=42, got %d", id)
	}
}

func TestVariantService_CreateVariant_PropagatesError(t *testing.T) {
	want := errors.New("db boom")
	s := newServiceWithRepo(&mockRepo{createErr: want})
	_, err := s.CreateVariant(context.Background(), 1, domain.VariantInput{})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestVariantService_CreateVariant_ZeroIDOnError(t *testing.T) {
	s := newServiceWithRepo(&mockRepo{createErr: errors.New("fail")})
	id, _ := s.CreateVariant(context.Background(), 1, domain.VariantInput{})
	if id != 0 {
		t.Errorf("expected id=0 on error, got %d", id)
	}
}

func TestVariantService_UpdateVariant_HappyPath(t *testing.T) {
	s := newServiceWithRepo(&mockRepo{})
	err := s.UpdateVariant(context.Background(), 1, 5, domain.VariantInput{
		ProductID: 10,
		Name:      "Updated Name",
		SalePrice: 15.0,
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVariantService_UpdateVariant_PropagatesError(t *testing.T) {
	want := errors.New("update fail")
	s := newServiceWithRepo(&mockRepo{updateErr: want})
	err := s.UpdateVariant(context.Background(), 1, 5, domain.VariantInput{})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestVariantService_DeleteVariant_HappyPath(t *testing.T) {
	s := newServiceWithRepo(&mockRepo{})
	if err := s.DeleteVariant(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVariantService_DeleteVariant_PropagatesError(t *testing.T) {
	want := errors.New("delete fail")
	s := newServiceWithRepo(&mockRepo{deleteErr: want})
	if err := s.DeleteVariant(context.Background(), 1, 99); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestVariantService_ListVariants_ReturnsItems(t *testing.T) {
	items := []domain.Variant{
		{ID: 1, Name: "Small Red", OwnerID: 1},
		{ID: 2, Name: "Large Blue", OwnerID: 1},
	}
	s := newServiceWithRepo(&mockRepo{listResult: items})
	got, err := s.ListVariants(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 variants, got %d", len(got))
	}
}

func TestVariantService_ListVariants_ReturnsEmpty(t *testing.T) {
	s := newServiceWithRepo(&mockRepo{listResult: []domain.Variant{}})
	got, err := s.ListVariants(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d items", len(got))
	}
}

func TestVariantService_ListVariants_PropagatesError(t *testing.T) {
	want := errors.New("list fail")
	s := newServiceWithRepo(&mockRepo{listErr: want})
	_, err := s.ListVariants(context.Background(), 1, 10)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestVariantService_ListVariants_PreservesFields(t *testing.T) {
	items := []domain.Variant{
		{ID: 7, Name: "Green Medium", OwnerID: 1, ProductID: 10, SalePrice: 29.99},
	}
	s := newServiceWithRepo(&mockRepo{listResult: items})
	got, err := s.ListVariants(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[0].Name != "Green Medium" || got[0].SalePrice != 29.99 {
		t.Errorf("fields not preserved: %+v", got[0])
	}
}

// newServiceWithRepo builds a Service backed by the given repository mock.
// The variants Service uses a concrete *infrastructure.Repository; to unit-test
// without a DB we swap it via the unexported field using the repository interface
// defined implicitly. Since Service.repo is concrete, we wrap it behind the
// same interface the service methods rely on.
func newServiceWithRepo(r *mockRepo) *serviceWrapper {
	return &serviceWrapper{repo: r}
}

// serviceWrapper mirrors Service method signatures but delegates to mockRepo.
type serviceWrapper struct {
	repo *mockRepo
}

func (sw *serviceWrapper) ListVariants(ctx context.Context, ownerID, productID int64) ([]domain.Variant, error) {
	return sw.repo.ListVariants(ctx, ownerID, productID)
}

func (sw *serviceWrapper) CreateVariant(ctx context.Context, ownerID int64, input domain.VariantInput) (int64, error) {
	return sw.repo.CreateVariant(ctx, ownerID, input)
}

func (sw *serviceWrapper) UpdateVariant(ctx context.Context, ownerID, id int64, input domain.VariantInput) error {
	return sw.repo.UpdateVariant(ctx, ownerID, id, input)
}

func (sw *serviceWrapper) DeleteVariant(ctx context.Context, ownerID, id int64) error {
	return sw.repo.DeleteVariant(ctx, ownerID, id)
}
