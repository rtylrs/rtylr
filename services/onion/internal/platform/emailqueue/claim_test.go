package emailqueue_test

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/platform/emailqueue"
	sqlc "github.com/voxire/onion/internal/platform/emailqueue/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/testutil"
)

// The invariant behind the whole pool: at most one worker can move a given job
// out of 'pending', and only the worker that won may acknowledge or requeue it.
// Every test here is about a claim token having to be honoured, so a worker
// whose claim was taken away cannot change the row's state.
//
// The claim is global rather than per test, so anything these helpers win that
// belongs to another package's fixture is handed straight back.

func claimToken() sql.NullString {
	return sql.NullString{String: uuid.New().String(), Valid: true}
}

// enqueueN queues n jobs under one run-unique subject, so a test finds exactly
// its own rows and can clean up after itself.
func enqueueN(ctx context.Context, t *testing.T, db *sql.DB, n int) []uint64 {
	t.Helper()

	subject := fmt.Sprintf("claim-%d", time.Now().UnixNano())
	to := fmt.Sprintf("%s@test.local", subject)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM email_jobs WHERE subject = ?", subject)
	})

	for i := 0; i < n; i++ {
		if err := emailqueue.Enqueue(ctx, db, emailqueue.Job{
			// Highest priority, so a job another package's test left pending in
			// the shared database cannot starve these out of the claim order.
			Priority: 10,
			To:       to,
			Subject:  subject,
			HTML:     "<p>body</p>",
		}); err != nil {
			t.Fatalf("enqueue %d: %v", i, err)
		}
	}

	ids := idsBySubject(ctx, t, db, subject)
	if len(ids) != n {
		t.Fatalf("queued %d jobs, read back %d", n, len(ids))
	}
	return ids
}

func idsBySubject(ctx context.Context, t *testing.T, db *sql.DB, subject string) []uint64 {
	t.Helper()

	rows, err := db.QueryContext(ctx, "SELECT id FROM email_jobs WHERE subject = ? ORDER BY id", subject)
	if err != nil {
		t.Fatalf("read back ids: %v", err)
	}
	defer rows.Close()

	var ids []uint64
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("scan id: %v", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate ids: %v", err)
	}
	return ids
}

func idSet(ids []uint64) map[uint64]bool {
	m := make(map[uint64]bool, len(ids))
	for _, id := range ids {
		m[id] = true
	}
	return m
}

// claimOwnRow claims until it wins one of this test's own jobs, releasing any
// row of another package's it picks up on the way.
func claimOwnRow(ctx context.Context, t *testing.T, q *sqlc.Queries, mine map[uint64]bool) sql.NullString {
	t.Helper()

	// Bounded so a shared database full of other packages' fixtures cannot
	// wedge the test. A handful of attempts is generous in practice.
	for i := 0; i < 50; i++ {
		token := claimToken()
		res, err := q.ClaimJob(ctx, token)
		if err != nil {
			t.Fatalf("ClaimJob: %v", err)
		}
		n, err := res.RowsAffected()
		if err != nil {
			t.Fatalf("RowsAffected: %v", err)
		}
		if n == 0 {
			t.Fatal("nothing left to claim")
		}
		if n != 1 {
			t.Fatalf("a single claim updated %d rows, want 1", n)
		}
		job, err := q.GetClaimedJob(ctx, token)
		if err != nil {
			t.Fatalf("GetClaimedJob: %v", err)
		}
		if !mine[job.ID] {
			if _, err := q.ReleaseClaim(ctx, token); err != nil {
				t.Fatalf("ReleaseClaim: %v", err)
			}
			continue
		}
		return token
	}
	t.Fatal("never won a claim on one of this test's own jobs")
	return sql.NullString{}
}

type jobState struct {
	Status   string
	JobID    sql.NullString
	Attempts int
}

// stateOf reads the columns the fence is defined over, so a test asserts on the
// row and not only on the row count an UPDATE reported.
func stateOf(ctx context.Context, t *testing.T, db *sql.DB, id uint64) jobState {
	t.Helper()

	var s jobState
	if err := db.QueryRowContext(ctx,
		"SELECT status, job_id, attempts FROM email_jobs WHERE id = ?", id).
		Scan(&s.Status, &s.JobID, &s.Attempts); err != nil {
		t.Fatalf("read state of job %d: %v", id, err)
	}
	return s
}

