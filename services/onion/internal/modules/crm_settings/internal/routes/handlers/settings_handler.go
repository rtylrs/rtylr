package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/crm_settings/internal/application"
	"github.com/voxire/onion/internal/modules/crm_settings/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func SettingsGetRoutes() []string {
	return []string{"/crm/settings"}
}

func SettingsUpdateRoutes() []string {
	return []string{"/crm/settings"}
}

type SettingsPayload struct {
	ID             int64  `json:"id"`
	OwnerID        int64  `json:"owner_id"`
	DefaultConsent bool   `json:"default_consent"`
	FromEmail      string `json:"from_email"`
	FromName       string `json:"from_name"`
	SmsSenderID    string `json:"sms_sender_id"`
	Timezone       string `json:"timezone"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// -- GET ---------------------------------------------------------------------

type GetSettingsHandler struct {
	Service *application.Service
}

type GetSettingsRequest struct{}

type GetSettingsResponse struct {
	Settings SettingsPayload `json:"settings"`
}

func (h *GetSettingsHandler) Serve(ctx context.Context, _ *GetSettingsRequest) (GetSettingsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetSettingsResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetSettingsResponse{}, errorModels.PermissionDeniedServerError()
	}
	settings, err := h.Service.Get(ctx, businessID)
	if err != nil {
		return GetSettingsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get CRM settings",
		}
	}

	createdAt := ""
	if !settings.CreatedAt.IsZero() {
		createdAt = settings.CreatedAt.Format("2006-01-02T15:04:05Z")
	}

	updatedAt := ""
	if !settings.UpdatedAt.IsZero() {
		updatedAt = settings.UpdatedAt.Format("2006-01-02T15:04:05Z")
	}

	return GetSettingsResponse{
		Settings: SettingsPayload{
			ID:             settings.ID,
			OwnerID:        settings.OwnerID,
			DefaultConsent: settings.DefaultConsent,
			FromEmail:      settings.FromEmail,
			FromName:       settings.FromName,
			SmsSenderID:    settings.SmsSenderID,
			Timezone:       settings.Timezone,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
		},
	}, nil
}

// -- UPDATE ------------------------------------------------------------------

type UpdateSettingsHandler struct {
	Service *application.Service
}

type UpdateSettingsRequest struct {
	DefaultConsent bool   `json:"default_consent"`
	FromEmail      string `json:"from_email"`
	FromName       string `json:"from_name"`
	SmsSenderID    string `json:"sms_sender_id"`
	Timezone       string `json:"timezone"`
}

type UpdateSettingsResponse struct {
	Message string `json:"message"`
}

func (h *UpdateSettingsHandler) Serve(ctx context.Context, req *UpdateSettingsRequest) (UpdateSettingsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateSettingsResponse{}, errorModels.PermissionDeniedServerError()
	}

	timezone := req.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateSettingsResponse{}, errorModels.PermissionDeniedServerError()
	}
	if err := h.Service.Upsert(ctx, domain.SettingsInput{
		OwnerID:        device.User.UserID,
		BusinessID:     businessID,
		DefaultConsent: req.DefaultConsent,
		FromEmail:      req.FromEmail,
		FromName:       req.FromName,
		SmsSenderID:    req.SmsSenderID,
		Timezone:       timezone,
	}); err != nil {
		return UpdateSettingsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update CRM settings",
		}
	}

	return UpdateSettingsResponse{Message: "CRM settings updated successfully"}, nil
}
