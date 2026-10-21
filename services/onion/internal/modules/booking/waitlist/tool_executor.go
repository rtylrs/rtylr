package booking_waitlist

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/booking/waitlist/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const listBookingWaitlistToolName = "booking.list_waitlist"

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	if name != listBookingWaitlistToolName {
		return nil, false, nil
	}

	ownerID, err := bookingWaitlistToolOwnerID(ctx)
	if err != nil {
		return nil, true, err
	}
	status := strings.ToLower(strings.TrimSpace(bookingWaitlistToolString(firstBookingWaitlistToolValue(input, "status"))))

	entries, err := m.service.List(ctx, ownerID)
	if err != nil {
		return nil, true, err
	}

	payloads := make([]map[string]any, 0, len(entries))
	for _, entry := range entries {
		if status != "" && !strings.EqualFold(entry.Status, status) {
			continue
		}
		payloads = append(payloads, bookingWaitlistToolPayload(entry))
	}

	return map[string]any{
		"entries":  payloads,
		"returned": len(payloads),
		"filters": map[string]any{
			"status": status,
		},
	}, true, nil
}

func bookingWaitlistToolOwnerID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, errors.New("authentication is required")
	}
	businessID, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return 0, fmt.Errorf("business scope could not be resolved: %w", err)
	}
	return businessID, nil
}

func firstBookingWaitlistToolValue(input map[string]any, keys ...string) any {
	for _, key := range keys {
		value, ok := input[key]
		if !ok || value == nil {
			continue
		}
		if s, ok := value.(string); ok && strings.TrimSpace(s) == "" {
			continue
		}
		return value
	}
	return nil
}

func bookingWaitlistToolString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		return v.String()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

func bookingWaitlistToolPayload(entry domain.WaitlistEntry) map[string]any {
	return map[string]any{
		"id":             entry.ID,
		"service_id":     entry.ServiceID,
		"customer_name":  entry.CustomerName,
		"customer_email": entry.CustomerEmail,
		"customer_phone": entry.CustomerPhone,
		"preferred_date": entry.PreferredDate,
		"preferred_time": entry.PreferredTime,
		"status":         entry.Status,
		"created_at":     entry.CreatedAt,
	}
}
