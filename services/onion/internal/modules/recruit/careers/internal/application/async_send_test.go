package application

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
	emailPkg "github.com/voxire/onion/internal/platform/email"
)

// sendObservation records the state of the context at the moment a send
// actually runs. It has to be captured inside the hook: the send functions
// cancel their detached context on return, so asserting afterwards would
// report "canceled" even when the send ran on a live context.
type sendObservation struct {
	err         error
	hasDeadline bool
}

// capturedSends swaps the send/IsEnabled seams for the duration of a test and
// returns a buffered channel of observations. The provider client is
// package-private, so without these seams a regression test could not tell a
// detached send from one wired to the canceled request context.
func capturedSends(t *testing.T) <-chan sendObservation {
	t.Helper()

	sent := make(chan sendObservation, 4)
	previousSend, previousEnabled := sendEmail, emailEnabled
	sendEmail = func(ctx context.Context, _ emailPkg.SendParams) (string, error) {
		_, hasDeadline := ctx.Deadline()
		sent <- sendObservation{err: ctx.Err(), hasDeadline: hasDeadline}
		return "test-email-id", nil
	}
	emailEnabled = func() bool { return true }
	t.Cleanup(func() {
		sendEmail = previousSend
		emailEnabled = previousEnabled
	})
	return sent
}

// canceledRequestContext models the context a handler passes around: net/http
// cancels r.Context() as soon as ServeHTTP returns, which for a fire-and-forget
// goroutine can be before the send even starts.
func canceledRequestContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func receiveSend(t *testing.T, sent <-chan sendObservation) {
	t.Helper()

	select {
	case observation := <-sent:
		if observation.err != nil {
			t.Fatalf("email send ran on a canceled context: %v", observation.err)
		}
		if !observation.hasDeadline {
			t.Error("detached send context has no deadline; a stuck provider would leak the goroutine")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("email send never ran")
	}
}

func TestSendSubmissionEmailsOutlivesCanceledRequest(t *testing.T) {
	sent := capturedSends(t)

	service := &Service{}
	custom := domain.Customization{
		OwnerID:            101,
		BusinessID:         202,
		BrandName:          "Acme",
		EmailNotifications: true,
		NotificationEmail:  "hr@acme.test",
	}
	career := domain.RecruitCareer{ID: 303, Title: "Backend Engineer"}

	service.SendSubmissionEmails(
		canceledRequestContext(),
		domain.Submission{ID: 404},
		career,
		custom,
		"Ada Lovelace",
		"ada@example.com",
		"https://recruit.example.test/admin/submissions",
	)

	receiveSend(t, sent) // candidate confirmation
	receiveSend(t, sent) // HR notification
}

func TestSendApplicationNotificationOutlivesCanceledRequest(t *testing.T) {
	sent := capturedSends(t)

	(&Service{}).sendApplicationNotification(canceledRequestContext(), 7, domain.ApplicationInput{
		JobID: 3,
		Name:  "Ada Lovelace",
		Email: "ada@example.com",
	})

	receiveSend(t, sent)
}

func TestDetachSendKeepsContextValues(t *testing.T) {
	type contextKey struct{}
	parent, cancel := context.WithCancel(context.WithValue(context.Background(), contextKey{}, "kept"))
	cancel()

	detached, stop := detachSend(parent)
	defer stop()

	if err := detached.Err(); err != nil {
		t.Fatalf("detached context canceled with the parent: %v", err)
	}
	if got := detached.Value(contextKey{}); got != "kept" {
		t.Fatalf("detached context value = %v, want kept", got)
	}
}
