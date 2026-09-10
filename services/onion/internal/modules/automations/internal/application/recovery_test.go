package application

import (
	"context"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/testutil"
)

// A run left 'running' by a dead worker is invisible to the due query, which
// only reads 'pending' and 'failed'. Without recovery it sits there forever.
func TestProcessDueDeadLettersAnAbandonedRun(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	seedRun(t, db, `[{"type":"send_email"}]`, true, `{"customer_id":882010}`)

	// Claimed, then abandoned an hour ago.
	if _, err := db.ExecContext(ctx,
		`UPDATE automation_runs SET status = 'running', attempts = 1,
		 updated_at = DATE_SUB(NOW(), INTERVAL 60 MINUTE) WHERE id = ?`, execRun); err != nil {
		t.Fatalf("abandon run: %v", err)
	}

	if err := NewExecutor(db).ProcessDue(ctx); err != nil {
		t.Fatalf("ProcessDue: %v", err)
	}

	status, lastErr := runStatus(t, db)
	if status != "dead" {
		t.Errorf("status = %q, want dead", status)
	}
	if !strings.Contains(lastErr, "abandoned") {
		t.Errorf("last_error = %q, want it to explain the run was abandoned", lastErr)
	}

	// Not re-executed. An automation may already have sent its email, and the
	// recipient cannot dedupe one, which is why this is dead rather than
	// requeued.
	if got := queuedEmails(t, db); got != 0 {
		t.Errorf("%d emails queued while recovering an abandoned run, want 0", got)
	}
}

// A run claimed moments ago is being executed right now by another worker.
// Recovery must not touch it, or two workers run the same automation.
func TestProcessDueLeavesARecentlyClaimedRunAlone(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	seedRun(t, db, `[{"type":"send_email"}]`, true, `{"customer_id":882010}`)

	if _, err := db.ExecContext(ctx,
		`UPDATE automation_runs SET status = 'running', attempts = 1, updated_at = NOW() WHERE id = ?`,
		execRun); err != nil {
		t.Fatalf("claim run: %v", err)
	}

	if err := NewExecutor(db).ProcessDue(ctx); err != nil {
		t.Fatalf("ProcessDue: %v", err)
	}

	if status, _ := runStatus(t, db); status != "running" {
		t.Errorf("status = %q, want running: a live worker's run was taken from it", status)
	}
	if got := queuedEmails(t, db); got != 0 {
		t.Errorf("%d emails queued, want 0", got)
	}
}
