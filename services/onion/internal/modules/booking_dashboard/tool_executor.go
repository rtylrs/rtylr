package booking_dashboard

import (
	"context"
	"errors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
)

const getBookingDashboardToolName = "booking.get_dashboard"

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, _ map[string]any) (map[string]any, bool, error) {
	if name != getBookingDashboardToolName {
		return nil, false, nil
	}
	ownerID, err := bookingDashboardToolOwnerID(ctx)
	if err != nil {
		return nil, true, err
	}
	dashboard, err := m.service.GetDashboard(ctx, ownerID)
	if err != nil {
		return nil, true, err
	}
	return map[string]any{"dashboard": dashboard}, true, nil
}

func bookingDashboardToolOwnerID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, errors.New("authentication is required")
	}
	return authUser.UserID, nil
}
