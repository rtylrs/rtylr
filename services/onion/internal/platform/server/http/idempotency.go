package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	idempotencyHeader             = "Idempotency-Key"
	idempotencyTTL                = 30 * time.Minute
	idempotencyMaxLen             = 128
	idempotencyMaxStoredBodyBytes = 256 << 10
	idempotencyCacheWriteTimeout  = 2 * time.Second
)

// idempotencyRecord is what we store in the cache for a completed request.
type idempotencyRecord struct {
	Status  int         `json:"status"`
	Headers http.Header `json:"headers"`
	Body    []byte      `json:"body"`
}

type idempotencyRedisStore interface {
	SetNx(ctx context.Context, key string, val []byte, ttl time.Duration) bool
	Get(ctx context.Context, key string) []byte
	Put(ctx context.Context, key string, val []byte, ttl time.Duration)
	Delete(ctx context.Context, keys []string)
}

// IdempotencyMiddleware deduplicates mutating requests that carry an
// Idempotency-Key header. On the first call the handler runs normally and
// the response is stored for 30 minutes; any subsequent request with the same
// key from the same principal within that window gets the stored response
// without touching the DB.
//
// When Redis is enabled the store is Redis (shared across all replicas), with a
// SetNx reservation so exactly one request executes even under a horizontally
// scaled deployment — a retry that lands on a different replica replays the
// stored response instead of re-executing the handler. Without Redis it falls
// back to the in-process cache (correct for a single replica).
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
			if _, writeErr := w.Write([]byte(`{"error":"Idempotency-Key too long (max 128 chars)"}`)); writeErr != nil {
				logger.Ctx(r.Context()).Warn("failed to write idempotency key length response", logger.Error(writeErr))
			}
			return
		}

		cacheKey := buildIdempotencyKey(r, key)
		if config.IsRedisEnabled() {
			serveIdempotentRedis(w, r, next, cacheKey)
			return
		}
		serveIdempotentLocal(w, r, next, cacheKey)
	})
}

// serveIdempotentRedis gives correct cross-replica idempotency. SetNx atomically
// reserves the key so exactly one request runs; concurrent or retried duplicates
// (on any replica) either replay the completed response or get 409 while it is
// still in flight. The reservation is released if the handler fails or panics so
// the client can retry with a fresh attempt.
func serveIdempotentRedis(w http.ResponseWriter, r *http.Request, next http.Handler, cacheKey string) {
	rc := cacher.GetRedisCache()
	pending := []byte(`{"status":0}`)

	reserved, cached := reserveIdempotencyKey(r.Context(), rc, cacheKey, pending)
	if !reserved {
		if cached != nil {
			var rec idempotencyRecord
			if err := json.Unmarshal(cached, &rec); err == nil {
				if rec.Status >= 200 {
					writeIdempotentReplay(w, rec)
					return
				}
				// Owned by another in-flight request (possibly on another
				// replica) — tell the client to retry shortly.
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				if _, writeErr := w.Write([]byte(`{"error":"a request with this Idempotency-Key is still being processed"}`)); writeErr != nil {
					logger.Ctx(r.Context()).Warn("failed to write idempotency conflict response", logger.Error(writeErr))
				}
				return
			}
		}
		// Do not execute a mutating handler without a reservation. A failed
		// reservation is either another in-flight request or a brief cache race;
		// both are safe to retry, while running now could duplicate the mutation.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		if _, writeErr := w.Write([]byte(`{"error":"a request with this Idempotency-Key is still being processed"}`)); writeErr != nil {
			logger.Ctx(r.Context()).Warn("failed to write idempotency conflict response", logger.Error(writeErr))
		}
		return
	}

	stored := false
	if reserved {
		// Release the reservation unless we store a success, so an error or a
		// panic doesn't leave the key wedged for the full TTL.
		defer func() {
			if !stored {
				cacheCtx, cancel := idempotencyCacheContext(r.Context())
				defer cancel()
				rc.Delete(cacheCtx, []string{cacheKey})
			}
		}()
	}

	rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
	next.ServeHTTP(rec, r)

	if rec.status >= 200 && rec.status < 300 {
		storeIdempotencyRecord(r.Context(), cacheKey, rec, func(cacheCtx context.Context, key string, payload []byte) {
			rc.Put(cacheCtx, key, payload, idempotencyTTL)
			stored = true
		})
	}
}

