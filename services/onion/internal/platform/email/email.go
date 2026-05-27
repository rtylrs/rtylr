package email

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/resend/resend-go/v2"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

var client *resend.Client

func Initialize() {
	var apiKey string
	if config.GlobalConfig != nil {
		apiKey = config.GlobalConfig.ResendKey
	} else {
		apiKey = os.Getenv("RESEND_API_KEY")
	}
	if apiKey == "" {
		logger.Global().Warn("RESEND_API_KEY not set, email sending will be disabled")
		return
	}

	client = resend.NewClient(apiKey)
	logger.Global().Info("Resend email client initialized")
}

func IsEnabled() bool {
	return client != nil
}

type SendParams struct {
	To      []string
	Subject string
	HTML    string
	Text    string
	From    string // optional, defaults to noreply@voxire.com
	ReplyTo string // optional
}

func Send(ctx context.Context, params SendParams) (string, error) {
	if client == nil {
		logger.Ctx(ctx).Warn("Email sending disabled (no RESEND_API_KEY)")
		return "", errors.New("email sending disabled")
	}

	from := params.From
	if from == "" {
		from = "rtylr <noreply@voxire.com>"
	}

	req := &resend.SendEmailRequest{
		From:    from,
		To:      params.To,
		Subject: params.Subject,
		Html:    params.HTML,
		Text:    params.Text,
	}

	if params.ReplyTo != "" {
		req.ReplyTo = params.ReplyTo
	}

	sent, err := client.Emails.Send(req)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to send email via Resend",
			logger.Error(err),
			logger.String("to", strings.Join(params.To, ",")),
			logger.String("subject", params.Subject),
		)
		return "", fmt.Errorf("failed to send email: %w", err)
	}

	logger.Ctx(ctx).Info("Email sent via Resend",
		logger.String("id", sent.Id),
		logger.String("to", strings.Join(params.To, ",")),
		logger.String("subject", params.Subject),
	)

	return sent.Id, nil
}

// SendInvitation sends a team invitation email.
func SendInvitation(ctx context.Context, toEmail, inviterName, businessName, inviteLink string) (string, error) {
	html := fmt.Sprintf(`
		<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 480px; margin: 0 auto; padding: 40px 20px;">
			<h2 style="font-size: 20px; font-weight: 600; color: #111; margin: 0 0 8px 0;">You've been invited to rtylr</h2>
			<p style="font-size: 14px; color: #666; line-height: 1.6; margin: 0 0 24px 0;">
				%s has invited you to join <strong>%s</strong> on rtylr.
			</p>
			<a href="%s" style="display: inline-block; background: #4583C4; color: #fff; padding: 12px 24px; border-radius: 8px; text-decoration: none; font-size: 14px; font-weight: 600;">
				Accept invitation
			</a>
			<p style="font-size: 12px; color: #999; margin-top: 24px; line-height: 1.5;">
				If you didn't expect this invitation, you can ignore this email.
			</p>
		</div>
	`, inviterName, businessName, inviteLink)

	return Send(ctx, SendParams{
		To:      []string{toEmail},
		Subject: fmt.Sprintf("%s invited you to %s on rtylr", inviterName, businessName),
		HTML:    html,
	})
}

// SendPasswordReset sends a password reset email.
func SendPasswordReset(ctx context.Context, toEmail, resetLink string) (string, error) {
	html := fmt.Sprintf(`
		<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 480px; margin: 0 auto; padding: 40px 20px;">
			<h2 style="font-size: 20px; font-weight: 600; color: #111; margin: 0 0 8px 0;">Reset your password</h2>
			<p style="font-size: 14px; color: #666; line-height: 1.6; margin: 0 0 24px 0;">
				We received a request to reset your password. Click the button below to choose a new one.
			</p>
			<a href="%s" style="display: inline-block; background: #4583C4; color: #fff; padding: 12px 24px; border-radius: 8px; text-decoration: none; font-size: 14px; font-weight: 600;">
				Reset password
			</a>
			<p style="font-size: 12px; color: #999; margin-top: 24px; line-height: 1.5;">
				This link expires in 1 hour. If you didn't request this, you can ignore this email.
			</p>
		</div>
	`, resetLink)

	return Send(ctx, SendParams{
		To:      []string{toEmail},
		Subject: "Reset your rtylr password",
		HTML:    html,
	})
}

// WelcomeHTML returns the HTML body for a welcome email.
func WelcomeHTML(name string) string {
	greeting := "there"
	if name != "" {
		greeting = name
	}
	return fmt.Sprintf(`
		<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 520px; margin: 0 auto; padding: 40px 20px; color: #111;">
			<h2 style="font-size: 22px; font-weight: 700; margin: 0 0 16px 0;">Hey %s, welcome to rtylr.</h2>
			<p style="font-size: 15px; color: #444; line-height: 1.7; margin: 0 0 12px 0;">
				I'm Mohamad, founder of rtylr. Really glad you're here.
			</p>
			<p style="font-size: 15px; color: #444; line-height: 1.7; margin: 0 0 12px 0;">
				rtylr puts your whole business in one place — POS, HR, CRM, Finance, ERP, and Flow. No juggling five tools. No copy-pasting between apps.
			</p>
			<p style="font-size: 15px; color: #444; line-height: 1.7; margin: 0 0 24px 0;">
				A few things worth knowing as you get started:
			</p>
			<ul style="font-size: 15px; color: #444; line-height: 1.9; margin: 0 0 24px 0; padding-left: 20px;">
				<li>Set up your business profile first — everything else pulls from it</li>
				<li>Invite your team from the HR or Team settings</li>
				<li>The POS works on any device, no app install needed</li>
				<li>If you get stuck, just reply to this email — goes straight to me</li>
			</ul>
			<a href="https://hub.rtylr.com" style="display: inline-block; background: #4583C4; color: #fff; padding: 13px 28px; border-radius: 8px; text-decoration: none; font-size: 15px; font-weight: 600;">
				Get started
			</a>
			<p style="font-size: 13px; color: #999; margin-top: 32px; line-height: 1.6;">
				Mohamad<br>
				Founder, rtylr
			</p>
		</div>
	`, greeting)
}

