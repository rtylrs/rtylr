package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/business/internal/application"
	"github.com/voxire/onion/internal/platform/database/dbclient"
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
	// Service is wired for symmetry with the rest of the module even
	// though the public fetch bypasses it — the query shape is too
	// simple to justify threading through a service method that
	// would never be reused internally (every other service method
	// is owner-scoped).
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
	// Direct query — the business module doesn't own the public
	// `businesses` row read today (its service is scoped to
	// `business_profile` rows). Since this is a single tight query
	// over always-public columns, the hand-rolled read here is
	// cleaner than expanding the service surface.
	const q = `SELECT id, name, COALESCE(logo_url, ''), COALESCE(currency, 'USD'), COALESCE(timezone, 'UTC')
	           FROM businesses WHERE id = ? AND is_active = TRUE LIMIT 1`

	var res PublicBusinessResponse
	row := dbclient.Pool().QueryRowContext(ctx, q, req.BusinessID)
	if err := row.Scan(&res.ID, &res.Name, &res.LogoURL, &res.Currency, &res.Timezone); err != nil {
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
	return res, nil
}