func reserveIdempotencyKey(ctx context.Context, store idempotencyRedisStore, key string, pending []byte) (bool, []byte) {
	for attempt := 0; attempt < 2; attempt++ {
		if store.SetNx(ctx, key, pending, idempotencyTTL) {
			return true, nil
		}
		if cached := store.Get(ctx, key); cached != nil {
			return false, cached
		}
	}
	return false, nil
}

// serveIdempotentLocal is the single-replica fallback used when Redis is off:
// dedup against the in-process cache. There is no cross-replica guarantee here
// (there is only one replica), so a simple get-or-run is sufficient.
func serveIdempotentLocal(w http.ResponseWriter, r *http.Request, next http.Handler, cacheKey string) {
	lc := cacher.GetLocalCache(r.Context())

	if cached := lc.Get(r.Context(), cacheKey); cached != nil {
		var rec idempotencyRecord
		if err := json.Unmarshal(cached, &rec); err == nil {
			writeIdempotentReplay(w, rec)
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
		storeIdempotencyRecord(r.Context(), cacheKey, rec, func(cacheCtx context.Context, key string, payload []byte) {
			lc.Put(cacheCtx, key, payload, idempotencyTTL)
		})
	}
}

func idempotencyCacheContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), idempotencyCacheWriteTimeout)
}

// writeIdempotentReplay replays a stored response, tagging it so clients can tell
// it came from the idempotency cache rather than a fresh execution.
func writeIdempotentReplay(w http.ResponseWriter, rec idempotencyRecord) {
	for hk, hvs := range rec.Headers {
		for _, hv := range hvs {
			w.Header().Set(hk, hv)
		}
	}
	w.Header().Set("Idempotent-Replayed", "true")
	w.WriteHeader(rec.Status)
	if _, writeErr := io.Copy(w, bytes.NewReader(rec.Body)); writeErr != nil {
		logger.Global().Warn("failed to write idempotent replay body", logger.Error(writeErr))
	}
}

func buildIdempotencyKey(r *http.Request, key string) string {
	principal := r.RemoteAddr
	if authUser, ok := authPkg.GetAuthUser(r.Context()); ok {
		principal = fmt.Sprintf("u%d", authUser.UserID)
	}
	return fmt.Sprintf("idem:%s:%s", principal, key)
}

func storeIdempotencyRecord(ctx context.Context, cacheKey string, rec *responseRecorder, persist func(context.Context, string, []byte)) {
	body := rec.buf.Bytes()
	if len(body) > idempotencyMaxStoredBodyBytes {
		logger.Ctx(ctx).Warn("skipping idempotency cache for oversized response body",
			logger.Int("bytes", len(body)),
			logger.Int("limit", idempotencyMaxStoredBodyBytes),
		)
		return
	}

	entry := idempotencyRecord{
		Status:  rec.status,
		Headers: idempotencyHeadersForReplay(rec.Header()),
		Body:    body,
	}
	payload, err := json.Marshal(entry)
	if err != nil {
		logger.Ctx(ctx).Warn("failed to marshal idempotency record", logger.Error(err))
		return
	}

	cacheCtx, cancel := idempotencyCacheContext(ctx)
	defer cancel()
	persist(cacheCtx, cacheKey, payload)
}

func idempotencyHeadersForReplay(h http.Header) http.Header {
	out := make(http.Header)
	for _, key := range []string{"Content-Type", "Content-Encoding"} {
		if values := h.Values(key); len(values) > 0 {
			out[key] = values
		}
	}
	return out
}

// responseRecorder buffers the response so we can cache it after the handler
// runs. It forwards all writes to the underlying ResponseWriter immediately.
type responseRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	buf         bytes.Buffer
}

func (rr *responseRecorder) Unwrap() http.ResponseWriter { return rr.ResponseWriter }

func (rr *responseRecorder) Flush() {
	if flusher, ok := rr.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (rr *responseRecorder) WriteHeader(code int) {
	if rr.wroteHeader {
		return
	}
	rr.wroteHeader = true
	rr.status = code
	rr.ResponseWriter.WriteHeader(code)
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	if !rr.wroteHeader {
		rr.wroteHeader = true
	}
	rr.buf.Write(b)
	return rr.ResponseWriter.Write(b)
}
