package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/testutil"
)

// A run left 'running' by a dead worker is invisible to the due query, which
// only reads 'pending' and 'failed'. Without recovery it sits there forever.
//
// It is requeued and re-executed rather than dead-lettered, which is safe
// because every effect an action performs carries a key naming it.
func TestProcessDueRequeuesAndRerunsAnAbandonedRun(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	seedRun(t, db, `[{"type":"send_email"}]`, true, `{"customer_id":882010}`)

	// Claimed, then abandoned an hour ago, having done nothing.
	if _, err := db.ExecContext(ctx,
		`UPDATE automation_runs SET status = 'running', attempts = 1,
		 updated_at = DATE_SUB(NOW(), INTERVAL 60 MINUTE) WHERE id = ?`, execRun); err != nil {
		t.Fatalf("abandon run: %v", err)
	}

	// One pass: recovery requeues it and the same pass reads it as due.
	if err := NewExecutor(db).ProcessDue(ctx); err != nil {
		t.Fatalf("ProcessDue: %v", err)
	}

	status, lastErr := runStatus(t, db)
	if status != "succeeded" {
		t.Fatalf("status = %q (%s), want succeeded: the abandoned run was not retried", status, lastErr)
	}
	if got := queuedEmails(t, db); got != 1 {
		t.Errorf("%d emails queued, want 1: the retried run did not do its work", got)
	}
}

// The case the dedupe key exists for, and the reason abandoned runs used to be
// dead-lettered instead of retried.
//
// A worker that died *after* queueing the email leaves exactly the same row
// behind as one that died before. Nothing in the run can tell them apart, so
// the retry has to be safe rather than clever: the second enqueue is
// recognised by its key and does nothing, and the customer gets one message.
func TestAnAbandonedRunThatAlreadySentDoesNotSendAgain(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	seedRun(t, db, `[{"type":"send_email"}]`, true, `{"customer_id":882010}`)

	executor := NewExecutor(db)

	// First pass sends and completes.
	if err := executor.ProcessDue(ctx); err != nil {
		t.Fatalf("first ProcessDue: %v", err)
	}
	if got := queuedEmails(t, db); got != 1 {
		t.Fatalf("%d emails after the first pass, want 1", got)
	}

	// Rewind the run to look abandoned mid-execution, as a worker dying just
	// after the enqueue and just before the mark would have left it.
	if _, err := db.ExecContext(ctx,
		`UPDATE automation_runs SET status = 'running', completed_at = NULL,
		 updated_at = DATE_SUB(NOW(), INTERVAL 60 MINUTE) WHERE id = ?`, execRun); err != nil {
		t.Fatalf("abandon run: %v", err)
	}

	if err := executor.ProcessDue(ctx); err != nil {
		t.Fatalf("second ProcessDue: %v", err)
	}

	if status, lastErr := runStatus(t, db); status != "succeeded" {
		t.Errorf("status = %q (%s), want succeeded", status, lastErr)
	}
	if got := queuedEmails(t, db); got != 1 {
		t.Errorf("%d emails after recovering a run that had already sent, want 1: the customer got a duplicate", got)
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
