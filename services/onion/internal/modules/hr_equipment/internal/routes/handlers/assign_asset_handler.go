package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_equipment/internal/application"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
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

func (h *AssignAssetHandler) Validate(_ context.Context, req *AssignAssetRequest) *errorModels.ServerError {
	if req.EmployeeID == 0 {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "employee_id is required"}
	}
	return nil
}

func (h *AssignAssetHandler) Serve(ctx context.Context, req *AssignAssetRequest) (AssignAssetResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return AssignAssetResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return AssignAssetResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	assetID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return AssignAssetResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid asset id"}
	}

	err := h.Service.AssignAsset(ctx, businessID, assetID, req.EmployeeID)
	if err != nil {
		if errors.Is(err, domain.ErrAssetNotFound) {
			return AssignAssetResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Asset not found"}
		}
		if errors.Is(err, domain.ErrAssetAlreadyAssigned) {
			return AssignAssetResponse{}, &errorModels.ServerError{HTTPCode: http.StatusConflict, Code: http.StatusConflict, Message: "Asset is already assigned"}
		}
		logger.Ctx(ctx).Error("failed to assign asset", logger.Any("owner_id", device.User.UserID), logger.Any("asset_id", assetID), logger.Error(err))
		return AssignAssetResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to assign asset"}
	}

	return AssignAssetResponse{Success: true}, nil
}
