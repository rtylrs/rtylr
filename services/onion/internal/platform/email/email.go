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
	"github.com/voxire/onion/internal/platform/emailqueue"
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

func (e *RateLimitError) RetryDelay() time.Duration { return e.RetryAfter }

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
	return emailqueue.Enqueue(ctx, db, emailqueue.Job{
		Priority: priority,
		To:       to,
		Subject:  subject,
		HTML:     html,
		Text:     htmlToText(html),
		From:     params.From,
		ReplyTo:  params.ReplyTo,
	})
}

// Email job priorities for EnqueueTemplate; higher is delivered first.
const (
	PriorityCritical = emailqueue.PriorityCritical
	PriorityHigh     = emailqueue.PriorityHigh
	PriorityNormal   = emailqueue.PriorityNormal
)

var (
	// Raw-text elements are not presentation markup: stripping only their tags
	// would leak the entire stylesheet/script into the plaintext alternative.
	htmlRawTextElementPattern = regexp.MustCompile(`(?is)<(?:style|script)\b[^>]*>.*?(?:</(?:style|script)\s*>|$)`)
	htmlTagPattern            = regexp.MustCompile(`<[^>]+>`)
)

func htmlToText(markup string) string {
	markup = htmlRawTextElementPattern.ReplaceAllString(markup, " ")
	text := htmlTagPattern.ReplaceAllString(markup, " ")
	return strings.TrimSpace(stdhtml.UnescapeString(text))
}
