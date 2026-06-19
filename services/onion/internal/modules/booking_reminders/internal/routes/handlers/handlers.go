package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/booking_reminders/internal/application"
	"github.com/voxire/onion/internal/modules/booking_reminders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type Payload struct {
	ID            int64  `json:"id"`
	OwnerID       int64  `json:"owner_id"`
	AppointmentID int64  `json:"appointment_id"`
	Type          string `json:"type"`
	ScheduledAt   string `json:"scheduled_at"`
	SentAt        string `json:"sent_at"`
	CreatedAt     string `json:"created_at"`
}

func toPayload(r domain.Reminder) Payload {
	return Payload{
		ID: r.ID, OwnerID: r.OwnerID, AppointmentID: r.AppointmentID, Type: r.Type,
		ScheduledAt: r.ScheduledAt.Format("2006-01-02T15:04:05Z07:00"),
		SentAt:      r.SentAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func ListRoutes() []string   { return []string{"/book/reminders"} }
func CreateRoutes() []string { return []string{"/book/reminders"} }

type (
	ListHandler  struct{ Service *application.Service }
	ListRequest  struct{}
	ListResponse struct {
		Reminders []Payload `json:"reminders"`
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
	items, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to list reminders"}
	}
	out := make([]Payload, 0, len(items))
	for _, i := range items {
		out = append(out, toPayload(i))
	}
	return ListResponse{Reminders: out}, nil
}

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		AppointmentID int64  `json:"appointment_id"`
		Type          string `json:"type"`
		ScheduledAt   string `json:"scheduled_at"`
	}
)

type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.AppointmentID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Appointment ID is required"}
	}
	req.Type = strings.TrimSpace(req.Type)
	if req.Type == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Type is required"}
	}
	req.ScheduledAt = strings.TrimSpace(req.ScheduledAt)
	if req.ScheduledAt == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "scheduled_at is required"}
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
	id, err := h.Service.Create(ctx, businessID, domain.ReminderInput{AppointmentID: req.AppointmentID, Type: req.Type, ScheduledAt: req.ScheduledAt})
	if err != nil {
		if errors.Is(err, domain.ErrAppointmentNotFound) {
			return CreateResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, HTTPCode: http.StatusNotFound, Message: "Appointment not found"}
		}
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to create reminder"}
	}
	return CreateResponse{ID: id, Message: "Reminder created"}, nil
}
