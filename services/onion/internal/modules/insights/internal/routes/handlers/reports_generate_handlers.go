package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
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

func (h *GenerateRevenueReportHandler) Validate(_ context.Context, req *GenerateRevenueReportRequest) *errorModels.ServerError {
	const layout = "2006-01-02"
	req.Format = strings.TrimSpace(req.Format)
	req.From = strings.TrimSpace(req.From)
	req.To = strings.TrimSpace(req.To)
	if req.From != "" {
		if _, err := time.Parse(layout, req.From); err != nil {
			return &errorModels.ServerError{
				HTTPCode: http.StatusBadRequest,
				Code:     http.StatusBadRequest,
				Message:  "Invalid from date format, expected YYYY-MM-DD",
			}
		}
	}
	if req.To != "" {
		if _, err := time.Parse(layout, req.To); err != nil {
			return &errorModels.ServerError{
				HTTPCode: http.StatusBadRequest,
				Code:     http.StatusBadRequest,
				Message:  "Invalid to date format, expected YYYY-MM-DD",
			}
		}
	}
	if req.From != "" && req.To != "" {
		from, _ := time.Parse(layout, req.From)
		to, _ := time.Parse(layout, req.To)
		if to.Before(from) {
			return &errorModels.ServerError{
				HTTPCode: http.StatusBadRequest,
				Code:     http.StatusBadRequest,
				Message:  "to date must not be before from date",
			}
		}
	}
	return nil
}

func (h *GenerateRevenueReportHandler) Serve(ctx context.Context, _ *GenerateRevenueReportRequest) (GenerateRevenueReportResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GenerateRevenueReportResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GenerateRevenueReportResponse{}, errorModels.PermissionDeniedServerError()
	}
	return h.Service.QueueRevenueReport(ctx, businessID), nil
}
