package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/application"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
	hronboardingv1 "github.com/voxire/voxie/gen/go/hr/onboarding/v1"
)

func CreateTemplateRoutes() []string { return []string{"/hr/onboarding/templates"} }

type CreateTemplateHandler struct{ Service *application.Service }

type CreateTemplateRequest struct {
	Name        string `json:"name"`
	TriggerType string `json:"trigger_type"`
	RoleID      int64  `json:"role_id"`
}

type CreateTemplateResponse struct {
	Template *hronboardingv1.ChecklistTemplatePayload `json:"template"`
}

func (h *CreateTemplateHandler) Validate(_ context.Context, req *CreateTemplateRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Name) == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *CreateTemplateHandler) Serve(ctx context.Context, req *CreateTemplateRequest) (CreateTemplateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateTemplateResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateTemplateResponse{}, errorModels.PermissionDeniedServerError()
	}

	created, err := h.Service.CreateTemplate(ctx, businessID, domain.CreateTemplateParams{
		Name:        req.Name,
		TriggerType: req.TriggerType,
		RoleID:      req.RoleID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create onboarding template", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateTemplateResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to create template"}
	}

	return CreateTemplateResponse{Template: templateToProto(*created)}, nil
}
