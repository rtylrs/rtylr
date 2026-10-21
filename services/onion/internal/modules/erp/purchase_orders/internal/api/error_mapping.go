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
	case errors.Is(err, domain.ErrPurchaseOrderInvalidTransition):
		return purchaseOrderStatusError(http.StatusConflict, "The purchase order cannot move to that status")
	case errors.Is(err, domain.ErrInvalidReceivedQuantity):
		return purchaseOrderStatusError(http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrInvalidPurchaseOrder):
		return purchaseOrderStatusError(http.StatusBadRequest, "Invalid purchase order")
	default:
		return purchaseOrderStatusError(http.StatusInternalServerError, fallback)
	}
}
