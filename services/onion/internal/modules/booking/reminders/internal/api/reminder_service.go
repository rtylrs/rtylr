package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/booking/reminders/internal/application"
	"github.com/voxire/onion/internal/modules/booking/reminders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	bookingreminderv1 "github.com/voxire/voxie/gen/go/booking/reminder/v1"
)

type ReminderAPI struct {
	service *application.Service
}

func NewReminderAPI(service *application.Service) *ReminderAPI {
	return &ReminderAPI{service: service}
}

func (a *ReminderAPI) ListReminders(ctx context.Context, _ *bookingreminderv1.ListRemindersRequest) (*bookingreminderv1.ListRemindersResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	items, err := a.service.List(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list reminders")
	}
	return &bookingreminderv1.ListRemindersResponse{Reminders: newReminders(items)}, nil
}

func (a *ReminderAPI) CreateReminder(ctx context.Context, req *bookingreminderv1.CreateReminderRequest) (*bookingreminderv1.CreateReminderResponse, error) {
	reminderType := strings.TrimSpace(req.GetType())
	if req.GetAppointmentId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Appointment ID is required")
	}
	if reminderType == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Type is required")
	}
	scheduledAt := strings.TrimSpace(req.GetScheduledAt())
	if scheduledAt == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "scheduled_at is required")
	}

	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	id, err := a.service.Create(ctx, businessID, domain.ReminderInput{
		AppointmentID: req.GetAppointmentId(),
		Type:          reminderType,
		ScheduledAt:   scheduledAt,
	})
	if err != nil {
		if errors.Is(err, domain.ErrAppointmentNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Appointment not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create reminder")
	}
	return &bookingreminderv1.CreateReminderResponse{Id: id, Message: "Reminder created"}, nil
}

func newReminders(items []domain.Reminder) []*bookingreminderv1.Reminder {
	out := make([]*bookingreminderv1.Reminder, 0, len(items))
	for _, item := range items {
		out = append(out, &bookingreminderv1.Reminder{
			Id:            item.ID,
			OwnerId:       item.OwnerID,
			BusinessId:    item.BusinessID,
			AppointmentId: item.AppointmentID,
			Type:          item.Type,
			ScheduledAt:   formatTime(item.ScheduledAt),
			SentAt:        formatTime(item.SentAt),
			CreatedAt:     formatTime(item.CreatedAt),
		})
	}
	return out
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
