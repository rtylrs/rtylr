package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/booking_appointments/internal/application"
	"github.com/voxire/onion/internal/modules/booking_appointments/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// appointmentsCacheScope namespaces the per-owner appointment list caches.
// booking_appointments is written only within this module
// (Create/Update/Delete), so a short TTL plus invalidation on those
// mutations keeps both the full list and today views consistent.
const appointmentsCacheScope = "booking_appointments"

func invalidateAppointmentsCache(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(appointmentsCacheScope, ownerID))
	}
}

func ListRoutes() []string   { return []string{"/book/appointments"} }
func TodayRoutes() []string  { return []string{"/book/appointments/today"} }
func CreateRoutes() []string { return []string{"/book/appointments"} }
func UpdateRoutes() []string { return []string{"/book/appointments"} }
func DeleteRoutes() []string { return []string{"/book/appointments"} }

type (
	ListHandler  struct{ Service *application.Service }
	ListRequest  struct{}
	ListResponse struct {
		Appointments []domain.Appointment `json:"appointments"`
	}
)

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}
	res, cacheErr := cacher.Of[ListResponse]().GetOrComputeSingle(
		ctx,
		cacheKeys.Scoped(appointmentsCacheScope, businessID, "list"),
		func(ctx context.Context) (ListResponse, error) {
			items, err := h.Service.List(ctx, businessID)
			if err != nil {
				return ListResponse{}, err
			}
			return ListResponse{Appointments: items}, nil
		},
		cacher.WithExpiry(ctx, 3*time.Minute),
	)
	if cacheErr == nil {
		return res, nil
	}

	items, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to list appointments"}
	}
	return ListResponse{Appointments: items}, nil
}

type (
	TodayHandler  struct{ Service *application.Service }
	TodayRequest  struct{}
	TodayResponse struct {
		Appointments []domain.Appointment `json:"appointments"`
	}
)

func (h *TodayHandler) Serve(ctx context.Context, _ *TodayRequest) (TodayResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return TodayResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return TodayResponse{}, errorModels.PermissionDeniedServerError()
	}
	// Date-stamp the cache detail so a cache written yesterday can never be
	// served today, independent of the TTL.
	todayKey := "today:" + time.Now().Format("2006-01-02")
	res, cacheErr := cacher.Of[TodayResponse]().GetOrComputeSingle(
		ctx,
		cacheKeys.Scoped(appointmentsCacheScope, businessID, todayKey),
		func(ctx context.Context) (TodayResponse, error) {
			items, err := h.Service.ListToday(ctx, businessID)
			if err != nil {
				return TodayResponse{}, err
			}
			return TodayResponse{Appointments: items}, nil
		},
		cacher.WithExpiry(ctx, 3*time.Minute),
	)
	if cacheErr == nil {
		return res, nil
	}

	items, err := h.Service.ListToday(ctx, businessID)
	if err != nil {
		return TodayResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to list today appointments"}
	}
	return TodayResponse{Appointments: items}, nil
}

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		ServiceID     int64  `json:"service_id"`
		ResourceID    int64  `json:"resource_id"`
		StaffID       int64  `json:"staff_id"`
		CustomerName  string `json:"customer_name"`
		CustomerEmail string `json:"customer_email"`
		CustomerPhone string `json:"customer_phone"`
		StartTime     string `json:"start_time"`
		EndTime       string `json:"end_time"`
		Status        string `json:"status"`
		Notes         string `json:"notes"`
	}
)

