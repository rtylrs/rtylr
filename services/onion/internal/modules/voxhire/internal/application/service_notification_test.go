package application

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	emailPkg "github.com/voxire/onion/internal/platform/email"
)

// stubNotificationSend swaps the notification sender for one that waits until
// the request context has been canceled and then reports the state of the
// context it was handed. It returns the functions that release the sender and
// that read the observed context error.
func stubNotificationSend(t *testing.T) (release func(), observed func() error) {
	t.Helper()

	original := sendNotificationEmail
	t.Cleanup(func() { sendNotificationEmail = original })

	releaseCh := make(chan struct{})
	observedCh := make(chan error, 1)
	sendNotificationEmail = func(ctx context.Context, _ emailPkg.SendParams) (string, error) {
		<-releaseCh
		observedCh <- ctx.Err()
		return "test-id", nil
	}

	return func() { close(releaseCh) }, func() error {
		select {
		case err := <-observedCh:
			return err
		case <-time.After(2 * time.Second):
			t.Fatal("notification email was never attempted")
			return nil
		}
	}
}

func TestSubmitCompanyLead_NotificationOutlivesRequestContext(t *testing.T) {
	release, observed := stubNotificationSend(t)

	svc := &Service{repo: &mockRepo{companyID: 42}}
	reqCtx, cancel := context.WithCancel(context.Background())
	if _, err := svc.SubmitCompanyLead(reqCtx, domain.CompanyLeadInput{
		CompanyName: "Acme Corp",
		ContactName: "Ada Lovelace",
		WorkEmail:   "ada@acme.com",
		Role:        "Engineer",
	}); err != nil {
		t.Fatalf("SubmitCompanyLead: %v", err)
	}

	// The handler returns immediately after SubmitCompanyLead; net/http cancels
	// the request context right after that.
	cancel()
	release()

	if err := observed(); err != nil {
		t.Fatalf("notification context was canceled with the request: %v", err)
	}
}

func TestSubmitTalentApplication_NotificationOutlivesRequestContext(t *testing.T) {
	release, observed := stubNotificationSend(t)

	svc := &Service{repo: &mockRepo{talentID: 7}}
	reqCtx, cancel := context.WithCancel(context.Background())
	if _, err := svc.SubmitTalentApplication(reqCtx, domain.TalentApplicationInput{
		FullName: "Grace Hopper",
		Email:    "grace@example.com",
		Track:    "engineering",
		Consent:  true,
	}); err != nil {
		t.Fatalf("SubmitTalentApplication: %v", err)
	}

	cancel()
	release()

	if err := observed(); err != nil {
		t.Fatalf("notification context was canceled with the request: %v", err)
	}
}

// The career-application notification must outlive the request but stay within
// the bounded background timeout instead of hanging on a stalled provider.
func TestSubmitCareerApplication_NotificationContextIsBounded(t *testing.T) {
	original := sendNotificationEmail
	t.Cleanup(func() { sendNotificationEmail = original })

	type observedCall struct {
		err      error
		deadline time.Time
		hasDL    bool
	}
	observedCh := make(chan observedCall, 1)
	sendNotificationEmail = func(ctx context.Context, _ emailPkg.SendParams) (string, error) {
		deadline, hasDL := ctx.Deadline()
		observedCh <- observedCall{err: ctx.Err(), deadline: deadline, hasDL: hasDL}
		return "test-id", nil
	}

	svc := &Service{repo: &mockRepo{}}
	reqCtx, cancel := context.WithCancel(context.Background())
	if _, err := svc.SubmitCareerApplication(reqCtx, domain.CareerApplicationInput{
		JobID: 7,
		Name:  "Ada Lovelace",
		Email: "ada@example.com",
	}); err != nil {
		t.Fatalf("SubmitCareerApplication: %v", err)
	}

	// The handler returns immediately after SubmitCareerApplication; net/http
	// cancels the request context right after that.
	cancel()

	select {
	case call := <-observedCh:
		if call.err != nil {
			t.Fatalf("notification context was canceled with the request: %v", call.err)
		}
		if !call.hasDL {
			t.Fatal("notification context has no deadline")
		}
		if remaining := time.Until(call.deadline); remaining <= 0 || remaining > careerApplicationNotifyTimeout {
			t.Fatalf("notification deadline in %s, want within (0, %s]", remaining, careerApplicationNotifyTimeout)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("notification email was never attempted")
	}
}
