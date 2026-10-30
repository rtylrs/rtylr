package application

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

type replyScript struct {
	mu      sync.Mutex
	replies []string
	prompts []string
}

func (s *replyScript) CompleteText(_ context.Context, prompt string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prompts = append(s.prompts, prompt)
	if len(s.replies) == 0 {
		return "", errors.New("script exhausted")
	}
	reply := s.replies[0]
	s.replies = s.replies[1:]
	return reply, nil
}

type proposal struct {
	runID, businessID int64
	tool              string
	input             map[string]any
}

type recordingStore struct {
	mu        sync.Mutex
	proposals []proposal
	err       error
}

func (s *recordingStore) ProposeAction(_ context.Context, runID, businessID int64, tool string, input map[string]any, _ string, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	s.proposals = append(s.proposals, proposal{runID, businessID, tool, input})
	return nil
}

type countingExecutor struct {
	mu    sync.Mutex
	calls []string
}

func (e *countingExecutor) ExecuteCaretTool(_ context.Context, name string, _ map[string]any) (map[string]any, bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.calls = append(e.calls, name)
	return map[string]any{"tickets": []any{map[string]any{"id": 7, "subject": "Refund please"}}}, true, nil
}

func newTestBuddy(ai TextCompleter, store BuddyActionStore, exec ToolExecutor) *BuddyAgent {
	return &BuddyAgent{AI: ai, Store: store, Quota: &testBuddyQuota{remaining: 100}, Executors: func() []ToolExecutor { return []ToolExecutor{exec} }}
}

type testBuddyQuota struct {
	remaining int
	actual    []int64
	reserved  []int64
}

func (q *testBuddyQuota) ReserveQuota(_ context.Context, business, user int64) (bool, string, int64, int64, int64, int64) {
	if business <= 0 || user != 9 || q.remaining <= 0 {
		return false, "", 0, 0, 0, 0
	}
	q.remaining--
	return true, "", 0, 0, 0, 100
}

func (q *testBuddyQuota) ReconcileUsage(_ context.Context, _, _ int64, actual, reserved int64) {
	q.actual = append(q.actual, actual)
	q.reserved = append(q.reserved, reserved)
}

