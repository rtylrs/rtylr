package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/booking/appointments/internal/application"
	"github.com/voxire/onion/internal/modules/booking/appointments/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	bookingappointmentv1 "github.com/voxire/voxie/gen/go/booking/appointment/v1"
)

const appointmentsCacheScope = "booking_appointments"

type AppointmentAPI struct {
	service *application.Service
}

func NewAppointmentAPI(service *application.Service) *AppointmentAPI {
	return &AppointmentAPI{service: service}
}

func (a *AppointmentAPI) ListAppointments(ctx context.Context, _ *bookingappointmentv1.ListAppointmentsRequest) (*bookingappointmentv1.ListAppointmentsResponse, error) {
	authUser, businessID, errModel := resolveBookingUser(ctx)
	if errModel != nil {
		return nil, errModel
	}
	_ = authUser

	res, cacheErr := cacher.Of[[]domain.Appointment]().GetOrComputeSingle(
		ctx,
		cacheKeys.Scoped(appointmentsCacheScope, businessID, "list")+"|domain=v2",
		func(ctx context.Context) ([]domain.Appointment, error) {
			return a.service.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, 3*time.Minute),
	)
	if cacheErr == nil {
		return &bookingappointmentv1.ListAppointmentsResponse{Appointments: newAppointments(res)}, nil
	}

	items, err := a.service.List(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list appointments")
	}
	return &bookingappointmentv1.ListAppointmentsResponse{Appointments: newAppointments(items)}, nil
}

func (a *AppointmentAPI) ListTodayAppointments(ctx context.Context, _ *bookingappointmentv1.ListTodayAppointmentsRequest) (*bookingappointmentv1.ListTodayAppointmentsResponse, error) {
	_, businessID, errModel := resolveBookingUser(ctx)
	if errModel != nil {
		return nil, errModel
	}

	todayKey := "today:" + time.Now().UTC().Format("2006-01-02")
	res, cacheErr := cacher.Of[[]domain.Appointment]().GetOrComputeSingle(
		ctx,
		cacheKeys.Scoped(appointmentsCacheScope, businessID, todayKey)+"|domain=v2",
		func(ctx context.Context) ([]domain.Appointment, error) {
			return a.service.ListToday(ctx, businessID)
		},
		cacher.WithExpiry(ctx, 3*time.Minute),
	)
	if cacheErr == nil {
		return &bookingappointmentv1.ListTodayAppointmentsResponse{Appointments: newAppointments(res)}, nil
	}

	items, err := a.service.ListToday(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list today appointments")
	}
	return &bookingappointmentv1.ListTodayAppointmentsResponse{Appointments: newAppointments(items)}, nil
}

func (a *AppointmentAPI) CreateAppointment(ctx context.Context, req *bookingappointmentv1.CreateAppointmentRequest) (*bookingappointmentv1.CreateAppointmentResponse, error) {
	input, errModel := createAppointmentInput(req)
	if errModel != nil {
		return nil, errModel
	}

	_, businessID, errModel := resolveBookingUser(ctx)
	if errModel != nil {
		return nil, errModel
	}
	id, err := a.service.Create(ctx, businessID, input)
	if err != nil {
		if errors.Is(err, domain.ErrTimeSlotUnavailable) {
			return nil, onekiterrors.New(http.StatusConflict, err.Error())
		}
		if errors.Is(err, domain.ErrInvalidStatus) {
			return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
		}
		if errors.Is(err, domain.ErrInvalidReference) {
			return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create appointment")
	}
	invalidateAppointmentsCache(ctx, businessID)
	return &bookingappointmentv1.CreateAppointmentResponse{Id: id, Message: "Appointment created"}, nil
}

func (a *AppointmentAPI) UpdateAppointment(ctx context.Context, req *bookingappointmentv1.UpdateAppointmentRequest) (*bookingappointmentv1.UpdateAppointmentResponse, error) {
	if req.GetId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "ID is required")
	}
	input, errModel := updateAppointmentInput(req)
	if errModel != nil {
		return nil, errModel
	}

	_, businessID, errModel := resolveBookingUser(ctx)
	if errModel != nil {
		return nil, errModel
	}
	if err := a.service.Update(ctx, businessID, req.GetId(), input); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, onekiterrors.New(http.StatusNotFound, "Appointment not found")
		}
		if errors.Is(err, domain.ErrTimeSlotUnavailable) {
			return nil, onekiterrors.New(http.StatusConflict, err.Error())
		}
		if errors.Is(err, domain.ErrInvalidStatus) {
			return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
		}
		if errors.Is(err, domain.ErrInvalidReference) {
			return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update appointment")
	}
	invalidateAppointmentsCache(ctx, businessID)
	return &bookingappointmentv1.UpdateAppointmentResponse{Message: "Appointment updated"}, nil
}

