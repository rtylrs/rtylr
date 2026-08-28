package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/support_settings/internal/application"
	"github.com/voxire/onion/internal/modules/support_settings/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	supportsettingsv1 "github.com/voxire/voxie/gen/go/support/settings/v1"
)

type SettingsAPI struct {
	service *application.Service
}

func NewSettingsAPI(service *application.Service) *SettingsAPI {
	return &SettingsAPI{service: service}
}

func (a *SettingsAPI) GetSettings(ctx context.Context, _ *supportsettingsv1.GetSettingsRequest) (*supportsettingsv1.GetSettingsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	settings, err := a.service.Get(ctx, device.User.BusinessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get support settings")
	}
	return &supportsettingsv1.GetSettingsResponse{Settings: newSettings(settings)}, nil
}

func (a *SettingsAPI) UpdateSettings(ctx context.Context, req *supportsettingsv1.UpdateSettingsRequest) (*supportsettingsv1.UpdateSettingsResponse, error) {
	if req.GetSlaHours() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "SLA hours must be greater than 0")
	}
	if req.GetSlaHours() > 1000 {
		return nil, onekiterrors.New(http.StatusBadRequest, "SLA hours must not exceed 1000")
	}

	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	err := a.service.Upsert(ctx, domain.SettingsInput{
		BusinessID:        device.User.BusinessID,
		AutoAssign:        req.GetAutoAssign(),
		SLAHours:          int(req.GetSlaHours()),
		BusinessHours:     req.GetBusinessHours(),
		NotificationEmail: req.GetNotificationEmail(),
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidSLAHours) {
			return nil, onekiterrors.New(http.StatusBadRequest, "SLA hours must be between 1 and 1000")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update support settings")
	}
	return &supportsettingsv1.UpdateSettingsResponse{Message: "Support settings updated"}, nil
}

func newSettings(settings *domain.Settings) *supportsettingsv1.Settings {
	if settings == nil {
		return nil
	}
	return &supportsettingsv1.Settings{
		Id:                settings.ID,
		BusinessId:        settings.BusinessID,
		AutoAssign:        settings.AutoAssign,
		SlaHours:          int32(settings.SLAHours), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		BusinessHours:     settings.BusinessHours,
		NotificationEmail: settings.NotificationEmail,
		CreatedAt:         formatTime(settings.CreatedAt),
		UpdatedAt:         formatTime(settings.UpdatedAt),
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
