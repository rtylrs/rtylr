package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr_equipment/internal/application"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
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

func (h *CreateAssetHandler) Validate(_ context.Context, req *CreateAssetRequest) *errorModels.ServerError {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "name is required"}
	}
	req.Category = strings.TrimSpace(req.Category)
	if req.Category == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "category is required"}
	}
	if req.PurchasePrice < 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "purchase_price must be non-negative"}
	}
	return nil
}

func (h *CreateAssetHandler) Serve(ctx context.Context, req *CreateAssetRequest) (CreateAssetResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateAssetResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateAssetResponse{}, errorModels.PermissionDeniedServerError()
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
		return CreateAssetResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to create asset"}
	}

	return CreateAssetResponse{Asset: assetToProto(*asset)}, nil
}
