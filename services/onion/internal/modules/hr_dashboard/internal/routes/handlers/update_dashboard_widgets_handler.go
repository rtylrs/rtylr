package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_dashboard/internal/application"
	"github.com/voxire/onion/internal/modules/hr_dashboard/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
)

func UpdateDashboardWidgetsRoutes() []string { return []string{"/hr/dashboard/widgets"} }

type UpdateDashboardWidgetsHandler struct{ Service *application.Service }

type UpdateDashboardWidgetsRequest struct {
	Widgets []*DashboardWidgetPayload `json:"widgets"`
}

type UpdateDashboardWidgetsResponse struct {
	Success bool `json:"success"`
}

func (h *UpdateDashboardWidgetsHandler) Serve(ctx context.Context, req *UpdateDashboardWidgetsRequest) (UpdateDashboardWidgetsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateDashboardWidgetsResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateDashboardWidgetsResponse{}, errorModels.PermissionDeniedServerError()
	}

	widgets := make([]*domain.DashboardWidget, 0, len(req.Widgets))
	for _, w := range req.Widgets {
		widgets = append(widgets, &domain.DashboardWidget{
			OwnerID:    device.User.UserID,
			BusinessID: businessID,
			RoleID:     w.RoleID,
			WidgetType: w.WidgetType,
			PositionX:  w.PositionX,
			PositionY:  w.PositionY,
			Width:      w.Width,
			Height:     w.Height,
			Config:     []byte(w.Config),
		})
	}

	err := h.Service.UpsertWidgets(ctx, businessID, widgets)
	if err != nil {
		logger.Ctx(ctx).Error("failed to update dashboard widgets", logger.Any("business_id", businessID), logger.Error(err))
		return UpdateDashboardWidgetsResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to update dashboard widgets"}
	}

	return UpdateDashboardWidgetsResponse{Success: true}, nil
}
