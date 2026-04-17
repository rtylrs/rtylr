package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/application"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
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
	Item *hrv1.ChecklistItemPayload `json:"item"`
}

func (h *AddItemToTemplateHandler) Validate(_ context.Context, req *AddItemToTemplateRequest) *errorModels.ServerError {
	if req.Title == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "title is required"}
	}
	return nil
}

func (h *AddItemToTemplateHandler) Serve(ctx context.Context, req *AddItemToTemplateRequest) (AddItemToTemplateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return AddItemToTemplateResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	templateID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return AddItemToTemplateResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid template id"}
	}

	item, err := h.Service.AddItemToTemplate(ctx, device.User.UserID, templateID, domain.CreateTemplateItemParams{
		Title:              req.Title,
		Description:        req.Description,
		AssigneeRole:       req.AssigneeRole,
		DueOffsetDays:      req.DueOffsetDays,
		RequiresAttachment: req.RequiresAttachment,
	})
	if err != nil {
		if err == domain.ErrTemplateNotFound {
			return AddItemToTemplateResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Template not found"}
		}
		logger.Ctx(ctx).Error("failed to add item to onboarding template", logger.Any("owner_id", device.User.UserID), logger.Any("template_id", templateID), logger.Error(err))
		return AddItemToTemplateResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to add item to template"}
	}

	return AddItemToTemplateResponse{Item: checklistItemToProto(*item)}, nil
}
