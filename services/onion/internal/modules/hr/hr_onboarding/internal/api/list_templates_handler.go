package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr/hr_onboarding/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_onboarding/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hronboardingv1 "github.com/voxire/voxie/gen/go/hr/onboarding/v1"
)

// Mirrors the default/max clamp applied in application.Service.ListTemplates
// so an unbounded or absent caller-provided limit never reaches the query
// layer, and so the response echoes the value actually used.
const (
	defaultListTemplatesLimit int32 = 20
	maxListTemplatesLimit     int32 = 100
)

type ListTemplatesHandler struct{ Service *application.Service }

type ListTemplatesRequest struct {
	Limit       int32   `json:"limit"`
	Offset      int32   `json:"offset"`
	TriggerType *string `json:"trigger_type"`
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

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListTemplatesResponse{}, onekiterrors.PermissionDenied()
	}
	if req == nil {
		req = &ListTemplatesRequest{}
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultListTemplatesLimit
	}
	if limit > maxListTemplatesLimit {
		limit = maxListTemplatesLimit
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	templates, total, err := h.Service.ListTemplates(ctx, businessID, domain.ListTemplatesParams{
		Limit:       limit,
		Offset:      offset,
		TriggerType: req.TriggerType,
	})
	if err != nil {
		if errors.Is(err, domain.ErrChecklistAccessDenied) {
			return ListTemplatesResponse{}, onekiterrors.PermissionDenied()
		}
		logger.Ctx(ctx).Error("failed to list onboarding templates", logger.Any("business_id", businessID), logger.Error(err))
		return ListTemplatesResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list templates")
	}

	payloads := make([]*hronboardingv1.ChecklistTemplatePayload, 0, len(templates))
	for _, t := range templates {
		payloads = append(payloads, templateToProto(t))
	}

	return ListTemplatesResponse{Templates: payloads, Total: total, Limit: limit, Offset: offset}, nil
}
