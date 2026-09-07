package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/notifications/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cursor"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubnotificationv1 "github.com/voxire/voxie/gen/go/hub/notification/v1"
)

type NotificationAPI struct {
	Service *application.Service
}

func NewNotificationAPI(service *application.Service) *NotificationAPI {
	return &NotificationAPI{Service: service}
}

func (a *NotificationAPI) ListNotifications(ctx context.Context, req *hubnotificationv1.ListNotificationsRequest) (*hubnotificationv1.ListNotificationsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	page, err := a.Service.ListPage(ctx, businessID, device.User.UserID, req.GetLimit(), req.GetCursor())
	if err != nil {
		if errors.Is(err, cursor.ErrInvalid) {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid cursor")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list notifications")
	}

	payloads := make([]*hubnotificationv1.Notification, 0, len(page.Notifications))
	for _, n := range page.Notifications {
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

	return &hubnotificationv1.ListNotificationsResponse{Notifications: payloads, NextCursor: page.NextCursor}, nil
}

func (a *NotificationAPI) MarkAllRead(ctx context.Context, _ *hubnotificationv1.MarkAllReadRequest) (*hubnotificationv1.MarkAllReadResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.MarkAllRead(ctx, businessID, device.User.UserID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to mark all notifications as read")
	}

	return &hubnotificationv1.MarkAllReadResponse{Message: "All notifications marked as read"}, nil
}

func (a *NotificationAPI) MarkNotificationRead(ctx context.Context, req *hubnotificationv1.MarkNotificationReadRequest) (*hubnotificationv1.MarkNotificationReadResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	id := req.GetId()
	if id <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid notification id")
	}

	if err := a.Service.MarkRead(ctx, businessID, device.User.UserID, id); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to mark notification as read")
	}

	return &hubnotificationv1.MarkNotificationReadResponse{Message: "Notification marked as read"}, nil
}

func (a *NotificationAPI) MarkNotificationsRead(ctx context.Context, req *hubnotificationv1.MarkNotificationsReadRequest) (*hubnotificationv1.MarkNotificationsReadResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	ids := req.GetIds()
	if len(ids) == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "At least one notification id is required")
	}

	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if err := a.Service.MarkRead(ctx, businessID, device.User.UserID, id); err != nil {
			return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to mark notifications as read")
		}
	}

	return &hubnotificationv1.MarkNotificationsReadResponse{Message: "Notifications marked as read"}, nil
}
