package email

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	stdhtml "html"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/resend/resend-go/v2"

	"github.com/voxire/onion/internal/platform/config"
	sqlc "github.com/voxire/onion/internal/platform/email/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

var client *resend.Client

var outboundLimiter struct {
	sync.Mutex
	next time.Time
}

const minOutboundInterval = 100 * time.Millisecond

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

// RateLimitError preserves the provider's retry advice for queue workers.
type RateLimitError struct {
	RetryAfter time.Duration
	Cause      error
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("email provider rate limited the request; retry after %s: %v", e.RetryAfter, e.Cause)
}

func (e *RateLimitError) Unwrap() error { return e.Cause }

func Send(ctx context.Context, params SendParams) (string, error) {
	if client == nil {
		logger.Ctx(ctx).Warn("Email sending disabled (no RESEND_API_KEY)")
		return "", errors.New("email sending disabled")
	}
	if err := waitForOutboundSlot(ctx); err != nil {
		return "", err
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

	sent, err := client.Emails.SendWithContext(ctx, req)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to send email via Resend",
			logger.Error(err),
			logger.String("to", strings.Join(params.To, ",")),
			logger.String("subject", params.Subject),
		)
		var providerRateLimit *resend.RateLimitError
		if errors.As(err, &providerRateLimit) {
			delaySeconds := 60
			if parsed, parseErr := strconv.Atoi(providerRateLimit.RetryAfter); parseErr == nil && parsed > 0 {
				delaySeconds = parsed
			}
			return "", &RateLimitError{
				RetryAfter: time.Duration(delaySeconds) * time.Second,
				Cause:      err,
			}
		}
		return "", fmt.Errorf("failed to send email: %w", err)
	}

	logger.Ctx(ctx).Info("Email sent via Resend",
		logger.String("id", sent.Id),
		logger.String("to", strings.Join(params.To, ",")),
		logger.String("subject", params.Subject),
	)

	return sent.Id, nil
}

func waitForOutboundSlot(ctx context.Context) error {
	outboundLimiter.Lock()
	now := time.Now()
	scheduled := now
	if outboundLimiter.next.After(now) {
		scheduled = outboundLimiter.next
	}
	outboundLimiter.next = scheduled.Add(minOutboundInterval)
	outboundLimiter.Unlock()

	if wait := time.Until(scheduled); wait > 0 {
		timer := time.NewTimer(wait)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
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
	return enqueueTemplate(ctx, db, priority, to, subject, tmpl, data, opts...)
}

// EnqueueTemplateTx inserts the rendered job into the caller's transaction so
// business writes and their notification are committed or rolled back together.
func EnqueueTemplateTx(ctx context.Context, tx *sql.Tx, priority int, to, subject string, tmpl Template, data any, opts ...Option) error {
	return enqueueTemplate(ctx, tx, priority, to, subject, tmpl, data, opts...)
}

type dbtx interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func enqueueTemplate(ctx context.Context, db dbtx, priority int, to, subject string, tmpl Template, data any, opts ...Option) error {
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
		text:     htmlToText(html),
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
	text     string
	from     string
	replyTo  string
}

func enqueue(ctx context.Context, db dbtx, p enqueueParams) error {
	return sqlc.New(db).EnqueueEmail(ctx, sqlc.EnqueueEmailParams{
		Priority:  int8(p.priority), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		ToEmail:   p.to,
		Subject:   p.subject,
		Html:      p.html,
		TextBody:  sql.NullString{String: p.text, Valid: p.text != ""},
		FromEmail: sql.NullString{String: p.from, Valid: p.from != ""},
		ReplyTo:   sql.NullString{String: p.replyTo, Valid: p.replyTo != ""},
	})
}

var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

func htmlToText(markup string) string {
	text := htmlTagPattern.ReplaceAllString(markup, " ")
	return strings.TrimSpace(stdhtml.UnescapeString(text))
}
