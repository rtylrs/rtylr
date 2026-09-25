package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/quotes/internal/application"
	"github.com/voxire/onion/internal/modules/quotes/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func UpdateQuoteStatusRoutes() []string { return []string{"/finance/quotes/{id}/status"} }

type UpdateQuoteStatusHandler struct{ Service *application.Service }

type UpdateQuoteStatusRequest struct {
	Status string `json:"status"`
}

type UpdateQuoteStatusResponse struct {
	Message string `json:"message"`
}

func (h *UpdateQuoteStatusHandler) Validate(_ context.Context, req *UpdateQuoteStatusRequest) error {
	status := strings.TrimSpace(req.Status)
	if status == "" {
		return onekiterrors.New(http.StatusBadRequest, "status is required")
	}
	if _, ok := domain.ValidQuoteStatuses[status]; !ok {
		return onekiterrors.New(http.StatusBadRequest, "invalid quote status")
	}
	return nil
}

func (h *UpdateQuoteStatusHandler) Serve(ctx context.Context, req *UpdateQuoteStatusRequest) (UpdateQuoteStatusResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateQuoteStatusResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateQuoteStatusResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UpdateQuoteStatusResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid quote id")
	}

	if err := h.Service.UpdateStatus(ctx, businessID, id, strings.TrimSpace(req.Status)); err != nil {
		if errors.Is(err, domain.ErrQuoteNotFound) {
			return UpdateQuoteStatusResponse{}, onekiterrors.New(http.StatusNotFound, "Quote not found")
		}
		return UpdateQuoteStatusResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update quote status")
	}

	return UpdateQuoteStatusResponse{Message: "Quote status updated successfully"}, nil
}
