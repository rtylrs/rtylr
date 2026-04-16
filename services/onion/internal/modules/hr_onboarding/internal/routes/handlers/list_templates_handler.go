package handlers

import (
	"context"
	"net/http"

	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/application"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

func ListTemplatesRoutes() []string { return []string{"/hr/onboarding/templates"} }

type ListTemplatesHandler struct{ Service *application.Service }

type ListTemplatesRequest struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListTemplatesResponse struct {
	Templates []*hrv1.ChecklistTemplatePayload `json:"templates"`
	Total     int64                            `json:"total"`
	Limit     int32                            `json:"limit"`
	Offset    int32                            `json:"offset"`
}

func (h *ListTemplatesHandler) Serve(ctx context.Context, req *ListTemplatesRequest) (ListTemplatesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListTemplatesResponse{}, errorModels.PermissionDeniedServerError()
	}

	templates, total, err := h.Service.ListTemplates(ctx, device.User.UserID, domain.ListTemplatesParams{
		Limit:  req.Limit,
		Offset: req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list onboarding templates", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListTemplatesResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to list templates"}
	}

	payloads := make([]*hrv1.ChecklistTemplatePayload, 0, len(templates))
	for _, t := range templates {
		payloads = append(payloads, templateToProto(t))
	}

	return ListTemplatesResponse{Templates: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}
