package buddy

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/caret/buddy/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

const runnerBusinessID = 920502

type fakeAccess struct {
	err    error
	called int
}

func (f *fakeAccess) GetEffectiveAccess(context.Context, int64, int64) (*authorization.EffectiveAccess, error) {
	f.called++
	if f.err != nil {
		return nil, f.err
	}
	return &authorization.EffectiveAccess{}, nil
}

type fakeAgent struct {
	user    *authPkg.AuthUser
	task    Task
	called  int
	outcome Outcome
	err     error
	panics  bool
}

func (f *fakeAgent) Run(ctx context.Context, task Task) (Outcome, error) {
	f.called++
	f.task = task
	f.user, _ = authPkg.GetAuthUser(ctx)
	if f.panics {
		panic("boom")
	}
	return f.outcome, f.err
}

// runOnce queues one run for a fresh buddy, claims it and waits for the
// worker goroutine to save the result.
func runOnce(t *testing.T, agent Agent, access authPkg.EffectiveAccessResolver) (status, report, runErr string) {
	t.Helper()
	db := testutil.OpenDB(t)
	clean := func() {
		for _, q := range []string{
			"DELETE FROM caret_buddy_actions WHERE business_id = ?",
			"DELETE FROM caret_buddy_runs WHERE business_id = ?",
			"DELETE FROM caret_buddies WHERE business_id = ?",
		} {
			if _, err := db.Exec(q, runnerBusinessID); err != nil {
				t.Fatalf("clean: %v", err)
			}
		}
	}
	clean()
	t.Cleanup(clean)

	m := NewModule(context.Background(), db)
	ctx := context.Background()
	if err := m.repo.UpsertBuddy(ctx, domain.Buddy{
		BusinessID: runnerBusinessID, RunAsUserID: 88, Enabled: true,
		Instructions: "check the inbox", IntervalMinutes: 15, DailyRunCap: 24,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.repo.EnqueueDue(ctx, 100); err != nil {
		t.Fatal(err)
	}
	if !m.claimAndRun(ctx, agent, access, make(chan struct{}, 1)) {
		t.Fatal("expected a queued run to be claimed")
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var rep, e *string
		if err := db.QueryRow("SELECT status, report, error FROM caret_buddy_runs WHERE business_id = ?", runnerBusinessID).Scan(&status, &rep, &e); err != nil {
			t.Fatal(err)
		}
		if status != "running" {
			if rep != nil {
				report = *rep
			}
			if e != nil {
				runErr = *e
			}
			return status, report, runErr
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("run never finished")
	return
}

func TestRunnerRunsTheAgentAsTheBuddyUserAndSavesTheReport(t *testing.T) {
	agent := &fakeAgent{outcome: Outcome{Report: "inbox is clear", ToolCalls: 2, Tokens: 400}}
	status, report, _ := runOnce(t, agent, &fakeAccess{})
	if status != "done" || report != "inbox is clear" {
		t.Fatalf("status=%s report=%q", status, report)
	}
	if agent.user == nil || agent.user.UserID != 88 || agent.user.BusinessID != runnerBusinessID || agent.user.EffectiveAccess == nil {
		t.Fatalf("agent identity = %+v, want user 88 in business %d with resolved access", agent.user, runnerBusinessID)
	}
	if agent.task.Instructions != "check the inbox" || agent.task.BusinessID != runnerBusinessID || agent.task.RunID == 0 {
		t.Fatalf("task = %+v", agent.task)
	}
}

func TestRunnerNeverRunsTheAgentWhenAccessIsLost(t *testing.T) {
	agent := &fakeAgent{}
	status, _, runErr := runOnce(t, agent, &fakeAccess{err: errors.New("no membership")})
	if status != "failed" || agent.called != 0 {
		t.Fatalf("status=%s agentCalls=%d; a buddy without access must not run", status, agent.called)
	}
	if runErr == "" {
		t.Fatal("the failure reason must be saved")
	}
}

func TestRunnerFailsTheRunWhenTheAgentErrorsOrPanics(t *testing.T) {
	status, _, runErr := runOnce(t, &fakeAgent{err: errors.New("model call failed")}, &fakeAccess{})
	if status != "failed" || runErr != "model call failed" {
		t.Fatalf("agent error: status=%s error=%q", status, runErr)
	}
	status, _, runErr = runOnce(t, &fakeAgent{panics: true}, &fakeAccess{})
	if status != "failed" || runErr != reportPanicText {
		t.Fatalf("agent panic: status=%s error=%q", status, runErr)
	}
}