// WelcomeSubject returns the subject line for a welcome email.
func WelcomeSubject() string { return "Welcome to rtylr" }

// SendWelcome sends a welcome email directly (use EnqueueWelcome for queue-based sending).
func SendWelcome(ctx context.Context, toEmail, name string) (string, error) {
	return Send(ctx, SendParams{
		To:      []string{toEmail},
		Subject: WelcomeSubject(),
		HTML:    WelcomeHTML(name),
		From:    "Mohamad at rtylr <mohamad@voxire.com>",
		ReplyTo: "mohamad@voxire.com",
	})
}

// EnqueueWelcome enqueues a welcome email at normal priority.
func EnqueueWelcome(ctx context.Context, db *sql.DB, toEmail, name string) error {
	return enqueue(ctx, db, enqueueParams{
		priority: priorityNormal,
		to:       toEmail,
		subject:  WelcomeSubject(),
		html:     WelcomeHTML(name),
		from:     "Mohamad at rtylr <mohamad@voxire.com>",
		replyTo:  "mohamad@voxire.com",
	})
}

// EnqueuePasswordReset enqueues a password-reset email at critical priority.
func EnqueuePasswordReset(ctx context.Context, db *sql.DB, toEmail, resetLink string) error {
	html := fmt.Sprintf(`
		<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 480px; margin: 0 auto; padding: 40px 20px;">
			<h2 style="font-size: 20px; font-weight: 600; color: #111; margin: 0 0 8px 0;">Reset your password</h2>
			<p style="font-size: 14px; color: #666; line-height: 1.6; margin: 0 0 24px 0;">
				We received a request to reset your password. Click the button below to choose a new one.
			</p>
			<a href="%s" style="display: inline-block; background: #4583C4; color: #fff; padding: 12px 24px; border-radius: 8px; text-decoration: none; font-size: 14px; font-weight: 600;">
				Reset password
			</a>
			<p style="font-size: 12px; color: #999; margin-top: 24px; line-height: 1.5;">
				This link expires in 1 hour. If you didn't request this, you can ignore this email.
			</p>
		</div>
	`, resetLink)
	return enqueue(ctx, db, enqueueParams{
		priority: priorityCritical,
		to:       toEmail,
		subject:  "Reset your rtylr password",
		html:     html,
	})
}

// EnqueueInvitation enqueues a team invitation email at high priority.
func EnqueueInvitation(ctx context.Context, db *sql.DB, toEmail, inviterName, businessName, inviteLink string) error {
	html := fmt.Sprintf(`
		<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 480px; margin: 0 auto; padding: 40px 20px;">
			<h2 style="font-size: 20px; font-weight: 600; color: #111; margin: 0 0 8px 0;">You've been invited to rtylr</h2>
			<p style="font-size: 14px; color: #666; line-height: 1.6; margin: 0 0 24px 0;">
				%s has invited you to join <strong>%s</strong> on rtylr.
			</p>
			<a href="%s" style="display: inline-block; background: #4583C4; color: #fff; padding: 12px 24px; border-radius: 8px; text-decoration: none; font-size: 14px; font-weight: 600;">
				Accept invitation
			</a>
			<p style="font-size: 12px; color: #999; margin-top: 24px; line-height: 1.5;">
				If you didn't expect this invitation, you can ignore this email.
			</p>
		</div>
	`, inviterName, businessName, inviteLink)
	return enqueue(ctx, db, enqueueParams{
		priority: priorityHigh,
		to:       toEmail,
		subject:  fmt.Sprintf("%s invited you to %s on rtylr", inviterName, businessName),
		html:     html,
	})
}

const (
	priorityCritical = 10
	priorityHigh     = 5
	priorityNormal   = 1
)

type enqueueParams struct {
	priority int
	to       string
	subject  string
	html     string
	from     string
	replyTo  string
}

func enqueue(ctx context.Context, db *sql.DB, p enqueueParams) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO email_jobs (priority, to_email, subject, html, from_email, reply_to)
		VALUES (?, ?, ?, ?, ?, ?)`,
		p.priority, p.to, p.subject, p.html, p.from, p.replyTo,
	)
	return err
}

// SendInvoice sends an invoice email.
func SendInvoice(ctx context.Context, toEmail, customerName, invoiceNumber, amount, dueDate, viewLink string) (string, error) {
	html := fmt.Sprintf(`
		<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 480px; margin: 0 auto; padding: 40px 20px;">
			<h2 style="font-size: 20px; font-weight: 600; color: #111; margin: 0 0 8px 0;">Invoice %s</h2>
			<p style="font-size: 14px; color: #666; line-height: 1.6; margin: 0 0 4px 0;">
				Hi %s, here's your invoice for <strong>%s</strong>.
			</p>
			<p style="font-size: 14px; color: #666; line-height: 1.6; margin: 0 0 24px 0;">
				Due by %s.
			</p>
			<a href="%s" style="display: inline-block; background: #4583C4; color: #fff; padding: 12px 24px; border-radius: 8px; text-decoration: none; font-size: 14px; font-weight: 600;">
				View invoice
			</a>
		</div>
	`, invoiceNumber, customerName, amount, dueDate, viewLink)

	return Send(ctx, SendParams{
		To:      []string{toEmail},
		Subject: fmt.Sprintf("Invoice %s - %s", invoiceNumber, amount),
		HTML:    html,
	})
}
