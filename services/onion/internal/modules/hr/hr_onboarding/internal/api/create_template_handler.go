package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr/hr_onboarding/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_onboarding/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hronboardingv1 "github.com/voxire/voxie/gen/go/hr/onboarding/v1"
)

type CreateTemplateHandler struct{ Service *application.Service }

type CreateTemplateRequest struct {
	Name        string `json:"name"`
	TriggerType string `json:"trigger_type"`
	RoleID      int64  `json:"role_id"`
}

type CreateTemplateResponse struct {
	Template *hronboardingv1.ChecklistTemplatePayload `json:"template"`
}

func (h *CreateTemplateHandler) Validate(_ context.Context, req *CreateTemplateRequest) error {
	if req == nil {
		return onekiterrors.New(http.StatusBadRequest, "request body is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	if err := domain.ValidateCreateTemplateParams(domain.CreateTemplateParams{
		Name: req.Name, TriggerType: req.TriggerType,
	}); err != nil {
		return onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	return nil
}

func (h *CreateTemplateHandler) Serve(ctx context.Context, req *CreateTemplateRequest) (CreateTemplateResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateTemplateResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireOnboardingManagePermission(ctx); err != nil {
		return CreateTemplateResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateTemplateResponse{}, onekiterrors.PermissionDenied()
	}

	created, err := h.Service.CreateTemplate(ctx, businessID, domain.CreateTemplateParams{
		Name:        req.Name,
		TriggerType: req.TriggerType,
		RoleID:      req.RoleID,
	})
	if err != nil {
		if errors.Is(err, domain.ErrChecklistAccessDenied) {
			return CreateTemplateResponse{}, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrEmptyTemplateName) || errors.Is(err, domain.ErrInvalidTriggerType) {
			return CreateTemplateResponse{}, onekiterrors.New(http.StatusBadRequest, err.Error())
		}
		logger.Ctx(ctx).Error("failed to create onboarding template", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateTemplateResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create template")
	}

	return CreateTemplateResponse{Template: templateToProto(*created)}, nil
}
