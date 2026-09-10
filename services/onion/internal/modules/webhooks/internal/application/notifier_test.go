package application

import (
	"context"
	"database/sql"
	"testing"

	"github.com/voxire/onion/internal/testutil"
)

const (
	notifyTenant  int64 = 885001
	notifyCreator int64 = 885010
	notifyHook    int64 = 885020
)

const notifyEmail = "hook-owner@test.local"

// seedFailing builds a registration whose endpoint has died, with the creator
// as an active member of its business. membershipStatus and deadAgeHours are
// the two things the tests vary.
func seedFailing(t *testing.T, db *sql.DB, membershipStatus string, deadAgeHours, deadCount int) {
	t.Helper()
	ctx := context.Background()

	clearFailing(t, db)
	t.Cleanup(func() { clearFailing(t, db) })

	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			t.Fatalf("seed (%s): %v", query, err)
		}
	}

	exec(`INSERT INTO users (id, email, password_hash, full_name, business_id)
	      VALUES (?, ?, 'fixture-auth-disabled', 'Hook Owner', ?)`, notifyCreator, notifyEmail, notifyTenant)
	// authorization_business_members has a foreign key to businesses, so the
	// tenant has to exist as a row rather than just as an id.
	exec(`INSERT INTO businesses (id, name, owner_id) VALUES (?, 'Notify Test Co', ?)`,
		notifyTenant, notifyCreator)
	if membershipStatus != "" {
		exec(`INSERT INTO authorization_business_members
		        (business_id, user_id, dash_role_key, status, authorization_version)
		      VALUES (?, ?, 'owner', ?, 1)`, notifyTenant, notifyCreator, membershipStatus)
	}
	exec(`INSERT INTO hub_webhooks (id, owner_id, business_id, url, events, secret, is_active)
	      VALUES (?, ?, ?, 'https://example.test/hook', JSON_ARRAY('order_created'), 's', TRUE)`,
		notifyHook, notifyCreator, notifyTenant)

	for i := 0; i < deadCount; i++ {
		eventID := 885030 + int64(i)
		exec(`INSERT INTO platform_events (id, business_id, event_type, subject_type, subject_id, payload, dedupe_key)
		      VALUES (?, ?, 'order_created', 'order', ?, JSON_OBJECT('order_id', ?), ?)`,
			eventID, notifyTenant, 500+i, 500+i, "notify-key-"+string(rune('a'+i)))
		exec(`INSERT INTO webhook_deliveries (id, business_id, webhook_id, event_id, status, attempts, updated_at)
		      VALUES (?, ?, ?, ?, 'dead', 5, DATE_SUB(NOW(), INTERVAL ? HOUR))`,
			885040+int64(i), notifyTenant, notifyHook, eventID, deadAgeHours)
	}
}

func clearFailing(t *testing.T, db *sql.DB) {
	t.Helper()
	ctx := context.Background()
	for _, stmt := range []string{
		"DELETE FROM webhook_deliveries WHERE business_id = ?",
		"DELETE FROM platform_events WHERE business_id = ?",
		"DELETE FROM hub_webhooks WHERE business_id = ?",
		"DELETE FROM authorization_business_members WHERE business_id = ?",
	} {
		_, _ = db.ExecContext(ctx, stmt, notifyTenant)
	}
	_, _ = db.ExecContext(ctx, "DELETE FROM businesses WHERE id = ?", notifyTenant)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", notifyCreator)
	_, _ = db.ExecContext(ctx, "DELETE FROM email_jobs WHERE to_email = ?", notifyEmail)
}

func queuedNotifications(t *testing.T, db *sql.DB) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM email_jobs WHERE to_email = ?", notifyEmail).Scan(&n); err != nil {
		t.Fatalf("count notifications: %v", err)
	}
	return n
}

