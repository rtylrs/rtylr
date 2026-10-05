package application

import (
	"context"
	"fmt"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/logger"
)

var voxireNotifyEmails = []string{"info@voxire.com", "amounehabed@gmail.com"}

// SaveApplication stores a Voxire career application and fires a
// notification email to the team. The email is best-effort — a failed
// send never blocks the applicant.
func (s *Service) SaveApplication(ctx context.Context, in domain.ApplicationInput) (int64, error) {
	id, err := s.repo.SaveApplication(ctx, in)
	if err != nil {
		return 0, err
	}

	go s.sendApplicationNotification(ctx, id, in)

	return id, nil
}

func (s *Service) sendApplicationNotification(ctx context.Context, id int64, in domain.ApplicationInput) {
	// The caller spawns this on the request context, which net/http cancels
	// as soon as the handler returns. Detach (values preserved) with a bound
	// so the applicant's request finishing cannot abort the notification.
	ctx, cancel := detachSend(ctx)
	defer cancel()

	if !emailEnabled() {
		return
	}

	subject := fmt.Sprintf("New job application from %s (Job #%d)", in.Name, in.JobID)

	cvLine := "(not provided)"
	if in.CVKey != "" {
		cvLine = in.CVKey
	}

	html := fmt.Sprintf(`
<div style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;max-width:600px;margin:0 auto;padding:40px 20px;background:#f9f9f9;">
  <div style="background:#111118;border-radius:12px;padding:32px;color:#f0f0f6;">
    <h2 style="font-size:20px;font-weight:600;margin:0 0 4px 0;color:#63c6bd;">New job application</h2>
    <p style="font-size:13px;color:rgba(240,240,246,0.55);margin:0 0 28px 0;">Application #%d via voxire.com/careers</p>

    <table style="width:100%%;border-collapse:collapse;">
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;width:140px;">Name</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;font-weight:500;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Email</td><td style="padding:8px 0;font-size:15px;"><a href="mailto:%s" style="color:#63c6bd;text-decoration:none;">%s</a></td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Phone</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Job ID</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%d</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">CV</td><td style="padding:8px 0;font-size:13px;color:rgba(240,240,246,0.7);">%s</td></tr>
    </table>

    %s
    %s

    <a href="mailto:%s?subject=Re%%3A%%20Your%%20Voxire%%20Application" style="display:inline-block;margin-top:24px;background:#63c6bd;color:#111118;padding:12px 24px;border-radius:8px;text-decoration:none;font-size:14px;font-weight:600;">
      Reply to applicant
    </a>
  </div>
</div>`,
		id,
		in.Name,
		in.Email, in.Email,
		in.Phone,
		in.JobID,
		cvLine,
		coverBlock(in.CoverLetter),
		summaryBlock(in.ExperienceSummary),
		in.Email,
	)

	_, emailErr := sendEmail(ctx, emailPkg.SendParams{
		To:      voxireNotifyEmails,
		Subject: subject,
		HTML:    html,
		From:    "Voxire Careers <noreply@voxire.com>",
		ReplyTo: in.Email,
	})
	if emailErr != nil {
		logger.Ctx(ctx).Warn("Failed to send application notification email",
			logger.Error(emailErr),
			logger.Any("application_id", id),
		)
	}
}

func coverBlock(s string) string {
	if s == "" {
		return ""
	}
	return fmt.Sprintf(`
    <div style="margin-top:20px;padding:16px;background:rgba(255,255,255,0.05);border-radius:8px;border-left:3px solid #63c6bd;">
      <p style="font-size:13px;color:rgba(240,240,246,0.55);margin:0 0 6px 0;">Cover letter</p>
      <p style="font-size:14px;color:#f0f0f6;margin:0;line-height:1.6;white-space:pre-wrap;">%s</p>
    </div>`, s)
}

func summaryBlock(s string) string {
	if s == "" {
		return ""
	}
	return fmt.Sprintf(`
    <div style="margin-top:16px;padding:16px;background:rgba(255,255,255,0.05);border-radius:8px;border-left:3px solid #519ad1;">
      <p style="font-size:13px;color:rgba(240,240,246,0.55);margin:0 0 6px 0;">Experience summary</p>
      <p style="font-size:14px;color:#f0f0f6;margin:0;line-height:1.6;white-space:pre-wrap;">%s</p>
    </div>`, s)
}
