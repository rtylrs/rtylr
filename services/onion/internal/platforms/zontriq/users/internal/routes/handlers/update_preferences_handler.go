package handlers

import (
	"context"
	"net/http"

	auth "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	usersv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/users/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/users/internal/application"
	"github.com/voxire/onion/internal/platforms/zontriq/users/internal/domain"
)

func UpdatePreferencesRoutes() []string {
	return []string{
		"/zontriq/user/preferences",
	}
}

type UpdatePreferencesHandler struct {
	UsersService *application.Service
}

func (h *UpdatePreferencesHandler) Serve(ctx context.Context, req *usersv1.UpdatePreferencesRequest) (usersv1.UpdatePreferencesResponse, *errorModels.ServerError) {
	device := auth.DeviceCtx(ctx)

	if device.User == nil {
		return usersv1.UpdatePreferencesResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	userID := int(device.User.UserID)

	zontriqUser, err := h.UsersService.GetZontriqUserByUserID(ctx, userID)
	if err != nil {
		return usersv1.UpdatePreferencesResponse{}, &errorModels.ServerError{
			Code:    http.StatusNotFound,
			Message: "Zontriq user not found",
		}
	}

	if req.Preferences == nil {
		return usersv1.UpdatePreferencesResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "preferences is required",
		}
	}

	prefs := domain.Preferences{
		DefaultVoice:         req.Preferences.DefaultVoice,
		DefaultNiche:         req.Preferences.DefaultNiche,
		NotifyVideoReady:     req.Preferences.NotifyVideoReady,
		NotifyPublishSuccess: req.Preferences.NotifyPublishSuccess,
		NotifyPublishFailure: req.Preferences.NotifyPublishFailure,
		NotifyQuotaWarning:   req.Preferences.NotifyQuotaWarning,
		NotifyWeeklyReport:   req.Preferences.NotifyWeeklyReport,
	}

	updated, err := h.UsersService.UpdatePreferences(ctx, zontriqUser.ID, prefs)
	if err != nil {
		return usersv1.UpdatePreferencesResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update preferences",
		}
	}

	return usersv1.UpdatePreferencesResponse{
		Preferences: &usersv1.Preferences{
			DefaultVoice:         updated.DefaultVoice,
			DefaultNiche:         updated.DefaultNiche,
			NotifyVideoReady:     updated.NotifyVideoReady,
			NotifyPublishSuccess: updated.NotifyPublishSuccess,
			NotifyPublishFailure: updated.NotifyPublishFailure,
			NotifyQuotaWarning:   updated.NotifyQuotaWarning,
			NotifyWeeklyReport:   updated.NotifyWeeklyReport,
		},
	}, nil
}
