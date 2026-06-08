package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ─── List Notifications ───────────────────────────────────────────────────────

func ListNotificationsRoutes() []string { return []string{"/flow/notifications"} }

type ListNotificationsRequest struct{}

type ListNotificationsResponse struct {
	Notifications []notificationPayload `json:"notifications"`
}

type notificationPayload struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	TaskID        string `json:"taskId,omitempty"`
	TaskTitle     string `json:"taskTitle"`
	ActorInitials string `json:"actorInitials"`
	Project       string `json:"project"`
	ProjectColor  string `json:"projectColor"`
	Read          bool   `json:"read"`
	Timestamp     string `json:"timestamp"`
}

type ListNotificationsHandler struct{ FlowService *application.Service }

func (h *ListNotificationsHandler) Serve(ctx context.Context, _ *ListNotificationsRequest) (ListNotificationsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListNotificationsResponse{}, errorModels.PermissionDeniedServerError()
	}

	notifs, err := h.FlowService.ListNotifications(ctx, authUser.UserID)
	if err != nil {
		return ListNotificationsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	payloads := make([]notificationPayload, 0, len(notifs))
	for _, n := range notifs {
		p := notificationPayload{
			ID:            strconv.FormatInt(n.ID, 10),
			Type:          n.Type,
			Title:         n.Title,
			Description:   n.Description,
			TaskTitle:     n.TaskTitle,
			ActorInitials: n.ActorInitials,
			Project:       n.Project,
			ProjectColor:  n.ProjectColor,
			Read:          n.IsRead,
			Timestamp:     n.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if n.TaskID != nil {
			p.TaskID = strconv.FormatInt(*n.TaskID, 10)
		}
		payloads = append(payloads, p)
	}
	return ListNotificationsResponse{Notifications: payloads}, nil
}

// ─── Mark Notification Read ───────────────────────────────────────────────────

func MarkNotificationReadRoutes() []string { return []string{"/flow/notifications/{id}/read"} }

type MarkNotificationReadRequest struct {
	ID string `path:"id"`
}

type MarkNotificationReadResponse struct{}

func (r *MarkNotificationReadRequest) Validate() *errorModels.ServerError {
	if strings.TrimSpace(r.ID) == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Notification id is required"}
	}
	if n, err := strconv.ParseInt(r.ID, 10, 64); err != nil || n <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid notification id"}
	}
	return nil
}

type MarkNotificationReadHandler struct{ FlowService *application.Service }

func (h *MarkNotificationReadHandler) Serve(ctx context.Context, r *MarkNotificationReadRequest) (MarkNotificationReadResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return MarkNotificationReadResponse{}, errorModels.PermissionDeniedServerError()
	}

	id, err := strconv.ParseInt(r.ID, 10, 64)
	if err != nil || id <= 0 {
		return MarkNotificationReadResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid notification id"}
	}
	if err := h.FlowService.MarkNotificationRead(ctx, id, authUser.UserID); err != nil {
		return MarkNotificationReadResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}
	return MarkNotificationReadResponse{}, nil
}

// ─── Mark All Notifications Read ──────────────────────────────────────────────

func MarkAllNotificationsReadRoutes() []string { return []string{"/flow/notifications/read-all"} }

type MarkAllNotificationsReadRequest struct{}

type MarkAllNotificationsReadResponse struct{}

type MarkAllNotificationsReadHandler struct{ FlowService *application.Service }

func (h *MarkAllNotificationsReadHandler) Serve(ctx context.Context, _ *MarkAllNotificationsReadRequest) (MarkAllNotificationsReadResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return MarkAllNotificationsReadResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.FlowService.MarkAllNotificationsRead(ctx, authUser.UserID); err != nil {
		return MarkAllNotificationsReadResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}
	return MarkAllNotificationsReadResponse{}, nil
}

// ─── Get User Settings ────────────────────────────────────────────────────────

func GetUserSettingsRoutes() []string { return []string{"/flow/settings"} }

type GetUserSettingsRequest struct{}

type GetUserSettingsResponse struct {
	TaskAssigned       bool `json:"taskAssigned"`
	TaskComments       bool `json:"taskComments"`
	Mentions           bool `json:"mentions"`
	StatusChanges      bool `json:"statusChanges"`
	WeeklyDigest       bool `json:"weeklyDigest"`
	EmailNotifications bool `json:"emailNotifications"`
}

type GetUserSettingsHandler struct{ FlowService *application.Service }

func (h *GetUserSettingsHandler) Serve(ctx context.Context, _ *GetUserSettingsRequest) (GetUserSettingsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return GetUserSettingsResponse{}, errorModels.PermissionDeniedServerError()
	}

	s, err := h.FlowService.GetUserSettings(ctx, authUser.UserID)
	if err != nil {
		return GetUserSettingsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	return GetUserSettingsResponse{
		TaskAssigned:       s.TaskAssigned,
		TaskComments:       s.TaskComments,
		Mentions:           s.Mentions,
		StatusChanges:      s.StatusChanges,
		WeeklyDigest:       s.WeeklyDigest,
		EmailNotifications: s.EmailNotifications,
	}, nil
}

// ─── Update User Settings ─────────────────────────────────────────────────────

func UpdateUserSettingsRoutes() []string { return []string{"/flow/settings"} }

type UpdateUserSettingsRequest struct {
	TaskAssigned       bool `json:"task_assigned"`
	TaskComments       bool `json:"task_comments"`
	Mentions           bool `json:"mentions"`
	StatusChanges      bool `json:"status_changes"`
	WeeklyDigest       bool `json:"weekly_digest"`
	EmailNotifications bool `json:"email_notifications"`
}

type UpdateUserSettingsResponse struct{}

func (r *UpdateUserSettingsRequest) Validate() *errorModels.ServerError { return nil }

type UpdateUserSettingsHandler struct{ FlowService *application.Service }

func (h *UpdateUserSettingsHandler) Serve(ctx context.Context, r *UpdateUserSettingsRequest) (UpdateUserSettingsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateUserSettingsResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.FlowService.UpdateUserSettings(ctx, domain.UserSettings{
		UserID:             authUser.UserID,
		TaskAssigned:       r.TaskAssigned,
		TaskComments:       r.TaskComments,
		Mentions:           r.Mentions,
		StatusChanges:      r.StatusChanges,
		WeeklyDigest:       r.WeeklyDigest,
		EmailNotifications: r.EmailNotifications,
	}); err != nil {
		return UpdateUserSettingsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}
	return UpdateUserSettingsResponse{}, nil
}
