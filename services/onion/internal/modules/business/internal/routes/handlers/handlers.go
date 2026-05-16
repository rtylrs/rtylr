package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/business/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func BusinessMeRoutes() []string { return []string{"/business/me"} }

type (
	GetProfileHandler struct{ Service *application.Service }
	GetProfileRequest struct{}
)

type GetProfileResponse struct {
	ID           int64  `json:"id"`
	SiteName     string `json:"site_name"`
	LogoURL      string `json:"logo_url"`
	ContactEmail string `json:"contact_email"`
	ContactPhone string `json:"contact_phone"`
	Address      string `json:"address"`
	Currency     string `json:"currency"`
	Timezone     string `json:"timezone"`
	Plan         string `json:"plan"`
}

func (h *GetProfileHandler) Serve(ctx context.Context, _ *GetProfileRequest) (GetProfileResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetProfileResponse{}, errorModels.PermissionDeniedServerError()
	}

	res, err := cacher.Of[GetProfileResponse]().GetOrComputeSingle(
		ctx,
		cacheKeys.BusinessConfig(device.User.BusinessID),
		func(ctx context.Context) (GetProfileResponse, error) {
			profile, err := h.Service.GetProfile(ctx, device.User.UserID)
			if err != nil {
				return GetProfileResponse{}, err
			}
			return GetProfileResponse{
				ID:           profile.ID,
				SiteName:     profile.SiteName,
				LogoURL:      profile.LogoURL,
				ContactEmail: profile.ContactEmail,
				ContactPhone: profile.ContactPhone,
				Address:      profile.Address,
				Currency:     profile.Currency,
				Timezone:     profile.Timezone,
				Plan:         profile.Plan,
			}, nil
		},
		cacher.WithExpiry(ctx, 10*time.Minute),
	)
	if err != nil {
		return GetProfileResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to get business profile",
		}
	}
	return res, nil
}
