package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/kpis/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	hubv1 "github.com/voxire/voxie/gen/go/hub/v1"
)

func KPIRoutes() []string { return []string{"/hub/kpis"} }

type (
	KPIHandler struct{ Service *application.Service }
	KPIRequest struct{}
)

type KPIResponse = hubv1.GetKPIsResponse

func (h *KPIHandler) Serve(ctx context.Context, _ *KPIRequest) (KPIResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return KPIResponse{}, errorModels.PermissionDeniedServerError()
	}

	res, err := cacher.Of[KPIResponse]().GetOrComputeSingle(
		ctx,
		cacheKeys.BusinessKPIs(device.User.UserID),
		func(ctx context.Context) (KPIResponse, error) {
			kpi, err := h.Service.GetKPIs(ctx, device.User.UserID)
			if err != nil {
				return KPIResponse{}, err
			}
			return KPIResponse{
				Revenue:       kpi.Revenue,
				OrderCount:    kpi.OrderCount,
				StaffCount:    kpi.StaffCount,
				CustomerCount: kpi.CustomerCount,
				LowStockCount: kpi.LowStockCount,
				ActiveApps:    kpi.ActiveApps,
			}, nil
		},
		cacher.WithExpiry(ctx, 2*time.Minute),
	)
	if err != nil {
		return KPIResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to get KPIs",
		}
	}
	return res, nil
}
