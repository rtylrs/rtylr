package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	slogbetterstack "github.com/samber/slog-betterstack"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/slack"
)

type ctxKey struct{}

type Logger struct {
	*slog.Logger
	ctx context.Context
}

type Field struct {
	Key   string
	Value interface{}
}

// SetupLogger initialises the global slog default. If BETTERSTACK_SOURCE_TOKEN
// is set, warn+error records are shipped to Better Stack in addition to JSON
// stdout. Call after loading environment variables.
func SetupLogger() {
	// Start the Slack notifier first so the slack handler below has somewhere to
	// send (no-op unless SLACK_LOGGING_ENABLED + a bot token are set).
	slack.Init()

	stdoutHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})

	// Error-level records are mirrored to the Slack alerts channel (batched).
	handlers := []slog.Handler{stdoutHandler, slackErrorHandler{}}

	var token, ingestingHost string
	if config.GlobalConfig != nil {
		token = config.GlobalConfig.BetterStackSourceToken
		ingestingHost = config.GlobalConfig.BetterStackIngestingHost
	} else {
		token = os.Getenv("BETTERSTACK_SOURCE_TOKEN")
		ingestingHost = os.Getenv("BETTERSTACK_INGESTING_HOST")
	}

	if strings.TrimSpace(token) != "" {
		endpoint := betterStackEndpoint(ingestingHost)
		if endpoint == "" {
			endpoint = slogbetterstack.BetterstackEndpoint
		}
		h := newBetterStackHandler(strings.TrimSpace(token), endpoint)
		handlers = append(handlers, &dedupHandler{inner: h, window: 30 * time.Second})
	}

	// Tag every record with which runmode emitted it (api / worker / upload),
	// so logs are filterable per service. Set SERVICE_NAME per task definition;
	// defaults to "onion" for local/dev.
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "onion"
	}

	slog.SetDefault(slog.New(&multiHandler{handlers: handlers}).With(slog.String("service", serviceName)))
}

// ServiceName returns the configured runmode label (api/worker/upload), or
// "onion" if unset. Used by request/security middleware so their explicit
// "service" field matches the base attribute.
func ServiceName() string {
	if n := os.Getenv("SERVICE_NAME"); n != "" {
		return n
	}
	return "onion"
}

// slackErrorHandler forwards error-level records to the Slack alerts channel. It
// is a no-op when Slack logging is disabled.
type slackErrorHandler struct{}

func (slackErrorHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= slog.LevelError
}

func (slackErrorHandler) Handle(_ context.Context, r slog.Record) error {
	if !slack.Enabled() {
		return nil
	}
	var b strings.Builder
	b.WriteString("🔴 ")
	b.WriteString(r.Message)
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "body_preview" {
			return true
		}
		b.WriteString("\n• ")
		b.WriteString(a.Key)
		b.WriteString(": ")
		b.WriteString(a.Value.String())
		return true
	})
	slack.Post(slack.Alerts, b.String())
	return nil
}

func (h slackErrorHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h slackErrorHandler) WithGroup(_ string) slog.Handler      { return h }

var activeBetterStackHandler *betterStackHandler

// Flush waits for Better Stack records that have already been accepted by the
// logger. Stdout and Slack are synchronous and need no shutdown hook.
func Flush() {
	if activeBetterStackHandler != nil {
		activeBetterStackHandler.Flush()
	}
}

func betterStackEndpoint(ingestingHost string) string {
	ingestingHost = strings.TrimSpace(ingestingHost)
	if ingestingHost == "" || strings.Contains(ingestingHost, "://") {
		return ingestingHost
	}
	return "https://" + ingestingHost
}

// multiHandler fans out to multiple slog.Handlers. Each handler's own Enabled
// check gates whether it receives a given record.
type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			_ = h.Handle(ctx, r)
		}
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}

// dedupHandler wraps a slog.Handler and suppresses repeated event fingerprints
// within the configured window so Better Stack doesn't receive duplicate bursts.
type dedupHandler struct {
	inner  slog.Handler
	window time.Duration
	attrs  []slog.Attr

	mu   sync.Mutex
	seen map[string]time.Time
}

type betterStackHandler struct {
	level    slog.Level
	token    string
	endpoint string
	timeout  time.Duration
	attrs    []slog.Attr
	groups   []string
	delivery *betterStackDelivery
	send     func(context.Context, string, string, time.Duration, []byte) error
}

type betterStackDelivery struct {
	mu sync.Mutex
	// activeFlushes counts Flush calls that have not returned yet. Handle drops
	// records while it is non-zero so Flush never waits on records logged during
	// shutdown. Counting instead of a plain bool lets delivery resume when a
	// Flush returns instead of staying disabled for the rest of the process —
	// runServer() is re-entered after a crash and Flush runs on every return.
	activeFlushes int
	pending       sync.WaitGroup
}

func newBetterStackHandler(token, endpoint string) *betterStackHandler {
	h := &betterStackHandler{
		level:    slog.LevelWarn,
		token:    token,
		endpoint: endpoint,
		timeout:  10 * time.Second,
		delivery: &betterStackDelivery{},
	}
	h.send = sendBetterStack
	activeBetterStackHandler = h
	return h
}
