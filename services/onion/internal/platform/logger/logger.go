package logger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"

	"github.com/voxire/onion/internal/platform/utils/alerts"
)

var ErrAlertsNotInitialized = errors.New("alerts service not initialized")

type ctxKey struct{}

type Logger struct {
	*slog.Logger
	ctx context.Context
}

type Field struct {
	Key   string
	Value interface{}
}

type Config struct {
	AlertsEnabled  bool
	AlertsChannels map[string]string
}

var (
	alertsSvc     *alerts.Service
	alertsMu      sync.RWMutex
	alertsEnabled bool
)

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

func InitAlertsFromEnv() {
	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	alertsChannelID := os.Getenv("SLACK_ALERTS_CHANNEL_ID")
	enabled := os.Getenv("SLACK_LOGGING_ENABLED") == "true"

	if slackBotToken != "" && alertsChannelID != "" {
		alertsMu.Lock()

		alertsSvc = alerts.NewAlertsService(context.Background(), slackBotToken)
		alertsSvc.AddChannel("alerts", alertsChannelID)

		alertsEnabled = enabled

		alertsMu.Unlock()
	}
}

func ConfigureAlerts(cfg Config) {
	alertsMu.Lock()
	defer alertsMu.Unlock()

	if cfg.AlertsEnabled && len(cfg.AlertsChannels) > 0 {
		slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
		if slackBotToken != "" {
			alertsSvc = alerts.NewAlertsService(context.Background(), slackBotToken)
			for name, channelID := range cfg.AlertsChannels {
				alertsSvc.AddChannel(name, channelID)
			}

			alertsEnabled = true
		}
	} else {
		alertsSvc = nil
		alertsEnabled = false
	}
}

func AddAlertsChannel(name, channelID string) {
	alertsMu.Lock()
	defer alertsMu.Unlock()

	if alertsSvc != nil {
		alertsSvc.AddChannel(name, channelID)
	}
}

func SendAlertsMessage(channelName, message string) error {
	alertsMu.RLock()
	defer alertsMu.RUnlock()

	if alertsSvc == nil {
		return ErrAlertsNotInitialized
	}

	return alertsSvc.SendMessage(context.Background(), channelName, message)
}

func sendNotifications(ctx context.Context, level, message string, fields []Field) {
	formattedMessage := fmt.Sprintf("%s: %s", level, message)
	if len(fields) > 0 {
		formattedMessage += "\n"
		for _, f := range fields {
			formattedMessage += fmt.Sprintf("• %s: %v\n", f.Key, f.Value)
		}
	}

	select {
	case <-ctx.Done():
		return
	default:
	}

	go func() {
		alertsMu.RLock()

		currentAlertsSvc := alertsSvc
		currentAlertsEnabled := alertsEnabled

		alertsMu.RUnlock()

		if currentAlertsEnabled && currentAlertsSvc != nil {
			if err := currentAlertsSvc.SendMessage(ctx, "alerts", formattedMessage); err != nil {
				log.Printf("Failed to send alerts notification: %v", err)
			}
		}
	}()
}

func (l *Logger) Info(msg string, fields ...Field) {
	l.LogAttrs(l.ctx, slog.LevelInfo, "\""+msg+"\"", convertFields(fields)...)
}

func (l *Logger) Error(msg string, fields ...Field) {
	l.LogAttrs(l.ctx, slog.LevelError, msg, convertFields(fields)...)
	sendNotifications(l.ctx, "[Error]", msg, fields)
}

func (l *Logger) Warn(msg string, fields ...Field) {
	l.LogAttrs(l.ctx, slog.LevelWarn, msg, convertFields(fields)...)
	sendNotifications(l.ctx, "[Warn]", msg, fields)
}

func (l *Logger) Fatal(msg string, fields ...Field) {
	l.LogAttrs(l.ctx, slog.LevelError, msg, convertFields(fields)...)
	sendNotifications(l.ctx, "[Fatal]", msg, fields)
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
