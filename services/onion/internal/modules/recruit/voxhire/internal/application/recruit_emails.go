package application

import (
	"context"
	"fmt"
	"html"
	"net/mail"
	"strings"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/logger"
)

// crlfStripper removes CR/LF so untrusted applicant fields can't inject
// extra email headers via the Subject or Reply-To.
var crlfStripper = strings.NewReplacer("\r", "", "\n", "")

// safeReplyTo returns the email only if it parses as a single valid
// address; otherwise empty so no header is set from untrusted input.
func safeReplyTo(email string) string {
	addr, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil {
		return ""
	}
	return addr.Address
}

// notifyCareerApplication emails the team when a candidate applies to a
// specific opening through hire.voxire.com.
func (s *Service) notifyCareerApplication(ctx context.Context, id int64, app domain.CareerApplicationInput) {
	salaryLine := "-"
	if app.SalaryExpectation > 0 {
		salaryLine = fmt.Sprintf("$%.0f", app.SalaryExpectation)
	}

	emailHTML := fmt.Sprintf(`
<div style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;max-width:600px;margin:0 auto;padding:40px 20px;background:#f9f9f9;">
  <div style="background:#111118;border-radius:12px;padding:32px;color:#f0f0f6;">
    <h2 style="font-size:20px;font-weight:600;margin:0 0 4px 0;color:#63c6bd;">New application</h2>
    <p style="font-size:13px;color:rgba(240,240,246,0.55);margin:0 0 28px 0;">Application #%d via hire.voxire.com (job #%d)</p>

    <table style="width:100%%;border-collapse:collapse;">
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;width:160px;">Name</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;font-weight:500;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Email</td><td style="padding:8px 0;font-size:15px;"><a href="mailto:%s" style="color:#63c6bd;text-decoration:none;">%s</a></td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Phone</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Location</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Salary expectation</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
    </table>
  </div>
</div>`,
		id,
		app.JobID,
		html.EscapeString(app.Name),
		html.EscapeString(app.Email), html.EscapeString(app.Email),
		html.EscapeString(orDash(app.Phone)),
		html.EscapeString(orDash(app.Location)),
		salaryLine,
	)

	subject := fmt.Sprintf("[VoxHire] New application from %s", crlfStripper.Replace(app.Name))

	if _, err := sendNotificationEmail(ctx, emailPkg.SendParams{
		To:      notifyEmails,
		Subject: subject,
		HTML:    emailHTML,
		From:    "VoxHire Careers <noreply@voxire.com>",
		ReplyTo: safeReplyTo(app.Email),
	}); err != nil {
		logger.Ctx(ctx).Warn("Failed to send VoxHire career application notification email",
			logger.Error(err),
			logger.Any("application_id", id),
		)
	}
}
