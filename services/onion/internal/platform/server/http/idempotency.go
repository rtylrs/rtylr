package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	idempotencyHeader = "Idempotency-Key"
	idempotencyTTL    = 30 * time.Minute
	idempotencyMaxLen = 128
)

// idempotencyRecord is what we store in the cache for a completed request.
type idempotencyRecord struct {
	Status  int         `json:"status"`
	Headers http.Header `json:"headers"`
	Body    []byte      `json:"body"`
}

// IdempotencyMiddleware deduplicates mutating requests that carry an
// Idempotency-Key header. On the first call the handler runs normally and
// the response is stored in BigCache for 30 minutes. Any subsequent request
// with the same key from the same principal within that window gets the
// stored response without touching the DB.
//
// Scope: POST, PUT, PATCH only. GET/DELETE are idempotent by nature.
// Key scoping: authenticated user ID when available, otherwise remote IP.
// Header absent: passthrough — callers that don't need dedup pay nothing.
func IdempotencyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
			next.ServeHTTP(w, r)
			return
		}

		key := strings.TrimSpace(r.Header.Get(idempotencyHeader))
		if key == "" {
			next.ServeHTTP(w, r)
			return
		}
		if len(key) > idempotencyMaxLen {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"Idempotency-Key too long (max 128 chars)"}`))
			return
		}

		cacheKey := buildIdempotencyKey(r, key)
		lc := cacher.GetLocalCache(r.Context())

		if cached := lc.Get(r.Context(), cacheKey); cached != nil {
			var rec idempotencyRecord
			if err := json.Unmarshal(cached, &rec); err == nil {
				for hk, hvs := range rec.Headers {
					for _, hv := range hvs {
						w.Header().Set(hk, hv)
					}
				}
				w.Header().Set("Idempotent-Replayed", "true")
				w.WriteHeader(rec.Status)
				_, _ = w.Write(rec.Body)
				return
			}
			logger.Ctx(r.Context()).Warn("failed to unmarshal idempotency record, reprocessing",
				logger.String("key", cacheKey),
			)
		}

		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		// Only cache successful responses — errors may be transient and the
		// client should be able to retry and get a fresh attempt.
		if rec.status >= 200 && rec.status < 300 {
			entry := idempotencyRecord{
				Status:  rec.status,
				Headers: w.Header().Clone(),
				Body:    rec.buf.Bytes(),
			}
			if b, err := json.Marshal(entry); err == nil {
				lc.Put(r.Context(), cacheKey, b, idempotencyTTL)
			}
		}
	})
}

func buildIdempotencyKey(r *http.Request, key string) string {
	principal := r.RemoteAddr
	if authUser, ok := authPkg.GetAuthUser(r.Context()); ok {
		principal = fmt.Sprintf("u%d", authUser.UserID)
	}
	return fmt.Sprintf("idem:%s:%s", principal, key)
}

// responseRecorder buffers the response so we can cache it after the handler
// runs. It forwards all writes to the underlying ResponseWriter immediately.
type responseRecorder struct {
	http.ResponseWriter
	status int
	buf    bytes.Buffer
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.status = code
	rr.ResponseWriter.WriteHeader(code)
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	rr.buf.Write(b)
	return rr.ResponseWriter.Write(b)
}
