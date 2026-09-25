package api

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/crm_settings/internal/application"
	"github.com/voxire/onion/internal/modules/crm_settings/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	crmsettingsv1 "github.com/voxire/voxie/gen/go/crm/settings/v1"
)

type SettingsAPI struct {
	service *application.Service
}

func NewSettingsAPI(service *application.Service) *SettingsAPI {
	return &SettingsAPI{service: service}
}

func (a *SettingsAPI) GetSettings(ctx context.Context, _ *crmsettingsv1.GetSettingsRequest) (*crmsettingsv1.GetSettingsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	settings, err := a.service.Get(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get CRM settings")
	}

	return &crmsettingsv1.GetSettingsResponse{Settings: newSettings(settings)}, nil
}

func (a *SettingsAPI) UpdateSettings(ctx context.Context, req *crmsettingsv1.UpdateSettingsRequest) (*crmsettingsv1.UpdateSettingsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if !authPkg.IsBusinessAdmin(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	timezone := req.GetTimezone()
	if timezone == "" {
		timezone = "UTC"
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.service.Upsert(ctx, domain.SettingsInput{
		OwnerID:        device.User.UserID,
		BusinessID:     businessID,
		DefaultConsent: req.GetDefaultConsent(),
		FromEmail:      req.GetFromEmail(),
		FromName:       req.GetFromName(),
		SmsSenderID:    req.GetSmsSenderId(),
		Timezone:       timezone,
	}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update CRM settings")
	}

	return &crmsettingsv1.UpdateSettingsResponse{Message: "CRM settings updated successfully"}, nil
}

func newSettings(settings domain.Settings) *crmsettingsv1.Settings {
	return &crmsettingsv1.Settings{
		Id:             settings.ID,
		OwnerId:        settings.OwnerID,
		DefaultConsent: settings.DefaultConsent,
		FromEmail:      settings.FromEmail,
		FromName:       settings.FromName,
		SmsSenderId:    settings.SmsSenderID,
		Timezone:       settings.Timezone,
		CreatedAt:      formatTime(settings.CreatedAt),
		UpdatedAt:      formatTime(settings.UpdatedAt),
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("2006-01-02T15:04:05Z")
}
