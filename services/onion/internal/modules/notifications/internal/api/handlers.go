package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/notifications/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubnotificationv1 "github.com/voxire/voxie/gen/go/hub/notification/v1"
)

func NotificationRoutes() []string { return []string{"/notifications"} }
func ReadAllRoutes() []string      { return []string{"/notifications/read-all"} }

// -- List --

type (
	ListHandler struct{ Service *application.Service }
	ListRequest struct{}
)

type ListResponse struct {
	Notifications []*hubnotificationv1.Notification `json:"notifications"`
}

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListResponse{}, onekiterrors.PermissionDenied()
	}

	notifications, err := h.Service.List(ctx, businessID, device.User.UserID)
	if err != nil {
		return ListResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list notifications")
	}

	payloads := make([]*hubnotificationv1.Notification, 0, len(notifications))
	for _, n := range notifications {
		payloads = append(payloads, &hubnotificationv1.Notification{
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

func (h *MarkReadHandler) Validate(_ context.Context, req *MarkReadRequest) error {
	if req.ID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "Notification id is required")
	}
	return nil
}

func (h *MarkReadHandler) Serve(ctx context.Context, req *MarkReadRequest) (MarkReadResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return MarkReadResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return MarkReadResponse{}, onekiterrors.PermissionDenied()
	}

	if err := h.Service.MarkRead(ctx, businessID, device.User.UserID, req.ID); err != nil {
		return MarkReadResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to mark notification as read")
	}

	return MarkReadResponse{Message: "Notification marked as read"}, nil
}

// -- Mark Read by path param: POST /notifications/{id}/read --

func MarkReadByIDRoutes() []string { return []string{"/notifications/{id}/read"} }

type MarkReadByIDHandler struct{ Service *application.Service }

type MarkReadByIDRequest struct{}

func (h *MarkReadByIDHandler) Validate(_ context.Context, _ *MarkReadByIDRequest) error {
	return nil
}

func (h *MarkReadByIDHandler) Serve(ctx context.Context, _ *MarkReadByIDRequest) (MarkReadResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return MarkReadResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return MarkReadResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return MarkReadResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid notification id")
	}

	if err := h.Service.MarkRead(ctx, businessID, device.User.UserID, id); err != nil {
		return MarkReadResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to mark notification as read")
	}

	return MarkReadResponse{Message: "Notification marked as read"}, nil
}

// -- Mark All Read --

type MarkAllReadHandler struct{ Service *application.Service }

type MarkAllReadRequest struct{}

type MarkAllReadResponse struct {
	Message string `json:"message"`
}

func (h *MarkAllReadHandler) Serve(ctx context.Context, _ *MarkAllReadRequest) (MarkAllReadResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return MarkAllReadResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return MarkAllReadResponse{}, onekiterrors.PermissionDenied()
	}

	if err := h.Service.MarkAllRead(ctx, businessID, device.User.UserID); err != nil {
		return MarkAllReadResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to mark all notifications as read")
	}

	return MarkAllReadResponse{Message: "All notifications marked as read"}, nil
}
