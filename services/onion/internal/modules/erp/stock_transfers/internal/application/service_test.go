package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/erp/stock_transfers/internal/domain"
)

type mockRepo struct {
	createID    int64
	createErr   error
	createCalls int
	updErr      error
}

func (m *mockRepo) ListStockTransfers(_ context.Context, _ int64, _, _ int) ([]domain.StockTransfer, bool, error) {
	return nil, false, nil
}

func (m *mockRepo) GetByID(_ context.Context, _, _ int64) (domain.StockTransfer, error) {
	return domain.StockTransfer{}, nil
}

func (m *mockRepo) CreateStockTransfer(_ context.Context, _ int64, _ domain.StockTransferInput, _ ...audit.Event) (int64, error) {
	m.createCalls++
	return m.createID, m.createErr
}

func (m *mockRepo) UpdateStockTransferStatus(_ context.Context, _, _ int64, _ domain.StockTransferStatus, _ ...audit.Event) error {
	return m.updErr
}

func TestCreateStockTransfer_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.CreateStockTransfer(context.Background(), 1, validStockTransferInput()); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestCreateStockTransfer_RejectsNonPositiveQuantityBeforeRepository(t *testing.T) {
	repo := &mockRepo{}
	s := &Service{repo: repo}
	_, err := s.CreateStockTransfer(context.Background(), 1, domain.StockTransferInput{
		FromWarehouseID: 1,
		ToWarehouseID:   2,
		Items: []domain.StockTransferItemInput{
			{ProductID: 10, Quantity: 0},
		},
	})
	if !errors.Is(err, domain.ErrInvalidStockTransfer) {
		t.Fatalf("expected ErrInvalidStockTransfer, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("repository CreateStockTransfer called %d times, want 0", repo.createCalls)
	}
}

func validStockTransferInput() domain.StockTransferInput {
	return domain.StockTransferInput{
		FromWarehouseID: 1,
		ToWarehouseID:   2,
		Items: []domain.StockTransferItemInput{
			{ProductID: 10, Quantity: 1},
		},
	}
}

func TestUpdateStockTransferStatus_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{updErr: want}}
	if err := s.UpdateStockTransferStatus(context.Background(), 1, 5, ""); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
