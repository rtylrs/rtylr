package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr_equipment/internal/application"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrequipmentv1 "github.com/voxire/voxie/gen/go/hr/equipment/v1"
)

func CreateAssetRoutes() []string { return []string{"/hr/equipment/assets"} }

type CreateAssetHandler struct{ Service *application.Service }

type CreateAssetRequest struct {
	Name          string  `json:"name"`
	Category      string  `json:"category"`
	SerialNumber  string  `json:"serial_number"`
	LocationID    int64   `json:"location_id"`
	PurchaseDate  string  `json:"purchase_date"`
	PurchasePrice float64 `json:"purchase_price"`
	Notes         string  `json:"notes"`
}

type CreateAssetResponse struct {
	Asset *hrequipmentv1.AssetPayload `json:"asset"`
}

func (h *CreateAssetHandler) Validate(_ context.Context, req *CreateAssetRequest) error {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	req.Category = strings.TrimSpace(req.Category)
	if req.Category == "" {
		return onekiterrors.New(http.StatusBadRequest, "category is required")
	}
	if req.PurchasePrice < 0 {
		return onekiterrors.New(http.StatusBadRequest, "purchase_price must be non-negative")
	}
	return nil
}

func (h *CreateAssetHandler) Serve(ctx context.Context, req *CreateAssetRequest) (CreateAssetResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateAssetResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateAssetResponse{}, onekiterrors.PermissionDenied()
	}

	asset, err := h.Service.CreateAsset(ctx, businessID, domain.CreateAssetParams{
		Name:          req.Name,
		Category:      req.Category,
		SerialNumber:  req.SerialNumber,
		LocationID:    req.LocationID,
		PurchaseDate:  req.PurchaseDate,
		PurchasePrice: req.PurchasePrice,
		Notes:         req.Notes,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create asset", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateAssetResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create asset")
	}

	return CreateAssetResponse{Asset: assetToProto(*asset)}, nil
}