func (a *AppointmentAPI) DeleteAppointment(ctx context.Context, req *bookingappointmentv1.DeleteAppointmentRequest) (*bookingappointmentv1.DeleteAppointmentResponse, error) {
	if req.GetId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "ID is required")
	}

	_, businessID, errModel := resolveBookingUser(ctx)
	if errModel != nil {
		return nil, errModel
	}
	if err := a.service.Delete(ctx, businessID, req.GetId()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, onekiterrors.New(http.StatusNotFound, "Appointment not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete appointment")
	}
	invalidateAppointmentsCache(ctx, businessID)
	return &bookingappointmentv1.DeleteAppointmentResponse{Message: "Appointment deleted"}, nil
}

func createAppointmentInput(req *bookingappointmentv1.CreateAppointmentRequest) (domain.AppointmentInput, error) {
	if req.GetServiceId() == 0 {
		return domain.AppointmentInput{}, onekiterrors.New(http.StatusBadRequest, "service_id is required")
	}
	customerName := strings.TrimSpace(req.GetCustomerName())
	if customerName == "" {
		return domain.AppointmentInput{}, onekiterrors.New(http.StatusBadRequest, "customer_name is required")
	}
	startTime := strings.TrimSpace(req.GetStartTime())
	endTime := strings.TrimSpace(req.GetEndTime())
	if startTime == "" || endTime == "" {
		return domain.AppointmentInput{}, onekiterrors.New(http.StatusBadRequest, "Start and end time are required")
	}
	if errModel := validateAppointmentTimes(startTime, endTime, true); errModel != nil {
		return domain.AppointmentInput{}, errModel
	}
	return domain.AppointmentInput{
		ServiceID:     req.GetServiceId(),
		ResourceID:    req.GetResourceId(),
		StaffID:       req.GetStaffId(),
		CustomerName:  customerName,
		CustomerEmail: req.GetCustomerEmail(),
		CustomerPhone: req.GetCustomerPhone(),
		StartTime:     startTime,
		EndTime:       endTime,
		Status:        req.GetStatus(),
		Notes:         req.GetNotes(),
	}, nil
}

func updateAppointmentInput(req *bookingappointmentv1.UpdateAppointmentRequest) (domain.AppointmentInput, error) {
	startTime := strings.TrimSpace(req.GetStartTime())
	endTime := strings.TrimSpace(req.GetEndTime())
	if errModel := validateAppointmentTimes(startTime, endTime, true); errModel != nil {
		return domain.AppointmentInput{}, errModel
	}
	return domain.AppointmentInput{
		ServiceID:     req.GetServiceId(),
		ResourceID:    req.GetResourceId(),
		StaffID:       req.GetStaffId(),
		CustomerName:  req.GetCustomerName(),
		CustomerEmail: req.GetCustomerEmail(),
		CustomerPhone: req.GetCustomerPhone(),
		StartTime:     startTime,
		EndTime:       endTime,
		Status:        req.GetStatus(),
		Notes:         req.GetNotes(),
	}, nil
}

func validateAppointmentTimes(startTime, endTime string, requireBoth bool) error {
	if requireBoth && (startTime == "" || endTime == "") {
		return onekiterrors.New(http.StatusBadRequest, "Start and end time are required")
	}

	var start, end time.Time
	if startTime != "" {
		parsed, err := time.Parse(time.RFC3339, startTime)
		if err != nil {
			return onekiterrors.New(http.StatusBadRequest, "start_time must be in RFC3339 format")
		}
		start = parsed
	}
	if endTime != "" {
		parsed, err := time.Parse(time.RFC3339, endTime)
		if err != nil {
			return onekiterrors.New(http.StatusBadRequest, "end_time must be in RFC3339 format")
		}
		end = parsed
	}
	if startTime != "" && endTime != "" && !start.Before(end) {
		return onekiterrors.New(http.StatusBadRequest, "start_time must be before end_time")
	}
	return nil
}

func resolveBookingUser(ctx context.Context) (*authPkg.AuthUser, int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, 0, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, 0, onekiterrors.PermissionDenied()
	}
	return authUser, businessID, nil
}

func invalidateAppointmentsCache(ctx context.Context, ownerID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(appointmentsCacheScope, ownerID))
}

func newAppointments(items []domain.Appointment) []*bookingappointmentv1.Appointment {
	out := make([]*bookingappointmentv1.Appointment, 0, len(items))
	for _, item := range items {
		out = append(out, &bookingappointmentv1.Appointment{
			Id:            item.ID,
			OwnerId:       item.OwnerID,
			BusinessId:    item.BusinessID,
			ServiceId:     item.ServiceID,
			ResourceId:    item.ResourceID,
			StaffId:       item.StaffID,
			CustomerName:  item.CustomerName,
			CustomerEmail: item.CustomerEmail,
			CustomerPhone: item.CustomerPhone,
			StartTime:     formatTime(item.StartTime),
			EndTime:       formatTime(item.EndTime),
			Status:        item.Status,
			Notes:         item.Notes,
			CreatedAt:     formatTime(item.CreatedAt),
			UpdatedAt:     formatTime(item.UpdatedAt),
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
