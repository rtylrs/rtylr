package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/recurring_invoices/internal/application"
	"github.com/voxire/onion/internal/modules/recurring_invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
)

func PauseRoutes() []string { return []string{"/finance/recurring-invoices/{id}/pause"} }

type PauseHandler struct{ Service *application.Service }

type PauseRequest struct{}

type PauseResponse struct {
	Message string `json:"message"`
}

func (h *PauseHandler) Serve(ctx context.Context, _ *PauseRequest) (PauseResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return PauseResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return PauseResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return PauseResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid recurring invoice id"}
	}

	if err := h.Service.Pause(ctx, businessID, id); err != nil {
		if errors.Is(err, domain.ErrRecurringInvoiceNotFound) {
			return PauseResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Recurring invoice not found"}
		}
		logger.Ctx(ctx).Error("failed to pause recurring invoice", logger.Any("owner_id", device.User.UserID), logger.Any("recurring_invoice_id", id), logger.Error(err))
		return PauseResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to pause recurring invoice"}
	}

	return PauseResponse{Message: "Recurring invoice paused"}, nil
}
