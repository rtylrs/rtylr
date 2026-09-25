package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/team/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestBulkCreateHubNotificationSetsBusinessID(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := infrastructure.NewRepository(db)

	const (
		ownerID    int64 = 990731
		businessID int64 = 990732
		userID     int64 = 990733
		title            = "hub-notification-business-probe"
	)

	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM hub_notifications WHERE user_id = ? AND title = ?", userID, title)
	})

	if err := repo.BulkCreateHubNotification(ctx, ownerID, businessID, []int64{userID}, title, "body", "info", "https://example.com/team"); err != nil {
		t.Fatalf("BulkCreateHubNotification: %v", err)
	}

	var got int64
	if err := db.QueryRowContext(ctx, "SELECT business_id FROM hub_notifications WHERE user_id = ? AND title = ?", userID, title).Scan(&got); err != nil {
		t.Fatalf("read hub_notifications.business_id: %v", err)
	}
	if got != businessID {
		t.Fatalf("hub_notifications.business_id = %d, want %d", got, businessID)
	}
}
