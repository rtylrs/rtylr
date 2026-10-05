package api

import (
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/erp/internal/stock"
	"github.com/voxire/onion/internal/modules/erp/stock_transfers/internal/domain"
)

func stockTransferMutationError(err error) error {
	switch {
	case errors.Is(err, domain.ErrStockTransferNotFound):
		return stockTransferStatusError(http.StatusNotFound, "Stock transfer not found")
	case errors.Is(err, domain.ErrStockTransferReferenceNotFound):
		return stockTransferStatusError(http.StatusNotFound, "Stock transfer reference not found")
	case errors.Is(err, domain.ErrInvalidStockTransfer):
		return stockTransferStatusError(http.StatusBadRequest, err.Error())
	case errors.Is(err, stock.ErrInsufficientStock):
		return stockTransferStatusError(http.StatusConflict, "Insufficient stock")
	default:
		return stockTransferStatusError(http.StatusInternalServerError, "Failed to update stock transfer status")
	}
}
