package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/careers/internal/domain"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/logger"
)

// asyncSendTimeout bounds an email send that was detached from the request
// that triggered it, so a stuck provider call cannot leak a goroutine.
const asyncSendTimeout = 10 * time.Second

// Seams for tests: the provider client is package-private, so tests swap
// these to observe the context an async send actually runs with.
var (
	sendEmail    = emailPkg.Send
	emailEnabled = emailPkg.IsEnabled
)

// detachSend returns a context for an asynchronous send. The caller's context
// is the HTTP request's, and net/http cancels it as soon as the handler
// returns — the goroutine would otherwise observe "context canceled" on both
// the outbound rate-limit wait and the provider call. Values (logger, auth)
// are preserved; only cancellation is dropped, and the send stays bounded.
func detachSend(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), asyncSendTimeout)
}

// Emails for the recruit product. Two flows:
//
//   - Candidate confirmation → sent immediately on POST /recruit/public/apply,
//     reassuring the candidate their application was received and telling
//     them what happens next.
//   - HR new-submission ping → sent to the business's notification_email,
//     linking directly to the admin submissions inbox for review.
//
// Both are best-effort: if Resend is misconfigured or returns an error,
// the request completes successfully and the failure is logged. We
// never want a flaky email provider to block a candidate's apply.

// SendSubmissionEmails fires both emails in parallel (via goroutines)
// so the HTTP handler doesn't wait on SMTP to return. Each send detaches
// from the request context (detachSend) so the response finishing doesn't
// cancel it. Errors are logged but swallowed — the submission is already
// saved to the DB.
func (s *Service) SendSubmissionEmails(
	ctx context.Context,
	submission domain.Submission,
	career domain.RecruitCareer,
	custom domain.Customization,
	candidateName, candidateEmail string,
	adminURL string,
) {
	go s.sendCandidateConfirmation(ctx, custom, career, candidateName, candidateEmail)
	if custom.EmailNotifications && custom.NotificationEmail != "" {
		go s.sendHRNewSubmission(ctx, custom, career, submission, candidateName, candidateEmail, adminURL)
	}
}

func (s *Service) sendCandidateConfirmation(
	ctx context.Context,
	custom domain.Customization,
	career domain.RecruitCareer,
	candidateName, candidateEmail string,
) {
	ctx, cancel := detachSend(ctx)
	defer cancel()

	if !emailEnabled() || candidateEmail == "" {
		return
	}

	subject := fmt.Sprintf("We got your application — %s at %s", career.Title, custom.BrandName)
	html := candidateConfirmationHTML(custom, career, candidateName)
	text := candidateConfirmationText(custom, career, candidateName)

	// ReplyTo is the business's notification_email so a candidate who
	// hits reply reaches HR, not noreply@.
	replyTo := custom.NotificationEmail

	if _, err := sendEmail(ctx, emailPkg.SendParams{
		To:      []string{candidateEmail},
		Subject: subject,
		HTML:    html,
		Text:    text,
		ReplyTo: replyTo,
	}); err != nil {
		logger.Ctx(ctx).Error("recruit: candidate confirmation email failed",
			logger.Any("owner_id", custom.OwnerID),
			logger.Any("career_id", career.ID),
			logger.Error(err),
		)
	}
}

func (s *Service) sendHRNewSubmission(
	ctx context.Context,
	custom domain.Customization,
	career domain.RecruitCareer,
	submission domain.Submission,
	candidateName, candidateEmail, adminURL string,
) {
	ctx, cancel := detachSend(ctx)
	defer cancel()

	if !emailEnabled() {
		return
	}

	subject := fmt.Sprintf("New application for %s — %s", career.Title, candidateName)
	html := hrNewSubmissionHTML(custom, career, submission, candidateName, candidateEmail, adminURL)
	text := hrNewSubmissionText(custom, career, submission, candidateName, candidateEmail, adminURL)

	// ReplyTo the candidate — HR hits reply and reaches them directly,
	// no copy-paste of the email address.
	if _, err := sendEmail(ctx, emailPkg.SendParams{
		To:      []string{custom.NotificationEmail},
		Subject: subject,
		HTML:    html,
		Text:    text,
		ReplyTo: candidateEmail,
	}); err != nil {
		logger.Ctx(ctx).Error("recruit: HR new-submission email failed",
			logger.Any("owner_id", custom.OwnerID),
			logger.Any("career_id", career.ID),
			logger.Error(err),
		)
	}
}

// ── templates ──────────────────────────────────────────────────────
//
// Inline HTML keeps mail-client rendering predictable (no external
// stylesheets, no remote images the client will strip). Apple-system
// font stack matches the pattern in platform/email/email.go.

