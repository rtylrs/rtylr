package application

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/websitelead/internal/domain"
	emailPkg "github.com/voxire/onion/internal/platform/email"
)

type mockRepo struct {
	createID     int64
	createErr    error
	createGot    domain.LeadInput
	subscribeID  int64
	subscribeErr error
	subscribeGot domain.SubscriberInput
}

func (m *mockRepo) Create(_ context.Context, input domain.LeadInput) (int64, error) {
	m.createGot = input
	return m.createID, m.createErr
}

func (m *mockRepo) CreateSubscriber(_ context.Context, input domain.SubscriberInput) (int64, error) {
	m.subscribeGot = input
	return m.subscribeID, m.subscribeErr
}

// captureEmails swaps the sendEmail seam for a recorder and returns a waiter
// that blocks until n sends have happened or the deadline passes.
func captureEmails(t *testing.T, n int) (*[]emailPkg.SendParams, func()) {
	t.Helper()
	var mu sync.Mutex
	var got []emailPkg.SendParams
	done := make(chan struct{}, 16)
	orig := sendEmail
	sendEmail = func(_ context.Context, p emailPkg.SendParams) (string, error) {
		mu.Lock()
		got = append(got, p)
		mu.Unlock()
		done <- struct{}{}
		return "id", nil
	}
	t.Cleanup(func() { sendEmail = orig })
	wait := func() {
		for i := 0; i < n; i++ {
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatalf("expected %d emails, got %d", n, len(got))
			}
		}
	}
	return &got, wait
}

func testNotify() Notify {
	return Notify{
		LeadEmails:       []string{"leads@example.com", "second@example.com"},
		NewsletterEmails: []string{"news@example.com"},
	}
}

func TestSubmit_Delegates(t *testing.T) {
	// A successful submit fires two detached emails (notification and
	// auto-reply). Wait for both so no goroutine outlives the seam swap.
	_, wait := captureEmails(t, 2)
	m := &mockRepo{createID: 7}
	s := &Service{repo: m, notify: testNotify()}
	id, err := s.Submit(context.Background(), domain.LeadInput{Name: "Alice", Email: "alice@example.com"})
	if err != nil || id != 7 {
		t.Errorf("id=%d err=%v", id, err)
	}
	if m.createGot.Name != "Alice" {
		t.Errorf("repo got %+v", m.createGot)
	}
	wait()
}

