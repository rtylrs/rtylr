package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/user/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// OnboardBusinessRoutes is the authenticated endpoint a freshly-registered owner
// hits from the dash onboarding screen to create their first business. Registration
// no longer auto-creates a business, so the user lands here with business_id = 0.
// It creates the business, sets it as the user's current business, and returns a
// fresh JWT carrying the new business_id so the frontend can proceed.
func OnboardBusinessRoutes() []string { return []string{"/onboarding/business"} }

type (
	OnboardBusinessHandler  struct{ UserService *application.Service }
	OnboardBusinessRequest  struct{ Name string `json:"name"` }
	OnboardBusinessResponse struct{ Token string `json:"token"` }
)

func (h *OnboardBusinessHandler) Validate(_ context.Context, req *OnboardBusinessRequest) *errorModels.ServerError {
	if len(req.Name) == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *OnboardBusinessHandler) Serve(ctx context.Context, req *OnboardBusinessRequest) (OnboardBusinessResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return OnboardBusinessResponse{}, errorModels.PermissionDeniedServerError()
	}

	token, err := h.UserService.OnboardCreateBusiness(ctx, device.User.UserID, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrBusinessNameRequired):
			return OnboardBusinessResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Business name is required"}
		case errors.Is(err, application.ErrBusinessNameTaken):
			return OnboardBusinessResponse{}, &errorModels.ServerError{Code: http.StatusConflict, Message: "That business name is already taken"}
		default:
			return OnboardBusinessResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to create business"}
		}
	}
	return OnboardBusinessResponse{Token: token}, nil
}
