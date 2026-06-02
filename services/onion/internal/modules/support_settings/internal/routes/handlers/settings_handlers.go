package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/support_settings/internal/application"
	"github.com/voxire/onion/internal/modules/support_settings/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func GetSettingsRoutes() []string    { return []string{"/support/settings"} }
func UpdateSettingsRoutes() []string { return []string{"/support/settings"} }

type (
	GetSettingsHandler  struct{ Service *application.Service }
	GetSettingsRequest  struct{}
	GetSettingsResponse struct {
		Settings *domain.Settings `json:"settings"`
	}
)

func (h *GetSettingsHandler) Serve(ctx context.Context, _ *GetSettingsRequest) (GetSettingsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetSettingsResponse{}, errorModels.PermissionDeniedServerError()
	}
	s, err := h.Service.Get(ctx, device.User.BusinessID)
	if err != nil {
		return GetSettingsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get support settings"}
	}
	return GetSettingsResponse{Settings: s}, nil
}

type (
	UpdateSettingsHandler struct{ Service *application.Service }
	UpdateSettingsRequest struct {
		AutoAssign        bool   `json:"auto_assign"`
		SLAHours          int    `json:"sla_hours"`
		BusinessHours     string `json:"business_hours"`
		NotificationEmail string `json:"notification_email"`
	}
)

type UpdateSettingsResponse struct {
	Message string `json:"message"`
}

func (h *UpdateSettingsHandler) Validate(_ context.Context, req *UpdateSettingsRequest) *errorModels.ServerError {
	if req.SLAHours <= 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "SLA hours must be greater than 0",
		}
	}
	if req.SLAHours > 1000 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "SLA hours must not exceed 1000",
		}
	}
	return nil
}

func (h *UpdateSettingsHandler) Serve(ctx context.Context, req *UpdateSettingsRequest) (UpdateSettingsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateSettingsResponse{}, errorModels.PermissionDeniedServerError()
	}
	err := h.Service.Upsert(ctx, domain.SettingsInput{
		BusinessID: device.User.BusinessID, AutoAssign: req.AutoAssign,
		SLAHours: req.SLAHours, BusinessHours: req.BusinessHours,
		NotificationEmail: req.NotificationEmail,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidSLAHours) {
			return UpdateSettingsResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "SLA hours must be between 1 and 1000"}
		}
		return UpdateSettingsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update support settings"}
	}
	return UpdateSettingsResponse{Message: "Support settings updated"}, nil
}
