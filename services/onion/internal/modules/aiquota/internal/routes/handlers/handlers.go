package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/aiquota/internal/application"
	"github.com/voxire/onion/internal/modules/aiquota/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ─── GET /ai/usage ────────────────────────────────────────────────────────────

func UsageRoutes() []string { return []string{"/ai/usage"} }

type UsageHandler struct{ Service *application.Service }

type UsageRequest struct{}

// UsageResponse embeds the status so it serializes flat.
type UsageResponse struct {
	domain.QuotaStatus
}

func (h *UsageHandler) Serve(ctx context.Context, _ *UsageRequest) (UsageResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UsageResponse{}, errorModels.PermissionDeniedServerError()
	}

	st, err := h.Service.Status(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return UsageResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to load AI usage",
		}
	}
	return UsageResponse{st}, nil
}

// ─── GET /ai/quota-settings ───────────────────────────────────────────────────

func GetQuotaSettingsRoutes() []string { return []string{"/ai/quota-settings"} }

type GetQuotaSettingsHandler struct{ Service *application.Service }

type GetQuotaSettingsRequest struct{}

type QuotaSettingsResponse struct {
	Mode string `json:"mode"`
}

func (h *GetQuotaSettingsHandler) Serve(ctx context.Context, _ *GetQuotaSettingsRequest) (QuotaSettingsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return QuotaSettingsResponse{}, errorModels.PermissionDeniedServerError()
	}

	mode, err := h.Service.GetMode(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return QuotaSettingsResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to load AI quota settings",
		}
	}
	return QuotaSettingsResponse{Mode: string(mode)}, nil
}

// ─── PUT /ai/quota-settings ───────────────────────────────────────────────────

func UpdateQuotaSettingsRoutes() []string { return []string{"/ai/quota-settings"} }

type UpdateQuotaSettingsHandler struct{ Service *application.Service }

type UpdateQuotaSettingsRequest struct {
	Mode string `json:"mode"`
}

func (h *UpdateQuotaSettingsHandler) Validate(_ context.Context, req *UpdateQuotaSettingsRequest) *errorModels.ServerError {
	m := domain.Mode(req.Mode)
	if m != domain.ModePool && m != domain.ModePerPerson {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "mode must be 'pool' or 'per_person'",
		}
	}
	return nil
}

func (h *UpdateQuotaSettingsHandler) Serve(ctx context.Context, req *UpdateQuotaSettingsRequest) (QuotaSettingsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return QuotaSettingsResponse{}, errorModels.PermissionDeniedServerError()
	}
	// Only owners/admins can change how the team's AI allowance is allocated.
	role := strings.ToLower(device.User.Role)
	if role != "owner" && role != "admin" {
		return QuotaSettingsResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.SetMode(ctx, device.User.BusinessID, device.User.UserID, domain.Mode(req.Mode)); err != nil {
		return QuotaSettingsResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to update AI quota settings",
		}
	}
	return QuotaSettingsResponse{Mode: string(domain.Mode(req.Mode).Normalize())}, nil
}
