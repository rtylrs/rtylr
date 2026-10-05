package application

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/voxire/onion/internal/testutil"
)

// A delivery left 'delivering' by a dead worker is invisible to the due query,
// which only reads 'pending' and 'failed'. Without recovery the subscriber
// never hears about that event.
//
// Unlike an abandoned automation run, this is requeued rather than
// dead-lettered: webhooks are at-least-once and every request carries the ids
// a subscriber dedupes on.
func TestProcessDueRequeuesAndRedeliversAnAbandonedDelivery(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	seedDelivery(t, db, server.URL, true, 0)

	// Claimed, then abandoned an hour ago.
	if _, err := db.ExecContext(ctx,
		`UPDATE webhook_deliveries SET status = 'delivering', attempts = 1,
		 updated_at = DATE_SUB(NOW(), INTERVAL 60 MINUTE) WHERE id = ?`, sendDelivery); err != nil {
		t.Fatalf("abandon delivery: %v", err)
	}

	sender := newTestSender(db)
	sender.client = server.Client()

	// One pass: recovery requeues it, and the same pass reads it as due.
	if err := sender.ProcessDue(ctx); err != nil {
		t.Fatalf("ProcessDue: %v", err)
	}

	status, responseStatus, lastErr := deliveryState(t, db)
	if status != "delivered" {
		t.Fatalf("status = %q (%s), want delivered after recovery", status, lastErr)
	}
	if !responseStatus.Valid || responseStatus.Int16 != http.StatusOK {
		t.Errorf("response_status = %v, want 200", responseStatus)
	}
	if got := calls.Load(); got != 1 {
		t.Errorf("subscriber called %d times, want 1", got)
	}
}

// A delivery claimed moments ago is in flight on another worker. Recovery must
// not touch it, or the subscriber gets the same event twice for no reason.
func TestProcessDueLeavesARecentlyClaimedDeliveryAlone(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	seedDelivery(t, db, server.URL, true, 0)

	if _, err := db.ExecContext(ctx,
		`UPDATE webhook_deliveries SET status = 'delivering', attempts = 1, updated_at = NOW() WHERE id = ?`,
		sendDelivery); err != nil {
		t.Fatalf("claim delivery: %v", err)
	}

	sender := newTestSender(db)
	sender.client = server.Client()
	if err := sender.ProcessDue(ctx); err != nil {
		t.Fatalf("ProcessDue: %v", err)
	}

	if status, _, _ := deliveryState(t, db); status != "delivering" {
		t.Errorf("status = %q, want delivering: a live worker's delivery was taken from it", status)
	}
	if got := calls.Load(); got != 0 {
		t.Errorf("subscriber called %d times, want 0", got)
	}
}

// An abandoned delivery whose attempts are already spent has nowhere to go but
// the dead letter, so recovery must not hand it back to the queue forever.
func TestProcessDueDeadLettersAnAbandonedDeliveryWithNoAttemptsLeft(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	seedDelivery(t, db, server.URL, true, 0)

	// max_attempts defaults to 5.
	if _, err := db.ExecContext(ctx,
		`UPDATE webhook_deliveries SET status = 'delivering', attempts = 5,
		 updated_at = DATE_SUB(NOW(), INTERVAL 60 MINUTE) WHERE id = ?`, sendDelivery); err != nil {
		t.Fatalf("abandon delivery: %v", err)
	}

	sender := newTestSender(db)
	sender.client = server.Client()
	if err := sender.ProcessDue(ctx); err != nil {
		t.Fatalf("ProcessDue: %v", err)
	}

	status, _, lastErr := deliveryState(t, db)
	if status != "dead" {
		t.Errorf("status = %q, want dead once attempts are spent", status)
	}
	if !strings.Contains(lastErr, "abandoned") {
		t.Errorf("last_error = %q, want it to explain the delivery was abandoned", lastErr)
	}
	if got := calls.Load(); got != 0 {
		t.Errorf("subscriber called %d times, want 0", got)
	}
}
