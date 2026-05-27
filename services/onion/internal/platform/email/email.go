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
	html, err := Render("invitation", struct {
		Subject      string
		InviterName  string
		BusinessName string
		InviteLink   string
	}{
		Subject:      fmt.Sprintf("%s invited you to %s on rtylr", inviterName, businessName),
		InviterName:  inviterName,
		BusinessName: businessName,
		InviteLink:   inviteLink,
	})
	if err != nil {
		return "", err
	}
	return Send(ctx, SendParams{
		To:      []string{toEmail},
		Subject: fmt.Sprintf("%s invited you to %s on rtylr", inviterName, businessName),
		HTML:    html,
	})
}

// SendPasswordReset sends a password reset email.
func SendPasswordReset(ctx context.Context, toEmail, resetLink string) (string, error) {
	html, err := Render("password_reset", struct {
		Subject   string
		ResetLink string
	}{Subject: "Reset your rtylr password", ResetLink: resetLink})
	if err != nil {
		return "", err
	}
	return Send(ctx, SendParams{
		To:      []string{toEmail},
		Subject: "Reset your rtylr password",
		HTML:    html,
	})
}

// SendWelcome sends a welcome email directly (use EnqueueWelcome for queue-based sending).
func SendWelcome(ctx context.Context, toEmail, name string) (string, error) {
	html, err := welcomeHTML(name)
	if err != nil {
		return "", err
	}
	return Send(ctx, SendParams{
		To:      []string{toEmail},
		Subject: "Welcome to rtylr",
		HTML:    html,
		From:    "Mohamad at rtylr <mohamad@voxire.com>",
		ReplyTo: "mohamad@voxire.com",
	})
}

// EnqueueWelcome enqueues a welcome email at normal priority.
func EnqueueWelcome(ctx context.Context, db *sql.DB, toEmail, name string) error {
	html, err := welcomeHTML(name)
	if err != nil {
		return err
	}
	return enqueue(ctx, db, enqueueParams{
		priority: priorityNormal,
		to:       toEmail,
		subject:  "Welcome to rtylr",
		html:     html,
		from:     "Mohamad at rtylr <mohamad@voxire.com>",
		replyTo:  "mohamad@voxire.com",
	})
}

// EnqueuePasswordReset enqueues a password-reset email at critical priority.
func EnqueuePasswordReset(ctx context.Context, db *sql.DB, toEmail, resetLink string) error {
	html, err := Render("password_reset", struct {
		Subject   string
		ResetLink string
	}{Subject: "Reset your rtylr password", ResetLink: resetLink})
	if err != nil {
		return err
	}
	return enqueue(ctx, db, enqueueParams{
		priority: priorityCritical,
		to:       toEmail,
		subject:  "Reset your rtylr password",
		html:     html,
	})
}

// EnqueueInvitation enqueues a team invitation email at high priority.
func EnqueueInvitation(ctx context.Context, db *sql.DB, toEmail, inviterName, businessName, inviteLink string) error {
	subject := fmt.Sprintf("%s invited you to %s on rtylr", inviterName, businessName)
	html, err := Render("invitation", struct {
		Subject      string
		InviterName  string
		BusinessName string
		InviteLink   string
	}{Subject: subject, InviterName: inviterName, BusinessName: businessName, InviteLink: inviteLink})
	if err != nil {
		return err
	}
	return enqueue(ctx, db, enqueueParams{
		priority: priorityHigh,
		to:       toEmail,
		subject:  subject,
		html:     html,
	})
}

func welcomeHTML(name string) (string, error) {
	greeting := "there"
	if name != "" {
		greeting = name
	}
	return Render("welcome", struct {
		Subject string
		Name    string
	}{Subject: "Welcome to rtylr", Name: greeting})
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
	subject := fmt.Sprintf("Invoice %s - %s", invoiceNumber, amount)
	html, err := Render("invoice", struct {
		Subject       string
		CustomerName  string
		InvoiceNumber string
		Amount        string
		DueDate       string
		ViewLink      string
	}{
		Subject:       subject,
		CustomerName:  customerName,
		InvoiceNumber: invoiceNumber,
		Amount:        amount,
		DueDate:       dueDate,
		ViewLink:      viewLink,
	})
	if err != nil {
		return "", err
	}
	return Send(ctx, SendParams{
		To:      []string{toEmail},
		Subject: subject,
		HTML:    html,
	})
}
