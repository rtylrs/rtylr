package http

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	idempotencyHeader = "Idempotency-Key"

	// idempotencyTTL is the retention window for a COMPLETED response. It must
	// comfortably exceed a client's retry window; 30 minutes is the contract
	// already documented on the middleware and is unchanged.
	idempotencyTTL = 30 * time.Minute

	// idempotencyClaimTTL is the lease on an IN-FLIGHT claim and is deliberately
	// much shorter than idempotencyTTL. A holder that crashes mid-handler never
	// runs its release, so the lease is what stops a crashed claim from wedging
	// the key for the full retention window.
	//
	// Takeover policy: once the lease expires the next requester's SET NX
	// succeeds and it executes the handler. That is at-least-once, chosen over
	// wedging, and it is safe here because the lease is strictly longer than
	// RequestTimeout — the first attempt is cancelled by the request deadline
	// before any takeover can occur, so an expired lease always means "crashed",
	// never "still running". Keep this invariant if either value changes.
	idempotencyClaimTTL = 2 * RequestTimeout

	// idempotencyRetryAfterSeconds is the Retry-After advertised to a client that
	// arrives while a claim is held. It matches the claim lease scale so a
	// client that polls lands after a crashed holder's lease expires.
	idempotencyRetryAfterSeconds = 2

	idempotencyMaxLen             = 128
	idempotencyMaxStoredBodyBytes = 256 << 10
	// Tied to the upstream body limit: a body that is too large to buffer here
	// is one RequestBodyLimitMiddleware has already rejected.
	idempotencyMaxFingerprintBody = DefaultMaxRequestBodyBytes
	idempotencyCacheWriteTimeout  = 2 * time.Second
)

// idempotencyRecord is what we store in the cache. Status 0 means "claimed, no
// result yet" — a distinct state from a completed response so a reader never
// mistakes an in-flight claim for a replayable answer.
type idempotencyRecord struct {
	Status  int         `json:"status"`
	Headers http.Header `json:"headers"`
	Body    []byte      `json:"body"`
	// Claim identifies the attempt that owns an in-flight claim. Release is
	// conditional on still seeing our own token, so a holder that comes back
	// after its lease was taken over cannot delete the successor's claim.
	Claim string `json:"claim,omitempty"`
	// Fingerprint binds the record to one specific request. A record without
	// one (written before this field existed) replays unconditionally.
	Fingerprint string `json:"fingerprint,omitempty"`
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
// key, from the same principal in the same tenant, carrying the same
// fingerprint, gets the stored response without touching the DB.
//
// When Redis is enabled the store is Redis (shared across all replicas), with a
// SET NX claim so exactly one request executes even under a horizontally scaled
// deployment. Without Redis it falls back to the in-process cache (correct for
// a single replica).
//
// In-flight duplicates get 409 with Retry-After — never a second execution and
// never a silent success. A key reused with a *different* request also gets 409
// rather than the first response.
//
// Scope: POST, PUT, PATCH only. GET/DELETE are idempotent by nature.
// Key scoping: tenant + authenticated user ID when available, otherwise remote
// IP. Header absent: passthrough — callers that don't need dedup pay nothing.
func IdempotencyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
			next.ServeHTTP(w, r)
			return
		}

		key := strings.TrimSpace(r.Header.Get(idempotencyHeader))
		if key == "" {
			if requiresIdempotencyKey(r) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusPreconditionRequired)
				if _, writeErr := w.Write([]byte(`{"error":"Idempotency-Key is required for this operation"}`)); writeErr != nil {
					logger.Ctx(r.Context()).Warn("failed to write required idempotency key response", logger.Error(writeErr))
				}
				return
			}
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
		if useRedisIdempotencyStore() {
			serveIdempotentRedis(w, r, next, cacheKey)
			return
		}
		serveIdempotentLocal(w, r, next, cacheKey)
	})
}

