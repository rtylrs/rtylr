package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/quotes/internal/application"
	"github.com/voxire/onion/internal/modules/quotes/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func UpdateQuoteByIDRoutes() []string { return []string{"/finance/quotes/{id}"} }

type UpdateQuoteByIDHandler struct{ Service *application.Service }

type UpdateQuoteByIDRequest struct {
	CustomerName  string  `json:"customer_name"`
	CustomerEmail string  `json:"customer_email"`
	Status        string  `json:"status"`
	Items         string  `json:"items"`
	Subtotal      float64 `json:"subtotal"`
	Tax           float64 `json:"tax"`
	Discount      float64 `json:"discount"`
	Total         float64 `json:"total"`
	ValidUntil    string  `json:"valid_until"`
	Notes         string  `json:"notes"`
}

type UpdateQuoteByIDResponse struct {
	Message string `json:"message"`
}

func (h *UpdateQuoteByIDHandler) Validate(_ context.Context, req *UpdateQuoteByIDRequest) error {
	if strings.TrimSpace(req.CustomerName) == "" {
		return onekiterrors.New(http.StatusBadRequest, "customerName is required")
	}
	if req.ValidUntil != "" {
		if _, err := time.Parse("2006-01-02", req.ValidUntil); err != nil {
			return onekiterrors.New(http.StatusBadRequest, "validUntil must be in YYYY-MM-DD format")
		}
	}
	return nil
}

func (h *UpdateQuoteByIDHandler) Serve(ctx context.Context, req *UpdateQuoteByIDRequest) (UpdateQuoteByIDResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateQuoteByIDResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateQuoteByIDResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateQuoteByIDResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid quote id")
	}

	input := domain.QuoteInput{
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		Status:        req.Status,
		Items:         req.Items,
		Subtotal:      req.Subtotal,
		Tax:           req.Tax,
		Discount:      req.Discount,
		Total:         req.Total,
		ValidUntil:    req.ValidUntil,
		Notes:         req.Notes,
	}

	if err := h.Service.Update(ctx, businessID, id, input); err != nil {
		if errors.Is(err, domain.ErrQuoteNotFound) {
			return UpdateQuoteByIDResponse{}, onekiterrors.New(http.StatusNotFound, "Quote not found")
		}
		if errors.Is(err, domain.ErrQuoteNotEditable) {
			return UpdateQuoteByIDResponse{}, onekiterrors.New(http.StatusConflict, "Quote is not editable")
		}
		logger.Ctx(ctx).Error("failed to update quote", logger.Any("owner_id", device.User.UserID), logger.Any("quote_id", id), logger.Error(err))
		return UpdateQuoteByIDResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update quote")
	}

	return UpdateQuoteByIDResponse{Message: "Quote updated successfully"}, nil
}
