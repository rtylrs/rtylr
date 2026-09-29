package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_setup/internal/application"
	"github.com/voxire/onion/internal/modules/hr_setup/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func CompleteSetupStepRoutes() []string { return []string{"/hr/setup/steps/{stepKey}/complete"} }

type CompleteSetupStepHandler struct{ Service *application.Service }

type CompleteSetupStepRequest struct{}

type CompleteSetupStepResponse struct{}

func (h *CompleteSetupStepHandler) Serve(ctx context.Context, _ *CompleteSetupStepRequest) (CompleteSetupStepResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CompleteSetupStepResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireHRSetupManagePermission(ctx); err != nil {
		return CompleteSetupStepResponse{}, err
	}

	stepKey := chi.URLParamFromCtx(ctx, "stepKey")

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CompleteSetupStepResponse{}, onekiterrors.PermissionDenied()
	}

	err := h.Service.CompleteSetupStep(ctx, businessID, stepKey)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidStepKey) {
			return CompleteSetupStepResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid step key")
		}
		logger.Ctx(ctx).Error("failed to complete setup step", logger.Any("owner_id", device.User.UserID), logger.String("step_key", stepKey), logger.Error(err))
		return CompleteSetupStepResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to complete setup step")
	}

	return CompleteSetupStepResponse{}, nil
}
