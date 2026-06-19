package handlers

import (
	"context"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// GET /insights/reports/templates

func GetReportTemplatesRoutes() []string { return []string{"/insights/reports/templates"} }

type GetReportTemplatesHandler struct{ Service *application.Service }

type GetReportTemplatesRequest struct{}

type GetReportTemplatesResponse struct {
	Templates []application.ReportTemplatePayload `json:"templates"`
}

func (h *GetReportTemplatesHandler) Serve(ctx context.Context, _ *GetReportTemplatesRequest) (GetReportTemplatesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetReportTemplatesResponse{}, errorModels.PermissionDeniedServerError()
	}
	_, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetReportTemplatesResponse{}, errorModels.PermissionDeniedServerError()
	}
	return GetReportTemplatesResponse{Templates: h.Service.ListReportTemplates(ctx)}, nil
}
