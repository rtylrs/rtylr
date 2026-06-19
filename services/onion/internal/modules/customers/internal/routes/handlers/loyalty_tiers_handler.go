package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	crmloyaltyv1 "github.com/voxire/voxie/gen/go/crm/loyalty/v1"
)

func LoyaltyTiersListRoutes() []string {
	return []string{"/crm/loyalty/tiers"}
}

func LoyaltyTiersCreateRoutes() []string {
	return []string{"/crm/loyalty/tiers"}
}

func LoyaltyTiersUpdateRoutes() []string {
	return []string{"/crm/loyalty/tiers"}
}

func LoyaltyTiersDeleteRoutes() []string {
	return []string{"/crm/loyalty/tiers/{id}"}
}

func newLoyaltyTierPayload(t domain.LoyaltyTier) *crmloyaltyv1.LoyaltyTier {
	return &crmloyaltyv1.LoyaltyTier{
		Id:          t.ID,
		OwnerId:     t.OwnerID,
		Name:        t.Name,
		MinPoints:   int64(t.MinPoints),
		DiscountPct: t.DiscountPct,
		Perks:       string(t.Perks),
		SortOrder:   int64(t.SortOrder),
		CreatedAt:   t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// -- LIST --------------------------------------------------------------------

type ListLoyaltyTiersHandler struct {
	CRMService *application.Service
}

type ListLoyaltyTiersRequest struct{}

type ListLoyaltyTiersResponse struct {
	Tiers []*crmloyaltyv1.LoyaltyTier `json:"tiers"`
}

func (h *ListLoyaltyTiersHandler) Serve(ctx context.Context, _ *ListLoyaltyTiersRequest) (ListLoyaltyTiersResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListLoyaltyTiersResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListLoyaltyTiersResponse{}, errorModels.PermissionDeniedServerError()
	}

	tiers, err := h.CRMService.ListLoyaltyTiers(ctx, businessID)
	if err != nil {
		return ListLoyaltyTiersResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list loyalty tiers",
		}
	}

	payloads := make([]*crmloyaltyv1.LoyaltyTier, len(tiers))
	for i, t := range tiers {
		payloads[i] = newLoyaltyTierPayload(t)
	}

	return ListLoyaltyTiersResponse{Tiers: payloads}, nil
}

// -- CREATE ------------------------------------------------------------------

type CreateLoyaltyTierHandler struct {
	CRMService *application.Service
}

type CreateLoyaltyTierRequest struct {
	Name        string          `json:"name"`
	MinPoints   int             `json:"min_points"`
	DiscountPct float64         `json:"discount_pct"`
	Perks       json.RawMessage `json:"perks"`
	SortOrder   int             `json:"sort_order"`
}

type CreateLoyaltyTierResponse struct {
	TierID  int64  `json:"tier_id"`
	Message string `json:"message"`
}

func (h *CreateLoyaltyTierHandler) Validate(_ context.Context, req *CreateLoyaltyTierRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Tier name is required",
		}
	}

	if req.MinPoints < 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Minimum points must not be negative",
		}
	}

	if req.DiscountPct < 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Discount percentage must not be negative",
		}
	}

	return nil
}

func (h *CreateLoyaltyTierHandler) Serve(ctx context.Context, req *CreateLoyaltyTierRequest) (CreateLoyaltyTierResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateLoyaltyTierResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateLoyaltyTierResponse{}, errorModels.PermissionDeniedServerError()
	}

	perks := []byte(req.Perks)
	if perks == nil {
		perks = []byte("[]")
	}

	id, err := h.CRMService.CreateLoyaltyTier(ctx, domain.LoyaltyTierInput{
		OwnerID:     device.User.UserID,
		BusinessID:  businessID,
		Name:        req.Name,
		MinPoints:   req.MinPoints,
		DiscountPct: req.DiscountPct,
		Perks:       perks,
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		return CreateLoyaltyTierResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create loyalty tier",
		}
	}

	return CreateLoyaltyTierResponse{TierID: id, Message: "Loyalty tier created successfully"}, nil
}

// -- UPDATE ------------------------------------------------------------------

type UpdateLoyaltyTierHandler struct {
	CRMService *application.Service
}

type UpdateLoyaltyTierRequest struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	MinPoints   int             `json:"min_points"`
	DiscountPct float64         `json:"discount_pct"`
	Perks       json.RawMessage `json:"perks"`
	SortOrder   int             `json:"sort_order"`
}

type UpdateLoyaltyTierResponse struct {
	Message string `json:"message"`
}

func (h *UpdateLoyaltyTierHandler) Validate(_ context.Context, req *UpdateLoyaltyTierRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Tier ID is required",
		}
	}

	if req.Name == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Tier name is required",
		}
	}

	if req.MinPoints < 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Minimum points must not be negative",
		}
	}

	if req.DiscountPct < 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Discount percentage must not be negative",
		}
	}

	return nil
}

func (h *UpdateLoyaltyTierHandler) Serve(ctx context.Context, req *UpdateLoyaltyTierRequest) (UpdateLoyaltyTierResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateLoyaltyTierResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateLoyaltyTierResponse{}, errorModels.PermissionDeniedServerError()
	}

	perks := []byte(req.Perks)
	if perks == nil {
		perks = []byte("[]")
	}

	if err := h.CRMService.UpdateLoyaltyTier(ctx, domain.LoyaltyTierUpdateInput{
		ID:          req.ID,
		OwnerID:     device.User.UserID,
		BusinessID:  businessID,
		Name:        req.Name,
		MinPoints:   req.MinPoints,
		DiscountPct: req.DiscountPct,
		Perks:       perks,
		SortOrder:   req.SortOrder,
	}); err != nil {
		if errors.Is(err, domain.ErrTierNotFound) {
			return UpdateLoyaltyTierResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Loyalty tier not found",
			}
		}

		return UpdateLoyaltyTierResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update loyalty tier",
		}
	}

	return UpdateLoyaltyTierResponse{Message: "Loyalty tier updated successfully"}, nil
}

// -- DELETE ------------------------------------------------------------------

type DeleteLoyaltyTierHandler struct {
	CRMService *application.Service
}

type DeleteLoyaltyTierRequest struct{}

type DeleteLoyaltyTierResponse struct {
	Message string `json:"message"`
}

func (h *DeleteLoyaltyTierHandler) Serve(ctx context.Context, _ *DeleteLoyaltyTierRequest) (DeleteLoyaltyTierResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteLoyaltyTierResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteLoyaltyTierResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil || id <= 0 {
		return DeleteLoyaltyTierResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid tier ID",
		}
	}

	if err := h.CRMService.DeleteLoyaltyTier(ctx, businessID, id); err != nil {
		if errors.Is(err, domain.ErrTierNotFound) {
			return DeleteLoyaltyTierResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusNotFound,
				Code:     http.StatusNotFound,
				Message:  "Loyalty tier not found",
			}
		}

		if errors.Is(err, domain.ErrTierInUse) {
			return DeleteLoyaltyTierResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusConflict,
				Code:     http.StatusConflict,
				Message:  "Loyalty tier has active members and cannot be deleted",
			}
		}

		return DeleteLoyaltyTierResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete loyalty tier",
		}
	}

	return DeleteLoyaltyTierResponse{Message: "Loyalty tier deleted successfully"}, nil
}
