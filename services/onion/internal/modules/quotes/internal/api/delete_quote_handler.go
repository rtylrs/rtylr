package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/quotes/internal/application"
	"github.com/voxire/onion/internal/modules/quotes/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func DeleteQuoteByIDRoutes() []string { return []string{"/finance/quotes/{id}"} }

type DeleteQuoteByIDHandler struct{ Service *application.Service }

type DeleteQuoteByIDRequest struct{}

type DeleteQuoteByIDResponse struct {
	Message string `json:"message"`
}

func (h *DeleteQuoteByIDHandler) Serve(ctx context.Context, _ *DeleteQuoteByIDRequest) (DeleteQuoteByIDResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return DeleteQuoteByIDResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteQuoteByIDResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return DeleteQuoteByIDResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid quote id")
	}

	if err := h.Service.Delete(ctx, businessID, id); err != nil {
		if errors.Is(err, domain.ErrQuoteNotFound) {
			return DeleteQuoteByIDResponse{}, onekiterrors.New(http.StatusNotFound, "Quote not found")
		}
		if errors.Is(err, domain.ErrQuoteNotEditable) {
			return DeleteQuoteByIDResponse{}, onekiterrors.New(http.StatusConflict, "Quote is not editable")
		}
		logger.Ctx(ctx).Error("failed to delete quote", logger.Any("owner_id", device.User.UserID), logger.Any("quote_id", id), logger.Error(err))
		return DeleteQuoteByIDResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete quote")
	}

	return DeleteQuoteByIDResponse{Message: "Quote deleted successfully"}, nil
}
