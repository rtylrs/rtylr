package http

import (
	"context"
	"errors"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

var (
	ErrCrashRecovered = errors.New("crash recovered")
	ErrEncodeFailed   = errors.New("encode failed")
)

func CrashNotifierMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
			defer cancel()

			defer func(_ context.Context) {
				if recovered := recover(); recovered != nil {
					stackTrace := debug.Stack()
					logger.Ctx(ctx).Error("Onion crashed", logger.Error(ErrCrashRecovered), logger.String("stack_trace", string(stackTrace)))
				}
			}(ctx)

			next.ServeHTTP(w, r)

			defer func(_ context.Context) {
				if err := recover(); err != nil {
					logger.Ctx(ctx).Error("Failed to encode error response", logger.Error(ErrEncodeFailed))
				}
			}(ctx)
		})
	}
}
