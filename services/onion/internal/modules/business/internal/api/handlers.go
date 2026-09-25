package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/business/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func BusinessMeRoutes() []string { return []string{"/business/me"} }

type (
	GetProfileHandler struct{ Service *application.Service }
	GetProfileRequest struct{}
)

type GetProfileResponse struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	SiteName     string `json:"site_name"`
	LogoURL      string `json:"logo_url"`
	ContactEmail string `json:"contact_email"`
	ContactPhone string `json:"contact_phone"`
	Address      string `json:"address"`
	Currency     string `json:"currency"`
	Timezone     string `json:"timezone"`
	Plan         string `json:"plan"`
}

func (h *GetProfileHandler) Serve(ctx context.Context, _ *GetProfileRequest) (GetProfileResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetProfileResponse{}, onekiterrors.PermissionDenied()
	}

	profileScopeID := device.User.BusinessID
	if profileScopeID == 0 {
		profileScopeID = device.User.UserID
	}

	profile, err := h.Service.GetProfile(ctx, profileScopeID)
	if err != nil {
		return GetProfileResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get business profile")
	}

	return GetProfileResponse{
		ID:           profile.ID,
		Name:         profile.Name,
		SiteName:     profile.SiteName,
		LogoURL:      profile.LogoURL,
		ContactEmail: profile.ContactEmail,
		ContactPhone: profile.ContactPhone,
		Address:      profile.Address,
		Currency:     profile.Currency,
		Timezone:     profile.Timezone,
		Plan:         profile.Plan,
	}, nil
}