func candidateConfirmationHTML(c domain.Customization, career domain.RecruitCareer, name string) string {
	greeting := "Hi"
	if name != "" {
		greeting = "Hi " + firstWord(name)
	}
	return fmt.Sprintf(`
<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 520px; margin: 0 auto; padding: 40px 24px; color: #18181b;">
  <h2 style="font-size: 22px; margin: 0 0 16px;">Application received</h2>
  <p style="font-size: 15px; line-height: 1.6;">%s, thanks for applying to
    <strong>%s</strong> at <strong>%s</strong>.</p>
  <p style="font-size: 15px; line-height: 1.6;">We've logged your application and the
    hiring team will review it. If there's a fit, you'll hear back within a few
    business days — reply to this email with questions.</p>
  <p style="font-size: 13px; color: #6b7280; margin-top: 32px;">Sent on behalf of %s.</p>
</div>`, escape(greeting), escape(career.Title), escape(c.BrandName), escape(c.BrandName))
}

func candidateConfirmationText(c domain.Customization, career domain.RecruitCareer, name string) string {
	greeting := "Hi there"
	if name != "" {
		greeting = "Hi " + firstWord(name)
	}
	return fmt.Sprintf(
		"%s,\n\nThanks for applying to %s at %s. We've logged your application and the hiring team will review it. If there's a fit, you'll hear back within a few business days.\n\n— %s",
		greeting, career.Title, c.BrandName, c.BrandName,
	)
}

func hrNewSubmissionHTML(
	c domain.Customization, career domain.RecruitCareer, s domain.Submission,
	name, email, adminURL string,
) string {
	phoneRow := ""
	if s.Phone != "" {
		phoneRow = fmt.Sprintf(`<tr><td style="color:#6b7280;padding:4px 12px 4px 0;">Phone</td><td>%s</td></tr>`, escape(s.Phone))
	}
	resumeRow := ""
	if s.ResumeURL != "" {
		resumeRow = `<tr><td style="color:#6b7280;padding:4px 12px 4px 0;">Resume</td><td>attached / linked (see admin)</td></tr>`
	}
	coverBlock := ""
	if s.CoverLetter != "" {
		coverBlock = fmt.Sprintf(`
  <div style="margin-top: 24px; padding: 16px; background: #f5f5f4; border-radius: 8px;">
    <div style="font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 6px;">Note from the candidate</div>
    <div style="font-size: 14px; line-height: 1.6; white-space: pre-wrap;">%s</div>
  </div>`, escape(s.CoverLetter))
	}

	return fmt.Sprintf(`
<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 560px; margin: 0 auto; padding: 40px 24px; color: #18181b;">
  <div style="font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.06em;">New application · %s</div>
  <h2 style="font-size: 22px; margin: 8px 0 4px;">%s</h2>
  <p style="font-size: 14px; color: #6b7280; margin: 0 0 20px;">applied to <strong style="color:#18181b">%s</strong></p>

  <table style="font-size: 14px; border-collapse: collapse;">
    <tr><td style="color:#6b7280;padding:4px 12px 4px 0;">Email</td><td><a href="mailto:%s" style="color:#18181b;">%s</a></td></tr>
    %s
    %s
  </table>
  %s

  <div style="margin-top: 32px;">
    <a href="%s" style="display:inline-block;padding:10px 18px;background:#18181b;color:white;
                         text-decoration:none;border-radius:8px;font-size:14px;font-weight:600;">
      Review in dashboard
    </a>
  </div>

  <p style="font-size: 12px; color: #9ca3af; margin-top: 32px;">
    Reply to this email to reach the candidate directly.
  </p>
</div>`,
		escape(c.BrandName),
		escape(name),
		escape(career.Title),
		escape(email), escape(email),
		phoneRow, resumeRow,
		coverBlock,
		escape(adminURL),
	)
}

func hrNewSubmissionText(
	c domain.Customization, career domain.RecruitCareer, s domain.Submission,
	name, email, adminURL string,
) string {
	phone := ""
	if s.Phone != "" {
		phone = "\nPhone: " + s.Phone
	}
	cover := ""
	if s.CoverLetter != "" {
		cover = "\n\nNote from the candidate:\n" + s.CoverLetter
	}
	return fmt.Sprintf(
		"New application at %s\n\nCandidate: %s\nEmail: %s%s\nRole: %s%s\n\nReview in dashboard: %s\n",
		c.BrandName, name, email, phone, career.Title, cover, adminURL,
	)
}

// firstWord returns the first whitespace-separated token. Used so
// "Hi Jamie" reads naturally even when the candidate typed their full
// legal name.
func firstWord(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\t' {
			return s[:i]
		}
	}
	return s
}

// escape is a minimal HTML escaper for user-controlled fields in the
// inline templates above. Enough for our attack surface (emails sent
// only to authenticated-to-their-own-data recipients).
func escape(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '<':
			out = append(out, "&lt;"...)
		case '>':
			out = append(out, "&gt;"...)
		case '&':
			out = append(out, "&amp;"...)
		case '"':
			out = append(out, "&quot;"...)
		case '\'':
			out = append(out, "&#39;"...)
		default:
			out = append(out, s[i])
		}
	}
	return string(out)
}
