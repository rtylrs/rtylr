package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_time/internal/application"
	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func ExportBillableRoutes() []string { return []string{"/hr/time-entries/export-billable"} }

type ExportBillableHandler struct{ Service *application.Service }

type ExportBillableRequest struct{}

type ExportBillableResponse struct {
	BatchID      int64                       `json:"batch_id"`
	Items        []domain.BillableExportItem `json:"items"`
	TotalMinutes int32                       `json:"total_minutes"`
	Count        int32                       `json:"count"`
}

func (h *ExportBillableHandler) Validate(_ context.Context, _ *ExportBillableRequest) error {
	return nil
}

func (h *ExportBillableHandler) Serve(ctx context.Context, _ *ExportBillableRequest) (ExportBillableResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ExportBillableResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ExportBillableResponse{}, onekiterrors.PermissionDenied()
	}

	items, batchID, err := h.Service.ExportBillableEntries(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to export billable entries",
			logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ExportBillableResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to export billable entries")
	}

	if items == nil {
		items = []domain.BillableExportItem{}
	}

	var totalMinutes int32
	for _, item := range items {
		totalMinutes += item.DurationMinutes
	}

	return ExportBillableResponse{
		BatchID:      batchID,
		Items:        items,
		TotalMinutes: totalMinutes,
		Count:        int32(len(items)),
	}, nil
}
