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

func ResumeRoutes() []string { return []string{"/finance/recurring-invoices/{id}/resume"} }

type ResumeHandler struct{ Service *application.Service }

type ResumeRequest struct{}

type ResumeResponse struct {
	Message string `json:"message"`
}

func (h *ResumeHandler) Serve(ctx context.Context, _ *ResumeRequest) (ResumeResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ResumeResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ResumeResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return ResumeResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid recurring invoice id"}
	}

	if err := h.Service.Resume(ctx, businessID, id); err != nil {
		if errors.Is(err, domain.ErrRecurringInvoiceNotFound) {
			return ResumeResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Recurring invoice not found"}
		}
		logger.Ctx(ctx).Error("failed to resume recurring invoice", logger.Any("owner_id", device.User.UserID), logger.Any("recurring_invoice_id", id), logger.Error(err))
		return ResumeResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to resume recurring invoice"}
	}

	return ResumeResponse{Message: "Recurring invoice resumed"}, nil
}
