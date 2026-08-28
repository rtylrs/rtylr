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
	"github.com/voxire/onion/internal/platform/onekiterrors"
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

func (h *ListLoyaltyTiersHandler) Serve(ctx context.Context, _ *ListLoyaltyTiersRequest) (ListLoyaltyTiersResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListLoyaltyTiersResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListLoyaltyTiersResponse{}, onekiterrors.PermissionDenied()
	}

	tiers, err := h.CRMService.ListLoyaltyTiers(ctx, businessID)
	if err != nil {
		return ListLoyaltyTiersResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list loyalty tiers")
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

func (h *CreateLoyaltyTierHandler) Validate(_ context.Context, req *CreateLoyaltyTierRequest) error {
	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "Tier name is required")
	}

	if req.MinPoints < 0 {
		return onekiterrors.New(http.StatusBadRequest, "Minimum points must not be negative")
	}

	if req.DiscountPct < 0 {
		return onekiterrors.New(http.StatusBadRequest, "Discount percentage must not be negative")
	}

	return nil
}

func (h *CreateLoyaltyTierHandler) Serve(ctx context.Context, req *CreateLoyaltyTierRequest) (CreateLoyaltyTierResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateLoyaltyTierResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:loyalty:manage"); err != nil {
		return CreateLoyaltyTierResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateLoyaltyTierResponse{}, onekiterrors.PermissionDenied()
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
		return CreateLoyaltyTierResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create loyalty tier")
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

func (h *UpdateLoyaltyTierHandler) Validate(_ context.Context, req *UpdateLoyaltyTierRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Tier ID is required")
	}

	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "Tier name is required")
	}

	if req.MinPoints < 0 {
		return onekiterrors.New(http.StatusBadRequest, "Minimum points must not be negative")
	}

	if req.DiscountPct < 0 {
		return onekiterrors.New(http.StatusBadRequest, "Discount percentage must not be negative")
	}

	return nil
}

func (h *UpdateLoyaltyTierHandler) Serve(ctx context.Context, req *UpdateLoyaltyTierRequest) (UpdateLoyaltyTierResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateLoyaltyTierResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:loyalty:manage"); err != nil {
		return UpdateLoyaltyTierResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateLoyaltyTierResponse{}, onekiterrors.PermissionDenied()
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
			return UpdateLoyaltyTierResponse{}, onekiterrors.New(http.StatusNotFound, "Loyalty tier not found")
		}

		return UpdateLoyaltyTierResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update loyalty tier")
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

func (h *DeleteLoyaltyTierHandler) Serve(ctx context.Context, _ *DeleteLoyaltyTierRequest) (DeleteLoyaltyTierResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteLoyaltyTierResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:loyalty:manage"); err != nil {
		return DeleteLoyaltyTierResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteLoyaltyTierResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil || id <= 0 {
		return DeleteLoyaltyTierResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid tier ID")
	}

	if err := h.CRMService.DeleteLoyaltyTier(ctx, businessID, id); err != nil {
		if errors.Is(err, domain.ErrTierNotFound) {
			return DeleteLoyaltyTierResponse{}, onekiterrors.New(http.StatusNotFound, "Loyalty tier not found")
		}

		if errors.Is(err, domain.ErrTierInUse) {
			return DeleteLoyaltyTierResponse{}, onekiterrors.New(http.StatusConflict, "Loyalty tier has active members and cannot be deleted")
		}

		return DeleteLoyaltyTierResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete loyalty tier")
	}

	return DeleteLoyaltyTierResponse{Message: "Loyalty tier deleted successfully"}, nil
}