type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.ServiceID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "service_id is required"}
	}
	req.CustomerName = strings.TrimSpace(req.CustomerName)
	if req.CustomerName == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "customer_name is required"}
	}
	req.StartTime = strings.TrimSpace(req.StartTime)
	req.EndTime = strings.TrimSpace(req.EndTime)
	if req.StartTime == "" || req.EndTime == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Start and end time are required"}
	}
	start, err := time.Parse("2006-01-02T15:04:05Z07:00", req.StartTime)
	if err != nil {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "start_time must be in RFC3339 format"}
	}
	end, err := time.Parse("2006-01-02T15:04:05Z07:00", req.EndTime)
	if err != nil {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "end_time must be in RFC3339 format"}
	}
	if !start.Before(end) {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "start_time must be before end_time"}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.Create(ctx, businessID, domain.AppointmentInput{
		ServiceID: req.ServiceID, ResourceID: req.ResourceID, StaffID: req.StaffID,
		CustomerName: req.CustomerName, CustomerEmail: req.CustomerEmail, CustomerPhone: req.CustomerPhone,
		StartTime: req.StartTime, EndTime: req.EndTime, Status: req.Status, Notes: req.Notes,
	})
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to create appointment"}
	}
	invalidateAppointmentsCache(ctx, businessID)
	return CreateResponse{ID: id, Message: "Appointment created"}, nil
}

type (
	UpdateHandler struct{ Service *application.Service }
	UpdateRequest struct {
		ID            int64  `json:"id"`
		ServiceID     int64  `json:"service_id"`
		ResourceID    int64  `json:"resource_id"`
		StaffID       int64  `json:"staff_id"`
		CustomerName  string `json:"customer_name"`
		CustomerEmail string `json:"customer_email"`
		CustomerPhone string `json:"customer_phone"`
		StartTime     string `json:"start_time"`
		EndTime       string `json:"end_time"`
		Status        string `json:"status"`
		Notes         string `json:"notes"`
	}
)

type UpdateResponse struct {
	Message string `json:"message"`
}

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"}
	}
	req.StartTime = strings.TrimSpace(req.StartTime)
	req.EndTime = strings.TrimSpace(req.EndTime)
	var start, end time.Time
	if req.StartTime != "" {
		parsed, err := time.Parse("2006-01-02T15:04:05Z07:00", req.StartTime)
		if err != nil {
			return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "start_time must be in RFC3339 format"}
		}
		start = parsed
	}
	if req.EndTime != "" {
		parsed, err := time.Parse("2006-01-02T15:04:05Z07:00", req.EndTime)
		if err != nil {
			return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "end_time must be in RFC3339 format"}
		}
		end = parsed
	}
	if req.StartTime != "" && req.EndTime != "" && !start.Before(end) {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "start_time must be before end_time"}
	}
	return nil
}

func (h *UpdateHandler) Serve(ctx context.Context, req *UpdateRequest) (UpdateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateResponse{}, errorModels.PermissionDeniedServerError()
	}
	if err := h.Service.Update(ctx, businessID, req.ID, domain.AppointmentInput{
		ServiceID: req.ServiceID, ResourceID: req.ResourceID, StaffID: req.StaffID,
		CustomerName: req.CustomerName, CustomerEmail: req.CustomerEmail, CustomerPhone: req.CustomerPhone,
		StartTime: req.StartTime, EndTime: req.EndTime, Status: req.Status, Notes: req.Notes,
	}); err != nil {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to update appointment"}
	}
	invalidateAppointmentsCache(ctx, businessID)
	return UpdateResponse{Message: "Appointment updated"}, nil
}

type (
	DeleteHandler struct{ Service *application.Service }
	DeleteRequest struct {
		ID int64 `json:"id"`
	}
)

type DeleteResponse struct {
	Message string `json:"message"`
}

func (h *DeleteHandler) Validate(_ context.Context, req *DeleteRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"}
	}
	return nil
}

func (h *DeleteHandler) Serve(ctx context.Context, req *DeleteRequest) (DeleteResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return DeleteResponse{}, errorModels.PermissionDeniedServerError()
	}
	if err := h.Service.Delete(ctx, businessID, req.ID); err != nil {
		return DeleteResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to delete appointment"}
	}
	invalidateAppointmentsCache(ctx, businessID)
	return DeleteResponse{Message: "Appointment deleted"}, nil
}
