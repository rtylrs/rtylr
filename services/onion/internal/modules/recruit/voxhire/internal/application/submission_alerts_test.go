package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
	emailPkg "github.com/voxire/onion/internal/platform/email"
)

// captureNotification returns a sender to inject into a Service and a
// function that waits for the email it sends.
func captureNotification(t *testing.T) (func(context.Context, emailPkg.SendParams) (string, error), func() emailPkg.SendParams) {
	t.Helper()
	sent := make(chan emailPkg.SendParams, 1)
	send := func(_ context.Context, p emailPkg.SendParams) (string, error) {
		select {
		case sent <- p:
		default:
		}
		return "test-id", nil
	}
	return send, func() emailPkg.SendParams {
		select {
		case p := <-sent:
			return p
		case <-time.After(2 * time.Second):
			t.Fatal("no notification email was sent")
			return emailPkg.SendParams{}
		}
	}
}

func assertAlert(t *testing.T, p emailPkg.SendParams, subjectPrefix, bodyFragment string) {
	t.Helper()
	if !strings.HasPrefix(p.Subject, subjectPrefix) {
		t.Fatalf("subject = %q, want prefix %q", p.Subject, subjectPrefix)
	}
	if !strings.Contains(p.HTML, bodyFragment) {
		t.Fatalf("email body does not contain %q", bodyFragment)
	}
}

func TestSubmitTalentApplication_SaveFailure_StillAlertsTeam(t *testing.T) {
	send, next := captureNotification(t)
	dbErr := errors.New("Error 1406: Data too long for column 'track'")
	svc := &Service{repo: &mockRepo{talentErr: dbErr}, sendEmail: send}

	_, err := svc.SubmitTalentApplication(context.Background(), domain.TalentApplicationInput{
		FullName: "Grace Hopper", Email: "talent-notsaved@example.com", CVKey: "voxhire/applications/x/cv.pdf",
	})
	if !errors.Is(err, dbErr) {
		t.Fatalf("err = %v, want the repository error", err)
	}

	p := next()
	assertAlert(t, p, noticeNotSaved.subjectPrefix(), "NOT saved")
	if !strings.Contains(p.HTML, "talent-notsaved@example.com") || !strings.Contains(p.HTML, "voxhire/applications/x/cv.pdf") {
		t.Fatal("not-saved alert must carry the applicant contact and CV key so they can be recovered")
	}
}

func TestSubmitTalentApplication_CVUploadFailed_AlertsTeam(t *testing.T) {
	send, next := captureNotification(t)
	svc := &Service{repo: &mockRepo{talentID: 12}, sendEmail: send}

	if _, err := svc.SubmitTalentApplication(context.Background(), domain.TalentApplicationInput{
		FullName: "Grace Hopper", Email: "talent-cvfail@example.com", CVUploadFailed: true,
	}); err != nil {
		t.Fatalf("SubmitTalentApplication: %v", err)
	}

	assertAlert(t, next(), noticeCVUploadFailed.subjectPrefix(), "Upload FAILED")
}

func TestSubmitTalentApplication_Clean_NoAlertPrefix(t *testing.T) {
	send, next := captureNotification(t)
	svc := &Service{repo: &mockRepo{talentID: 13}, sendEmail: send}

	if _, err := svc.SubmitTalentApplication(context.Background(), domain.TalentApplicationInput{
		FullName: "Grace Hopper", Email: "talent-clean@example.com", CVKey: "k.pdf",
	}); err != nil {
		t.Fatalf("SubmitTalentApplication: %v", err)
	}

	if p := next(); strings.HasPrefix(p.Subject, "[ACTION NEEDED") {
		t.Fatalf("clean submission flagged: %q", p.Subject)
	}
}

func TestSubmitCareerApplication_SaveFailure_StillAlertsTeam(t *testing.T) {
	send, next := captureNotification(t)
	dbErr := errors.New("db down")
	svc := &Service{repo: &mockRepo{careerAppErr: dbErr}, sendEmail: send}

	_, err := svc.SubmitCareerApplication(context.Background(), domain.CareerApplicationInput{
		JobID: 3, Name: "Khadija", Email: "career-notsaved@example.com", CVKey: "voxire/applications/3/cv.pdf",
	})
	if !errors.Is(err, dbErr) {
		t.Fatalf("err = %v, want the repository error", err)
	}

	p := next()
	assertAlert(t, p, noticeNotSaved.subjectPrefix(), "NOT saved")
	if !strings.Contains(p.HTML, "voxire/applications/3/cv.pdf") {
		t.Fatal("not-saved alert must carry the CV key")
	}
}

func TestSubmitCareerApplication_CVUploadFailed_AlertsTeam(t *testing.T) {
	send, next := captureNotification(t)
	svc := &Service{repo: &mockRepo{careerAppID: 21}, sendEmail: send}

	if _, err := svc.SubmitCareerApplication(context.Background(), domain.CareerApplicationInput{
		JobID: 3, Name: "Khadija", Email: "career-cvfail@example.com", CVUploadFailed: true,
	}); err != nil {
		t.Fatalf("SubmitCareerApplication: %v", err)
	}

	assertAlert(t, next(), noticeCVUploadFailed.subjectPrefix(), "Upload FAILED")
}
