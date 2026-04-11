package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/support_settings/internal/application"
	"github.com/voxire/onion/internal/modules/support_settings/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

type SettingsPayload struct {
	ID                int64  `json:"id"`
	BusinessID        int64  `json:"businessId"`
	AutoAssign        bool   `json:"autoAssign"`
	SLAHours          int    `json:"slaHours"`
	BusinessHours     string `json:"businessHours"`
	NotificationEmail string `json:"notificationEmail"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
}

func GetSettingsRoutes() []string    { return []string{"/support/settings"} }
func UpdateSettingsRoutes() []string { return []string{"/support/settings"} }

// -- GET ---------------------------------------------------------------------

type GetSettingsHandler struct{ Service *application.Service }

type GetSettingsRequest struct{}

type GetSettingsResponse struct {
	Settings SettingsPayload `json:"settings"`
}

func (h *GetSettingsHandler) Serve(ctx context.Context, _ *GetSettingsRequest) (GetSettingsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetSettingsResponse{}, errorModels.PermissionDeniedServerError()
	}
	s, err := h.Service.Get(ctx, device.User.UserID)
	if err != nil {
		return GetSettingsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get support settings"}
	}
	return GetSettingsResponse{Settings: SettingsPayload{
		ID: s.ID, BusinessID: s.BusinessID, AutoAssign: s.AutoAssign,
		SLAHours: s.SLAHours, BusinessHours: s.BusinessHours,
		NotificationEmail: s.NotificationEmail,
		CreatedAt:         s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}}, nil
}

// -- UPDATE ------------------------------------------------------------------

type UpdateSettingsHandler struct{ Service *application.Service }

type UpdateSettingsRequest struct {
	AutoAssign        bool   `json:"autoAssign"`
	SLAHours          int    `json:"slaHours"`
	BusinessHours     string `json:"businessHours"`
	NotificationEmail string `json:"notificationEmail"`
}

type UpdateSettingsResponse struct {
	Message string `json:"message"`
}

func (h *UpdateSettingsHandler) Serve(ctx context.Context, req *UpdateSettingsRequest) (UpdateSettingsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateSettingsResponse{}, errorModels.PermissionDeniedServerError()
	}
	err := h.Service.Upsert(ctx, domain.SettingsInput{
		BusinessID:        device.User.UserID,
		AutoAssign:        req.AutoAssign,
		SLAHours:          req.SLAHours,
		BusinessHours:     req.BusinessHours,
		NotificationEmail: req.NotificationEmail,
	})
	if err != nil {
		return UpdateSettingsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update support settings"}
	}
	return UpdateSettingsResponse{Message: "Support settings updated"}, nil
}
