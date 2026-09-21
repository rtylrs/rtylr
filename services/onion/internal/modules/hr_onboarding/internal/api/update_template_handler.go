package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/application"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func UpdateTemplateRoutes() []string { return []string{"/hr/onboarding/templates/{id}"} }

type UpdateTemplateHandler struct{ Service *application.Service }

type UpdateTemplateRequest struct {
	Name        string `json:"name"`
	TriggerType string `json:"trigger_type"`
	RoleID      int64  `json:"role_id"`
}

type UpdateTemplateResponse struct {
	Success bool `json:"success"`
}

func (h *UpdateTemplateHandler) Validate(_ context.Context, req *UpdateTemplateRequest) error {
	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	return nil
}

func (h *UpdateTemplateHandler) Serve(ctx context.Context, req *UpdateTemplateRequest) (UpdateTemplateResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateTemplateResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateTemplateResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateTemplateResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid template id")
	}

	err := h.Service.UpdateTemplate(ctx, businessID, id, domain.UpdateTemplateParams{
		Name:        req.Name,
		TriggerType: req.TriggerType,
		RoleID:      req.RoleID,
	})
	if err != nil {
		if errors.Is(err, domain.ErrTemplateNotFound) {
			return UpdateTemplateResponse{}, onekiterrors.New(http.StatusNotFound, "Template not found")
		}
		logger.Ctx(ctx).Error("failed to update onboarding template", logger.Any("owner_id", device.User.UserID), logger.Any("template_id", id), logger.Error(err))
		return UpdateTemplateResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update template")
	}

	return UpdateTemplateResponse{Success: true}, nil
}
