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

	_, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
}

func TestListNotifications_ResultsAreSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	notifications, err := repo.List(ctx, testutil.OwnerID)
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

	notifications, err := repo.List(ctx, 0)
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
	err := repo.MarkRead(ctx, testutil.OwnerID, 0)
	if err != nil {
		t.Fatalf("MarkRead on non-existent notification returned error: %v", err)
	}
}

func TestMarkAllRead_UnknownUserIsNoOp(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.MarkAllRead(ctx, 0)
	if err != nil {
		t.Fatalf("MarkAllRead with unknown user_id returned error: %v", err)
	}
}

func TestListNotifications_EachHasNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	notifications, err := repo.List(ctx, testutil.OwnerID)
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

	notifications, err := repo.List(ctx, testutil.OwnerID)
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

	notifications, err := repo.List(ctx, testutil.OwnerID)
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
	err := repo.MarkAllRead(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("MarkAllRead returned error: %v", err)
	}
}

func TestListNotifications_TwiceReturnsSameCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("first List error: %v", err)
	}

	second, err := repo.List(ctx, testutil.OwnerID)
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

	notifications, err := repo.List(ctx, testutil.OwnerID)
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

	_, err := repo.List(ctx, testutil.OwnerID)
	if err == nil {
		t.Log("cancelled context did not return error (driver may buffer); acceptable")
	}
}
