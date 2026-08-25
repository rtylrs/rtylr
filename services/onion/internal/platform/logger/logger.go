package logger

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
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
	mu       sync.Mutex
	pending  sync.WaitGroup
	flushing bool
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

func (h *betterStackHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *betterStackHandler) Handle(_ context.Context, record slog.Record) error {
	payload := slogbetterstack.DefaultConverter(false, nil, h.attrs, h.groups, &record)
	body, err := json.Marshal([]map[string]any{payload})
	if err != nil {
		return err
	}

	h.delivery.mu.Lock()
	if h.delivery.flushing {
		h.delivery.mu.Unlock()
		return nil
	}
	h.delivery.pending.Add(1)
	h.delivery.mu.Unlock()
	go func() {
		defer h.delivery.pending.Done()
		// Request contexts may be canceled as soon as the request completes. Log
		// delivery must outlive the request so Flush can account for it.
		_ = h.send(context.Background(), h.endpoint, h.token, h.timeout, body)
	}()
	return nil
}

func (h *betterStackHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	combined := appendAttrsToGroup(h.groups, h.attrs, attrs...)
	return &betterStackHandler{
		level:    h.level,
		token:    h.token,
		endpoint: h.endpoint,
		timeout:  h.timeout,
		attrs:    combined,
		groups:   h.groups,
		delivery: h.delivery,
		send:     h.send,
	}
}

func (h *betterStackHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	groups := append([]string{}, h.groups...)
	groups = append(groups, name)
	return &betterStackHandler{
		level:    h.level,
		token:    h.token,
		endpoint: h.endpoint,
		timeout:  h.timeout,
		attrs:    h.attrs,
		groups:   groups,
		delivery: h.delivery,
		send:     h.send,
	}
}

func (h *betterStackHandler) Flush() {
	h.delivery.mu.Lock()
	h.delivery.flushing = true
	h.delivery.mu.Unlock()
	h.delivery.pending.Wait()
}

func appendAttrsToGroup(groups []string, existing []slog.Attr, attrs ...slog.Attr) []slog.Attr {
	if len(groups) == 0 {
		combined := make([]slog.Attr, 0, len(existing)+len(attrs))
		combined = append(combined, existing...)
		return append(combined, attrs...)
	}

	combined := append([]slog.Attr{}, existing...)
	for i, attr := range combined {
		if attr.Key == groups[0] && attr.Value.Kind() == slog.KindGroup {
			combined[i] = groupAttr(groups[0], appendAttrsToGroup(groups[1:], attr.Value.Group(), attrs...))
			return combined
		}
	}
	return append(combined, groupAttr(groups[0], appendAttrsToGroup(groups[1:], nil, attrs...)))
}

func groupAttr(name string, attrs []slog.Attr) slog.Attr {
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return slog.Group(name, args...)
}

func sendBetterStack(ctx context.Context, endpoint, token string, timeout time.Duration, body []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "onion")

	resp, err := (&http.Client{Timeout: timeout}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (d *dedupHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return d.inner.Enabled(ctx, level)
}

func (d *dedupHandler) Handle(ctx context.Context, r slog.Record) error {
	key := d.key(r)

	d.mu.Lock()
	if d.seen == nil {
		d.seen = make(map[string]time.Time)
	}
	last, ok := d.seen[key]
	now := time.Now()
	d.cleanupLocked(now)
	if ok && now.Sub(last) < d.window {
		d.mu.Unlock()
		return nil
	}
	d.seen[key] = now
	d.mu.Unlock()

	return d.inner.Handle(ctx, r)
}

func (d *dedupHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	combined := make([]slog.Attr, 0, len(d.attrs)+len(attrs))
	combined = append(combined, d.attrs...)
	combined = append(combined, attrs...)
	return &dedupHandler{inner: d.inner.WithAttrs(attrs), window: d.window, attrs: combined}
}

func (d *dedupHandler) WithGroup(name string) slog.Handler {
	return &dedupHandler{inner: d.inner.WithGroup(name), window: d.window, attrs: d.attrs}
}

func (d *dedupHandler) key(r slog.Record) string {
	var fingerprint strings.Builder
	fingerprint.WriteString(r.Level.String())
	fingerprint.WriteByte('|')
	fingerprint.WriteString(r.Message)
	for _, attr := range d.attrs {
		appendDedupAttr(&fingerprint, attr)
	}
	r.Attrs(func(attr slog.Attr) bool {
		appendDedupAttr(&fingerprint, attr)
		return true
	})
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fingerprint.String())))
}

func appendDedupAttr(b *strings.Builder, attr slog.Attr) {
	b.WriteByte('|')
	b.WriteString(attr.Key)
	b.WriteByte('=')
	b.WriteString(attr.Value.String())
}

func (d *dedupHandler) cleanupLocked(now time.Time) {
	for key, lastSeen := range d.seen {
		if now.Sub(lastSeen) >= d.window {
			delete(d.seen, key)
		}
	}
	const maxEntries = 4096
	for len(d.seen) > maxEntries {
		var oldestKey string
		var oldest time.Time
		for key, lastSeen := range d.seen {
			if oldestKey == "" || lastSeen.Before(oldest) {
				oldestKey = key
				oldest = lastSeen
			}
		}
		delete(d.seen, oldestKey)
	}
}

func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

func Global() *Logger {
	return &Logger{slog.Default(), context.Background()}
}

func Ctx(ctx context.Context) *Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok && l != nil {
		return &Logger{l, ctx}
	}

	return &Logger{slog.Default(), ctx}
}

func (l *Logger) Info(msg string, fields ...Field) {
	l.LogAttrs(l.ctx, slog.LevelInfo, msg, convertFields(fields)...)
}

func (l *Logger) Debug(msg string, fields ...Field) {
	l.LogAttrs(l.ctx, slog.LevelDebug, msg, convertFields(fields)...)
}

func (l *Logger) Error(msg string, fields ...Field) {
	l.LogAttrs(l.ctx, slog.LevelError, msg, convertFields(fields)...)
}

func (l *Logger) Warn(msg string, fields ...Field) {
	l.LogAttrs(l.ctx, slog.LevelWarn, msg, convertFields(fields)...)
}

func (l *Logger) Fatal(msg string, fields ...Field) {
	l.LogAttrs(l.ctx, slog.LevelError, msg, convertFields(fields)...)
	os.Exit(1)
}

func Any(key string, value interface{}) Field { return Field{Key: key, Value: value} }

func Bool(key string, value bool) Field { return Field{Key: key, Value: value} }

func Float64(key string, value float64) Field { return Field{Key: key, Value: value} }

func Int(key string, value int) Field { return Field{Key: key, Value: value} }

func String(key, value string) Field { return Field{Key: key, Value: value} }

func Error(err error) Field { return Field{Key: "error", Value: err} }

func Pretty(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "error: cannot pretty-print"
	}

	return string(b)
}

func convertFields(fields []Field) []slog.Attr {
	attrs := make([]slog.Attr, len(fields))
	for i, f := range fields {
		switch v := f.Value.(type) {
		case string:
			attrs[i] = slog.String(f.Key, v)
		case int:
			attrs[i] = slog.Int(f.Key, v)
		case int64:
			attrs[i] = slog.Int64(f.Key, v)
		case float64:
			attrs[i] = slog.Float64(f.Key, v)
		case bool:
			attrs[i] = slog.Bool(f.Key, v)
		case error:
			attrs[i] = slog.Any(f.Key, v)
		default:
			attrs[i] = slog.Any(f.Key, v)
		}
	}

	return attrs
}
