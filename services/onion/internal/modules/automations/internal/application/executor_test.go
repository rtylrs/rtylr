package application

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/testutil"
)

const (
	execTenant     int64  = 882001
	execOwner      int64  = 882000
	execCustomer   int64  = 882010
	execAutomation int64  = 882020
	execEvent      uint64 = 882030
	execRun        uint64 = 882040
)

// seedRun builds one complete order_created to send_email path: a customer
// with an address, an automation, the event, and the queued run the dispatcher
// would have written.
func seedRun(t *testing.T, db *sql.DB, actions string, active bool, payload string) {
	t.Helper()
	ctx := context.Background()

	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			t.Fatalf("seed (%s): %v", query, err)
		}
	}

	cleanup(t, db)
	t.Cleanup(func() { cleanup(t, db) })

	exec(`INSERT INTO crm_customers (id, owner_id, business_id, first_name, last_name, email)
	      VALUES (?, ?, ?, 'Ada', 'Lovelace', 'ada@example.test')`, execCustomer, execOwner, execTenant)
	exec(`INSERT INTO crm_automations (id, owner_id, business_id, name, trigger_type, actions, is_active, run_count)
	      VALUES (?, ?, ?, 'order thanks', 'order_created', ?, ?, 0)`,
		execAutomation, execOwner, execTenant, actions, active)
	exec(`INSERT INTO platform_events (id, business_id, event_type, subject_type, subject_id, payload, dedupe_key)
	      VALUES (?, ?, 'order_created', 'order', 501, ?, 'order_created:order:501')`,
		execEvent, execTenant, payload)
	exec(`INSERT INTO automation_runs (id, business_id, automation_id, event_id)
	      VALUES (?, ?, ?, ?)`, execRun, execTenant, execAutomation, execEvent)
}

func cleanup(t *testing.T, db *sql.DB) {
	t.Helper()
	ctx := context.Background()
	for _, stmt := range []string{
		"DELETE FROM automation_runs WHERE business_id = ?",
		"DELETE FROM platform_events WHERE business_id = ?",
		"DELETE FROM crm_automations WHERE business_id = ?",
		"DELETE FROM crm_customers WHERE business_id = ?",
	} {
		_, _ = db.ExecContext(ctx, stmt, execTenant)
	}
	_, _ = db.ExecContext(ctx, "DELETE FROM email_jobs WHERE to_email = 'ada@example.test'")
}

func runStatus(t *testing.T, db *sql.DB) (string, string) {
	t.Helper()
	var status string
	var lastError sql.NullString
	if err := db.QueryRowContext(context.Background(),
		"SELECT status, last_error FROM automation_runs WHERE id = ?", execRun).Scan(&status, &lastError); err != nil {
		t.Fatalf("read run: %v", err)
	}
	return status, lastError.String
}

func queuedEmails(t *testing.T, db *sql.DB) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM email_jobs WHERE to_email = 'ada@example.test'").Scan(&n); err != nil {
		t.Fatalf("count emails: %v", err)
	}
	return n
}

// The path #834 says does not exist: a user builds order_created to
// send_email, turns it on, and something happens.
func TestProcessDueSendsTheEmailAndRecordsTheRun(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	seedRun(t, db, `[{"type":"send_email","subject":"Thanks!","body":"We got your order."}]`, true,
		`{"order_id":501,"customer_id":882010}`)

	if err := NewExecutor(db).ProcessDue(ctx); err != nil {
		t.Fatalf("ProcessDue: %v", err)
	}

	if status, lastErr := runStatus(t, db); status != "succeeded" {
		t.Fatalf("run status = %q (%s), want succeeded", status, lastErr)
	}

	var subject, to string
	if err := db.QueryRowContext(ctx,
		"SELECT subject, to_email FROM email_jobs WHERE to_email = 'ada@example.test'").Scan(&subject, &to); err != nil {
		t.Fatalf("read queued email: %v", err)
	}
	if subject != "Thanks!" {
		t.Errorf("subject = %q, want the automation's own subject", subject)
	}

	// The counters #834 records as dead.
	var runCount int
	var lastRunAt sql.NullTime
	if err := db.QueryRowContext(ctx,
		"SELECT run_count, last_run_at FROM crm_automations WHERE id = ?", execAutomation).Scan(&runCount, &lastRunAt); err != nil {
		t.Fatalf("read automation: %v", err)
	}
	if runCount != 1 {
		t.Errorf("run_count = %d, want 1", runCount)
	}
	if !lastRunAt.Valid {
		t.Error("last_run_at was not written")
	}
}

