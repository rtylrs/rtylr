package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/suppliers/internal/domain"
	"github.com/voxire/onion/internal/platform/audit"
)

type mockRepo struct {
	createID  int64
	createErr error
	updateErr error
	deleteErr error
}

func (m *mockRepo) ListSuppliers(_ context.Context, _ int64, _, _ int) ([]domain.Supplier, bool, error) {
	return nil, false, nil
}

func (m *mockRepo) GetSupplierByID(_ context.Context, _, _ int64) (domain.Supplier, error) {
	return domain.Supplier{}, nil
}

func (m *mockRepo) CreateSupplier(_ context.Context, _ int64, _ domain.SupplierInput, _ ...audit.Event) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) UpdateSupplier(_ context.Context, _, _ int64, _ domain.SupplierInput, _ ...audit.Event) error {
	return m.updateErr
}

func (m *mockRepo) DeleteSupplier(_ context.Context, _, _ int64, _ ...audit.Event) error {
	return m.deleteErr
}

func TestCreateSupplier_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.CreateSupplier(context.Background(), 1, domain.SupplierInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateSupplier_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{updateErr: want}}
	if err := s.UpdateSupplier(context.Background(), 1, 5, domain.SupplierInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestDeleteSupplier_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.DeleteSupplier(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestCreateSupplier_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 9}}
	id, err := s.CreateSupplier(context.Background(), 1, domain.SupplierInput{})
	if err != nil || id != 9 {
		t.Errorf("id=%d err=%v", id, err)
	}
}
