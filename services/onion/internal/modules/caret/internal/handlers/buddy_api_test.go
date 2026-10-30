package handlers

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/caret/buddy"
	"github.com/voxire/onion/internal/modules/caret/internal/application"
	auth "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	pb "github.com/voxire/voxie/gen/go/hub/caret/v1"
)

type buddyFake struct {
	mu     sync.Mutex
	action buddy.Action
	saved  buddy.Settings
}

func (f *buddyFake) Routines(context.Context, int64, int64) ([]buddy.Routine, error) { return nil, nil }
func (f *buddyFake) UpdateRoutine(context.Context, int64, int64, int64, buddy.RoutineChange) (buddy.Routine, error) {
	return buddy.Routine{}, buddy.ErrRoutineNotFound
}
func (f *buddyFake) DeleteRoutine(context.Context, int64, int64, int64) error {
	return buddy.ErrRoutineNotFound
}

func (f *buddyFake) Settings(context.Context, int64) (buddy.Settings, error)        { return f.saved, nil }
func (f *buddyFake) Save(_ context.Context, s buddy.Settings) error                 { f.saved = s; return nil }
func (f *buddyFake) Runs(context.Context, int64) ([]buddy.Run, error)               { return nil, nil }
func (f *buddyFake) Actions(context.Context, int64, string) ([]buddy.Action, error) { return nil, nil }
func (f *buddyFake) Action(_ context.Context, b, id int64) (buddy.Action, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.action, nil
}
func (f *buddyFake) ClaimAction(_ context.Context, b, id, u int64, status string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.action.Status != "pending" || time.Now().After(f.action.ExpiresAt) {
		return false, nil
	}
	f.action.Status = status
	f.action.DecidedBy = u
	return true, nil
}
func (f *buddyFake) FinishAction(_ context.Context, b, id, u int64, status string, result json.RawMessage) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.action.Status = status
	f.action.Result = result
	return nil
}

type buddyExecutor struct {
	mu    sync.Mutex
	calls int
	user  int64
}

func (e *buddyExecutor) ExecuteCaretTool(ctx context.Context, _ string, _ map[string]any) (map[string]any, bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.calls++
	_, e.user = application.AuthenticatedIdentity(ctx)
	return map[string]any{"ok": true}, true, nil
}
func buddyTestContext(user int64) context.Context {
	return auth.WithAuthUser(context.Background(), &auth.AuthUser{UserID: user, BusinessID: 2, EffectiveAccess: &authorization.EffectiveAccess{IsAdmin: true}})
}
func TestBuddyDecisionExecutesOnceAsApprover(t *testing.T) {
	f := &buddyFake{action: buddy.Action{ID: 7, ToolName: "flow.create_task", Status: "pending", Input: json.RawMessage(`{"title":"Test"}`), ExpiresAt: time.Now().Add(time.Hour)}}
	exec := &buddyExecutor{}
	api := &BuddyAPI{Store: f, Executors: []application.ToolExecutor{exec}}
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = api.DecideBuddyAction(buddyTestContext(42), &pb.DecideBuddyActionRequest{ActionId: 7, Decision: "approve"})
		}()
	}
	wg.Wait()
	if exec.calls != 1 || exec.user != 42 || f.action.Status != "executed" || f.action.DecidedBy != 42 {
		t.Fatalf("calls=%d user=%d action=%+v", exec.calls, exec.user, f.action)
	}
}
func TestBuddySettingsRequireAdminAndUseCaller(t *testing.T) {
	f := &buddyFake{}
	a := &BuddyAPI{Store: f}
	req := &pb.PutBuddySettingsRequest{IntervalMinutes: 60, DailyRunCap: 10}
	if _, err := a.PutBuddySettings(auth.WithAuthUser(context.Background(), &auth.AuthUser{BusinessID: 2, UserID: 3}), req); err == nil {
		t.Fatal("member allowed")
	}
	if _, err := a.PutBuddySettings(buddyTestContext(42), req); err != nil {
		t.Fatal(err)
	}
	if f.saved.RunAsUserID != 42 || f.saved.BusinessID != 2 {
		t.Fatalf("settings: %+v", f.saved)
	}
}

func TestBuddyDecisionRejectsExpiredAndDenialDoesNotExecute(t *testing.T) {
	for _, tc := range []struct {
		name, decision string
		expires        time.Time
		wantStatus     string
		wantError      bool
	}{
		{"expired", "approve", time.Now().Add(-time.Minute), "pending", true},
		{"denied", "deny", time.Now().Add(time.Hour), "denied", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := &buddyFake{action: buddy.Action{ID: 7, ToolName: "flow.create_task", Status: "pending", Input: json.RawMessage(`{"title":"Test"}`), ExpiresAt: tc.expires}}
			exec := &buddyExecutor{}
			a := &BuddyAPI{Store: f, Executors: []application.ToolExecutor{exec}}
			_, err := a.DecideBuddyAction(buddyTestContext(42), &pb.DecideBuddyActionRequest{ActionId: 7, Decision: tc.decision})
			if (err != nil) != tc.wantError || f.action.Status != tc.wantStatus || exec.calls != 0 {
				t.Fatalf("err=%v status=%s calls=%d", err, f.action.Status, exec.calls)
			}
		})
	}
}
