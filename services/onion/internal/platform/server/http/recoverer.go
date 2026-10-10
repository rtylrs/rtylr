package http

import (
	"net/http"
	"runtime/debug"

	"github.com/voxire/onion/internal/platform/logger"
)

// recoverMiddleware turns a handler panic into a logged 500, so one bad
// request cannot take the process down. http.ErrAbortHandler is re-raised:
// it is how a handler deliberately aborts a response.
func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}
			if recovered == http.ErrAbortHandler { //nolint:errorlint // sentinel compared by identity, as net/http does
				panic(recovered)
			}
			logger.Ctx(r.Context()).Error("Recovered handler panic",
				logger.Any("panic", recovered),
				logger.String("stack", string(debug.Stack())))
			if r.Header.Get("Connection") != "Upgrade" {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