// A second pass must not send again. The run is no longer pending, so it is
// not re-claimed.
func TestProcessDueDoesNotResendACompletedRun(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	seedRun(t, db, `[{"type":"send_email"}]`, true, `{"customer_id":882010}`)

	executor := NewExecutor(db)
	if err := executor.ProcessDue(ctx); err != nil {
		t.Fatalf("first ProcessDue: %v", err)
	}
	if err := executor.ProcessDue(ctx); err != nil {
		t.Fatalf("second ProcessDue: %v", err)
	}

	if got := queuedEmails(t, db); got != 1 {
		t.Errorf("%d emails queued after two passes, want 1", got)
	}
}

// An action this build does not implement must fail loudly and execute
// nothing, rather than doing the half it understands. Silently doing nothing
// is the complaint in #834; silently doing half would be worse.
func TestProcessDueRefusesAnUnimplementedActionWithoutRunningTheRest(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	seedRun(t, db, `[{"type":"send_email"},{"type":"award_points"}]`, true, `{"customer_id":882010}`)

	if err := NewExecutor(db).ProcessDue(ctx); err != nil {
		t.Fatalf("ProcessDue: %v", err)
	}

	status, lastErr := runStatus(t, db)
	if status != "dead" {
		t.Errorf("run status = %q, want dead: an unimplementable run must not be retried", status)
	}
	if !strings.Contains(lastErr, "award_points") {
		t.Errorf("last_error = %q, want it to name the unimplemented action", lastErr)
	}
	if got := queuedEmails(t, db); got != 0 {
		t.Errorf("%d emails queued, want 0: the supported half ran even though the automation could not complete", got)
	}
}

// Switched off between the event and the run. Terminal, not retryable.
func TestProcessDueDeadLettersAnInactiveAutomation(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	seedRun(t, db, `[{"type":"send_email"}]`, false, `{"customer_id":882010}`)

	if err := NewExecutor(db).ProcessDue(ctx); err != nil {
		t.Fatalf("ProcessDue: %v", err)
	}

	if status, _ := runStatus(t, db); status != "dead" {
		t.Errorf("run status = %q, want dead", status)
	}
	if got := queuedEmails(t, db); got != 0 {
		t.Errorf("%d emails queued for an inactive automation, want 0", got)
	}
}

// send_email with nobody to send to cannot be retried into working.
func TestProcessDueDeadLettersWhenTheEventNamesNoCustomer(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	seedRun(t, db, `[{"type":"send_email"}]`, true, `{"order_id":501}`)

	if err := NewExecutor(db).ProcessDue(ctx); err != nil {
		t.Fatalf("ProcessDue: %v", err)
	}

	status, lastErr := runStatus(t, db)
	if status != "dead" {
		t.Errorf("run status = %q, want dead", status)
	}
	if !strings.Contains(lastErr, "customer") {
		t.Errorf("last_error = %q, want it to say a customer was missing", lastErr)
	}
}

func TestRetryDelayBacksOffAndCaps(t *testing.T) {
	cases := map[int]int64{1: 60, 2: 120, 3: 240, 20: maxRetrySeconds}
	for attempts, want := range cases {
		if got := retryDelaySeconds(attempts); got != want {
			t.Errorf("retryDelaySeconds(%d) = %d, want %d", attempts, got, want)
		}
	}
}

// An int64 crosses the wire as a string in this product, so a producer may
// well have written one either way.
func TestPayloadInt64ReadsBothNumberAndString(t *testing.T) {
	cases := map[string]int64{
		`{"customer_id":42}`:   42,
		`{"customer_id":"42"}`: 42,
		`{"customer_id":null}`: 0,
		`{}`:                   0,
		`not json`:             0,
	}
	for payload, want := range cases {
		if got := payloadInt64([]byte(payload), "customer_id"); got != want {
			t.Errorf("payloadInt64(%s) = %d, want %d", payload, got, want)
		}
	}
}
