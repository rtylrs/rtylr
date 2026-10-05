package notifications_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/voxire/onion/internal/modules/notifications"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateNotificationUsesTheRecipientBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const businessID = testutil.BusinessID
	title := "test-notification-" + uuid.NewString()
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM hub_notifications WHERE title = ?", title) })

	module := notifications.NewModule(ctx, db)
	if err := module.CreateNotification(ctx, notifications.NotificationInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: businessID,
		UserID:     testutil.OwnerID,
		Title:      title,
		Body:       "body",
		Type:       "info",
		Link:       "/team",
	}); err != nil {
		t.Fatalf("CreateNotification: %v", err)
	}

	var got int64
	if err := db.QueryRowContext(ctx, "SELECT business_id FROM hub_notifications WHERE title = ?", title).Scan(&got); err != nil {
		t.Fatalf("read created notification: %v", err)
	}
	if got != businessID {
		t.Fatalf("notification business_id = %d, want %d", got, businessID)
	}
}
