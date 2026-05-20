package logger

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"time"

	axiomSlog "github.com/axiomhq/axiom-go/adapters/slog"
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

var globalAxiomHandler *axiomSlog.Handler

// SetupLogger initialises the global slog default. If AXIOM_TOKEN and
// AXIOM_DATASET are set, warn+error records are shipped to Axiom in addition
// to JSON stdout. Call after loading environment variables.
func SetupLogger() {
	stdoutHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})

	if os.Getenv("AXIOM_TOKEN") == "" || os.Getenv("AXIOM_DATASET") == "" {
		slog.SetDefault(slog.New(stdoutHandler))
		return
	}

	h, err := axiomSlog.New(axiomSlog.SetLevel(slog.LevelWarn))
	if err != nil {
		slog.SetDefault(slog.New(stdoutHandler))
		return
	}

	globalAxiomHandler = h
	slog.SetDefault(slog.New(&multiHandler{handlers: []slog.Handler{
		stdoutHandler,
		&dedupHandler{inner: h, window: 30 * time.Second},
	}}))
}

// Flush drains any buffered Axiom events. Call on shutdown.
func Flush() {
	if globalAxiomHandler != nil {
		globalAxiomHandler.Close()
	}
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

// dedupHandler wraps a slog.Handler and suppresses repeated (level, message)
// pairs within the configured window so Axiom doesn't receive duplicate bursts.
type dedupHandler struct {
	inner  slog.Handler
	window time.Duration

	mu   sync.Mutex
	seen map[string]time.Time
}

func (d *dedupHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return d.inner.Enabled(ctx, level)
}

func (d *dedupHandler) Handle(ctx context.Context, r slog.Record) error {
	key := r.Level.String() + "|" + r.Message

	d.mu.Lock()
	if d.seen == nil {
		d.seen = make(map[string]time.Time)
	}
	last, ok := d.seen[key]
	now := time.Now()
	if ok && now.Sub(last) < d.window {
		d.mu.Unlock()
		return nil
	}
	d.seen[key] = now
	d.mu.Unlock()

	return d.inner.Handle(ctx, r)
}

func (d *dedupHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &dedupHandler{inner: d.inner.WithAttrs(attrs), window: d.window}
}

func (d *dedupHandler) WithGroup(name string) slog.Handler {
	return &dedupHandler{inner: d.inner.WithGroup(name), window: d.window}
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
