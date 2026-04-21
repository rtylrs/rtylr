package handlers

import (
	"context"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// POST /insights/reports/generate/revenue

func GenerateRevenueReportRoutes() []string { return []string{"/insights/reports/generate/revenue"} }

type GenerateRevenueReportHandler struct{ Service *application.Service }

type GenerateRevenueReportRequest struct {
	Format string `json:"format"`
	From   string `json:"from"`
	To     string `json:"to"`
}

type GenerateRevenueReportResponse = application.GenerateReportResult

func (h *GenerateRevenueReportHandler) Serve(ctx context.Context, _ *GenerateRevenueReportRequest) (GenerateRevenueReportResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GenerateRevenueReportResponse{}, errorModels.PermissionDeniedServerError()
	}
	return h.Service.QueueRevenueReport(ctx, device.User.UserID), nil
}
