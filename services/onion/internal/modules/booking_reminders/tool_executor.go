package booking_reminders

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/booking_reminders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const listBookingRemindersToolName = "booking.list_reminders"

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, _ map[string]any) (map[string]any, bool, error) {
	if name != listBookingRemindersToolName {
		return nil, false, nil
	}

	businessID, err := bookingRemindersToolBusinessID(ctx)
	if err != nil {
		return nil, true, err
	}

	reminders, err := m.service.List(ctx, businessID)
	if err != nil {
		return nil, true, err
	}

	payloads := make([]map[string]any, 0, len(reminders))
	unsent := 0
	for _, reminder := range reminders {
		if reminder.SentAt.IsZero() {
			unsent++
		}
		payloads = append(payloads, bookingReminderToolPayload(reminder))
	}

	return map[string]any{
		"reminders": payloads,
		"returned":  len(payloads),
		"unsent":    unsent,
	}, true, nil
}

func bookingRemindersToolBusinessID(ctx context.Context) (int64, error) {
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

func bookingReminderToolPayload(reminder domain.Reminder) map[string]any {
	return map[string]any{
		"id":             reminder.ID,
		"appointment_id": reminder.AppointmentID,
		"type":           reminder.Type,
		"scheduled_at":   reminder.ScheduledAt,
		"sent_at":        reminder.SentAt,
		"created_at":     reminder.CreatedAt,
	}
}
