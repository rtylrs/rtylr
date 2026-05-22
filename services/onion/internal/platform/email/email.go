package email

import (
	"context"
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

// SendWelcome sends a welcome email after registration.
func SendWelcome(ctx context.Context, toEmail, name string) (string, error) {
	greeting := "there"
	if name != "" {
		greeting = name
	}

	html := fmt.Sprintf(`
		<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 480px; margin: 0 auto; padding: 40px 20px;">
			<h2 style="font-size: 20px; font-weight: 600; color: #111; margin: 0 0 8px 0;">Welcome to rtylr</h2>
			<p style="font-size: 14px; color: #666; line-height: 1.6; margin: 0 0 24px 0;">
				Hey %s, your account is ready. rtylr gives you POS, ERP, HR, CRM, Finance, and Flow in one platform.
			</p>
			<a href="https://hub.rtylr.com" style="display: inline-block; background: #4583C4; color: #fff; padding: 12px 24px; border-radius: 8px; text-decoration: none; font-size: 14px; font-weight: 600;">
				Go to Hub
			</a>
		</div>
	`, greeting)

	return Send(ctx, SendParams{
		To:      []string{toEmail},
		Subject: "Welcome to rtylr",
		HTML:    html,
	})
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
