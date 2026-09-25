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
	"github.com/voxire/onion/internal/platform/tenant"
)

func AssignAssetRoutes() []string { return []string{"/hr/equipment/assets/{id}/assign"} }

type AssignAssetHandler struct{ Service *application.Service }

type AssignAssetRequest struct {
	EmployeeID int64 `json:"employee_id"`
}

type AssignAssetResponse struct {
	Success bool `json:"success"`
}

func (h *AssignAssetHandler) Validate(_ context.Context, req *AssignAssetRequest) error {
	if req.EmployeeID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "employee_id is required")
	}
	return nil
}

func (h *AssignAssetHandler) Serve(ctx context.Context, req *AssignAssetRequest) (AssignAssetResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return AssignAssetResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return AssignAssetResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	assetID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return AssignAssetResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid asset id")
	}

	err := h.Service.AssignAsset(ctx, businessID, assetID, req.EmployeeID)
	if err != nil {
		if errors.Is(err, domain.ErrAssetNotFound) {
			return AssignAssetResponse{}, onekiterrors.New(http.StatusNotFound, "Asset not found")
		}
		if errors.Is(err, domain.ErrAssetAlreadyAssigned) {
			return AssignAssetResponse{}, onekiterrors.New(http.StatusConflict, "Asset is already assigned")
		}
		logger.Ctx(ctx).Error("failed to assign asset", logger.Any("owner_id", device.User.UserID), logger.Any("asset_id", assetID), logger.Error(err))
		return AssignAssetResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to assign asset")
	}

	return AssignAssetResponse{Success: true}, nil
}
