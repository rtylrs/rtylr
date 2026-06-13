package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_menu/internal/application"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// Public read path for the customer-facing menu app. This endpoint
// is anonymous (registered in GetAnonymousRoutes) — a diner scans a
// QR code and loads the menu; no account is involved.
//
// The route takes `business_id` as a query param rather than a
// human-readable slug because `businesses` doesn't have a slug
// column today. Adding that column is additive and cheap (see plan
// P3 cross-cutting notes) — but the public-read contract works
// either way, so we keep scope small here.
//
// Trust boundary: this handler MUST only return data that's OK to
// show the public internet. Product media URLs are public storefront
// assets; keep private cost/supplier fields out of MenuItemPayload. If a
// future column adds cost-price or supplier info, the `toProto` in
// menu_handlers.go becomes a privacy hole — add a separate
// `toPublicProto` when that happens.

func PublicListMenuItemsRoutes() []string {
	return []string{"/pos/public/menu"}
}

type PublicListMenuItemsHandler struct{ Service *application.Service }

type PublicListMenuItemsRequest struct {
	BusinessID int64 `json:"business_id"`
	// Optional. Accepted now so menu.rtylr.com/<biz>/<loc> can pass
	// it through; server-side per-location filtering kicks in when
	// pos_menu_items get a location_id column. Until then this is
	// ignored so callers don't need to special-case multi-location
	// vs single-location tenants.
	LocationID int64 `json:"location_id"`
}

type PublicListMenuItemsResponse struct {
	Items []MenuItemPayload `json:"items"`
}

func (h *PublicListMenuItemsHandler) Validate(_ context.Context, req *PublicListMenuItemsRequest) *errorModels.ServerError {
	if req.BusinessID <= 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "business_id is required",
		}
	}
	return nil
}

func (h *PublicListMenuItemsHandler) Serve(ctx context.Context, req *PublicListMenuItemsRequest) (PublicListMenuItemsResponse, *errorModels.ServerError) {
	// No auth check — this is the whole point of the endpoint. Resolve
	// business_id to the canonical owner tenant before reading products.
	ownerID, err := h.Service.BusinessOwnerID(ctx, req.BusinessID, req.BusinessID)
	if err != nil {
		return PublicListMenuItemsResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to resolve public menu",
		}
	}

	items, err := h.Service.List(ctx, ownerID)
	if err != nil {
		return PublicListMenuItemsResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load public menu",
		}
	}

	payloads := make([]MenuItemPayload, len(items))
	for i, item := range items {
		payloads[i] = toMenuItemPayload(item)
	}
	return PublicListMenuItemsResponse{Items: payloads}, nil
}