// useRedisIdempotencyStore reports whether the shared Redis store is usable
// right now. It requires the config flag *and* a live client: cache.Initialize
// swaps in the no-op cache when its Redis probe fails
// (cacher.DisableRedisForProcess), and the no-op reports success for every
// SetNx while persisting nothing. Selecting the Redis path in that state would
// silently disable deduplication instead of falling back to the in-process
// store that still dedups a single replica.
func useRedisIdempotencyStore() bool {
	return config.IsRedisEnabled() && cacher.RedisAvailable()
}

// idempotencyConfirmingStore is implemented by cache stores that can report
// whether a write was actually persisted. Stores without it keep the previous
// assumption that Put succeeded.
type idempotencyConfirmingStore interface {
	PutConfirmed(ctx context.Context, key string, val []byte, ttl time.Duration) bool
}

// persistIdempotencyRecord stores the completed response and reports whether
// the write can be trusted. A false result makes the caller release the
// reservation so a retry re-executes instead of replaying the pending sentinel
// for the full TTL.
func persistIdempotencyRecord(ctx context.Context, store idempotencyRedisStore, key string, payload []byte) bool {
	if confirming, ok := store.(idempotencyConfirmingStore); ok {
		return confirming.PutConfirmed(ctx, key, payload, idempotencyTTL)
	}

	store.Put(ctx, key, payload, idempotencyTTL)

	return true
}

// serveIdempotentRedis gives correct cross-replica idempotency. SET NX
// atomically claims the key so exactly one request runs; concurrent or retried
// duplicates (on any replica) either replay the completed response or get 409
// with Retry-After while the claim is still in flight. Only the attempt that
// owns the claim releases it, so a handler that fails, panics, or returns after
// its lease was taken over cannot delete the successor's claim.
func serveIdempotentRedis(w http.ResponseWriter, r *http.Request, next http.Handler, cacheKey string) {
	serveIdempotentRedisWithStore(w, r, next, cacheKey, cacher.GetRedisCache())
}

func serveIdempotentRedisWithStore(w http.ResponseWriter, r *http.Request, next http.Handler, cacheKey string, rc idempotencyRedisStore) {
	fingerprint := idempotencyRequestFingerprint(r)
	claimPayload, claim, err := newIdempotencyClaim(fingerprint)
	if err != nil {
		// Without a claim token the release path cannot prove ownership, and
		// running an unclaimed mutation risks a double charge.
		logger.Ctx(r.Context()).Error("refusing to execute an unclaimed idempotent mutation",
			logger.Error(err),
		)
		writeIdempotencyUnavailable(w, r)

		return
	}

	reserved, cached := reserveIdempotencyKey(r.Context(), rc, cacheKey, claimPayload)
	if !reserved {
		if cached != nil {
			var rec idempotencyRecord
			if err := json.Unmarshal(cached, &rec); err == nil {
				if rec.Status >= 200 {
					if idempotencyFingerprintConflict(rec.Fingerprint, fingerprint) {
						writeIdempotencyKeyReuse(w, r)

						return
					}
					writeIdempotentReplay(w, rec)

					return
				}
				// Owned by another in-flight request (possibly on another
				// replica) — tell the client to retry shortly.
				writeIdempotencyInFlight(w, r)

				return
			}
		}
		// Do not execute a mutating handler without a reservation. A failed
		// reservation is either another in-flight request or a brief cache race;
		// both are safe to retry, while running now could duplicate the mutation.
		writeIdempotencyInFlight(w, r)

		return
	}

	stored := false
	// Release the claim unless we store a success, so an error or a panic
	// doesn't leave the key wedged. The token check is what keeps a holder that
	// returns after its lease expired from evicting the takeover attempt.
	defer func() {
		if stored {
			return
		}
		cacheCtx, cancel := idempotencyCacheContext(r.Context())
		defer cancel()
		releaseIdempotencyClaim(cacheCtx, rc, cacheKey, claim)
	}()

	rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
	next.ServeHTTP(rec, r)

	if rec.status >= 200 && rec.status < 300 {
		storeIdempotencyRecord(r.Context(), cacheKey, rec, fingerprint, func(cacheCtx context.Context, key string, payload []byte) {
			// Put cannot report failure, so confirm the write instead of
			// assuming it landed: an unconfirmed write leaves the pending
			// sentinel in place and every retry would be answered with 409
			// until the reservation TTL expires.
			stored = persistIdempotencyRecord(cacheCtx, rc, key, payload)
		})
	}
}

