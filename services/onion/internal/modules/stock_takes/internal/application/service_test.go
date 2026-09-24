package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/stock_takes/internal/domain"
	"github.com/voxire/onion/internal/platform/audit"
)

type mockRepo struct {
	createID  int64
	createErr error
	updErr    error
}

func (m *mockRepo) ListStockTakes(_ context.Context, _ int64, _, _ int) ([]domain.StockTake, bool, error) {
	return nil, false, nil
}

func (m *mockRepo) CreateStockTake(_ context.Context, _ int64, _ domain.StockTakeInput, _ ...audit.Event) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) UpdateStockTake(_ context.Context, _ int64, _ domain.StockTakeUpdateInput, _ ...audit.Event) error {
	return m.updErr
}

func TestCreateStockTake_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.CreateStockTake(context.Background(), 1, domain.StockTakeInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateStockTake_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{updErr: want}}
	if err := s.UpdateStockTake(context.Background(), 1, domain.StockTakeUpdateInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