// An endpoint that stops answering produces many dead deliveries. The owner
// needs to hear that their endpoint is failing, not receive one message per
// lost event, so three dead deliveries are one notification.
func TestNotifyFailingSendsOneMessagePerRegistration(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	seedFailing(t, db, "active", 1, 3)

	if err := NewNotifier(db).NotifyFailing(ctx); err != nil {
		t.Fatalf("NotifyFailing: %v", err)
	}

	if got := queuedNotifications(t, db); got != 1 {
		t.Fatalf("%d notifications for 3 dead deliveries on one registration, want 1", got)
	}

	var subject, body string
	if err := db.QueryRowContext(ctx,
		"SELECT subject, text_body FROM email_jobs WHERE to_email = ?", notifyEmail).Scan(&subject, &body); err != nil {
		t.Fatalf("read notification: %v", err)
	}
	if subject == "" {
		t.Error("notification has no subject")
	}
	// The message has to name the endpoint, or the owner cannot tell which of
	// their registrations broke.
	if !contains(body, "https://example.test/hook") {
		t.Errorf("body does not name the endpoint: %q", body)
	}
	if !contains(body, "3 events") {
		t.Errorf("body does not say how many events were lost: %q", body)
	}
}

// Run hourly over a query that keeps returning the same rows, so the dedupe
// key is what stops the owner being told every hour.
func TestNotifyFailingTellsTheOwnerOncePerDay(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	seedFailing(t, db, "active", 1, 1)

	notifier := NewNotifier(db)
	for pass := 1; pass <= 3; pass++ {
		if err := notifier.NotifyFailing(ctx); err != nil {
			t.Fatalf("pass %d: %v", pass, err)
		}
	}

	if got := queuedNotifications(t, db); got != 1 {
		t.Errorf("%d notifications after three passes, want 1", got)
	}
}

// A registration that broke long ago and was never fixed still has dead rows.
// Without the window it would generate a notification every day forever.
func TestNotifyFailingIgnoresFailuresOutsideTheWindow(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	seedFailing(t, db, "active", 72, 2)

	if err := NewNotifier(db).NotifyFailing(ctx); err != nil {
		t.Fatalf("NotifyFailing: %v", err)
	}

	if got := queuedNotifications(t, db); got != 0 {
		t.Errorf("%d notifications for failures three days old, want 0", got)
	}
}

// The creator has left the business. Telling them about a current tenant's
// failures would be sending somebody data they no longer have access to, and
// falling back to whoever else is around would tell someone about something
// they never set up.
func TestNotifyFailingTellsNobodyWhenTheCreatorHasLeft(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	seedFailing(t, db, "", 1, 1)

	if err := NewNotifier(db).NotifyFailing(ctx); err != nil {
		t.Fatalf("NotifyFailing: %v", err)
	}

	if got := queuedNotifications(t, db); got != 0 {
		t.Errorf("%d notifications sent to a departed creator, want 0", got)
	}
}

// Membership exists but is revoked, which is the same situation as having left.
func TestNotifyFailingSkipsARevokedMembership(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	seedFailing(t, db, "revoked", 1, 1)

	if err := NewNotifier(db).NotifyFailing(ctx); err != nil {
		t.Fatalf("NotifyFailing: %v", err)
	}

	if got := queuedNotifications(t, db); got != 0 {
		t.Errorf("%d notifications sent on a revoked membership, want 0", got)
	}
}

// One is the common case: an endpoint starts failing rather than failing in
// bulk, so the message should not read "1 events".
func TestDeliveryCountPhraseReadsNaturallyAtOne(t *testing.T) {
	if got := deliveryCountPhrase(1); got != "One event" {
		t.Errorf("deliveryCountPhrase(1) = %q, want \"One event\"", got)
	}
	if got := deliveryCountPhrase(4); got != "4 events" {
		t.Errorf("deliveryCountPhrase(4) = %q, want \"4 events\"", got)
	}
}

func contains(haystack, needle string) bool {
	return len(needle) > 0 && len(haystack) >= len(needle) && func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	}()
}
