package infrastructure_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestWorkspaces_ListNotifications_PaginationAndTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Sentinel user ID unlikely to collide with notifications created by
	// other tests or fixtures.
	const notifiedUser = int64(929292929)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM flow_notifications WHERE user_id = ?", notifiedUser)
	})
	for i := 0; i < 5; i++ {
		if err := repo.CreateNotification(ctx, domain.CreateNotificationInput{
			UserID: notifiedUser,
			Type:   "assigned",
			Title:  fmt.Sprintf("NotifPage Notification %d", i),
		}); err != nil {
			t.Fatalf("CreateNotification %d: %v", i, err)
		}
	}

	page1, total1, err := repo.ListNotifications(ctx, notifiedUser, 2, 0)
	if err != nil {
		t.Fatalf("ListNotifications page1: %v", err)
	}
	if total1 != 5 {
		t.Errorf("expected total=5, got %d", total1)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 notifications on page1, got %d", len(page1))
	}

	page2, total2, err := repo.ListNotifications(ctx, notifiedUser, 2, 2)
	if err != nil {
		t.Fatalf("ListNotifications page2: %v", err)
	}
	if total2 != 5 {
		t.Errorf("expected total=5 on page2, got %d", total2)
	}
	if len(page2) != 2 {
		t.Fatalf("expected 2 notifications on page2, got %d", len(page2))
	}
	for _, p1 := range page1 {
		for _, p2 := range page2 {
			if p1.ID == p2.ID {
				t.Errorf("notification %d appeared on both page1 and page2", p1.ID)
			}
		}
	}

	page3, total3, err := repo.ListNotifications(ctx, notifiedUser, 2, 4)
	if err != nil {
		t.Fatalf("ListNotifications page3: %v", err)
	}
	if total3 != 5 {
		t.Errorf("expected total=5 on page3, got %d", total3)
	}
	if len(page3) != 1 {
		t.Errorf("expected 1 remainder notification on page3, got %d", len(page3))
	}
}
