package emailqueue

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	sqlc "github.com/voxire/onion/internal/platform/emailqueue/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/testutil"
)

// The visibility timeout half of the lifecycle. A worker that dies holding a
// claim leaves the row in 'processing' with nothing looking at it, so the
// sweeper requeues it; and the worker it belonged to must not get to write the
// row's state afterwards. The two are what make the queue at-least-once
// without letting a dead worker's late write corrupt a live job.
//
// Internal test because recoverStuck is the sweeper under test and the exported
// surface deliberately does not expose it.
func TestRecoveredClaimRejectsTheWorkerThatLostIt(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	q := sqlc.New(db)

	subject := "lease-" + uuid.New().String()
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM email_jobs WHERE subject = ?", subject)
	})
	if err := Enqueue(ctx, db, Job{To: "lease@test.local", Subject: subject, HTML: "<p>body</p>"}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	var id uint64
	if err := db.QueryRowContext(ctx, "SELECT id FROM email_jobs WHERE subject = ?", subject).Scan(&id); err != nil {
		t.Fatalf("read back id: %v", err)
	}

	state := func() (string, sql.NullString, int) {
		t.Helper()
		var (
			status   string
			jobID    sql.NullString
			attempts int
		)
		if err := db.QueryRowContext(ctx,
			"SELECT status, job_id, attempts FROM email_jobs WHERE id = ?", id).
			Scan(&status, &jobID, &attempts); err != nil {
			t.Fatalf("read state: %v", err)
		}
		return status, jobID, attempts
	}

	stale := sql.NullString{String: uuid.New().String(), Valid: true}
	if res, err := q.ClaimJob(ctx, stale); err != nil {
		t.Fatalf("ClaimJob: %v", err)
	} else if n, _ := res.RowsAffected(); n != 1 {
		t.Fatalf("claim affected %d rows, want 1", n)
	}

	// Stand in for a worker that stopped making progress: nothing touches
	// updated_at between the claim and here, so ageing the row is the same
	// state a hung worker leaves behind.
	if _, err := db.ExecContext(ctx,
		"UPDATE email_jobs SET updated_at = DATE_SUB(NOW(), INTERVAL ? SECOND) WHERE id = ?",
		int64(claimLease/time.Second)+60, id); err != nil {
		t.Fatalf("age the claim: %v", err)
	}
	NewPool(db).recoverStuck(ctx)

	if status, jobID, attempts := state(); status != "pending" || jobID.Valid || attempts != 1 {
		t.Fatalf("after recovery job %d is status=%s job_id=%+v attempts=%d, want pending/NULL/1",
			id, status, jobID, attempts)
	}

	// The recovered worker's late ack must not finish a job somebody else now
	// owns. Unfenced, this is what marked a requeued job 'sent' while its new
	// owner was still sending it.
	if n, err := q.MarkJobSent(ctx, sqlc.MarkJobSentParams{Attempts: 1, ID: id, JobID: stale}); err != nil || n != 0 {
		t.Fatalf("stale ack updated %d rows (err %v), want 0", n, err)
	}
	// Its late failure must not requeue it either, and must not rewind the
	// attempt counter recovery just charged it.
	if n, err := q.RescheduleJob(ctx, sqlc.RescheduleJobParams{
		Status:    sqlc.EmailJobsStatusPending,
		Attempts:  0,
		LastError: sql.NullString{String: "late failure", Valid: true},
		JobID:     stale,
		DATEADD:   0,
		ID:        id,
	}); err != nil || n != 0 {
		t.Fatalf("stale nack updated %d rows (err %v), want 0", n, err)
	}
	if status, _, attempts := state(); status != "pending" || attempts != 1 {
		t.Errorf("job %d is status=%s attempts=%d after the stale writes, want pending/1", id, status, attempts)
	}

	// The retry is a normal claim, and its owner can finish the job.
	fresh := sql.NullString{String: uuid.New().String(), Valid: true}
	if res, err := q.ClaimJob(ctx, fresh); err != nil {
		t.Fatalf("re-claim: %v", err)
	} else if n, _ := res.RowsAffected(); n != 1 {
		t.Fatalf("re-claim affected %d rows, want 1", n)
	}
	if _, err := q.GetClaimedJob(ctx, fresh); err != nil {
		t.Fatalf("the re-claimed job is not readable by its new owner: %v", err)
	}
	if n, err := q.MarkJobSent(ctx, sqlc.MarkJobSentParams{Attempts: 2, ID: id, JobID: fresh}); err != nil || n != 1 {
		t.Fatalf("new owner's ack updated %d rows (err %v), want 1", n, err)
	}
	if status, jobID, attempts := state(); status != "sent" || jobID.Valid || attempts != 2 {
		t.Errorf("job %d ended as status=%s job_id=%+v attempts=%d, want sent/NULL/2",
			id, status, jobID, attempts)
	}
}
