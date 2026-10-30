package infrastructure

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/caret/buddy/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

const fixtureBusinessID = 920501

func newFixtureRepo(t *testing.T) (*Repository, *sql.DB) {
	t.Helper()
	db := testutil.OpenDB(t)
	clean := func() {
		for _, query := range []string{
			"DELETE FROM caret_buddy_actions WHERE business_id = ?",
			"DELETE FROM caret_buddy_runs WHERE business_id = ?",
			"DELETE FROM caret_buddy_routines WHERE business_id = ?",
			"DELETE FROM caret_buddies WHERE business_id = ?",
		} {
			if _, err := db.Exec(query, fixtureBusinessID); err != nil {
				t.Fatalf("clean fixtures: %v", err)
			}
		}
	}
	clean()
	t.Cleanup(clean)
	return NewRepository(db), db
}

func saveBuddy(t *testing.T, repo *Repository, cap int) domain.Buddy {
	t.Helper()
	ctx := context.Background()
	if err := repo.UpsertBuddy(ctx, domain.Buddy{
		BusinessID: fixtureBusinessID, RunAsUserID: 77, Enabled: true,
		Instructions: "watch tickets", IntervalMinutes: 15, DailyRunCap: cap,
	}); err != nil {
		t.Fatalf("UpsertBuddy: %v", err)
	}
	buddy, err := repo.BuddyByBusiness(ctx, fixtureBusinessID)
	if err != nil {
		t.Fatalf("BuddyByBusiness: %v", err)
	}
	dueRoutine(t, repo)
	return buddy
}

// dueRoutine adds a routine owned by user 77 that is already due.
func dueRoutine(t *testing.T, repo *Repository) domain.Routine {
	t.Helper()
	due := time.Now().Add(-time.Minute)
	routine, err := repo.CreateRoutine(context.Background(), fixtureBusinessID, 77, domain.Routine{
		Name: "Inbox check", Instructions: "watch tickets", EveryMinutes: 15, NextRunAt: &due,
	})
	if err != nil {
		t.Fatalf("CreateRoutine: %v", err)
	}
	return routine
}

func countRuns(t *testing.T, db *sql.DB) int {
	t.Helper()
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM caret_buddy_runs WHERE business_id = ?", fixtureBusinessID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestEnqueueDueQueuesOnceThenWaitsForTheInterval(t *testing.T) {
	repo, db := newFixtureRepo(t)
	saveBuddy(t, repo, 24)
	ctx := context.Background()

	if _, err := repo.EnqueueDue(ctx, 100); err != nil {
		t.Fatalf("EnqueueDue: %v", err)
	}
	if got := countRuns(t, db); got != 1 {
		t.Fatalf("runs = %d, want 1", got)
	}
	if _, err := repo.EnqueueDue(ctx, 100); err != nil {
		t.Fatalf("second EnqueueDue: %v", err)
	}
	if got := countRuns(t, db); got != 1 {
		t.Fatalf("a routine that just ran queued again: runs = %d", got)
	}
}

