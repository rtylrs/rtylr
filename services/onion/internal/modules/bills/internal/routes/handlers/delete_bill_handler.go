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

func DeleteBillRoutes() []string { return []string{"/finance/bills/{id}"} }

type DeleteBillHandler struct{ Service *application.Service }

type DeleteBillRequest struct{}

type DeleteBillResponse struct {
	Message string `json:"message"`
}

func (h *DeleteBillHandler) Serve(ctx context.Context, _ *DeleteBillRequest) (DeleteBillResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return DeleteBillResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteBillResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return DeleteBillResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid bill id"}
	}

	if err := h.Service.DeleteBill(ctx, businessID, id); err != nil {
		if errors.Is(err, domain.ErrBillNotFound) {
			return DeleteBillResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Bill not found"}
		}
		logger.Ctx(ctx).Error("failed to delete bill", logger.Any("owner_id", device.User.UserID), logger.Any("bill_id", id), logger.Error(err))
		return DeleteBillResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to delete bill"}
	}

	return DeleteBillResponse{Message: "Bill deleted successfully"}, nil
}