// newIdempotencyClaim builds the in-flight sentinel. It marshals with a zero
// Status, which is how a later reader tells a live claim from a completed
// response.
func newIdempotencyClaim(fingerprint string) ([]byte, string, error) {
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return nil, "", err
	}
	claim := hex.EncodeToString(token[:])

	payload, err := json.Marshal(idempotencyRecord{Claim: claim, Fingerprint: fingerprint})
	if err != nil {
		return nil, "", err
	}

	return payload, claim, nil
}

// releaseIdempotencyClaim drops an in-flight claim this attempt still owns. A
// successor can only win SET NX once the key is gone, so if the stored value is
// still our own claim no successor has started and the delete is safe.
func releaseIdempotencyClaim(ctx context.Context, store idempotencyRedisStore, key, claim string) {
	if current := store.Get(ctx, key); current != nil {
		var rec idempotencyRecord
		if err := json.Unmarshal(current, &rec); err != nil || rec.Claim != claim {
			return
		}
	}
	store.Delete(ctx, []string{key})
}

func reserveIdempotencyKey(ctx context.Context, store idempotencyRedisStore, key string, pending []byte) (bool, []byte) {
	for attempt := 0; attempt < 2; attempt++ {
		if store.SetNx(ctx, key, pending, idempotencyClaimTTL) {
			return true, nil
		}
		if cached := store.Get(ctx, key); cached != nil {
			return false, cached
		}
	}
	return false, nil
}

// serveIdempotentLocal is the single-replica fallback used when Redis is off:
// dedup against the in-process cache. The keyed reservation closes the
// get-then-run race so concurrent retries cannot execute the handler twice. A
// duplicate waits for the holder and then replays its stored response, so
// there is no in-flight window to answer 409 with.
func serveIdempotentLocal(w http.ResponseWriter, r *http.Request, next http.Handler, cacheKey string) {
	reservation := acquireLocalIdempotencyReservation(cacheKey)
	defer reservation.release()

	fingerprint := idempotencyRequestFingerprint(r)
	lc := cacher.GetLocalCache(r.Context())

	if cached := lc.Get(r.Context(), cacheKey); cached != nil {
		var rec idempotencyRecord
		if err := json.Unmarshal(cached, &rec); err == nil {
			if idempotencyFingerprintConflict(rec.Fingerprint, fingerprint) {
				writeIdempotencyKeyReuse(w, r)

				return
			}
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
		storeIdempotencyRecord(r.Context(), cacheKey, rec, fingerprint, func(cacheCtx context.Context, key string, payload []byte) {
			lc.Put(cacheCtx, key, payload, idempotencyTTL)
		})
	}
}

type localIdempotencyReservation struct {
	mu   sync.Mutex
	refs int
}

var localIdempotencyReservations = struct {
	sync.Mutex
	entries map[string]*localIdempotencyReservation
}{entries: make(map[string]*localIdempotencyReservation)}

func acquireLocalIdempotencyReservation(key string) *localIdempotencyReservation {
	localIdempotencyReservations.Lock()
	reservation := localIdempotencyReservations.entries[key]
	if reservation == nil {
		reservation = &localIdempotencyReservation{}
		localIdempotencyReservations.entries[key] = reservation
	}
	reservation.refs++
	localIdempotencyReservations.Unlock()

	reservation.mu.Lock()
	return reservation
}

func (r *localIdempotencyReservation) release() {
	r.mu.Unlock()

	localIdempotencyReservations.Lock()
	r.refs--
	if r.refs == 0 {
		for key, candidate := range localIdempotencyReservations.entries {
			if candidate == r {
				delete(localIdempotencyReservations.entries, key)
				break
			}
		}
	}
	localIdempotencyReservations.Unlock()
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
	return fmt.Sprintf("idem:%s:%s", idempotencyPrincipal(r), key)
}

// idempotencyPrincipal scopes a key to one principal in one tenant. The business
// ID is part of the scope because a user can belong to several businesses;
// keying on the user alone would let one tenant's client
// key read another tenant's stored response.
func idempotencyPrincipal(r *http.Request) string {
	if authUser, ok := authPkg.GetAuthUser(r.Context()); ok {
		return fmt.Sprintf("b%d:u%d", authUser.BusinessID, authUser.UserID)
	}
	return "anon:" + anonymousIdempotencyPrincipal(r.RemoteAddr)
}

// idempotencyFingerprintConflict reports that a stored record belongs to a
// different request. An empty fingerprint on either side means it could not be
// computed (legacy record, or a body that could not be buffered); those replay
// as before rather than rejecting a request that cannot be proven to be a reuse.
func idempotencyFingerprintConflict(stored, incoming string) bool {
	return stored != "" && incoming != "" && stored != incoming
}

func anonymousIdempotencyPrincipal(remoteAddr string) string {
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}
	if remoteAddr == "" {
		return "anonymous"
	}
	return remoteAddr
}

