package booking_dashboard

import (
	"context"
	"errors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const getBookingDashboardToolName = "booking.get_dashboard"

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, _ map[string]any) (map[string]any, bool, error) {
	if name != getBookingDashboardToolName {
		return nil, false, nil
	}
	businessID, err := bookingDashboardToolBusinessID(ctx)
	if err != nil {
		return nil, true, err
	}
	dashboard, err := m.service.GetDashboard(ctx, businessID)
	if err != nil {
		return nil, true, err
	}
	return map[string]any{"dashboard": dashboard}, true, nil
}

func bookingDashboardToolBusinessID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, errors.New("authentication is required")
	}
	businessID, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return 0, err
	}
	return businessID, nil
}
