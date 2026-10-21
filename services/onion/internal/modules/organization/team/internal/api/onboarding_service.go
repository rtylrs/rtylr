package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/organization/team/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hubteamonboardingv1 "github.com/voxire/voxie/gen/go/hub/team/onboarding/v1"
)

type OnboardingAPI struct {
	Service *application.Service
}

func NewOnboardingAPI(service *application.Service) *OnboardingAPI {
	return &OnboardingAPI{Service: service}
}

func (a *OnboardingAPI) MarkOnboarded(ctx context.Context, _ *hubteamonboardingv1.MarkOnboardedRequest) (*hubteamonboardingv1.MarkOnboardedResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.MarkOnboarded(ctx, device.User.UserID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update")
	}

	return &hubteamonboardingv1.MarkOnboardedResponse{Message: "Onboarding complete"}, nil
}
