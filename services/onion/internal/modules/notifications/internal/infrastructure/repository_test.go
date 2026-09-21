package infrastructure_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/voxire/onion/internal/modules/notifications/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const (
	notificationsBusinessA int64 = 886001
	notificationsBusinessB int64 = 886002
	notificationsUser      int64 = 886003
)

func insertNotification(t *testing.T, db *sql.DB, businessID int64) int64 {
	t.Helper()
	res, err := db.ExecContext(context.Background(),
		`INSERT INTO hub_notifications (owner_id, business_id, user_id, title, is_read)
		 VALUES (?, ?, ?, 'tenant scoping test', FALSE)`,
		notificationsUser, businessID, notificationsUser)
	if err != nil {
		t.Fatalf("insert notification: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("insert notification id: %v", err)
	}
	return id
}

func notificationIsRead(t *testing.T, db *sql.DB, id int64) bool {
	t.Helper()
	var read bool
	if err := db.QueryRowContext(context.Background(),
		"SELECT is_read FROM hub_notifications WHERE id = ?", id).Scan(&read); err != nil {
		t.Fatalf("read notification %d: %v", id, err)
	}
	return read
}

// A user who belongs to two businesses acts in one of them. Marking a
// notification read must not mutate the other business's row, which is what the
// dropped tenant filter used to allow.
func TestMarkReadOnlyTouchesTheActingBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := infrastructure.NewRepository(db)

	clear := func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM hub_notifications WHERE user_id = ?", notificationsUser)
	}
	clear()
	t.Cleanup(clear)

	ownID := insertNotification(t, db, notificationsBusinessA)
	foreignID := insertNotification(t, db, notificationsBusinessB)

	if err := repo.MarkRead(ctx, notificationsBusinessA, notificationsUser, foreignID); err != nil {
		t.Fatalf("MarkRead(foreign): %v", err)
	}
	if notificationIsRead(t, db, foreignID) {
		t.Fatal("notification from another business was marked read")
	}

	if err := repo.MarkRead(ctx, notificationsBusinessA, notificationsUser, ownID); err != nil {
		t.Fatalf("MarkRead(own): %v", err)
	}
	if !notificationIsRead(t, db, ownID) {
		t.Fatal("the acting business's own notification was not marked read")
	}
}

func TestMarkAllReadOnlyTouchesTheActingBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := infrastructure.NewRepository(db)

	clear := func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM hub_notifications WHERE user_id = ?", notificationsUser)
	}
	clear()
	t.Cleanup(clear)

	ownID := insertNotification(t, db, notificationsBusinessA)
	foreignID := insertNotification(t, db, notificationsBusinessB)

	if err := repo.MarkAllRead(ctx, notificationsBusinessA, notificationsUser); err != nil {
		t.Fatalf("MarkAllRead: %v", err)
	}
	if !notificationIsRead(t, db, ownID) {
		t.Fatal("the acting business's own notification was not marked read")
	}
	if notificationIsRead(t, db, foreignID) {
		t.Fatal("notification from another business was marked read")
	}
}