func requiresIdempotencyKey(r *http.Request) bool {
	path := strings.TrimSuffix(r.URL.Path, "/")

	switch path {
	case "/pos/orders/payment", "/pos/stock/v1/transfer", "/finance/invoices/payment":
		return true
	}

	return (strings.HasPrefix(path, "/pos/orders/") && strings.HasSuffix(path, "/payment")) ||
		(strings.HasPrefix(path, "/finance/invoices/payments/v1/invoices/") && strings.HasSuffix(path, "/payment")) ||
		(strings.HasPrefix(path, "/finance/invoices/v1/invoices/") && strings.HasSuffix(path, "/send")) ||
		path == "/erp/stock-transfer/v1/stock-transfers" ||
		(strings.HasPrefix(path, "/erp/stock-transfer/v1/stock-transfers/") && strings.HasSuffix(path, "/status")) ||
		(strings.HasPrefix(path, "/pos/orders/") && strings.HasSuffix(path, "/pay"))
}

// idempotencyRequestFingerprint binds a key to one specific request, so a key
// reused with a different payload is rejected instead of answered with the first
// request's response.
//
// The body is fingerprinted after content-decoding, so a client that gzips and
// one that does not produce the same fingerprint for the same logical request,
// while a genuine body change still produces a different one. Method, path and
// query are included; the tenant and principal are already part of the store
// key. r.Body is left readable by the handler.
func idempotencyRequestFingerprint(r *http.Request) string {
	digest := sha256.New()
	writeFingerprintPart(digest, r.Method)
	writeFingerprintPart(digest, r.URL.Path)
	writeFingerprintPart(digest, r.URL.RawQuery)

	raw, ok := readIdempotencyRequestBody(r)
	if !ok {
		// The body could not be buffered, so no fingerprint can be trusted.
		// Returning "" disables the reuse check rather than comparing a value
		// that could collide.
		return ""
	}

	decoded, err := decodeIdempotencyRequestBody(r.Header.Get("Content-Encoding"), raw)
	if err != nil {
		decoded = raw
	}
	if _, err := digest.Write(decoded); err != nil {
		return ""
	}

	return hex.EncodeToString(digest.Sum(nil))
}

func writeFingerprintPart(digest io.Writer, part string) {
	// Length-prefixed so two different (method, path, query) triples cannot
	// produce the same concatenated input.
	_, _ = fmt.Fprintf(digest, "%d:%s\x00", len(part), part)
}

// readIdempotencyRequestBody buffers the body and puts it back, so the handler
// still reads the original wire bytes. Bounded by
// idempotencyMaxFingerprintBody; RequestBodyLimitMiddleware runs ahead of this
// middleware and caps the body lower.
func readIdempotencyRequestBody(r *http.Request) ([]byte, bool) {
	if r.Body == nil {
		return nil, true
	}

	raw, err := io.ReadAll(io.LimitReader(r.Body, idempotencyMaxFingerprintBody+1))
	_ = r.Body.Close()
	if err != nil {
		return nil, false
	}

	if len(raw) > idempotencyMaxFingerprintBody {
		// Too large to fingerprint. Drain the remainder so the handler is not
		// handed a silently truncated body, then report it as unfingerprintable.
		if rest, restErr := io.ReadAll(io.LimitReader(r.Body, idempotencyMaxFingerprintBody+1)); restErr == nil {
			raw = append(raw, rest...)
		}
		r.Body = io.NopCloser(bytes.NewReader(raw))

		return nil, false
	}

	r.Body = io.NopCloser(bytes.NewReader(raw))

	return raw, true
}

