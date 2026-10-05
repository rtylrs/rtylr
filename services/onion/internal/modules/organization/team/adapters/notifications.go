package teamadapters

import (
	"context"

	"github.com/voxire/onion/internal/modules/notifications"
	"github.com/voxire/onion/internal/modules/organization/team"
)

type teamNotificationAdapter struct{ module *notifications.Module }

func (a teamNotificationAdapter) CreateNotification(ctx context.Context, ownerID, businessID, userID int64, title, body, notificationType, link string) error {
	return a.module.CreateNotification(ctx, notifications.NotificationInput{
		OwnerID:    ownerID,
		BusinessID: businessID,
		UserID:     userID,
		Title:      title,
		Body:       body,
		Type:       notificationType,
		Link:       link,
	})
}

var _ team.NotificationPublisher = teamNotificationAdapter{}

func NewNotificationPublisher(module *notifications.Module) team.NotificationPublisher {
	return teamNotificationAdapter{module: module}
}
