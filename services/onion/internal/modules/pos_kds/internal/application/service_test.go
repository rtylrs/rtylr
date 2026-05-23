package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_kds/internal/domain"
)

type mockRepo struct {
	orders  []domain.KDSOrder
	listErr error
	upErr   error
	upIDErr error
}

func (m *mockRepo) ListActiveOrders(_ context.Context, _ int64) ([]domain.KDSOrder, error) {
	return m.orders, m.listErr
}

func (m *mockRepo) UpdateItemStatus(_ context.Context, _ domain.UpdateItemStatusInput) error {
	return m.upErr
}

func (m *mockRepo) UpdateItemStatusByItemID(_ context.Context, _, _ int64, _ string) error {
	return m.upIDErr
}

func TestListActiveOrdersByStation_NoStationReturnsAll(t *testing.T) {
	m := &mockRepo{orders: []domain.KDSOrder{{KitchenNotes: "grill"}, {KitchenNotes: "fryer"}}}
	s := &Service{repo: m}
	got, _ := s.ListActiveOrdersByStation(context.Background(), 1, "")
	if len(got) != 2 {
		t.Errorf("expected 2, got %d", len(got))
	}
}

func TestListActiveOrdersByStation_FiltersByStation(t *testing.T) {
	m := &mockRepo{orders: []domain.KDSOrder{
		{KitchenNotes: "Grill order"},
		{KitchenNotes: "Salad bar"},
		{KitchenNotes: "GRILL extra"},
	}}
	s := &Service{repo: m}
	got, _ := s.ListActiveOrdersByStation(context.Background(), 1, "grill")
	if len(got) != 2 {
		t.Errorf("expected 2 grill orders (case-insensitive), got %d", len(got))
	}
}

func TestListActiveOrdersByStation_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{listErr: want}}
	if _, err := s.ListActiveOrdersByStation(context.Background(), 1, "grill"); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateItemStatus_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.UpdateItemStatus(context.Background(), domain.UpdateItemStatusInput{}); err != nil {
		t.Errorf("unexpected: %v", err)
	}
	if err := s.UpdateItemStatusByItemID(context.Background(), 1, 5, "ready"); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}
