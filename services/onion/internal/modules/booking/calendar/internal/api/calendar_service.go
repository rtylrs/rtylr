package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/booking/calendar/internal/application"
	"github.com/voxire/onion/internal/modules/booking/calendar/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	bookingcalendarv1 "github.com/voxire/voxie/gen/go/booking/calendar/v1"
)

type CalendarAPI struct {
	service *application.Service
}

func NewCalendarAPI(service *application.Service) *CalendarAPI {
	return &CalendarAPI{service: service}
}

func (a *CalendarAPI) GetCalendar(ctx context.Context, req *bookingcalendarv1.GetCalendarRequest) (*bookingcalendarv1.GetCalendarResponse, error) {
	from := strings.TrimSpace(req.GetFromDate())
	to := strings.TrimSpace(req.GetToDate())
	if from == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "from is required")
	}
	if to == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "to is required")
	}

	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	days, err := a.service.GetCalendar(ctx, businessID, from, to)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get calendar")
	}

	return &bookingcalendarv1.GetCalendarResponse{Days: newCalendarDays(days)}, nil
}

func newCalendarDays(days []domain.CalendarDay) []*bookingcalendarv1.CalendarDay {
	out := make([]*bookingcalendarv1.CalendarDay, 0, len(days))
	for _, day := range days {
		appointments := make([]*bookingcalendarv1.CalendarAppointment, 0, len(day.Appointments))
		for _, appt := range day.Appointments {
			appointments = append(appointments, &bookingcalendarv1.CalendarAppointment{
				Id:           appt.ID,
				BusinessId:   appt.BusinessID,
				ServiceId:    appt.ServiceID,
				CustomerName: appt.CustomerName,
				StartTime:    appt.StartTime,
				EndTime:      appt.EndTime,
				Status:       appt.Status,
			})
		}
		out = append(out, &bookingcalendarv1.CalendarDay{
			Date:         day.Date,
			Appointments: appointments,
		})
	}
	return out
}
