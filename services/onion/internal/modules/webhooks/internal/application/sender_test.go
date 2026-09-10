package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/voxire/onion/internal/testutil"
)

const (
	sendTenant   int64  = 883001
	sendOwner    int64  = 883000
	sendWebhook  int64  = 883010
	sendEvent    uint64 = 883020
	sendDelivery uint64 = 883030
	sendSecret          = "whsec_test"
)

func seedDelivery(t *testing.T, db *sql.DB, url string, active bool, attempts int) {
	t.Helper()
	ctx := context.Background()

	clearDelivery(t, db)
	t.Cleanup(func() { clearDelivery(t, db) })

	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			t.Fatalf("seed (%s): %v", query, err)
		}
	}
	exec(`INSERT INTO hub_webhooks (id, owner_id, business_id, url, events, secret, is_active)
	      VALUES (?, ?, ?, ?, JSON_ARRAY('order_created'), ?, ?)`,
		sendWebhook, sendOwner, sendTenant, url, sendSecret, active)
	exec(`INSERT INTO platform_events (id, business_id, event_type, subject_type, subject_id, payload, dedupe_key)
	      VALUES (?, ?, 'order_created', 'order', 501, JSON_OBJECT('order_id', 501), 'order_created:order:501')`,
		sendEvent, sendTenant)
	exec(`INSERT INTO webhook_deliveries (id, business_id, webhook_id, event_id, attempts)
	      VALUES (?, ?, ?, ?, ?)`, sendDelivery, sendTenant, sendWebhook, sendEvent, attempts)
}

func clearDelivery(t *testing.T, db *sql.DB) {
	t.Helper()
	ctx := context.Background()
	for _, stmt := range []string{
		"DELETE FROM webhook_deliveries WHERE business_id = ?",
		"DELETE FROM platform_events WHERE business_id = ?",
		"DELETE FROM hub_webhooks WHERE business_id = ?",
	} {
		_, _ = db.ExecContext(ctx, stmt, sendTenant)
	}
}

func deliveryState(t *testing.T, db *sql.DB) (status string, responseStatus sql.NullInt16, lastError string) {
	t.Helper()
	var lastErr sql.NullString
	if err := db.QueryRowContext(context.Background(),
		"SELECT status, response_status, last_error FROM webhook_deliveries WHERE id = ?",
		sendDelivery).Scan(&status, &responseStatus, &lastErr); err != nil {
		t.Fatalf("read delivery: %v", err)
	}
	return status, responseStatus, lastErr.String
}

