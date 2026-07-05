package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_equipment/internal/application"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func UnassignAssetRoutes() []string { return []string{"/hr/equipment/assets/{id}/unassign"} }

type UnassignAssetHandler struct{ Service *application.Service }

type UnassignAssetRequest struct{}

type UnassignAssetResponse struct {
	Success bool `json:"success"`
}

func (h *UnassignAssetHandler) Serve(ctx context.Context, _ *UnassignAssetRequest) (UnassignAssetResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UnassignAssetResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	assetID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UnassignAssetResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid asset id")
	}

	err := h.Service.UnassignAsset(ctx, device.User.UserID, assetID)
	if err != nil {
		if errors.Is(err, domain.ErrAssetNotFound) {
			return UnassignAssetResponse{}, onekiterrors.New(http.StatusNotFound, "Asset not found")
		}
		logger.Ctx(ctx).Error("failed to unassign asset", logger.Any("owner_id", device.User.UserID), logger.Any("asset_id", assetID), logger.Error(err))
		return UnassignAssetResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to unassign asset")
	}

	return UnassignAssetResponse{Success: true}, nil
}
