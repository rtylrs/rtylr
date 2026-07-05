package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_setup/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hrsetupv1 "github.com/voxire/voxie/gen/go/hr/setup/v1"
)

type SetupAPI struct {
	Service *application.Service
}

func NewSetupAPI(service *application.Service) *SetupAPI {
	return &SetupAPI{Service: service}
}

func (a *SetupAPI) CompleteSetupStep(ctx context.Context, req *hrsetupv1.CompleteSetupStepRequest) (*hrsetupv1.CompleteSetupStepResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.Service.CompleteSetupStep(ctx, device.User.UserID, req.GetStepKey()); err != nil {
		logger.Ctx(ctx).Error("failed to complete setup step", logger.Any("owner_id", device.User.UserID), logger.String("step_key", req.GetStepKey()), logger.Error(err))
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to complete setup step")
	}
	return &hrsetupv1.CompleteSetupStepResponse{}, nil
}
