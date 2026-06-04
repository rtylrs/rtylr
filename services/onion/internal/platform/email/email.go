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
	sqlc "github.com/voxire/onion/internal/platform/email/internal/infrastructure/persistence/sqlc"
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

// Option customizes the optional fields of an outgoing email (From, Reply-To).
type Option func(*SendParams)

// WithFrom overrides the default From address.
func WithFrom(from string) Option { return func(p *SendParams) { p.From = from } }

// WithReplyTo sets the Reply-To address.
func WithReplyTo(replyTo string) Option { return func(p *SendParams) { p.ReplyTo = replyTo } }

// SendTemplate renders tmpl with data and sends it immediately via Resend. It is
// the single entry point for templated email: pass a Template constant from
// templates.go and its matching data struct (e.g. TemplateWelcome + WelcomeData).
func SendTemplate(ctx context.Context, to []string, subject string, tmpl Template, data any, opts ...Option) (string, error) {
	html, err := Render(tmpl, data)
	if err != nil {
		return "", err
	}
	params := SendParams{To: to, Subject: subject, HTML: html}
	for _, opt := range opts {
		opt(&params)
	}
	return Send(ctx, params)
}

// EnqueueTemplate renders tmpl with data and enqueues it on email_jobs at the
// given priority (PriorityCritical/High/Normal) for the background worker to
// deliver. Use it instead of SendTemplate when delivery can be deferred/retried.
func EnqueueTemplate(ctx context.Context, db *sql.DB, priority int, to, subject string, tmpl Template, data any, opts ...Option) error {
	html, err := Render(tmpl, data)
	if err != nil {
		return err
	}
	params := SendParams{Subject: subject, HTML: html}
	for _, opt := range opts {
		opt(&params)
	}
	return enqueue(ctx, db, enqueueParams{
		priority: priority,
		to:       to,
		subject:  subject,
		html:     html,
		from:     params.From,
		replyTo:  params.ReplyTo,
	})
}

// Email job priorities for EnqueueTemplate; higher is delivered first.
const (
	PriorityCritical = 10
	PriorityHigh     = 5
	PriorityNormal   = 1
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
	return sqlc.New(db).EnqueueEmail(ctx, sqlc.EnqueueEmailParams{
		Priority:  int8(p.priority),
		ToEmail:   p.to,
		Subject:   p.subject,
		Html:      p.html,
		FromEmail: sql.NullString{String: p.from, Valid: p.from != ""},
		ReplyTo:   sql.NullString{String: p.replyTo, Valid: p.replyTo != ""},
	})
}
