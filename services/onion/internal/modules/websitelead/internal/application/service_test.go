package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/websitelead/internal/domain"
)

type mockRepo struct {
	createID  int64
	createErr error
	createGot domain.LeadInput
}

func (m *mockRepo) Create(_ context.Context, input domain.LeadInput) (int64, error) {
	m.createGot = input
	return m.createID, m.createErr
}

func TestSubmit_Delegates(t *testing.T) {
	m := &mockRepo{createID: 7}
	s := &Service{repo: m}
	id, err := s.Submit(context.Background(), domain.LeadInput{Name: "Alice"})
	if err != nil || id != 7 {
		t.Errorf("id=%d err=%v", id, err)
	}
	if m.createGot.Name != "Alice" {
		t.Errorf("repo got %+v", m.createGot)
	}
}

func TestSubmit_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Submit(context.Background(), domain.LeadInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestAsyncLeadContextDetachesCancellationAndPreservesValues(t *testing.T) {
	type ctxKey struct{}
	key := ctxKey{}
	requestCtx, cancel := context.WithCancel(context.WithValue(context.Background(), key, "request-value"))
	cancel()

	taskCtx, taskCancel := asyncLeadContext(requestCtx)
	defer taskCancel()

	if err := taskCtx.Err(); err != nil {
		t.Fatalf("detached task context is already canceled: %v", err)
	}
	if got := taskCtx.Value(key); got != "request-value" {
		t.Fatalf("task context lost request value: %v", got)
	}
}
