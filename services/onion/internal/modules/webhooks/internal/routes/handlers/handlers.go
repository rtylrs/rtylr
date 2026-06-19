package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/webhooks/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubdeveloperv1 "github.com/voxire/voxie/gen/go/hub/developer/v1"
)

func WebhookRoutes() []string { return []string{"/webhooks"} }

// -- List --

type (
	ListHandler struct{ Service *application.Service }
	ListRequest struct{}
)

type ListResponse struct {
	Webhooks []*hubdeveloperv1.Webhook `json:"webhooks"`
}

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}

	webhooks, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to list webhooks",
		}
	}

	payloads := make([]*hubdeveloperv1.Webhook, 0, len(webhooks))
	for _, w := range webhooks {
		events := w.Events
		if events == nil {
			events = []string{}
		}
		wh := &hubdeveloperv1.Webhook{
			Id:        w.ID,
			Url:       w.URL,
			Events:    events,
			IsActive:  w.IsActive,
			CreatedAt: w.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if w.LastTriggeredAt != nil {
			wh.LastTriggeredAt = w.LastTriggeredAt.Format("2006-01-02T15:04:05Z07:00")
		}
		payloads = append(payloads, wh)
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
	if strings.TrimSpace(req.URL) == "" {
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

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}

	id, err := h.Service.Create(ctx, device.User.UserID, businessID, req.URL, req.Events)
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

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.Delete(ctx, businessID, req.ID); err != nil {
		return DeleteResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to delete webhook",
		}
	}

	return DeleteResponse{Message: "Webhook deleted"}, nil
}
