package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/business/internal/application"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

// Public business header for the customer-facing menu app. Sibling
// to `GET /pos/public/menu?business_id=` — together they give the
// menu page everything it needs (branding + items) without requiring
// a JWT. Only fields that are already safe to show the public
// internet are returned: business name, logo URL, currency, timezone.
// Contact info, plan, address, ownership — all stripped.

func PublicBusinessRoutes() []string {
	return []string{"/pos/public/business"}
}

type PublicBusinessHandler struct {
	Service *application.Service
}

type PublicBusinessRequest struct {
	BusinessID int64 `json:"business_id"`
}

type PublicBusinessResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	LogoURL  string `json:"logo_url"`
	Currency string `json:"currency"`
	Timezone string `json:"timezone"`
}

func (h *PublicBusinessHandler) Validate(_ context.Context, req *PublicBusinessRequest) *errorModels.ServerError {
	if req.BusinessID <= 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "business_id is required",
		}
	}
	return nil
}

func (h *PublicBusinessHandler) Serve(ctx context.Context, req *PublicBusinessRequest) (PublicBusinessResponse, *errorModels.ServerError) {
	res, err := h.Service.GetPublicBusiness(ctx, req.BusinessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PublicBusinessResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusNotFound,
				Code:     http.StatusNotFound,
				Message:  "Business not found",
			}
		}
		logger.Ctx(ctx).Error("public business lookup failed",
			logger.Any("business_id", req.BusinessID),
			logger.Error(err),
		)
		return PublicBusinessResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load business",
		}
	}
	return PublicBusinessResponse{
		ID:       res.ID,
		Name:     res.Name,
		LogoURL:  res.LogoURL,
		Currency: res.Currency,
		Timezone: res.Timezone,
	}, nil
}
