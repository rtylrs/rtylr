package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/warehouses/internal/domain"
)

type mockRepo struct {
	listResult []domain.Warehouse
	listErr    error
	createID   int64
	createErr  error
	updateErr  error
	deleteErr  error
}

func (m *mockRepo) ListWarehouses(_ context.Context, _ int64) ([]domain.Warehouse, error) {
	return m.listResult, m.listErr
}

func (m *mockRepo) GetWarehouseByID(_ context.Context, _, _ int64) (domain.Warehouse, error) {
	return domain.Warehouse{}, nil
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

func TestListWarehouses_ReturnsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{listResult: []domain.Warehouse{}}}
	got, err := s.ListWarehouses(context.Background(), 1)
	if err != nil || len(got) != 0 {
		t.Errorf("got %v err %v", got, err)
	}
}

func TestListWarehouses_ReturnsItems(t *testing.T) {
	items := []domain.Warehouse{{ID: 1, Name: "Main"}, {ID: 2, Name: "Secondary"}}
	s := &Service{repo: &mockRepo{listResult: items}}
	got, err := s.ListWarehouses(context.Background(), 1)
	if err != nil || len(got) != 2 {
		t.Errorf("got %v err %v", got, err)
	}
}

func TestListWarehouses_PropagatesError(t *testing.T) {
	want := errors.New("db fail")
	s := &Service{repo: &mockRepo{listErr: want}}
	if _, err := s.ListWarehouses(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateWarehouse_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.UpdateWarehouse(context.Background(), 1, 5, domain.WarehouseInput{Name: "New Name"}); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestDeleteWarehouse_PropagatesError(t *testing.T) {
	want := errors.New("not found")
	s := &Service{repo: &mockRepo{deleteErr: want}}
	if err := s.DeleteWarehouse(context.Background(), 1, 99); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestCreateWarehouse_ZeroIDOnError(t *testing.T) {
	s := &Service{repo: &mockRepo{createErr: errors.New("fail")}}
	id, _ := s.CreateWarehouse(context.Background(), 1, domain.WarehouseInput{})
	if id != 0 {
		t.Errorf("expected 0 on error, got %d", id)
	}
}

func TestListWarehouses_NamePreserved(t *testing.T) {
	items := []domain.Warehouse{{ID: 7, Name: "Warehouse A", City: "Beirut"}}
	s := &Service{repo: &mockRepo{listResult: items}}
	got, err := s.ListWarehouses(context.Background(), 1)
	if err != nil || got[0].Name != "Warehouse A" || got[0].City != "Beirut" {
		t.Errorf("fields not preserved: %v err=%v", got, err)
	}
}

func TestCreateWarehouse_WithFullInput(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 20}}
	id, err := s.CreateWarehouse(context.Background(), 1, domain.WarehouseInput{
		Name:    "Central",
		Address: "123 Main St",
		City:    "London",
		Country: "UK",
	})
	if err != nil || id != 20 {
		t.Errorf("id=%d err=%v", id, err)
	}
}
