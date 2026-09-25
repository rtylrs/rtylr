package api

import (
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/stock_transfers/internal/domain"
	"github.com/voxire/onion/internal/platform/stock_core"
)

func stockTransferMutationError(err error) error {
	switch {
	case errors.Is(err, domain.ErrStockTransferNotFound):
		return stockTransferStatusError(http.StatusNotFound, "Stock transfer not found")
	case errors.Is(err, domain.ErrStockTransferReferenceNotFound):
		return stockTransferStatusError(http.StatusNotFound, "Stock transfer reference not found")
	case errors.Is(err, domain.ErrInvalidStockTransfer):
		return stockTransferStatusError(http.StatusBadRequest, err.Error())
	case errors.Is(err, stock_core.ErrInsufficientStock):
		return stockTransferStatusError(http.StatusConflict, "Insufficient stock")
	default:
		return stockTransferStatusError(http.StatusInternalServerError, "Failed to update stock transfer status")
	}
}
