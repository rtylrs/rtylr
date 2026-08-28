package http

import (
	"encoding/json"
	"net/http"
	"runtime/debug"

	"github.com/voxire/onion/internal/platform/logger"
)

type SecureErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func SecureErrorMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			defer func() {
				if err := recover(); err != nil {
					logger.Ctx(ctx).Error("Panic recovered",
						logger.Any("error", err),
						logger.String("stack", string(debug.Stack())),
						logger.String("path", r.URL.Path),
						logger.String("method", r.Method),
						logger.String("ip", r.RemoteAddr),
					)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					response := SecureErrorResponse{
						Error:   "Internal Server Error",
						Message: "An unexpected error occurred. Please try again later.",
						Status:  "error",
					}

					if err := json.NewEncoder(w).Encode(response); err != nil {
						logger.Ctx(ctx).Error("Failed to write error response", logger.Error(err))
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func SanitizeError(err error) error {
	if err != nil {
		return &GenericError{message: "An error occurred"}
	}

	return err
}

type GenericError struct {
	message string
}

func (e *GenericError) Error() string {
	return e.message
}
