package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/brands/internal/domain"
)

type mockRepo struct {
	createID  int64
	createErr error
	updateErr error
	deleteErr error
	createGot domain.BrandInput
	updateGot struct {
		owner, id int64
		input     domain.BrandInput
	}
	deleteGot struct{ owner, id int64 }
}

func (m *mockRepo) ListBrands(_ context.Context, _ int64) ([]domain.Brand, error) {
	return nil, nil
}

func (m *mockRepo) ListBrandsPaginated(_ context.Context, _ int64, _, _ int32) ([]domain.Brand, error) {
	return nil, nil
}

func (m *mockRepo) CreateBrand(_ context.Context, _ int64, input domain.BrandInput) (int64, error) {
	m.createGot = input
	return m.createID, m.createErr
}

func (m *mockRepo) UpdateBrand(_ context.Context, owner, id int64, input domain.BrandInput) error {
	m.updateGot.owner = owner
	m.updateGot.id = id
	m.updateGot.input = input
	return m.updateErr
}

func (m *mockRepo) DeleteBrand(_ context.Context, owner, id int64) error {
	m.deleteGot.owner = owner
	m.deleteGot.id = id
	return m.deleteErr
}

func TestCreateBrand_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.CreateBrand(context.Background(), 1, domain.BrandInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateBrand_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{updateErr: want}}
	if err := s.UpdateBrand(context.Background(), 1, 5, domain.BrandInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestDeleteBrand_HappyPathDelegatesIDs(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	if err := s.DeleteBrand(context.Background(), 7, 42); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if m.deleteGot.owner != 7 || m.deleteGot.id != 42 {
		t.Errorf("delete got %+v", m.deleteGot)
	}
}
