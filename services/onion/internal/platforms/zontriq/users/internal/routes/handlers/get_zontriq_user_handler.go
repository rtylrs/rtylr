package handlers

import (
	"context"
	"net/http"
	"time"

	auth "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	usersv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/users/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/users/internal/application"
)

func GetZontriqUserRoutes() []string {
	return []string{
		"/zontriq/user",
	}
}

type GetZontriqUserHandler struct {
	UsersService *application.Service
}

func (h *GetZontriqUserHandler) Serve(ctx context.Context, _ *usersv1.GetZontriqUserRequest) (usersv1.GetZontriqUserResponse, *errorModels.ServerError) {
	device := auth.DeviceCtx(ctx)

	if device.User == nil {
		return usersv1.GetZontriqUserResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	userID := int(device.User.UserID)

	zontriqUser, err := h.UsersService.GetZontriqUserByUserID(ctx, userID)
	if err != nil {
		_, createErr := h.UsersService.CreateZontriqUser(ctx, userID, 1)
		if createErr != nil {
			return usersv1.GetZontriqUserResponse{}, &errorModels.ServerError{
				Code:    http.StatusInternalServerError,
				Message: "Failed to create zontriq user",
			}
		}

		zontriqUser, err = h.UsersService.GetZontriqUserByUserID(ctx, userID)
		if err != nil {
			return usersv1.GetZontriqUserResponse{}, &errorModels.ServerError{
				Code:    http.StatusInternalServerError,
				Message: "Failed to fetch created zontriq user",
			}
		}
	}

	usage, usageErr := h.UsersService.GetCurrentUsage(ctx, zontriqUser.ID)
	if usageErr != nil {
		usage = nil
	}

	prefs, prefsErr := h.UsersService.GetPreferences(ctx, zontriqUser.ID)
	if prefsErr != nil {
		prefs = nil
	}

	var trialEndsAt string
	if !zontriqUser.TrialEndsAt.IsZero() {
		trialEndsAt = zontriqUser.TrialEndsAt.Format(time.RFC3339)
	}

	protoUser := &usersv1.ZontriqUser{
		Id:     int64(zontriqUser.ID),
		UserId: int64(zontriqUser.UserID),
		Email:  zontriqUser.Email,
		Name:   zontriqUser.Name,
		Plan: &usersv1.Plan{
			Id:                int64(zontriqUser.Plan.ID),
			Name:              zontriqUser.Plan.Name,
			DisplayName:       zontriqUser.Plan.DisplayName,
			PriceCents:        int32(zontriqUser.Plan.PriceCents),
			MaxVideosPerMonth: int32(zontriqUser.Plan.MaxVideosPerMonth),
			MaxChannels:       int32(zontriqUser.Plan.MaxChannels),
			MaxSeries:         int32(zontriqUser.Plan.MaxSeries),
			MaxAutomations:    int32(zontriqUser.Plan.MaxAutomations),
		},
		Status:      int32(zontriqUser.Status),
		Onboarded:   zontriqUser.Onboarded,
		TrialEndsAt: trialEndsAt,
		CreatedAt:   zontriqUser.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   zontriqUser.UpdatedAt.Format(time.RFC3339),
	}

	var protoUsage *usersv1.Usage
	if usage != nil {
		protoUsage = &usersv1.Usage{
			Id:              int64(usage.ID),
			PeriodStart:     usage.PeriodStart.Format(time.RFC3339),
			PeriodEnd:       usage.PeriodEnd.Format(time.RFC3339),
			VideosGenerated: int32(usage.VideosGenerated),
			VideosPublished: int32(usage.VideosPublished),
		}
	}

	var protoPrefs *usersv1.Preferences
	if prefs != nil {
		protoPrefs = &usersv1.Preferences{
			DefaultVoice:         prefs.DefaultVoice,
			DefaultNiche:         prefs.DefaultNiche,
			NotifyVideoReady:     prefs.NotifyVideoReady,
			NotifyPublishSuccess: prefs.NotifyPublishSuccess,
			NotifyPublishFailure: prefs.NotifyPublishFailure,
			NotifyQuotaWarning:   prefs.NotifyQuotaWarning,
			NotifyWeeklyReport:   prefs.NotifyWeeklyReport,
		}
	}

	return usersv1.GetZontriqUserResponse{
		User:        protoUser,
		Usage:       protoUsage,
		Preferences: protoPrefs,
	}, nil
}
