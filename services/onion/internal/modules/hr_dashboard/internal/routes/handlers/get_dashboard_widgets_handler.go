package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_dashboard/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
)

func GetDashboardWidgetsRoutes() []string { return []string{"/hr/dashboard/widgets"} }

type GetDashboardWidgetsHandler struct{ Service *application.Service }

type GetDashboardWidgetsRequest struct{}

type DashboardWidgetPayload struct {
	ID         int64           `json:"id"`
	OwnerID    int64           `json:"owner_id"`
	RoleID     int64           `json:"role_id"`
	WidgetType string          `json:"widget_type"`
	PositionX  int32           `json:"position_x"`
	PositionY  int32           `json:"position_y"`
	Width      int32           `json:"width"`
	Height     int32           `json:"height"`
	Config     json.RawMessage `json:"config"`
	CreatedAt  int64           `json:"created_at"`
	UpdatedAt  int64           `json:"updated_at"`
}

type GetDashboardWidgetsResponse struct {
	Widgets []*DashboardWidgetPayload `json:"widgets"`
	Data    map[string]interface{}    `json:"data"`
}

func (h *GetDashboardWidgetsHandler) Serve(ctx context.Context, _ *GetDashboardWidgetsRequest) (GetDashboardWidgetsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetDashboardWidgetsResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetDashboardWidgetsResponse{}, errorModels.PermissionDeniedServerError()
	}

	widgets, err := h.Service.GetWidgets(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to get dashboard widgets", logger.Any("business_id", businessID), logger.Error(err))
		return GetDashboardWidgetsResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to get dashboard widgets"}
	}

	payloads := make([]*DashboardWidgetPayload, 0, len(widgets))
	for _, w := range widgets {
		config := json.RawMessage(w.Config)
		if len(config) == 0 {
			config = json.RawMessage("{}")
		}
		payloads = append(payloads, &DashboardWidgetPayload{
			ID:         w.ID,
			OwnerID:    w.OwnerID,
			RoleID:     w.RoleID,
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

	return GetDashboardWidgetsResponse{Widgets: payloads, Data: map[string]interface{}{}}, nil
}
