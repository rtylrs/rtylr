package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	flownotificationv1 "github.com/voxire/voxie/gen/go/flow/notification/v1"
)

type NotificationAPI struct {
	Service *application.Service
}

func NewNotificationAPI(service *application.Service) *NotificationAPI {
	return &NotificationAPI{Service: service}
}

func toNotification(n domain.Notification) *flownotificationv1.Notification {
	out := &flownotificationv1.Notification{
		Id:            n.ID,
		Type:          n.Type,
		Title:         n.Title,
		Description:   n.Description,
		TaskTitle:     n.TaskTitle,
		ActorId:       n.ActorID,
		ActorName:     n.ActorName,
		ActorInitials: n.ActorInitials,
		Project:       n.Project,
		ProjectColor:  n.ProjectColor,
		Read:          n.IsRead,
		Timestamp:     n.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if n.TaskID != nil {
		out.TaskId = strconv.FormatInt(*n.TaskID, 10)
	}
	return out
}

func (a *NotificationAPI) ListNotifications(ctx context.Context, req *flownotificationv1.ListNotificationsRequest) (*flownotificationv1.ListNotificationsResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	notifs, total, err := a.Service.ListNotifications(ctx, authUser.UserID, int64(limit), offset)
	if err != nil {
		return nil, workspaceInternalError(ctx, "ListNotifications failed", err)
	}

	out := make([]*flownotificationv1.Notification, 0, len(notifs))
	for _, n := range notifs {
		out = append(out, toNotification(n))
	}
	return &flownotificationv1.ListNotificationsResponse{
		Notifications: out,
		Pagination:    buildPagination(page, limit, total, len(notifs)),
	}, nil
}

func (a *NotificationAPI) MarkNotificationRead(ctx context.Context, req *flownotificationv1.MarkNotificationReadRequest) (*flownotificationv1.MarkNotificationReadResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	if req.GetId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid notification id")
	}

	if err := a.Service.MarkNotificationRead(ctx, req.GetId(), authUser.UserID); err != nil {
		return nil, workspaceInternalError(ctx, "MarkNotificationRead failed", err)
	}
	return &flownotificationv1.MarkNotificationReadResponse{}, nil
}

func (a *NotificationAPI) MarkAllNotificationsRead(ctx context.Context, _ *flownotificationv1.MarkAllNotificationsReadRequest) (*flownotificationv1.MarkAllNotificationsReadResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.MarkAllNotificationsRead(ctx, authUser.UserID); err != nil {
		return nil, workspaceInternalError(ctx, "MarkAllNotificationsRead failed", err)
	}
	return &flownotificationv1.MarkAllNotificationsReadResponse{}, nil
}