func TestEnqueueDueRespectsTheDailyCap(t *testing.T) {
	repo, db := newFixtureRepo(t)
	saveBuddy(t, repo, 1)
	ctx := context.Background()
	if _, err := repo.EnqueueDue(ctx, 100); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("UPDATE caret_buddy_routines SET next_run_at = NOW() WHERE business_id = ?", fixtureBusinessID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.EnqueueDue(ctx, 100); err != nil {
		t.Fatal(err)
	}
	if got := countRuns(t, db); got != 1 {
		t.Fatalf("runs = %d, want the cap of 1", got)
	}
	routines, _ := repo.Routines(ctx, fixtureBusinessID, 77)
	if len(routines) != 1 || routines[0].NextRunAt == nil || !routines[0].NextRunAt.After(time.Now().Add(10*time.Minute)) {
		t.Fatalf("a capped routine must still move to its next slot, got %+v", routines)
	}
}

func TestPausedRoutineIsNeverScheduled(t *testing.T) {
	repo, db := newFixtureRepo(t)
	saveBuddy(t, repo, 24)
	ctx := context.Background()
	routines, err := repo.Routines(ctx, fixtureBusinessID, 77)
	if err != nil || len(routines) != 1 {
		t.Fatalf("Routines: %v %v", routines, err)
	}
	paused := routines[0]
	paused.Enabled = false
	if err := repo.UpdateRoutine(ctx, paused); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.EnqueueDue(ctx, 100); err != nil {
		t.Fatal(err)
	}
	if got := countRuns(t, db); got != 0 {
		t.Fatalf("paused routine queued %d runs", got)
	}
}

func TestRoutinesBelongToTheirOwner(t *testing.T) {
	repo, _ := newFixtureRepo(t)
	saveBuddy(t, repo, 24)
	ctx := context.Background()
	routines, _ := repo.Routines(ctx, fixtureBusinessID, 77)
	if others, _ := repo.Routines(ctx, fixtureBusinessID, 78); len(others) != 0 {
		t.Fatalf("another member sees %d routines", len(others))
	}
	if _, err := repo.Routine(ctx, fixtureBusinessID, 78, routines[0].ID); err != ErrRoutineNotFound {
		t.Fatalf("another member read the routine: %v", err)
	}
	if err := repo.DeleteRoutine(ctx, fixtureBusinessID, 78, routines[0].ID); err != ErrRoutineNotFound {
		t.Fatalf("another member deleted the routine: %v", err)
	}
}

func TestClaimRunCarriesTheRoutine(t *testing.T) {
	repo, _ := newFixtureRepo(t)
	saveBuddy(t, repo, 24)
	ctx := context.Background()
	if _, err := repo.EnqueueDue(ctx, 100); err != nil {
		t.Fatal(err)
	}
	run, _, routine, err := repo.ClaimRun(ctx, time.Minute)
	if err != nil || run == nil {
		t.Fatalf("ClaimRun: %v %v", run, err)
	}
	if run.TriggerKind != domain.TriggerRoutine || routine == nil || routine.UserID != 77 || routine.Name != "Inbox check" {
		t.Fatalf("run %+v routine %+v", run, routine)
	}
}

func TestClaimRunHandsEachRunToOneClaimer(t *testing.T) {
	repo, _ := newFixtureRepo(t)
	saveBuddy(t, repo, 24)
	ctx := context.Background()
	if _, err := repo.EnqueueDue(ctx, 100); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	claimed := 0
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			run, buddy, _, err := repo.ClaimRun(ctx, time.Minute)
			if err != nil {
				t.Errorf("ClaimRun: %v", err)
				return
			}
			if run != nil {
				if buddy == nil || buddy.RunAsUserID != 77 || run.Attempts != 1 {
					t.Errorf("claimed run = %+v buddy = %+v", run, buddy)
				}
				mu.Lock()
				claimed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if claimed != 1 {
		t.Fatalf("one queued run was claimed %d times", claimed)
	}
}

func TestExpiredLeaseIsRequeuedThenFailedAfterTheLastAttempt(t *testing.T) {
	repo, db := newFixtureRepo(t)
	saveBuddy(t, repo, 24)
	ctx := context.Background()
	if _, err := repo.EnqueueDue(ctx, 100); err != nil {
		t.Fatal(err)
	}
	run, _, _, err := repo.ClaimRun(ctx, time.Minute)
	if err != nil || run == nil {
		t.Fatalf("claim: %v %v", run, err)
	}
	if err := repo.RecoverStaleRuns(ctx); err != nil {
		t.Fatal(err)
	}
	var status string
	_ = db.QueryRow("SELECT status FROM caret_buddy_runs WHERE id = ?", run.ID).Scan(&status)
	if status != "running" {
		t.Fatalf("a live lease was recovered: status = %s", status)
	}

	if _, err := db.Exec("UPDATE caret_buddy_runs SET lease_until = NOW() - INTERVAL 1 MINUTE WHERE id = ?", run.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.RecoverStaleRuns(ctx); err != nil {
		t.Fatal(err)
	}
	_ = db.QueryRow("SELECT status FROM caret_buddy_runs WHERE id = ?", run.ID).Scan(&status)
	if status != "queued" {
		t.Fatalf("expired lease should requeue, status = %s", status)
	}

	if _, err := db.Exec("UPDATE caret_buddy_runs SET status='running', attempts = ?, lease_until = NOW() - INTERVAL 1 MINUTE WHERE id = ?", domain.MaxRunAttempts, run.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.RecoverStaleRuns(ctx); err != nil {
		t.Fatal(err)
	}
	_ = db.QueryRow("SELECT status FROM caret_buddy_runs WHERE id = ?", run.ID).Scan(&status)
	if status != "failed" {
		t.Fatalf("a run out of attempts should fail, status = %s", status)
	}
}

func TestFinishRunStoresTheReportAndMarksFailures(t *testing.T) {
	repo, db := newFixtureRepo(t)
	saveBuddy(t, repo, 24)
	ctx := context.Background()
	_, _ = repo.EnqueueDue(ctx, 100)
	run, _, _, _ := repo.ClaimRun(ctx, time.Minute)

	if err := repo.FinishRun(ctx, run.ID, domain.RunResult{Report: "all quiet", ToolCalls: 3, Tokens: 900}); err != nil {
		t.Fatal(err)
	}
	var status, report string
	var calls int
	if err := db.QueryRow("SELECT status, report, tool_calls FROM caret_buddy_runs WHERE id = ?", run.ID).Scan(&status, &report, &calls); err != nil {
		t.Fatal(err)
	}
	if status != "done" || report != "all quiet" || calls != 3 {
		t.Fatalf("finished run = %s %q %d", status, report, calls)
	}

	if err := repo.FinishRun(ctx, run.ID, domain.RunResult{Error: "model down"}); err != nil {
		t.Fatal(err)
	}
	_ = db.QueryRow("SELECT status FROM caret_buddy_runs WHERE id = ?", run.ID).Scan(&status)
	if status != "failed" {
		t.Fatalf("status = %s, want failed", status)
	}
}

func TestProposeActionIsIdempotentPerRunAndCall(t *testing.T) {
	repo, db := newFixtureRepo(t)
	saveBuddy(t, repo, 24)
	ctx := context.Background()
	_, _ = repo.EnqueueDue(ctx, 100)
	run, _, _, _ := repo.ClaimRun(ctx, time.Minute)

	for i := 0; i < 2; i++ {
		if err := repo.ProposeAction(ctx, run.ID, fixtureBusinessID, "crm.create_customer", map[string]any{"name": "Dana"}, "fp-1", 72*time.Hour); err != nil {
			t.Fatalf("ProposeAction: %v", err)
		}
	}
	var n int
	var status string
	if err := db.QueryRow("SELECT COUNT(*), MIN(status) FROM caret_buddy_actions WHERE run_id = ?", run.ID).Scan(&n, &status); err != nil {
		t.Fatal(err)
	}
	if n != 1 || status != "pending" {
		t.Fatalf("actions = %d status = %s, want one pending", n, status)
	}

	if _, err := db.Exec("UPDATE caret_buddy_actions SET expires_at = NOW() - INTERVAL 1 HOUR WHERE run_id = ?", run.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.ExpireActions(ctx); err != nil {
		t.Fatal(err)
	}
	_ = db.QueryRow("SELECT status FROM caret_buddy_actions WHERE run_id = ?", run.ID).Scan(&status)
	if status != "expired" {
		t.Fatalf("status = %s, want expired", status)
	}
}

func TestClaimActionIsAtomicAndRejectsExpired(t *testing.T) {
	repo, db := newFixtureRepo(t)
	saveBuddy(t, repo, 24)
	ctx := context.Background()
	if _, err := repo.EnqueueDue(ctx, 100); err != nil {
		t.Fatal(err)
	}
	run, _, _, err := repo.ClaimRun(ctx, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.ProposeAction(ctx, run.ID, fixtureBusinessID, "flow.create_task", map[string]any{"title": "Test"}, "atomic", time.Hour); err != nil {
		t.Fatal(err)
	}
	var id int64
	if err := db.QueryRow("SELECT id FROM caret_buddy_actions WHERE run_id = ?", run.ID).Scan(&id); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	claims := make(chan bool, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, err := repo.ClaimAction(ctx, fixtureBusinessID, id, 77, "approved")
			if err != nil {
				t.Error(err)
			}
			claims <- ok
		}()
	}
	wg.Wait()
	close(claims)
	n := 0
	for ok := range claims {
		if ok {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("successful claims: %d", n)
	}
	if err := repo.FinishAction(ctx, fixtureBusinessID, id, 77, "executed", []byte(`{"ok":true}`)); err != nil {
		t.Fatal(err)
	}
	if err := repo.ProposeAction(ctx, run.ID, fixtureBusinessID, "flow.create_task", map[string]any{"title": "Test"}, "expired", time.Hour); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("UPDATE caret_buddy_actions SET expires_at = NOW() - INTERVAL 1 HOUR WHERE run_id = ? AND fingerprint = ?", run.ID, "expired"); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT id FROM caret_buddy_actions WHERE run_id = ? AND fingerprint = ?", run.ID, "expired").Scan(&id); err != nil {
		t.Fatal(err)
	}
	ok, err := repo.ClaimAction(ctx, fixtureBusinessID, id, 77, "denied")
	if err != nil || ok {
		t.Fatalf("expired claim = %v, %v", ok, err)
	}
}