// Eight workers draining twenty jobs at once. Nothing coordinates them, which
// is the point: the database has to make the claim safe on its own. Every job
// must be delivered exactly once.
func TestConcurrentClaimersDeliverEachJobExactlyOnce(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const (
		claimers = 8
		jobs     = 20
		// Bounds each worker so a shared database cannot wedge the test; a
		// clean drain takes one claim per job.
		maxRounds = 4 * jobs
	)
	ids := enqueueN(ctx, t, db, jobs)
	mine := idSet(ids)

	q := sqlc.New(db)
	var (
		mu      sync.Mutex
		claimed = make(map[uint64]string, jobs)
		wg      sync.WaitGroup
	)
	wg.Add(claimers)
	for w := 0; w < claimers; w++ {
		go func() {
			defer wg.Done()
			for round := 0; round < maxRounds; round++ {
				token := claimToken()
				res, err := q.ClaimJob(ctx, token)
				if err != nil {
					t.Errorf("ClaimJob: %v", err)
					return
				}
				n, err := res.RowsAffected()
				if err != nil {
					t.Errorf("RowsAffected: %v", err)
					return
				}
				if n == 0 {
					return // queue drained
				}
				if n != 1 {
					t.Errorf("a single claim updated %d rows, want 1", n)
					return
				}
				job, err := q.GetClaimedJob(ctx, token)
				if err != nil {
					t.Errorf("GetClaimedJob: %v", err)
					return
				}
				if !mine[job.ID] {
					// Another package's fixture: hand it back rather than
					// delivering a row its own test is asserting on.
					if _, err := q.ReleaseClaim(ctx, token); err != nil {
						t.Errorf("ReleaseClaim: %v", err)
					}
					continue
				}
				// Only the winner acknowledges, and only with its own token.
				if _, err := q.MarkJobSent(ctx, sqlc.MarkJobSentParams{
					Attempts: job.Attempts + 1,
					ID:       job.ID,
					JobID:    token,
				}); err != nil {
					t.Errorf("MarkJobSent: %v", err)
					return
				}

				mu.Lock()
				if prev, dup := claimed[job.ID]; dup {
					t.Errorf("job %d was delivered twice: by claim tokens %q and %q", job.ID, prev, token.String)
				}
				claimed[job.ID] = token.String
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if len(claimed) != jobs {
		t.Errorf("%d of %d jobs were delivered", len(claimed), jobs)
	}
	for _, id := range ids {
		if _, ok := claimed[id]; !ok {
			t.Errorf("job %d was never delivered", id)
		}
		if got := stateOf(ctx, t, db, id); got.Status != "sent" || got.Attempts != 1 {
			t.Errorf("job %d ended as status=%s attempts=%d, want sent/1", id, got.Status, got.Attempts)
		}
	}
}

// An acknowledgement from a worker that never won the claim must change
// nothing. Before the ack was fenced on the token it matched on id alone, so it
// stamped 'sent' on a job its real owner was still sending.
func TestAckFromAWorkerThatDidNotWinTheClaimIsRefused(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	q := sqlc.New(db)

	id := enqueueN(ctx, t, db, 1)[0]
	winner := claimOwnRow(ctx, t, q, idSet([]uint64{id}))

	loser := claimToken()
	n, err := q.MarkJobSent(ctx, sqlc.MarkJobSentParams{Attempts: 1, ID: id, JobID: loser})
	if err != nil {
		t.Fatalf("MarkJobSent: %v", err)
	}
	if n != 0 {
		t.Errorf("ack from a worker that never claimed the job updated %d rows, want 0", n)
	}
	if got := stateOf(ctx, t, db, id); got.Status != "processing" || got.JobID.String != winner.String {
		t.Errorf("job %d is status=%s job_id=%q, want processing under the winner's token %q",
			id, got.Status, got.JobID.String, winner.String)
	}

	// The winner's own ack is the one that counts, and repeating it is a no-op
	// rather than a second write.
	if n, err := q.MarkJobSent(ctx, sqlc.MarkJobSentParams{Attempts: 1, ID: id, JobID: winner}); err != nil || n != 1 {
		t.Fatalf("winner's ack updated %d rows (err %v), want 1", n, err)
	}
	if n, err := q.MarkJobSent(ctx, sqlc.MarkJobSentParams{Attempts: 2, ID: id, JobID: winner}); err != nil || n != 0 {
		t.Errorf("repeated ack updated %d rows (err %v), want 0", n, err)
	}
	if got := stateOf(ctx, t, db, id); got.Status != "sent" || got.Attempts != 1 {
		t.Errorf("job %d ended as status=%s attempts=%d, want sent/1", id, got.Status, got.Attempts)
	}
}

// The nack needs the same fence as the ack. Unfenced, a slow worker's failure
// requeued a job another worker had already sent, and rewound the attempt
// counter, so a job that could never be delivered retried forever.
func TestNackFromAWorkerThatDidNotWinTheClaimIsRefused(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	q := sqlc.New(db)

	id := enqueueN(ctx, t, db, 1)[0]
	winner := claimOwnRow(ctx, t, q, idSet([]uint64{id}))

	loser := claimToken()
	n, err := q.RescheduleJob(ctx, sqlc.RescheduleJobParams{
		Status:    sqlc.EmailJobsStatusPending,
		Attempts:  1,
		LastError: sql.NullString{String: "stale failure", Valid: true},
		JobID:     loser,
		DATEADD:   0,
		ID:        id,
	})
	if err != nil {
		t.Fatalf("RescheduleJob: %v", err)
	}
	if n != 0 {
		t.Errorf("nack from a worker that never claimed the job updated %d rows, want 0", n)
	}
	if got := stateOf(ctx, t, db, id); got.Status != "processing" || got.Attempts != 0 {
		t.Errorf("job %d is status=%s attempts=%d, want processing/0", id, got.Status, got.Attempts)
	}

	// The owner's nack does apply, and requeues by dropping the claim.
	if n, err := q.RescheduleJob(ctx, sqlc.RescheduleJobParams{
		Status:    sqlc.EmailJobsStatusPending,
		Attempts:  1,
		LastError: sql.NullString{String: "provider rejected the message", Valid: true},
		JobID:     winner,
		DATEADD:   0,
		ID:        id,
	}); err != nil || n != 1 {
		t.Fatalf("owner's nack updated %d rows (err %v), want 1", n, err)
	}
	if got := stateOf(ctx, t, db, id); got.Status != "pending" || got.JobID.Valid || got.Attempts != 1 {
		t.Errorf("job %d ended as status=%s job_id=%+v attempts=%d, want pending/NULL/1",
			id, got.Status, got.JobID, got.Attempts)
	}
}
