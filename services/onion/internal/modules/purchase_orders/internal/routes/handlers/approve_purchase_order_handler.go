package handlers

import (
	"context"
	stderrors "errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/purchase_orders/internal/application"
	"github.com/voxire/onion/internal/modules/purchase_orders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ─── Approve Purchase Order ───────────────────────────────────────────────────
//
// POST /erp/purchase-orders/{id}/approve transitions the PO status to `approved`.
// The path-parameter form is hand-wired because the generic Register* helpers
// do not parse chi path params into the request struct.

func ApprovePurchaseOrderRoutes() []string {
	return []string{"/erp/purchase-orders/{id}/approve"}
}

type ApprovePurchaseOrderHandler struct {
	Service *application.Service
}

type ApprovePurchaseOrderRequest struct{}

type ApprovePurchaseOrderResponse struct {
	Message string `json:"message"`
}

func (h *ApprovePurchaseOrderHandler) Serve(ctx context.Context, _ *ApprovePurchaseOrderRequest) (ApprovePurchaseOrderResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ApprovePurchaseOrderResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	idRaw := chi.URLParamFromCtx(ctx, "id")
	poID, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || poID <= 0 {
		return ApprovePurchaseOrderResponse{}, &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Invalid purchase order id",
		}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ApprovePurchaseOrderResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.ApprovePurchaseOrder(ctx, businessID, poID); err != nil {
		if stderrors.Is(err, domain.ErrPurchaseOrderNotFound) {
			return ApprovePurchaseOrderResponse{}, &errorModels.ServerError{
				Code:     http.StatusNotFound,
				HTTPCode: http.StatusNotFound,
				Message:  "Purchase order not found",
			}
		}

		if stderrors.Is(err, domain.ErrPurchaseOrderAlreadyApproved) {
			return ApprovePurchaseOrderResponse{}, &errorModels.ServerError{
				Code:     http.StatusConflict,
				HTTPCode: http.StatusConflict,
				Message:  "Purchase order is already approved",
			}
		}

		return ApprovePurchaseOrderResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to approve purchase order",
		}
	}

	return ApprovePurchaseOrderResponse{Message: "Purchase order approved successfully"}, nil
}
