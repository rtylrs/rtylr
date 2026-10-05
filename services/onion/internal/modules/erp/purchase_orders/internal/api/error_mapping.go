package api

import (
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/erp/internal/stock"
	"github.com/voxire/onion/internal/modules/erp/purchase_orders/internal/domain"
)

func purchaseOrderMutationError(err error, fallback string) error {
	switch {
	case errors.Is(err, domain.ErrPurchaseOrderNotFound):
		return purchaseOrderStatusError(http.StatusNotFound, "Purchase order not found")
	case errors.Is(err, domain.ErrPurchaseOrderReferenceNotFound):
		return purchaseOrderStatusError(http.StatusNotFound, "Purchase order reference not found")
	case errors.Is(err, domain.ErrPurchaseOrderAlreadyReceived):
		return purchaseOrderStatusError(http.StatusConflict, "Received purchase orders cannot be reopened")
	case errors.Is(err, stock.ErrInsufficientStock):
		return purchaseOrderStatusError(http.StatusConflict, "Insufficient stock")
	default:
		return purchaseOrderStatusError(http.StatusInternalServerError, fallback)
	}
}
