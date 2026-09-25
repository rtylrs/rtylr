package booking_resources

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/booking_resources/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const listBookingResourcesToolName = "booking.list_resources"

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, _ map[string]any) (map[string]any, bool, error) {
	if name != listBookingResourcesToolName {
		return nil, false, nil
	}

	businessID, err := bookingResourcesToolBusinessID(ctx)
	if err != nil {
		return nil, true, err
	}

	resources, err := m.service.List(ctx, businessID)
	if err != nil {
		return nil, true, err
	}

	payloads := make([]map[string]any, 0, len(resources))
	for _, resource := range resources {
		payloads = append(payloads, bookingResourceToolPayload(resource))
	}

	return map[string]any{
		"resources": payloads,
		"returned":  len(payloads),
	}, true, nil
}

func bookingResourcesToolBusinessID(ctx context.Context) (int64, error) {
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

func bookingResourceToolPayload(resource domain.Resource) map[string]any {
	return map[string]any{
		"id":        resource.ID,
		"name":      resource.Name,
		"type":      resource.Type,
		"capacity":  resource.Capacity,
		"is_active": resource.IsActive,
	}
}
