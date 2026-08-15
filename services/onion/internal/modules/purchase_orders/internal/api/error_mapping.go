package api

import (
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/purchase_orders/internal/domain"
	"github.com/voxire/onion/internal/platform/stock_core"
)

func purchaseOrderMutationError(err error, fallback string) error {
	switch {
	case errors.Is(err, domain.ErrPurchaseOrderNotFound):
		return purchaseOrderStatusError(http.StatusNotFound, "Purchase order not found")
	case errors.Is(err, domain.ErrPurchaseOrderReferenceNotFound):
		return purchaseOrderStatusError(http.StatusNotFound, "Purchase order reference not found")
	case errors.Is(err, domain.ErrPurchaseOrderAlreadyReceived):
		return purchaseOrderStatusError(http.StatusConflict, "Received purchase orders cannot be reopened")
	case errors.Is(err, stock_core.ErrInsufficientStock):
		return purchaseOrderStatusError(http.StatusConflict, "Insufficient stock")
	default:
		return purchaseOrderStatusError(http.StatusInternalServerError, fallback)
	}
}
