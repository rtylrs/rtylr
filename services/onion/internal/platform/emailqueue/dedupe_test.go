package emailqueue_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/platform/emailqueue"
	"github.com/voxire/onion/internal/testutil"
)

// The primitive an abandoned automation run relies on: a producer that can
// name its message enqueues idempotently, so being retried does not send the
// recipient a second copy.
func TestEnqueueWithADedupeKeyIsIdempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	key := fmt.Sprintf("test:dedupe:%d", time.Now().UnixNano())
	to := fmt.Sprintf("dedupe-%d@test.local", time.Now().UnixNano())

	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM email_jobs WHERE to_email = ?", to)
	})

	count := func() int {
		t.Helper()
		var n int
		if err := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM email_jobs WHERE to_email = ?", to).Scan(&n); err != nil {
			t.Fatalf("count: %v", err)
		}
		return n
	}

	job := emailqueue.Job{To: to, Subject: "s", HTML: "<p>b</p>", DedupeKey: key}

	for attempt := 1; attempt <= 3; attempt++ {
		if err := emailqueue.Enqueue(ctx, db, job); err != nil {
			t.Fatalf("Enqueue attempt %d: %v", attempt, err)
		}
	}
	if got := count(); got != 1 {
		t.Errorf("%d jobs queued after three enqueues under one key, want 1", got)
	}

	// A different key is a different message, not a duplicate.
	other := job
	other.DedupeKey = key + ":2"
	if err := emailqueue.Enqueue(ctx, db, other); err != nil {
		t.Fatalf("Enqueue under a second key: %v", err)
	}
	if got := count(); got != 2 {
		t.Errorf("%d jobs after a second distinct key, want 2", got)
	}
}

// Every existing caller sets no key and must keep queueing unconditionally.
// NULL does not collide in a unique index, which is what makes the column
// additive in practice rather than only on paper.
func TestEnqueueWithoutADedupeKeyStillQueuesEveryMessage(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	to := fmt.Sprintf("nodedupe-%d@test.local", time.Now().UnixNano())
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM email_jobs WHERE to_email = ?", to)
	})

	job := emailqueue.Job{To: to, Subject: "s", HTML: "<p>b</p>"}
	for attempt := 1; attempt <= 3; attempt++ {
		if err := emailqueue.Enqueue(ctx, db, job); err != nil {
			t.Fatalf("Enqueue attempt %d: %v", attempt, err)
		}
	}

	var n int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM email_jobs WHERE to_email = ?", to).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 3 {
		t.Errorf("%d jobs queued without a key, want 3: an unkeyed caller was deduplicated", n)
	}
}
