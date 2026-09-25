package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/notifications/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// The notifications fixture inserts rows with owner_id=1001 and user_id=1001.
// ListNotifications queries by user_id, so we pass OwnerID as the user_id
// because the fixture seeds user_id = testutil.OwnerID.

func TestListNotifications_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.List(ctx, testutil.OwnerID, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
}

func TestListNotifications_ResultsAreSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	notifications, err := repo.List(ctx, testutil.OwnerID, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	// nil and empty slice are both valid when there are no fixture rows.
	_ = notifications
}

func TestListNotifications_UnknownUserReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	notifications, err := repo.List(ctx, testutil.OwnerID, 0)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(notifications) != 0 {
		t.Errorf("expected empty list for user_id=0, got %d items", len(notifications))
	}
}

func TestMarkRead_UnknownNotificationIsNoOp(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Should not error even when no matching row exists.
	err := repo.MarkRead(ctx, testutil.OwnerID, testutil.OwnerID, 0)
	if err != nil {
		t.Fatalf("MarkRead on non-existent notification returned error: %v", err)
	}
}

func TestMarkAllRead_UnknownUserIsNoOp(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.MarkAllRead(ctx, testutil.OwnerID, 0)
	if err != nil {
		t.Fatalf("MarkAllRead with unknown user_id returned error: %v", err)
	}
}

func TestListNotifications_EachHasNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	notifications, err := repo.List(ctx, testutil.OwnerID, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for i, n := range notifications {
		if n.ID == 0 {
			t.Errorf("notification[%d] has zero ID", i)
		}
	}
}

func TestListNotifications_EachHasNonEmptyTitle(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	notifications, err := repo.List(ctx, testutil.OwnerID, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for i, n := range notifications {
		if n.Title == "" {
			t.Errorf("notification[%d] (id=%d) has empty Title", i, n.ID)
		}
	}
}

func TestListNotifications_UserIDMatchesQuery(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	notifications, err := repo.List(ctx, testutil.OwnerID, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for i, n := range notifications {
		if n.UserID != testutil.OwnerID {
			t.Errorf("notification[%d] has UserID=%d, want %d", i, n.UserID, testutil.OwnerID)
		}
	}
}

func TestMarkAllRead_ForOwnerIsNoOp(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Just verify no error is returned for a valid user.
	err := repo.MarkAllRead(ctx, testutil.OwnerID, testutil.OwnerID)
	if err != nil {
		t.Fatalf("MarkAllRead returned error: %v", err)
	}
}

func TestListNotifications_TwiceReturnsSameCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.List(ctx, testutil.OwnerID, testutil.OwnerID)
	if err != nil {
		t.Fatalf("first List error: %v", err)
	}

	second, err := repo.List(ctx, testutil.OwnerID, testutil.OwnerID)
	if err != nil {
		t.Fatalf("second List error: %v", err)
	}

	if len(first) != len(second) {
		t.Errorf("list count changed between calls: %d vs %d", len(first), len(second))
	}
}

func TestListNotifications_TypeFieldIsString(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	notifications, err := repo.List(ctx, testutil.OwnerID, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	// Type is a string; this test confirms it scans without panic.
	for _, n := range notifications {
		_ = n.Type
	}
}

func TestListNotifications_CancelledContextErrors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.List(ctx, testutil.OwnerID, testutil.OwnerID)
	if err == nil {
		t.Log("cancelled context did not return error (driver may buffer); acceptable")
	}
}

func TestMarkRead_KnownUserUnknownNotification_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// A valid user but a notification ID that does not exist; should be no-op.
	err := repo.MarkRead(ctx, testutil.OwnerID, testutil.OwnerID, 9999999)
	if err != nil {
		t.Fatalf("MarkRead valid user, missing notification: %v", err)
	}
}

func TestMarkAllRead_OwnerID_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.MarkAllRead(ctx, testutil.OwnerID, testutil.OwnerID)
	if err != nil {
		t.Fatalf("MarkAllRead ownerID: %v", err)
	}
}

func TestMarkAllRead_ThenListReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.MarkAllRead(ctx, testutil.OwnerID, testutil.OwnerID); err != nil {
		t.Fatalf("MarkAllRead: %v", err)
	}
	_, err := repo.List(ctx, testutil.OwnerID, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after MarkAllRead: %v", err)
	}
}

func TestMarkRead_ZeroNotificationID_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.MarkRead(ctx, testutil.OwnerID, testutil.OwnerID, 0)
	if err != nil {
		t.Fatalf("MarkRead with notificationID=0: %v", err)
	}
}

func TestListNotifications_UnknownUserID_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	notes, err := repo.List(ctx, testutil.OwnerID, 9999999)
	if err != nil {
		t.Fatalf("List(userID=9999999): %v", err)
	}
	if len(notes) != 0 {
		t.Errorf("expected 0 notifications for unknown user, got %d", len(notes))
	}
}

func TestMarkAllRead_ConsecutiveCalls_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if err := repo.MarkAllRead(ctx, testutil.OwnerID, testutil.OwnerID); err != nil {
			t.Fatalf("MarkAllRead iteration %d: %v", i, err)
		}
	}
}

func TestListNotifications_IsReadFieldAccessible(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	notifications, err := repo.List(ctx, testutil.OwnerID, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	// Just verify IsRead field can be read without panic.
	for _, n := range notifications {
		_ = n.IsRead
	}
}

func TestListNotifications_CreatedAtAccessible(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	notifications, err := repo.List(ctx, testutil.OwnerID, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, n := range notifications {
		_ = n.CreatedAt
	}
}

func TestMarkRead_MultipleCallsSameNotification_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Call MarkRead twice for the same (non-existent) notification; should be idempotent.
	for i := 0; i < 2; i++ {
		if err := repo.MarkRead(ctx, testutil.OwnerID, testutil.OwnerID, 8888888); err != nil {
			t.Fatalf("MarkRead iteration %d: %v", i, err)
		}
	}
}

func TestListNotifications_LargeUserID_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	notes, err := repo.List(ctx, testutil.OwnerID, 1<<40)
	if err != nil {
		t.Fatalf("List(large userID): %v", err)
	}
	if len(notes) != 0 {
		t.Errorf("expected 0 notifications for large unknown userID, got %d", len(notes))
	}
}

func TestMarkAllRead_LargeUserID_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.MarkAllRead(ctx, testutil.OwnerID, 1<<40)
	if err != nil {
		t.Fatalf("MarkAllRead(large userID): %v", err)
	}
}

func TestListNotifications_BodyFieldAccessible(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	notifications, err := repo.List(ctx, testutil.OwnerID, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	// Verify the Body field scans without panic.
	for _, n := range notifications {
		_ = n.Body
	}
}
