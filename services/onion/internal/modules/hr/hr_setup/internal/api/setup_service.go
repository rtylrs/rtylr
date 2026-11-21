package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr/hr_setup/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
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
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.Service.CompleteSetupStep(ctx, businessID, req.GetStepKey()); err != nil {
		logger.Ctx(ctx).Error("failed to complete setup step", logger.Any("business_id", businessID), logger.String("step_key", req.GetStepKey()), logger.Error(err))
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to complete setup step")
	}
	return &hrsetupv1.CompleteSetupStepResponse{}, nil
}
