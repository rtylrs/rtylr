package api

import (
	"context"
	"net/http"

	hrdashboardapp "github.com/voxire/onion/internal/modules/hr/hr_dashboard/internal/application"
	hrdashboarddomain "github.com/voxire/onion/internal/modules/hr/hr_dashboard/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_dashboard/stats"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrdashboardv1 "github.com/voxire/voxie/gen/go/hr/dashboard/v1"
)

type DashboardAPI struct {
	loadStats     stats.Loader
	WidgetService *hrdashboardapp.Service
}

func NewDashboardAPI(loadStats stats.Loader, widgetService *hrdashboardapp.Service) *DashboardAPI {
	return &DashboardAPI{loadStats: loadStats, WidgetService: widgetService}
}

func (a *DashboardAPI) GetHRDashboard(ctx context.Context, _ *hrdashboardv1.GetHRDashboardRequest) (*hrdashboardv1.GetHRDashboardResponse, error) {
	if a.loadStats == nil {
		return nil, onekiterrors.New(http.StatusNotImplemented, "Dashboard stats loader is not configured")
	}

	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	stats, err := a.loadStats(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get dashboard stats")
	}

	return &hrdashboardv1.GetHRDashboardResponse{
		Stats: &hrdashboardv1.DashboardStats{
			TotalStaff:     int64(stats.TotalStaff),
			OnShiftToday:   int64(stats.OnShiftToday),
			PendingTimeOff: int64(stats.PendingTimeOff),
			PayrollDraft:   int64(stats.PayrollDraft),
		},
	}, nil
}

func (a *DashboardAPI) GetDashboardWidgets(ctx context.Context, _ *hrdashboardv1.GetDashboardWidgetsRequest) (*hrdashboardv1.GetDashboardWidgetsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	widgets, err := a.WidgetService.GetWidgets(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to get dashboard widgets", logger.Any("business_id", businessID), logger.Error(err))
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get dashboard widgets")
	}

	payloads := make([]*hrdashboardv1.DashboardWidgetPayload, 0, len(widgets))
	for _, w := range widgets {
		config := w.Config
		if len(config) == 0 {
			config = []byte("{}")
		}
		payloads = append(payloads, &hrdashboardv1.DashboardWidgetPayload{
			Id:         w.ID,
			OwnerId:    w.OwnerID,
			RoleId:     w.RoleID,
			WidgetType: w.WidgetType,
			PositionX:  w.PositionX,
			PositionY:  w.PositionY,
			Width:      w.Width,
			Height:     w.Height,
			Config:     config,
			CreatedAt:  w.CreatedAt.Unix(),
			UpdatedAt:  w.UpdatedAt.Unix(),
		})
	}

	return &hrdashboardv1.GetDashboardWidgetsResponse{Widgets: payloads, Data: map[string]*hrdashboardv1.DashboardDataPayload{}}, nil
}

func (a *DashboardAPI) UpdateDashboardWidgets(ctx context.Context, req *hrdashboardv1.UpdateDashboardWidgetsRequest) (*hrdashboardv1.UpdateDashboardWidgetsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	widgets := make([]*hrdashboarddomain.DashboardWidget, 0, len(req.GetWidgets()))
	for _, w := range req.GetWidgets() {
		if w.GetPositionX() < 0 || w.GetPositionY() < 0 {
			return nil, onekiterrors.New(http.StatusBadRequest, "widget position values must be non-negative")
		}
		if w.GetWidth() < 0 || w.GetHeight() < 0 {
			return nil, onekiterrors.New(http.StatusBadRequest, "widget width and height must be non-negative")
		}
		widgets = append(widgets, &hrdashboarddomain.DashboardWidget{
			OwnerID:    device.User.UserID,
			BusinessID: businessID,
			RoleID:     w.GetRoleId(),
			WidgetType: w.GetWidgetType(),
			PositionX:  w.GetPositionX(),
			PositionY:  w.GetPositionY(),
			Width:      w.GetWidth(),
			Height:     w.GetHeight(),
			Config:     w.GetConfig(),
		})
	}

	if err := a.WidgetService.UpsertWidgets(ctx, businessID, widgets); err != nil {
		logger.Ctx(ctx).Error("failed to update dashboard widgets", logger.Any("business_id", businessID), logger.Error(err))
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update dashboard widgets")
	}

	return &hrdashboardv1.UpdateDashboardWidgetsResponse{Success: true}, nil
}
