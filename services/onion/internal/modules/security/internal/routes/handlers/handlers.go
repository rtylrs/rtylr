package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/security/internal/application"
	"github.com/voxire/onion/internal/modules/security/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func SecurityRoutes() []string { return []string{"/security"} }
func TwoFARoutes() []string    { return []string{"/security/2fa"} }

// -- Get Security Settings --

type GetSettingsHandler struct{ Service *application.Service }
type GetSettingsRequest struct{}

type SessionPayload struct {
	ID        string `json:"id"`
	IP        string `json:"ip"`
	UserAgent string `json:"userAgent"`
	CreatedAt string `json:"createdAt"`
	IsCurrent bool   `json:"isCurrent"`
}

type GetSettingsResponse struct {
	TwoFactorEnabled bool             `json:"twoFactorEnabled"`
	ActiveSessions   []SessionPayload `json:"activeSessions"`
}

func (h *GetSettingsHandler) Serve(ctx context.Context, _ *GetSettingsRequest) (GetSettingsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetSettingsResponse{}, errorModels.PermissionDeniedServerError()
	}

	settings, err := h.Service.GetSettings(ctx, device.User.UserID)
	if err != nil {
		return GetSettingsResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to get security settings",
		}
	}

	sessions := make([]SessionPayload, 0, len(settings.ActiveSessions))
	for _, s := range settings.ActiveSessions {
		sessions = append(sessions, SessionPayload{
			ID:        s.ID,
			IP:        s.IP,
			UserAgent: s.UserAgent,
			CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z"),
			IsCurrent: s.IsCurrent,
		})
	}

	return GetSettingsResponse{
		TwoFactorEnabled: settings.TwoFactorEnabled,
		ActiveSessions:   sessions,
	}, nil
}

// -- Toggle 2FA --

type Toggle2FAHandler struct{ Service *application.Service }

type Toggle2FARequest struct {
	Enabled bool `json:"enabled"`
}

type Toggle2FAResponse struct {
	Message string `json:"message"`
}

func (h *Toggle2FAHandler) Serve(ctx context.Context, req *Toggle2FARequest) (Toggle2FAResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return Toggle2FAResponse{}, errorModels.PermissionDeniedServerError()
	}

	_ = domain.TwoFactorInput{Enabled: req.Enabled}

	if err := h.Service.SetTwoFactor(ctx, device.User.UserID, req.Enabled); err != nil {
		return Toggle2FAResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to update 2FA settings",
		}
	}

	msg := "Two-factor authentication disabled"
	if req.Enabled {
		msg = "Two-factor authentication enabled"
	}

	return Toggle2FAResponse{Message: msg}, nil
}
