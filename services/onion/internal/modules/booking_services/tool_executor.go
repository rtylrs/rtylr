package booking_services

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/booking_services/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
)

const listBookingServicesToolName = "booking.list_services"

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, _ map[string]any) (map[string]any, bool, error) {
	if name != listBookingServicesToolName {
		return nil, false, nil
	}

	ownerID, err := bookingServicesToolOwnerID(ctx)
	if err != nil {
		return nil, true, err
	}

	services, err := m.service.List(ctx, ownerID)
	if err != nil {
		return nil, true, err
	}

	payloads := make([]map[string]any, 0, len(services))
	for _, service := range services {
		payloads = append(payloads, bookingServiceToolPayload(service))
	}

	return map[string]any{
		"services": payloads,
		"returned": len(payloads),
	}, true, nil
}

func bookingServicesToolOwnerID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, errors.New("authentication is required")
	}
	return authUser.UserID, nil
}

func bookingServiceToolPayload(service domain.BookingService) map[string]any {
	return map[string]any{
		"id":               service.ID,
		"name":             service.Name,
		"description":      service.Description,
		"duration_minutes": service.DurationMinutes,
		"price":            service.Price,
		"color":            service.Color,
		"is_active":        service.IsActive,
	}
}
