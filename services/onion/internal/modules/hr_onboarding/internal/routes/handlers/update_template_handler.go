package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/application"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
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

func (h *UpdateTemplateHandler) Validate(_ context.Context, req *UpdateTemplateRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *UpdateTemplateHandler) Serve(ctx context.Context, req *UpdateTemplateRequest) (UpdateTemplateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateTemplateResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateTemplateResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid template id"}
	}

	err := h.Service.UpdateTemplate(ctx, device.User.UserID, id, domain.UpdateTemplateParams{
		Name:        req.Name,
		TriggerType: req.TriggerType,
		RoleID:      req.RoleID,
	})
	if err != nil {
		if errors.Is(err, domain.ErrTemplateNotFound) {
			return UpdateTemplateResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Template not found"}
		}
		logger.Ctx(ctx).Error("failed to update onboarding template", logger.Any("owner_id", device.User.UserID), logger.Any("template_id", id), logger.Error(err))
		return UpdateTemplateResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to update template"}
	}

	return UpdateTemplateResponse{Success: true}, nil
}
