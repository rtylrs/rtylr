package infrastructure

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/stock_transfers/internal/domain"
)

func TestCreateStockTransfer_RejectsInvalidInputBeforeDatabase(t *testing.T) {
	repo := NewRepository(nil)
	_, err := repo.CreateStockTransfer(context.Background(), 1, domain.StockTransferInput{
		FromWarehouseID: 1,
		ToWarehouseID:   2,
		Items:           []domain.StockTransferItemInput{{ProductID: 3, Quantity: -1}},
	})
	if !errors.Is(err, domain.ErrInvalidStockTransfer) {
		t.Fatalf("expected ErrInvalidStockTransfer, got %v", err)
	}
}
