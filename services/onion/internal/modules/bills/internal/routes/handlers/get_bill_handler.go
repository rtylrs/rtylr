package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/bills/internal/application"
	"github.com/voxire/onion/internal/modules/bills/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
)

func GetBillRoutes() []string { return []string{"/finance/bills/{id}"} }

type GetBillHandler struct{ Service *application.Service }

type GetBillRequest struct{}

type GetBillResponse struct {
	Bill BillPayload `json:"bill"`
}

func (h *GetBillHandler) Serve(ctx context.Context, _ *GetBillRequest) (GetBillResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetBillResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetBillResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return GetBillResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid bill id"}
	}

	bill, err := h.Service.GetBill(ctx, businessID, id)
	if err != nil {
		if errors.Is(err, domain.ErrBillNotFound) {
			return GetBillResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Bill not found"}
		}
		logger.Ctx(ctx).Error("failed to get bill", logger.Any("owner_id", device.User.UserID), logger.Any("bill_id", id), logger.Error(err))
		return GetBillResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to get bill"}
	}

	return GetBillResponse{Bill: newPayload(bill)}, nil
}
