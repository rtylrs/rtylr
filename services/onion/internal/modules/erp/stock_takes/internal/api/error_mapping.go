package api

import (
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/erp/internal/stock"
	"github.com/voxire/onion/internal/modules/erp/stock_takes/internal/domain"
)

func stockTakeMutationError(err error) error {
	switch {
	case errors.Is(err, domain.ErrStockTakeNotFound):
		return stockTakeStatusError(http.StatusNotFound, "Stock take not found")
	case errors.Is(err, domain.ErrStockTakeReferenceNotFound):
		return stockTakeStatusError(http.StatusNotFound, "Stock take reference not found")
	case errors.Is(err, domain.ErrStockTakeIncomplete):
		return stockTakeStatusError(http.StatusConflict, "Every stock-take item requires a physical count")
	case errors.Is(err, stock.ErrInsufficientStock):
		return stockTakeStatusError(http.StatusConflict, "Insufficient stock")
	default:
		return stockTakeStatusError(http.StatusInternalServerError, "Failed to update stock take")
	}
}