func decodeIdempotencyRequestBody(contentEncoding string, raw []byte) ([]byte, error) {
	if len(raw) == 0 {
		return raw, nil
	}

	var decoder io.ReadCloser
	var err error
	switch strings.ToLower(strings.TrimSpace(contentEncoding)) {
	case "gzip", "x-gzip":
		decoder, err = gzip.NewReader(bytes.NewReader(raw))
	case "deflate":
		decoder, err = zlib.NewReader(bytes.NewReader(raw))
	default:
		// No stdlib decoder (br, zstd) or an identity request. Fingerprint the
		// raw bytes, which is still content-sensitive and stable per client.
		return nil, errUnsupportedContentEncoding
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = decoder.Close() }()

	return io.ReadAll(io.LimitReader(decoder, idempotencyMaxFingerprintBody))
}

var errUnsupportedContentEncoding = errors.New("unsupported content encoding")

func writeIdempotencyInFlight(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", strconv.Itoa(idempotencyRetryAfterSeconds))
	w.WriteHeader(http.StatusConflict)
	if _, writeErr := w.Write([]byte(`{"error":"a request with this Idempotency-Key is still being processed"}`)); writeErr != nil {
		logger.Ctx(r.Context()).Warn("failed to write idempotency conflict response", logger.Error(writeErr))
	}
}

// writeIdempotencyKeyReuse answers a client that reused a key for a different
// request. Replaying the stored response would report a success for a mutation
// the client never asked for.
func writeIdempotencyKeyReuse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusConflict)
	if _, writeErr := w.Write([]byte(`{"error":"this Idempotency-Key was already used with a different request"}`)); writeErr != nil {
		logger.Ctx(r.Context()).Warn("failed to write idempotency key reuse response", logger.Error(writeErr))
	}
}

func writeIdempotencyUnavailable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", strconv.Itoa(idempotencyRetryAfterSeconds))
	w.WriteHeader(http.StatusServiceUnavailable)
	if _, writeErr := w.Write([]byte(`{"error":"idempotency is temporarily unavailable, retry shortly"}`)); writeErr != nil {
		logger.Ctx(r.Context()).Warn("failed to write idempotency unavailable response", logger.Error(writeErr))
	}
}

func storeIdempotencyRecord(ctx context.Context, cacheKey string, rec *responseRecorder, fingerprint string, persist func(context.Context, string, []byte)) {
	body := rec.buf.Bytes()
	entry := idempotencyRecord{
		Status:      rec.status,
		Headers:     idempotencyHeadersForReplay(rec.Header()),
		Body:        body,
		Fingerprint: fingerprint,
	}
	if len(body) > idempotencyMaxStoredBodyBytes {
		logger.Ctx(ctx).Warn("storing idempotency tombstone for oversized response body",
			logger.Int("bytes", len(body)),
			logger.Int("limit", idempotencyMaxStoredBodyBytes),
		)
		entry.Body = nil
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

// idempotencyHeadersForReplay keeps only headers the stored body is
// self-describing by. Content-Encoding is deliberately dropped:
// responseCompressMiddleware wraps this middleware, so it has already set
// Content-Encoding: gzip on the shared header map by the time the recorder runs,
// while the bytes the recorder buffers are the UNCOMPRESSED ones. Persisting
// that header would replay a plain body labelled gzip — and the outer
// compressor skips re-encoding precisely because the header is already set.
// Encoding is re-derived from the replaying client's own Accept-Encoding.
func idempotencyHeadersForReplay(h http.Header) http.Header {
	out := make(http.Header)
	if values := h.Values("Content-Type"); len(values) > 0 {
		out["Content-Type"] = values
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
