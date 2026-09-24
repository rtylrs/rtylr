package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func GetConfigRoutes() []string { return []string{"/hr/performance/config"} }

type GetConfigHandler struct{ Service *application.Service }

type GetConfigRequest struct{}

type GetConfigResponse struct {
	Config *domain.PerformanceConfigRecord `json:"config"`
}

func (h *GetConfigHandler) Serve(ctx context.Context, _ *GetConfigRequest) (GetConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetConfigResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetConfigResponse{}, onekiterrors.PermissionDenied()
	}

	cfg, err := h.Service.GetPerformanceConfig(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to get performance config", logger.Any("business_id", businessID), logger.Error(err))
		return GetConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get performance config")
	}

	return GetConfigResponse{Config: cfg}, nil
}
