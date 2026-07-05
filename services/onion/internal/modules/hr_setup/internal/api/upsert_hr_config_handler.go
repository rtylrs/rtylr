package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_setup/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func UpsertHRConfigRoutes() []string { return []string{"/hr/config"} }

type UpsertHRConfigHandler struct{ Service *application.Service }

type UpsertHRConfigRequest struct {
	SetupCompletedAt *int64 `json:"setup_completed_at"`
}

type UpsertHRConfigResponse struct {
	Config *HRConfigPayload `json:"config"`
}

func (h *UpsertHRConfigHandler) Serve(ctx context.Context, req *UpsertHRConfigRequest) (UpsertHRConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpsertHRConfigResponse{}, onekiterrors.PermissionDenied()
	}

	cfg, err := h.Service.UpsertHRConfig(ctx, device.User.UserID, req.SetupCompletedAt)
	if err != nil {
		logger.Ctx(ctx).Error("failed to upsert hr config", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return UpsertHRConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update HR config")
	}

	return UpsertHRConfigResponse{Config: toConfigPayload(cfg)}, nil
}