func TestSubmit_PropagatesError(t *testing.T) {
	captureEmails(t, 0)
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}, notify: testNotify()}
	if _, err := s.Submit(context.Background(), domain.LeadInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestSubmit_NotificationGoesToConfiguredLeadList(t *testing.T) {
	got, wait := captureEmails(t, 2)
	s := &Service{repo: &mockRepo{createID: 42}, notify: testNotify()}
	_, err := s.Submit(context.Background(), domain.LeadInput{
		Name: "Rania", Email: "rania@example.com", Service: "Web Development", Budget: "$8K - $20K",
	})
	if err != nil {
		t.Fatal(err)
	}
	wait()
	var notification *emailPkg.SendParams
	for i := range *got {
		if strings.HasPrefix((*got)[i].Subject, "[Lead #42]") {
			notification = &(*got)[i]
		}
	}
	if notification == nil {
		t.Fatalf("no lead notification among %d emails", len(*got))
	}
	if strings.Join(notification.To, ",") != "leads@example.com,second@example.com" {
		t.Fatalf("notification To = %v", notification.To)
	}
	if notification.ReplyTo != "rania@example.com" {
		t.Fatalf("notification ReplyTo = %q", notification.ReplyTo)
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

func TestSubmit_AutoReplySpeaksForTheTeam(t *testing.T) {
	got, wait := captureEmails(t, 2)
	s := &Service{repo: &mockRepo{createID: 9}, notify: testNotify()}
	if _, err := s.Submit(context.Background(), domain.LeadInput{
		Name: "Rania", Email: "rania@example.com", Phone: "+961 70 123 456",
	}); err != nil {
		t.Fatal(err)
	}
	wait()
	var reply, notification *emailPkg.SendParams
	for i := range *got {
		switch {
		case (*got)[i].Subject == "We received your message - Voxire":
			reply = &(*got)[i]
		case strings.HasPrefix((*got)[i].Subject, "[Lead #9]"):
			notification = &(*got)[i]
		}
	}
	if reply == nil || notification == nil {
		t.Fatalf("expected a reply and a notification, got %d emails", len(*got))
	}
	if reply.From != "Voxire <info@voxire.com>" {
		t.Fatalf("reply From = %q", reply.From)
	}
	if reply.ReplyTo != "info@voxire.com" || strings.Join(reply.To, ",") != "rania@example.com" {
		t.Fatalf("reply routing To=%v ReplyTo=%q", reply.To, reply.ReplyTo)
	}
	for _, banned := range []string{"Abed", "\u2014", "&mdash;"} {
		if strings.Contains(reply.HTML, banned) {
			t.Fatalf("reply still contains %q", banned)
		}
		if strings.Contains(notification.HTML, banned) {
			t.Fatalf("notification still contains %q", banned)
		}
	}
	for _, want := range []string{"Thanks, Rania.", "24 hours", "wa.me/9613940708"} {
		if !strings.Contains(reply.HTML, want) {
			t.Fatalf("reply missing %q", want)
		}
	}
	if !strings.Contains(notification.HTML, "this+is+Voxire+following+up") {
		t.Fatal("notification WhatsApp prefill is not the neutral team wording")
	}
}

func TestSubmit_EscapesLeadFieldsInEmails(t *testing.T) {
	got, wait := captureEmails(t, 2)
	s := &Service{repo: &mockRepo{createID: 13}, notify: testNotify()}
	if _, err := s.Submit(context.Background(), domain.LeadInput{
		Name:    "Rania <b>x</b>",
		Email:   "rania@example.com",
		Company: "A & B",
		Project: "<script>alert(1)</script>",
		Budget:  "$8K - $20K",
		Phone:   "+961 70 123\"><b>x</b>",
	}); err != nil {
		t.Fatal(err)
	}
	wait()
	var reply, notification *emailPkg.SendParams
	for i := range *got {
		switch {
		case (*got)[i].Subject == "We received your message - Voxire":
			reply = &(*got)[i]
		case strings.HasPrefix((*got)[i].Subject, "[Lead #13]"):
			notification = &(*got)[i]
		}
	}
	if reply == nil || notification == nil {
		t.Fatalf("expected a reply and a notification, got %d emails", len(*got))
	}
	for _, banned := range []string{"<script>", "<b>x</b>"} {
		if strings.Contains(reply.HTML, banned) {
			t.Fatalf("reply still contains %q", banned)
		}
		if strings.Contains(notification.HTML, banned) {
			t.Fatalf("notification still contains %q", banned)
		}
	}
	for _, want := range []string{"&lt;script&gt;", "A &amp; B"} {
		if !strings.Contains(notification.HTML, want) {
			t.Fatalf("notification missing %q", want)
		}
	}
	// The WhatsApp quick-action href builds its phone segment from
	// attacker-controlled lead.Phone. It must never carry the raw
	// break-out sequence, and the digits are PathEscape'd before being
	// HTML-escaped, so the payload's quote/angle-bracket/slash characters
	// end up entirely percent-encoded (no raw HTML metacharacter survives
	// in the href), which is what proves the attribute cannot be broken
	// out of.
	if strings.Contains(notification.HTML, `"><b>x</b>`) {
		t.Fatalf("notification WhatsApp link still contains raw break-out sequence")
	}
	if !strings.Contains(notification.HTML, "96170123%22%3E%3Cb%3Ex%3C%2Fb%3E") {
		t.Fatalf("notification missing PathEscape'd WhatsApp phone segment")
	}
}
