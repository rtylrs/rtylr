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

func ConvertQuoteRoutes() []string { return []string{"/finance/quotes/{id}/convert"} }

type ConvertQuoteHandler struct{ Service *application.Service }

type ConvertQuoteRequest struct{}

type ConvertQuoteResponse struct {
	InvoiceID int64  `json:"invoice_id"`
	Message   string `json:"message"`
}

func (h *ConvertQuoteHandler) Serve(ctx context.Context, _ *ConvertQuoteRequest) (ConvertQuoteResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ConvertQuoteResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ConvertQuoteResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return ConvertQuoteResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid quote id")
	}

	invoiceID, err := h.Service.ConvertToInvoice(ctx, businessID, id)
	if err != nil {
		if errors.Is(err, domain.ErrQuoteNotFound) {
			return ConvertQuoteResponse{}, onekiterrors.New(http.StatusNotFound, "Quote not found")
		}
		if errors.Is(err, domain.ErrQuoteAlreadyConverted) {
			return ConvertQuoteResponse{}, onekiterrors.New(http.StatusConflict, "Quote already converted")
		}
		logger.Ctx(ctx).Error("failed to convert quote", logger.Any("owner_id", device.User.UserID), logger.Any("quote_id", id), logger.Error(err))
		return ConvertQuoteResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to convert quote")
	}

	return ConvertQuoteResponse{InvoiceID: invoiceID, Message: "Quote converted to invoice"}, nil
}
