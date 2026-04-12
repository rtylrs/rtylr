package handlers

import (
	"context"
	"net/http"

	hubv1 "github.com/voxire/voxie/gen/go/hub/v1"

	"github.com/voxire/onion/internal/modules/notifications/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func NotificationRoutes() []string    { return []string{"/notifications"} }
func ReadAllRoutes() []string         { return []string{"/notifications/read-all"} }

// -- List --

type ListHandler struct{ Service *application.Service }
type ListRequest struct{}

type ListResponse struct {
	Notifications []*hubv1.Notification `json:"notifications"`
}

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}

	notifications, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to list notifications",
		}
	}

	payloads := make([]*hubv1.Notification, 0, len(notifications))
	for _, n := range notifications {
		payloads = append(payloads, &hubv1.Notification{
			Id:        n.ID,
			Title:     n.Title,
			Body:      n.Body,
			Type:      n.Type,
			Read:      n.IsRead,
			Link:      n.Link,
			CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return ListResponse{Notifications: payloads}, nil
}

// -- Mark Read --

type MarkReadHandler struct{ Service *application.Service }

type MarkReadRequest struct {
	ID int64 `json:"id"`
}

type MarkReadResponse struct {
	Message string `json:"message"`
}

func (h *MarkReadHandler) Validate(_ context.Context, req *MarkReadRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Notification id is required",
		}
	}
	return nil
}

func (h *MarkReadHandler) Serve(ctx context.Context, req *MarkReadRequest) (MarkReadResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return MarkReadResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.MarkRead(ctx, device.User.UserID, req.ID); err != nil {
		return MarkReadResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to mark notification as read",
		}
	}

	return MarkReadResponse{Message: "Notification marked as read"}, nil
}

// -- Mark All Read --

type MarkAllReadHandler struct{ Service *application.Service }

type MarkAllReadRequest struct{}

type MarkAllReadResponse struct {
	Message string `json:"message"`
}

func (h *MarkAllReadHandler) Serve(ctx context.Context, _ *MarkAllReadRequest) (MarkAllReadResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return MarkAllReadResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.MarkAllRead(ctx, device.User.UserID); err != nil {
		return MarkAllReadResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to mark all notifications as read",
		}
	}

	return MarkAllReadResponse{Message: "All notifications marked as read"}, nil
}
