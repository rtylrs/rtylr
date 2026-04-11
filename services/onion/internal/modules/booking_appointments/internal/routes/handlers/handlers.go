package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/booking_appointments/internal/application"
	"github.com/voxire/onion/internal/modules/booking_appointments/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

type Payload struct {
	ID int64 `json:"id"`; OwnerID int64 `json:"ownerId"`; ServiceID int64 `json:"serviceId"`
	ResourceID int64 `json:"resourceId"`; StaffID int64 `json:"staffId"`
	CustomerName string `json:"customerName"`; CustomerEmail string `json:"customerEmail"`; CustomerPhone string `json:"customerPhone"`
	StartTime string `json:"startTime"`; EndTime string `json:"endTime"`; Status string `json:"status"`
	Notes string `json:"notes"`; CreatedAt string `json:"createdAt"`; UpdatedAt string `json:"updatedAt"`
}

func toPayload(a domain.Appointment) Payload {
	return Payload{ID: a.ID, OwnerID: a.OwnerID, ServiceID: a.ServiceID, ResourceID: a.ResourceID, StaffID: a.StaffID,
		CustomerName: a.CustomerName, CustomerEmail: a.CustomerEmail, CustomerPhone: a.CustomerPhone,
		StartTime: a.StartTime.Format("2006-01-02T15:04:05Z07:00"), EndTime: a.EndTime.Format("2006-01-02T15:04:05Z07:00"),
		Status: a.Status, Notes: a.Notes,
		CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), UpdatedAt: a.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")}
}

func ListRoutes() []string   { return []string{"/book/appointments"} }
func TodayRoutes() []string  { return []string{"/book/appointments/today"} }
func CreateRoutes() []string { return []string{"/book/appointments"} }
func UpdateRoutes() []string { return []string{"/book/appointments"} }
func DeleteRoutes() []string { return []string{"/book/appointments"} }

type ListHandler struct{ Service *application.Service }
type ListRequest struct{}
type ListResponse struct{ Appointments []Payload `json:"appointments"` }

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return ListResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	items, err := h.Service.List(ctx, authUser.UserID)
	if err != nil { return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to list appointments"} }
	out := make([]Payload, 0, len(items)); for _, i := range items { out = append(out, toPayload(i)) }
	return ListResponse{Appointments: out}, nil
}

type TodayHandler struct{ Service *application.Service }
type TodayRequest struct{}
type TodayResponse struct{ Appointments []Payload `json:"appointments"` }

func (h *TodayHandler) Serve(ctx context.Context, _ *TodayRequest) (TodayResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return TodayResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	items, err := h.Service.ListToday(ctx, authUser.UserID)
	if err != nil { return TodayResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to list today appointments"} }
	out := make([]Payload, 0, len(items)); for _, i := range items { out = append(out, toPayload(i)) }
	return TodayResponse{Appointments: out}, nil
}

type CreateHandler struct{ Service *application.Service }
type CreateRequest struct {
	ServiceID int64 `json:"serviceId"`; ResourceID int64 `json:"resourceId"`; StaffID int64 `json:"staffId"`
	CustomerName string `json:"customerName"`; CustomerEmail string `json:"customerEmail"`; CustomerPhone string `json:"customerPhone"`
	StartTime string `json:"startTime"`; EndTime string `json:"endTime"`; Status string `json:"status"`; Notes string `json:"notes"`
}
type CreateResponse struct { ID int64 `json:"id"`; Message string `json:"message"` }

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.StartTime == "" || req.EndTime == "" { return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Start and end time are required"} }; return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return CreateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	id, err := h.Service.Create(ctx, authUser.UserID, domain.AppointmentInput{
		ServiceID: req.ServiceID, ResourceID: req.ResourceID, StaffID: req.StaffID,
		CustomerName: req.CustomerName, CustomerEmail: req.CustomerEmail, CustomerPhone: req.CustomerPhone,
		StartTime: req.StartTime, EndTime: req.EndTime, Status: req.Status, Notes: req.Notes,
	})
	if err != nil { return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to create appointment"} }
	return CreateResponse{ID: id, Message: "Appointment created"}, nil
}

type UpdateHandler struct{ Service *application.Service }
type UpdateRequest struct {
	ID int64 `json:"id"`; ServiceID int64 `json:"serviceId"`; ResourceID int64 `json:"resourceId"`; StaffID int64 `json:"staffId"`
	CustomerName string `json:"customerName"`; CustomerEmail string `json:"customerEmail"`; CustomerPhone string `json:"customerPhone"`
	StartTime string `json:"startTime"`; EndTime string `json:"endTime"`; Status string `json:"status"`; Notes string `json:"notes"`
}
type UpdateResponse struct{ Message string `json:"message"` }

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) *errorModels.ServerError {
	if req.ID == 0 { return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"} }; return nil
}

func (h *UpdateHandler) Serve(ctx context.Context, req *UpdateRequest) (UpdateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	if err := h.Service.Update(ctx, authUser.UserID, req.ID, domain.AppointmentInput{
		ServiceID: req.ServiceID, ResourceID: req.ResourceID, StaffID: req.StaffID,
		CustomerName: req.CustomerName, CustomerEmail: req.CustomerEmail, CustomerPhone: req.CustomerPhone,
		StartTime: req.StartTime, EndTime: req.EndTime, Status: req.Status, Notes: req.Notes,
	}); err != nil {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to update appointment"}
	}
	return UpdateResponse{Message: "Appointment updated"}, nil
}

type DeleteHandler struct{ Service *application.Service }
type DeleteRequest struct{ ID int64 `json:"id"` }
type DeleteResponse struct{ Message string `json:"message"` }

func (h *DeleteHandler) Validate(_ context.Context, req *DeleteRequest) *errorModels.ServerError {
	if req.ID == 0 { return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"} }; return nil
}

func (h *DeleteHandler) Serve(ctx context.Context, req *DeleteRequest) (DeleteResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return DeleteResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	if err := h.Service.Delete(ctx, authUser.UserID, req.ID); err != nil {
		return DeleteResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to delete appointment"}
	}
	return DeleteResponse{Message: "Appointment deleted"}, nil
}
