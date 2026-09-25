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
	"time"

	slogbetterstack "github.com/samber/slog-betterstack"
)

func (h *betterStackHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *betterStackHandler) Handle(ctx context.Context, record slog.Record) error {
	payload := slogbetterstack.DefaultConverter(false, nil, h.attrs, h.groups, &record)
	body, err := json.Marshal([]map[string]any{payload})
	if err != nil {
		return err
	}

	h.delivery.mu.Lock()
	if h.delivery.activeFlushes > 0 {
		h.delivery.mu.Unlock()
		return nil
	}
	h.delivery.pending.Add(1)
	h.delivery.mu.Unlock()
	deliveryCtx := context.WithoutCancel(ctx)
	go func() {
		defer h.delivery.pending.Done()
		// Request contexts may be canceled as soon as the request completes. Log
		// delivery must outlive the request so Flush can account for it.
		_ = h.send(deliveryCtx, h.endpoint, h.token, h.timeout, body)
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
	h.delivery.activeFlushes++
	h.delivery.mu.Unlock()

	h.delivery.pending.Wait()

	h.delivery.mu.Lock()
	h.delivery.activeFlushes--
	h.delivery.mu.Unlock()
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

// exitProcess terminates the process. It is a variable so tests can exercise
// Fatal's shutdown path without killing the test binary.
var exitProcess = os.Exit

func (l *Logger) Fatal(msg string, fields ...Field) {
	l.LogAttrs(l.ctx, slog.LevelError, msg, convertFields(fields)...)
	// os.Exit skips every deferred call, including runServer's Flush, so the
	// record queued above must be delivered before the process dies.
	Flush()
	exitProcess(1)
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
