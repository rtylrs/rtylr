package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// GET /insights/settings

func GetSettingsRoutes() []string { return []string{"/insights/settings"} }

type GetSettingsHandler struct{ Service *application.Service }

type GetSettingsRequest struct{}

type GetSettingsResponse = application.SettingsPayload

func (h *GetSettingsHandler) Serve(ctx context.Context, _ *GetSettingsRequest) (GetSettingsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetSettingsResponse{}, errorModels.PermissionDeniedServerError()
	}
	payload, err := h.Service.GetSettings(ctx, device.User.UserID)
	if err != nil {
		return GetSettingsResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load insights settings",
		}
	}
	return payload, nil
}

// PUT /insights/settings — upsert for the one-row-per-owner config.

func UpdateSettingsRoutes() []string { return []string{"/insights/settings"} }

type UpdateSettingsHandler struct{ Service *application.Service }

type UpdateSettingsRequest struct {
	DefaultCurrency string `json:"default_currency"`
	FiscalYearStart string `json:"fiscal_year_start"`
	Timezone        string `json:"timezone"`
}

type UpdateSettingsResponse = application.SettingsPayload

func (h *UpdateSettingsHandler) Serve(ctx context.Context, req *UpdateSettingsRequest) (UpdateSettingsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateSettingsResponse{}, errorModels.PermissionDeniedServerError()
	}
	payload, err := h.Service.UpdateSettings(ctx, device.User.UserID, application.SettingsPayload{
		DefaultCurrency: req.DefaultCurrency,
		FiscalYearStart: req.FiscalYearStart,
		Timezone:        req.Timezone,
	})
	if err != nil {
		return UpdateSettingsResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to save insights settings",
		}
	}
	return payload, nil
}