func TestBuddyReadsRunAndWritesOnlyBecomeProposals(t *testing.T) {
	ai := &replyScript{replies: []string{
		`<tool>{"name":"support.list_tickets","input":{}}</tool><tool>{"name":"crm.create_customer","input":{"name":"Dana"}}</tool>`,
		`<report>Ticket 7 needs a refund decision. Proposed adding Dana.</report>`,
	}}
	store, exec := &recordingStore{}, &countingExecutor{}

	out, err := newTestBuddy(ai, store, exec).Run(context.Background(), BuddyTask{UserID: 9, RunID: 11, BusinessID: 5, Instructions: "Watch tickets"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(exec.calls) != 1 || exec.calls[0] != "support.list_tickets" {
		t.Fatalf("executed %v, want only the read", exec.calls)
	}
	if len(store.proposals) != 1 || store.proposals[0].tool != "crm.create_customer" || store.proposals[0].runID != 11 || store.proposals[0].businessID != 5 {
		t.Fatalf("proposals = %+v", store.proposals)
	}
	if out.Proposals != 1 || out.ToolCalls != 2 || !strings.Contains(out.Report, "Ticket 7") {
		t.Fatalf("outcome = %+v", out)
	}
	if !strings.Contains(ai.prompts[1], "Refund please") || !strings.Contains(ai.prompts[1], "Filed for a person to approve") {
		t.Fatalf("second prompt missing tool results:\n%s", ai.prompts[1])
	}
}

func TestBuddyRefusesForbiddenAndUnknownTools(t *testing.T) {
	ai := &replyScript{replies: []string{
		`<tool>{"name":"agents.run","input":{"tasks":["x"]}}</tool><tool>{"name":"memory.remember","input":{"key":"k","value":"v"}}</tool><tool>{"name":"nope.nothing","input":{}}</tool>`,
		`<report>Nothing to do.</report>`,
	}}
	store, exec := &recordingStore{}, &countingExecutor{}

	out, err := newTestBuddy(ai, store, exec).Run(context.Background(), BuddyTask{UserID: 9, RunID: 1, BusinessID: 1})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(exec.calls) != 0 || len(store.proposals) != 0 || out.ToolCalls != 0 {
		t.Fatalf("forbidden tools ran: calls=%v proposals=%v outcome=%+v", exec.calls, store.proposals, out)
	}
}

func TestBuddyFailsWhenProposalCannotBeSaved(t *testing.T) {
	ai := &replyScript{replies: []string{
		`<tool>{"name":"crm.create_customer","input":{"name":"Dana"}}</tool>`,
		`<report>Tried to add Dana.</report>`,
	}}
	store := &recordingStore{err: errors.New("db down")}

	out, err := newTestBuddy(ai, store, &countingExecutor{}).Run(context.Background(), BuddyTask{UserID: 9, RunID: 1, BusinessID: 1})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if out.Proposals != 0 || !strings.Contains(ai.prompts[1], "could not be saved") {
		t.Fatalf("a failed save must not count as a proposal: %+v\n%s", out, ai.prompts[1])
	}
}

func TestBuddyStopsAtTheRoundLimit(t *testing.T) {
	replies := make([]string, 0, buddyMaxRounds+1)
	for i := 0; i <= buddyMaxRounds; i++ {
		replies = append(replies, `<tool>{"name":"support.list_tickets","input":{"page":`+string(rune('1'+i))+`}}</tool>`)
	}
	ai := &replyScript{replies: replies}

	_, err := newTestBuddy(ai, &recordingStore{}, &countingExecutor{}).Run(context.Background(), BuddyTask{UserID: 9, RunID: 1, BusinessID: 1})
	if err == nil {
		t.Fatal("a model that never reports must fail the run")
	}
	if len(ai.prompts) != buddyMaxRounds+1 || !strings.Contains(ai.prompts[buddyMaxRounds], "No more tool calls are possible") {
		t.Fatalf("model called %d times; last round must be tool-free", len(ai.prompts))
	}
}

func TestBuddyReportsModelErrors(t *testing.T) {
	_, err := newTestBuddy(&replyScript{}, &recordingStore{}, &countingExecutor{}).Run(context.Background(), BuddyTask{UserID: 9, RunID: 1, BusinessID: 1})
	if err == nil {
		t.Fatal("expected the model failure to fail the run")
	}
}

func TestBuddyQuotaStopsBeforeNextModelCallAndRecordsUsage(t *testing.T) {
	ai := &replyScript{replies: []string{`<tool>{"name":"support.list_tickets","input":{}}</tool>`, `<report>done</report>`}}
	q := &testBuddyQuota{remaining: 1}
	agent := newTestBuddy(ai, &recordingStore{}, &countingExecutor{})
	agent.Quota = q
	out, err := agent.Run(context.Background(), BuddyTask{RunID: 1, BusinessID: 5, UserID: 9})
	if err == nil || !strings.Contains(err.Error(), "quota") || len(ai.prompts) != 1 {
		t.Fatalf("err=%v calls=%d", err, len(ai.prompts))
	}
	if out.Tokens <= 0 || len(q.actual) != 1 || q.actual[0] != out.Tokens || q.reserved[0] != 100 {
		t.Fatalf("out=%+v quota=%+v", out, q)
	}
}

func TestBuddyExhaustedQuotaNeverCallsModel(t *testing.T) {
	ai := &replyScript{replies: []string{"<report>done</report>"}}
	q := &testBuddyQuota{}
	agent := newTestBuddy(ai, &recordingStore{}, &countingExecutor{})
	agent.Quota = q
	out, err := agent.Run(context.Background(), BuddyTask{RunID: 1, BusinessID: 5, UserID: 9})
	if err == nil || !strings.Contains(err.Error(), "quota") || len(ai.prompts) != 0 || out.Tokens != 0 || len(q.actual) != 0 {
		t.Fatalf("out=%+v err=%v calls=%d quota=%+v", out, err, len(ai.prompts), q)
	}
}

func TestBuddyReconcilesFailedModelCall(t *testing.T) {
	q := &testBuddyQuota{remaining: 1}
	agent := newTestBuddy(&replyScript{}, &recordingStore{}, &countingExecutor{})
	agent.Quota = q
	out, err := agent.Run(context.Background(), BuddyTask{RunID: 1, BusinessID: 5, UserID: 9})
	if err == nil || out.Tokens <= 0 || len(q.actual) != 1 || q.actual[0] != out.Tokens {
		t.Fatalf("out=%+v err=%v quota=%+v", out, err, q)
	}
}
