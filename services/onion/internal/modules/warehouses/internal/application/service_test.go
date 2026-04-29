package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/warehouses/internal/domain"
)

type mockRepo struct {
	createID  int64
	createErr error
	updateErr error
	deleteErr error
}

func (m *mockRepo) ListWarehouses(_ context.Context, _ int64) ([]domain.Warehouse, error) {
	return nil, nil
}
func (m *mockRepo) CreateWarehouse(_ context.Context, _ int64, _ domain.WarehouseInput) (int64, error) {
	return m.createID, m.createErr
}
func (m *mockRepo) UpdateWarehouse(_ context.Context, _, _ int64, _ domain.WarehouseInput) error {
	return m.updateErr
}
func (m *mockRepo) DeleteWarehouse(_ context.Context, _, _ int64) error {
	return m.deleteErr
}

func TestCreateWarehouse_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.CreateWarehouse(context.Background(), 1, domain.WarehouseInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateWarehouse_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{updateErr: want}}
	if err := s.UpdateWarehouse(context.Background(), 1, 5, domain.WarehouseInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestDeleteWarehouse_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.DeleteWarehouse(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestCreateWarehouse_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 11}}
	id, err := s.CreateWarehouse(context.Background(), 1, domain.WarehouseInput{})
	if err != nil || id != 11 {
		t.Errorf("id=%d err=%v", id, err)
	}
}
