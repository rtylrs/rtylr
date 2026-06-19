package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/price_lists/internal/application"
	"github.com/voxire/onion/internal/modules/price_lists/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ─── List Price Lists ─────────────────────────────────────────────────────────

func ListPriceListsRoutes() []string {
	return []string{"/erp/price-lists"}
}

type ListPriceListsHandler struct {
	Service *application.Service
}

type ListPriceListsRequest struct{}

type PriceListPayload struct {
	ID        int64  `json:"id"`
	OwnerID   int64  `json:"owner_id"`
	Name      string `json:"name"`
	Currency  string `json:"currency"`
	IsDefault bool   `json:"is_default"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ListPriceListsResponse struct {
	PriceLists []PriceListPayload `json:"price_lists"`
}

func newPriceListPayload(p domain.PriceList) PriceListPayload {
	return PriceListPayload{
		ID:        p.ID,
		OwnerID:   p.OwnerID,
		Name:      p.Name,
		Currency:  p.Currency,
		IsDefault: p.IsDefault,
		CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *ListPriceListsHandler) Serve(ctx context.Context, _ *ListPriceListsRequest) (ListPriceListsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListPriceListsResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListPriceListsResponse{}, errorModels.PermissionDeniedServerError()
	}

	lists, err := h.Service.ListPriceLists(ctx, businessID)
	if err != nil {
		return ListPriceListsResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to list price lists",
		}
	}

	payloads := make([]PriceListPayload, 0, len(lists))
	for _, p := range lists {
		payloads = append(payloads, newPriceListPayload(p))
	}

	return ListPriceListsResponse{PriceLists: payloads}, nil
}

// ─── Create Price List ────────────────────────────────────────────────────────

func CreatePriceListRoutes() []string {
	return []string{"/erp/price-lists"}
}

type CreatePriceListHandler struct {
	Service *application.Service
}

type CreatePriceListRequest struct {
	Name      string `json:"name"`
	Currency  string `json:"currency"`
	IsDefault bool   `json:"is_default"`
}

type CreatePriceListResponse struct {
	PriceListID int64  `json:"price_list_id"`
	Message     string `json:"message"`
}

func (h *CreatePriceListHandler) Validate(_ context.Context, req *CreatePriceListRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Name) == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Price list name is required",
		}
	}

	return nil
}

func (h *CreatePriceListHandler) Serve(ctx context.Context, req *CreatePriceListRequest) (CreatePriceListResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreatePriceListResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreatePriceListResponse{}, errorModels.PermissionDeniedServerError()
	}

	id, err := h.Service.CreatePriceList(ctx, businessID, domain.PriceListInput{
		Name:      req.Name,
		Currency:  req.Currency,
		IsDefault: req.IsDefault,
	})
	if err != nil {
		return CreatePriceListResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to create price list",
		}
	}

	return CreatePriceListResponse{PriceListID: id, Message: "Price list created successfully"}, nil
}
