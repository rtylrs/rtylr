package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/quotes/internal/application"
	"github.com/voxire/onion/internal/modules/quotes/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func UpdateQuoteStatusRoutes() []string { return []string{"/finance/quotes/{id}/status"} }

type UpdateQuoteStatusHandler struct{ Service *application.Service }

type UpdateQuoteStatusRequest struct {
	Status string `json:"status"`
}

type UpdateQuoteStatusResponse struct {
	Message string `json:"message"`
}

func (h *UpdateQuoteStatusHandler) Validate(_ context.Context, req *UpdateQuoteStatusRequest) *errorModels.ServerError {
	status := strings.TrimSpace(req.Status)
	if status == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "status is required"}
	}
	if _, ok := domain.ValidQuoteStatuses[status]; !ok {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "invalid quote status"}
	}
	return nil
}

func (h *UpdateQuoteStatusHandler) Serve(ctx context.Context, req *UpdateQuoteStatusRequest) (UpdateQuoteStatusResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateQuoteStatusResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UpdateQuoteStatusResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid quote id"}
	}

	if err := h.Service.UpdateStatus(ctx, device.User.UserID, id, strings.TrimSpace(req.Status)); err != nil {
		return UpdateQuoteStatusResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to update quote status"}
	}

	return UpdateQuoteStatusResponse{Message: "Quote status updated successfully"}, nil
}
