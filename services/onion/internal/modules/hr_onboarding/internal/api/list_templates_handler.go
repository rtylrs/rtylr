package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/application"
	"github.com/voxire/onion/internal/modules/hr_onboarding/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hronboardingv1 "github.com/voxire/voxie/gen/go/hr/onboarding/v1"
)

func ListTemplatesRoutes() []string { return []string{"/hr/onboarding/templates"} }

type ListTemplatesHandler struct{ Service *application.Service }

type ListTemplatesRequest struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListTemplatesResponse struct {
	Templates []*hronboardingv1.ChecklistTemplatePayload `json:"templates"`
	Total     int64                                      `json:"total"`
	Limit     int32                                      `json:"limit"`
	Offset    int32                                      `json:"offset"`
}

func (h *ListTemplatesHandler) Serve(ctx context.Context, req *ListTemplatesRequest) (ListTemplatesResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListTemplatesResponse{}, onekiterrors.PermissionDenied()
	}

	templates, total, err := h.Service.ListTemplates(ctx, device.User.UserID, domain.ListTemplatesParams{
		Limit:  req.Limit,
		Offset: req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list onboarding templates", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListTemplatesResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list templates")
	}

	payloads := make([]*hronboardingv1.ChecklistTemplatePayload, 0, len(templates))
	for _, t := range templates {
		payloads = append(payloads, templateToProto(t))
	}

	return ListTemplatesResponse{Templates: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}
