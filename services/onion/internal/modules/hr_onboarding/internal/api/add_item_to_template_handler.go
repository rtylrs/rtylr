package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/application"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hronboardingv1 "github.com/voxire/voxie/gen/go/hr/onboarding/v1"
)

func AddItemToTemplateRoutes() []string { return []string{"/hr/onboarding/templates/{id}/items"} }

type AddItemToTemplateHandler struct{ Service *application.Service }

type AddItemToTemplateRequest struct {
	Title              string `json:"title"`
	Description        string `json:"description"`
	AssigneeRole       string `json:"assignee_role"`
	DueOffsetDays      int32  `json:"due_offset_days"`
	RequiresAttachment bool   `json:"requires_attachment"`
}

type AddItemToTemplateResponse struct {
	Item *hronboardingv1.ChecklistItemPayload `json:"item"`
}

func (h *AddItemToTemplateHandler) Validate(_ context.Context, req *AddItemToTemplateRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return onekiterrors.New(http.StatusBadRequest, "title is required")
	}
	return nil
}

func (h *AddItemToTemplateHandler) Serve(ctx context.Context, req *AddItemToTemplateRequest) (AddItemToTemplateResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return AddItemToTemplateResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireOnboardingManagePermission(ctx); err != nil {
		return AddItemToTemplateResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return AddItemToTemplateResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	templateID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return AddItemToTemplateResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid template id")
	}

	item, err := h.Service.AddItemToTemplate(ctx, businessID, templateID, domain.CreateTemplateItemParams{
		Title:              req.Title,
		Description:        req.Description,
		AssigneeRole:       req.AssigneeRole,
		DueOffsetDays:      req.DueOffsetDays,
		RequiresAttachment: req.RequiresAttachment,
	})
	if err != nil {
		if errors.Is(err, domain.ErrChecklistAccessDenied) {
			return AddItemToTemplateResponse{}, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrTemplateNotFound) {
			return AddItemToTemplateResponse{}, onekiterrors.New(http.StatusNotFound, "Template not found")
		}
		logger.Ctx(ctx).Error("failed to add item to onboarding template", logger.Any("owner_id", device.User.UserID), logger.Any("template_id", templateID), logger.Error(err))
		return AddItemToTemplateResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to add item to template")
	}

	return AddItemToTemplateResponse{Item: checklistItemToProto(*item)}, nil
}
