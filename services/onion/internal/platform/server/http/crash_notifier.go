package http

import (
	"context"
	"errors"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

var ErrCrashRecovered = errors.New("crash recovered")

func CrashNotifierMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
			defer cancel()

			defer func() {
				if recovered := recover(); recovered != nil {
					stackTrace := debug.Stack()
					logger.Ctx(ctx).Error("Onion crashed", logger.Error(ErrCrashRecovered), logger.String("stack_trace", string(stackTrace)))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
