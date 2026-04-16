package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/application"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

func UnassignAssetRoutes() []string { return []string{"/hr/equipment/assets/{id}/unassign"} }

type UnassignAssetHandler struct{ Service *application.Service }

type UnassignAssetRequest struct{}

type UnassignAssetResponse struct {
	Success bool `json:"success"`
}

func (h *UnassignAssetHandler) Serve(ctx context.Context, _ *UnassignAssetRequest) (UnassignAssetResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UnassignAssetResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	assetID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UnassignAssetResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid asset id"}
	}

	err := h.Service.UnassignAsset(ctx, device.User.UserID, assetID)
	if err != nil {
		if err == domain.ErrAssetNotFound {
			return UnassignAssetResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Asset not found"}
		}
		logger.Ctx(ctx).Error("failed to unassign asset", logger.Any("owner_id", device.User.UserID), logger.Any("asset_id", assetID), logger.Error(err))
		return UnassignAssetResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to unassign asset"}
	}

	return UnassignAssetResponse{Success: true}, nil
}
