package logger

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
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

func SetupLogger() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})
	slog.SetDefault(slog.New(handler))
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
	l.LogAttrs(l.ctx, slog.LevelInfo, "\""+msg+"\"", convertFields(fields)...)
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

func Flush() {}

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