// The path #787 says does not exist: a registered URL receives an event.
func TestProcessDueDeliversASignedEvent(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	var (
		mu       sync.Mutex
		gotBody  []byte
		gotSig   string
		gotType  string
		gotEvent string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		gotBody, gotSig = body, r.Header.Get(SignatureHeader)
		gotType, gotEvent = r.Header.Get(EventTypeHeader), r.Header.Get(EventIDHeader)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	seedDelivery(t, db, server.URL, true, 0)

	sender := NewSender(db)
	// The real client refuses loopback by design, which is what
	// TestTheDeliveryClientRefusesALoopbackDestination proves. Swapped here so
	// the queue behaviour can be tested against a local server.
	sender.client = server.Client()

	if err := sender.ProcessDue(ctx); err != nil {
		t.Fatalf("ProcessDue: %v", err)
	}

	status, responseStatus, lastErr := deliveryState(t, db)
	if status != "delivered" {
		t.Fatalf("status = %q (%s), want delivered", status, lastErr)
	}
	if !responseStatus.Valid || responseStatus.Int16 != http.StatusOK {
		t.Errorf("response_status = %v, want 200", responseStatus)
	}

	mu.Lock()
	defer mu.Unlock()

	if gotType != "order_created" {
		t.Errorf("%s = %q, want order_created", EventTypeHeader, gotType)
	}
	if gotEvent != fmt.Sprint(sendEvent) {
		t.Errorf("%s = %q, want %d", EventIDHeader, gotEvent, sendEvent)
	}

	// The signature a subscriber would verify.
	var timestamp int64
	var provided string
	if _, err := fmt.Sscanf(gotSig, "t=%d,v1=%s", &timestamp, &provided); err != nil {
		t.Fatalf("signature header %q is not parseable: %v", gotSig, err)
	}
	mac := hmac.New(sha256.New, []byte(sendSecret))
	mac.Write([]byte(fmt.Sprintf("%d.", timestamp)))
	mac.Write(gotBody)
	if want := hex.EncodeToString(mac.Sum(nil)); !hmac.Equal([]byte(provided), []byte(want)) {
		t.Errorf("signature does not verify against the delivered body")
	}

	// The envelope carries ids as strings, matching the rest of the product.
	var received envelope
	if err := json.Unmarshal(gotBody, &received); err != nil {
		t.Fatalf("body is not the documented envelope: %v", err)
	}
	if received.Type != "order_created" || received.Subject.Type != "order" || received.Subject.ID != "501" {
		t.Errorf("envelope = %+v, want the order_created subject", received)
	}
	if !strings.Contains(string(received.Data), "order_id") {
		t.Errorf("envelope data = %s, want the producer payload", received.Data)
	}

	// hub_webhooks.last_triggered_at, which nothing has ever written.
	var lastTriggered sql.NullTime
	if err := db.QueryRowContext(ctx,
		"SELECT last_triggered_at FROM hub_webhooks WHERE id = ?", sendWebhook).Scan(&lastTriggered); err != nil {
		t.Fatalf("read registration: %v", err)
	}
	if !lastTriggered.Valid {
		t.Error("last_triggered_at was not written")
	}
}

// A subscriber that errors is retried, not dead-lettered, until its attempts
// run out. A 404 during someone's deploy is as common as a wrong URL.
func TestProcessDueRetriesANonSuccessResponse(t *testing.T) {
	db := testutil.OpenDB(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	seedDelivery(t, db, server.URL, true, 0)

	sender := NewSender(db)
	sender.client = server.Client()
	if err := sender.ProcessDue(context.Background()); err != nil {
		t.Fatalf("ProcessDue: %v", err)
	}

	status, responseStatus, lastErr := deliveryState(t, db)
	if status != "failed" {
		t.Errorf("status = %q, want failed so it is retried", status)
	}
	if !responseStatus.Valid || responseStatus.Int16 != http.StatusServiceUnavailable {
		t.Errorf("response_status = %v, want 503 recorded", responseStatus)
	}
	if !strings.Contains(lastErr, "503") {
		t.Errorf("last_error = %q, want the status in it", lastErr)
	}
}

// Attempts spent, so the same failure now dead-letters rather than retrying
// forever. max_attempts defaults to 5.
func TestProcessDueDeadLettersWhenAttemptsAreSpent(t *testing.T) {
	db := testutil.OpenDB(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	seedDelivery(t, db, server.URL, true, 4)

	sender := NewSender(db)
	sender.client = server.Client()
	if err := sender.ProcessDue(context.Background()); err != nil {
		t.Fatalf("ProcessDue: %v", err)
	}

	if status, _, _ := deliveryState(t, db); status != "dead" {
		t.Errorf("status = %q, want dead once attempts are spent", status)
	}
}

// Switched off between fan-out and delivery. Terminal: delivering anyway would
// ignore the tenant's decision.
func TestProcessDueDeadLettersAnInactiveRegistration(t *testing.T) {
	db := testutil.OpenDB(t)

	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	seedDelivery(t, db, server.URL, false, 0)

	sender := NewSender(db)
	sender.client = server.Client()
	if err := sender.ProcessDue(context.Background()); err != nil {
		t.Fatalf("ProcessDue: %v", err)
	}

	if status, _, _ := deliveryState(t, db); status != "dead" {
		t.Errorf("status = %q, want dead", status)
	}
	if called {
		t.Error("an inactive registration was still called")
	}
}

// End to end through the queue with the real client: a registration pointing
// at an internal address is dead-lettered and nothing is sent. This is the
// tenant-supplied-URL SSRF, refused where it matters rather than in a unit.
func TestProcessDueDeadLettersAnInternalDestination(t *testing.T) {
	db := testutil.OpenDB(t)

	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// httptest binds to loopback, so its own URL is the internal destination.
	seedDelivery(t, db, server.URL, true, 0)

	// The real client, deliberately not substituted.
	if err := NewSender(db).ProcessDue(context.Background()); err != nil {
		t.Fatalf("ProcessDue: %v", err)
	}

	status, _, lastErr := deliveryState(t, db)
	if status != "dead" {
		t.Errorf("status = %q, want dead: an internal destination must not be retried", status)
	}
	if !strings.Contains(lastErr, "not reachable from this platform") {
		t.Errorf("last_error = %q, want it to say the destination was refused", lastErr)
	}
	if called {
		t.Error("the platform connected to an internal address supplied by a tenant")
	}
}
