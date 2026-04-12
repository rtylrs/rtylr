package handlers

import (
	"context"
	"net/http"

	suppv1 "github.com/voxire/voxie/gen/go/support/v1"

	"github.com/voxire/onion/internal/modules/support_settings/internal/application"
	"github.com/voxire/onion/internal/modules/support_settings/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func GetSettingsRoutes() []string    { return []string{"/support/settings"} }
func UpdateSettingsRoutes() []string { return []string{"/support/settings"} }

type GetSettingsHandler struct{ Service *application.Service }
type GetSettingsRequest struct{}
type GetSettingsResponse struct {
	Settings *suppv1.SupportSettings `json:"settings"`
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
	return GetSettingsResponse{Settings: &suppv1.SupportSettings{
		Id: s.ID, BusinessId: s.BusinessID, AutoAssign: s.AutoAssign,
		SlaHours: int64(s.SLAHours), BusinessHours: s.BusinessHours,
		NotificationEmail: s.NotificationEmail,
		CreatedAt:         s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}}, nil
}

type UpdateSettingsHandler struct{ Service *application.Service }
type UpdateSettingsRequest struct {
	AutoAssign        bool   `json:"auto_assign"`
	SLAHours          int    `json:"sla_hours"`
	BusinessHours     string `json:"business_hours"`
	NotificationEmail string `json:"notification_email"`
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
		BusinessID: device.User.UserID, AutoAssign: req.AutoAssign,
		SLAHours: req.SLAHours, BusinessHours: req.BusinessHours,
		NotificationEmail: req.NotificationEmail,
	})
	if err != nil {
		return UpdateSettingsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update support settings"}
	}
	return UpdateSettingsResponse{Message: "Support settings updated"}, nil
}
