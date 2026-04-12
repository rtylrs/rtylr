package handlers

import (
	"context"
	"net/http"

	hubv1 "github.com/voxire/voxie/gen/go/hub/v1"

	"github.com/voxire/onion/internal/modules/webhooks/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func WebhookRoutes() []string { return []string{"/webhooks"} }

// -- List --

type ListHandler struct{ Service *application.Service }
type ListRequest struct{}

type WebhookPayload struct {
	ID              int64      `json:"id"`
	URL             string     `json:"url"`
	Events          []string   `json:"events"`
	IsActive        bool       `json:"is_active"`
	LastTriggeredAt *time.Time `json:"last_triggered_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

type ListResponse struct {
	Webhooks []WebhookPayload `json:"webhooks"`
}

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}

	webhooks, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to list webhooks",
		}
	}

	payloads := make([]WebhookPayload, 0, len(webhooks))
	for _, w := range webhooks {
		events := w.Events
		if events == nil {
			events = []string{}
		}
		payloads = append(payloads, WebhookPayload{
			ID:              w.ID,
			URL:             w.URL,
			Events:          events,
			IsActive:        w.IsActive,
			LastTriggeredAt: w.LastTriggeredAt,
			CreatedAt:       w.CreatedAt,
		})
	}

	return ListResponse{Webhooks: payloads}, nil
}

// -- Create --

type CreateHandler struct{ Service *application.Service }

type CreateRequest struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.URL == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "URL is required",
		}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}

	id, err := h.Service.Create(ctx, device.User.UserID, req.URL, req.Events)
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to create webhook",
		}
	}

	return CreateResponse{ID: id, Message: "Webhook created"}, nil
}

// -- Delete --

type DeleteHandler struct{ Service *application.Service }

type DeleteRequest struct {
	ID int64 `json:"id"`
}

type DeleteResponse struct {
	Message string `json:"message"`
}

func (h *DeleteHandler) Validate(_ context.Context, req *DeleteRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Webhook id is required",
		}
	}
	return nil
}

func (h *DeleteHandler) Serve(ctx context.Context, req *DeleteRequest) (DeleteResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.Delete(ctx, device.User.UserID, req.ID); err != nil {
		return DeleteResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to delete webhook",
		}
	}

	return DeleteResponse{Message: "Webhook deleted"}, nil
}
