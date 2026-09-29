package http

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
)

// idempotencyMemoryStore is a real (mutex-guarded) stand-in for Redis so the
// reservation is actually atomic the way SET NX is.
type idempotencyMemoryStore struct {
	mu        sync.Mutex
	entries   map[string][]byte
	claimTTLs map[string]time.Duration
}

func newIdempotencyMemoryStore() *idempotencyMemoryStore {
	return &idempotencyMemoryStore{
		entries:   make(map[string][]byte),
		claimTTLs: make(map[string]time.Duration),
	}
}

func (s *idempotencyMemoryStore) SetNx(_ context.Context, key string, val []byte, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.entries[key]; exists {
		return false
	}
	s.entries[key] = append([]byte(nil), val...)
	s.claimTTLs[key] = ttl

	return true
}

func (s *idempotencyMemoryStore) Get(_ context.Context, key string) []byte {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.entries[key]
}

func (s *idempotencyMemoryStore) Put(_ context.Context, key string, val []byte, _ time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[key] = append([]byte(nil), val...)
}

func (s *idempotencyMemoryStore) PutConfirmed(ctx context.Context, key string, val []byte, ttl time.Duration) bool {
	s.Put(ctx, key, val, ttl)

	return true
}

func (s *idempotencyMemoryStore) Delete(_ context.Context, keys []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, key := range keys {
		delete(s.entries, key)
		delete(s.claimTTLs, key)
	}
}

func (s *idempotencyMemoryStore) claimTTL(key string) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.claimTTLs[key]
}

func withIdempotencyAuth(r *http.Request, userID, businessID int64) *http.Request {
	ctx := context.WithValue(r.Context(), authPkg.AuthUserKey, &authPkg.AuthUser{
		UserID:     userID,
		BusinessID: businessID,
	})

	return r.WithContext(ctx)
}

func postPayment(t *testing.T, body []byte) *http.Request {
	t.Helper()

	return httptest.NewRequest(http.MethodPost, "/pos/orders/42/payment", bytes.NewReader(body))
}

// A second requester must never execute the handler while the first one holds
// the claim, and must get the stored response once the first one finishes.
func TestIdempotencyClaimSerializesConcurrentSameKeyRequests(t *testing.T) {
	store := newIdempotencyMemoryStore()

	started := make(chan struct{})
	release := make(chan struct{})
	var runs int32
	var mu sync.Mutex

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		runs++
		mu.Unlock()
		close(started)
		<-release
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(`{"paid":true}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	})

	const cacheKey = "idem:b1:u7:concurrent"

	firstRec := httptest.NewRecorder()
	firstDone := make(chan struct{})
	go func() {
		defer close(firstDone)
		serveIdempotentRedisWithStore(firstRec, postPayment(t, []byte(`{"amount":10}`)), next, cacheKey, store)
	}()

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("first request never reached the handler")
	}

	secondRec := httptest.NewRecorder()
	serveIdempotentRedisWithStore(secondRec, postPayment(t, []byte(`{"amount":10}`)), next, cacheKey, store)

	if secondRec.Code != http.StatusConflict {
		t.Fatalf("in-flight duplicate status = %d, want %d", secondRec.Code, http.StatusConflict)
	}
	if secondRec.Header().Get("Retry-After") == "" {
		t.Fatal("in-flight duplicate must carry a Retry-After hint")
	}
	mu.Lock()
	runsAfterDuplicate := runs
	mu.Unlock()
	if runsAfterDuplicate != 1 {
		t.Fatalf("handler runs during in-flight duplicate = %d, want 1", runsAfterDuplicate)
	}

	close(release)
	<-firstDone

	thirdRec := httptest.NewRecorder()
	serveIdempotentRedisWithStore(thirdRec, postPayment(t, []byte(`{"amount":10}`)), next, cacheKey, store)

	if thirdRec.Code != http.StatusCreated {
		t.Fatalf("replay status = %d, want %d", thirdRec.Code, http.StatusCreated)
	}
	if thirdRec.Header().Get("Idempotent-Replayed") != "true" {
		t.Fatal("replay after completion was not tagged as replayed")
	}
	if thirdRec.Body.String() != firstRec.Body.String() {
		t.Fatalf("replay body = %q, want %q", thirdRec.Body.String(), firstRec.Body.String())
	}

	mu.Lock()
	defer mu.Unlock()
	if runs != 1 {
		t.Fatalf("handler runs = %d, want 1 (double-applied mutation)", runs)
	}
}

// A claim is a lease, not the retention window: a holder that dies mid-flight
// must not wedge the key for the full 30 minute response TTL.
func TestIdempotencyClaimLeaseIsShorterThanResponseRetention(t *testing.T) {
	store := newIdempotencyMemoryStore()

	blocked := make(chan struct{})
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		<-blocked
		w.WriteHeader(http.StatusOK)
	})

	const cacheKey = "idem:b1:u7:lease"

	done := make(chan struct{})
	go func() {
		defer close(done)
		serveIdempotentRedisWithStore(httptest.NewRecorder(), postPayment(t, []byte(`{"amount":1}`)), next, cacheKey, store)
	}()

	// Wait for the claim to land.
	deadline := time.Now().Add(2 * time.Second)
	for store.Get(context.Background(), cacheKey) == nil && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if store.Get(context.Background(), cacheKey) == nil {
		t.Fatal("no claim was written")
	}

	if claimTTL := store.claimTTL(cacheKey); claimTTL != idempotencyClaimTTL {
		t.Fatalf("claim TTL = %s, want %s", claimTTL, idempotencyClaimTTL)
	}
	if idempotencyClaimTTL >= idempotencyTTL {
		t.Fatalf("claim TTL %s must be shorter than the response retention %s", idempotencyClaimTTL, idempotencyTTL)
	}
	if idempotencyClaimTTL <= RequestTimeout {
		t.Fatalf("claim TTL %s must exceed the %s request deadline or a live request loses its claim", idempotencyClaimTTL, RequestTimeout)
	}

	close(blocked)
	<-done
}

// gzip and plain encodings of the same logical request must fingerprint
// identically, and the handler must still receive the untouched wire bytes.
func TestIdempotencyFingerprintIgnoresContentEncoding(t *testing.T) {
	payload := []byte(`{"amount":1234,"currency":"EUR"}`)

	var gzipped bytes.Buffer
	writer := gzip.NewWriter(&gzipped)
	if _, err := writer.Write(payload); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}

	plainReq := postPayment(t, payload)
	compressedReq := postPayment(t, gzipped.Bytes())
	compressedReq.Header.Set("Content-Encoding", "gzip")

	plainFingerprint := idempotencyRequestFingerprint(plainReq)
	compressedFingerprint := idempotencyRequestFingerprint(compressedReq)

	if plainFingerprint == "" || compressedFingerprint == "" {
		t.Fatalf("fingerprints = %q / %q, want both computed", plainFingerprint, compressedFingerprint)
	}
	if plainFingerprint != compressedFingerprint {
		t.Fatalf("gzip and plain fingerprints differ:\n plain = %s\n gzip  = %s", plainFingerprint, compressedFingerprint)
	}

	restored, err := io.ReadAll(compressedReq.Body)
	if err != nil {
		t.Fatalf("read restored body: %v", err)
	}
	if !bytes.Equal(restored, gzipped.Bytes()) {
		t.Fatal("fingerprinting must hand the handler the original wire bytes, not the decoded ones")
	}

	restoredPlain, err := io.ReadAll(plainReq.Body)
	if err != nil {
		t.Fatalf("read restored plain body: %v", err)
	}
	if !bytes.Equal(restoredPlain, payload) {
		t.Fatalf("plain body after fingerprinting = %q, want %q", restoredPlain, payload)
	}
}

func TestIdempotencyFingerprintStillSeparatesGenuineDifferences(t *testing.T) {
	base := postPayment(t, []byte(`{"amount":1234}`))
	baseFingerprint := idempotencyRequestFingerprint(base)

	otherAmount := idempotencyRequestFingerprint(postPayment(t, []byte(`{"amount":1235}`)))
	if otherAmount == baseFingerprint {
		t.Fatal("a different body must produce a different fingerprint")
	}

	otherPath := httptest.NewRequest(http.MethodPost, "/pos/orders/43/payment", bytes.NewReader([]byte(`{"amount":1234}`)))
	if idempotencyRequestFingerprint(otherPath) == baseFingerprint {
		t.Fatal("a different path must produce a different fingerprint")
	}

	otherQuery := httptest.NewRequest(http.MethodPost, "/pos/orders/42/payment?mode=split", bytes.NewReader([]byte(`{"amount":1234}`)))
	if idempotencyRequestFingerprint(otherQuery) == baseFingerprint {
		t.Fatal("a different query must produce a different fingerprint")
	}

	otherMethod := httptest.NewRequest(http.MethodPatch, "/pos/orders/42/payment", bytes.NewReader([]byte(`{"amount":1234}`)))
	if idempotencyRequestFingerprint(otherMethod) == baseFingerprint {
		t.Fatal("a different method must produce a different fingerprint")
	}
}

// Reusing a key for a different body must not hand back the first response.
func TestIdempotencyRejectsKeyReuseWithDifferentBody(t *testing.T) {
	store := newIdempotencyMemoryStore()

	var runs int
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		runs++
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(`{"paid":true}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	})

	const cacheKey = "idem:b1:u7:reuse"

	firstRec := httptest.NewRecorder()
	serveIdempotentRedisWithStore(firstRec, postPayment(t, []byte(`{"amount":10}`)), next, cacheKey, store)

	secondRec := httptest.NewRecorder()
	serveIdempotentRedisWithStore(secondRec, postPayment(t, []byte(`{"amount":999}`)), next, cacheKey, store)

	if secondRec.Code != http.StatusConflict {
		t.Fatalf("key reuse status = %d, want %d", secondRec.Code, http.StatusConflict)
	}
	if secondRec.Body.String() == firstRec.Body.String() {
		t.Fatal("key reuse replayed the first response instead of rejecting the mismatch")
	}
	if runs != 1 {
		t.Fatalf("handler runs = %d, want 1", runs)
	}
}

// A user can belong to several businesses, so the store key must include the
// tenant: otherwise the same client key in two tenants replays across them.
func TestIdempotencyKeyIsScopedPerTenant(t *testing.T) {
	tenantA := withIdempotencyAuth(postPayment(t, nil), 7, 1)
	tenantB := withIdempotencyAuth(postPayment(t, nil), 7, 2)

	headerA := tenantA.Clone(tenantA.Context())
	headerA.Header.Set(idempotencyHeader, "shared-key")
	headerB := tenantB.Clone(tenantB.Context())
	headerB.Header.Set(idempotencyHeader, "shared-key")

	if buildIdempotencyKey(headerA, "shared-key") == buildIdempotencyKey(headerB, "shared-key") {
		t.Fatalf("two tenants share the idempotency key %q", buildIdempotencyKey(headerA, "shared-key"))
	}
}

func TestIdempotencyKeyIsScopedPerTenantEndToEnd(t *testing.T) {
	store := newIdempotencyMemoryStore()

	var mu sync.Mutex
	bodies := map[int64]string{}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authUser, ok := authPkg.GetAuthUser(r.Context())
		if !ok {
			t.Error("handler lost the authenticated principal")
			return
		}
		mu.Lock()
		bodies[authUser.BusinessID] = `tenant ` + string(rune('0'+authUser.BusinessID))
		mu.Unlock()
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(`{"business":` + string(rune('0'+authUser.BusinessID)) + `}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	})

	keyA := withIdempotencyAuth(postPayment(t, nil), 7, 1)
	keyB := withIdempotencyAuth(postPayment(t, nil), 7, 2)

	recA := httptest.NewRecorder()
	serveIdempotentRedisWithStore(recA, keyA, next, buildIdempotencyKey(keyA, "shared-key"), store)
	recB := httptest.NewRecorder()
	serveIdempotentRedisWithStore(recB, keyB, next, buildIdempotencyKey(keyB, "shared-key"), store)

	if recA.Body.String() != `{"business":1}` {
		t.Fatalf("tenant 1 body = %q", recA.Body.String())
	}
	if recB.Body.String() != `{"business":2}` {
		t.Fatalf("tenant 2 replayed another tenant's response: %q", recB.Body.String())
	}
	if recB.Header().Get("Idempotent-Replayed") != "" {
		t.Fatal("tenant 2 was served a replay instead of its own stored response")
	}
}

// A record written before the fingerprint existed must still replay.
func TestIdempotencyReplaysRecordsWithoutFingerprint(t *testing.T) {
	store := newIdempotencyMemoryStore()

	legacy := idempotencyRecord{Status: http.StatusCreated, Body: []byte(`{"legacy":true}`)}
	payload, err := json.Marshal(legacy)
	if err != nil {
		t.Fatalf("marshal legacy record: %v", err)
	}
	store.Put(context.Background(), "idem:b1:u7:legacy", payload, time.Minute)

	runs := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		runs++
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	serveIdempotentRedisWithStore(rec, postPayment(t, []byte(`{"amount":10}`)), next, "idem:b1:u7:legacy", store)

	if rec.Code != http.StatusCreated || rec.Body.String() != `{"legacy":true}` {
		t.Fatalf("legacy record replay = %d %q", rec.Code, rec.Body.String())
	}
	if runs != 0 {
		t.Fatalf("legacy replay executed the handler %d times", runs)
	}
}

// The stored record must carry the fingerprint so a later reuse is detectable.
func TestStoredIdempotencyRecordCarriesFingerprint(t *testing.T) {
	store := newIdempotencyMemoryStore()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(`{"paid":true}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	})

	const cacheKey = "idem:b1:u7:stored"
	serveIdempotentRedisWithStore(httptest.NewRecorder(), postPayment(t, []byte(`{"amount":10}`)), next, cacheKey, store)

	var stored idempotencyRecord
	if err := json.Unmarshal(store.Get(context.Background(), cacheKey), &stored); err != nil {
		t.Fatalf("stored payload is not an idempotency record: %v", err)
	}
	if stored.Status != http.StatusCreated {
		t.Fatalf("stored status = %d, want %d", stored.Status, http.StatusCreated)
	}
	if stored.Fingerprint == "" {
		t.Fatal("stored record has no fingerprint, so a later key reuse cannot be detected")
	}
}

// A late holder must not delete a takeover claim that already replaced it.
func TestIdempotencyClaimReleaseKeepsSuccessorClaim(t *testing.T) {
	store := newIdempotencyMemoryStore()

	blocked := make(chan struct{})
	started := make(chan struct{})
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-blocked
		w.WriteHeader(http.StatusInternalServerError)
	})

	const cacheKey = "idem:b1:u7:takeover"

	done := make(chan struct{})
	go func() {
		defer close(done)
		serveIdempotentRedisWithStore(httptest.NewRecorder(), postPayment(t, []byte(`{"amount":10}`)), next, cacheKey, store)
	}()
	<-started

	// Simulate the lease expiring and a second attempt taking the key over.
	takeover := []byte(`{"status":0,"claim":"successor"}`)
	store.Put(context.Background(), cacheKey, takeover, idempotencyClaimTTL)

	close(blocked)
	<-done

	if got := store.Get(context.Background(), cacheKey); string(got) != string(takeover) {
		t.Fatalf("claim = %q, want the successor claim %q to survive", got, takeover)
	}
}

// responseCompressMiddleware wraps the idempotency middleware, so it sets
// Content-Encoding on the shared header map before the recorder buffers the
// (uncompressed) body. A stored record that keeps that header replays a plain
// body labelled gzip.
func TestIdempotentReplayDoesNotLieAboutItsContentEncoding(t *testing.T) {
	store := newIdempotencyMemoryStore()
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(`{"paid":true}`)); err != nil {
			t.Errorf("write: %v", err)
		}
	})
	chain := responseCompressMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveIdempotentRedisWithStore(w, r, inner, "idem:b1:u7:encoding", store)
	}))

	live := postPayment(t, []byte(`{"amount":10}`))
	live.Header.Set(idempotencyHeader, "encoding")
	live.Header.Set("Accept-Encoding", "gzip")
	liveRec := httptest.NewRecorder()
	chain.ServeHTTP(liveRec, live)
	if liveRec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("precondition: live response should be gzipped, got %q", liveRec.Header().Get("Content-Encoding"))
	}

	replay := postPayment(t, []byte(`{"amount":10}`))
	replay.Header.Set(idempotencyHeader, "encoding")
	replayRec := httptest.NewRecorder()
	chain.ServeHTTP(replayRec, replay)
	if got := replayRec.Header().Get("Content-Encoding"); got != "" {
		t.Fatalf("replay declared Content-Encoding %q to a client that did not accept it", got)
	}
	if replayRec.Body.String() != `{"paid":true}` {
		t.Fatalf("replay body = %q, want the plain stored body", replayRec.Body.String())
	}

	gzipped := postPayment(t, []byte(`{"amount":10}`))
	gzipped.Header.Set(idempotencyHeader, "encoding")
	gzipped.Header.Set("Accept-Encoding", "gzip")
	gzRec := httptest.NewRecorder()
	chain.ServeHTTP(gzRec, gzipped)
	if gzRec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("gzip replay Content-Encoding = %q", gzRec.Header().Get("Content-Encoding"))
	}
	reader, err := gzip.NewReader(bytes.NewReader(gzRec.Body.Bytes()))
	if err != nil {
		t.Fatalf("gzip replay body is not valid gzip: %v", err)
	}
	decoded, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read gzip replay body: %v", err)
	}
	if string(decoded) != `{"paid":true}` {
		t.Fatalf("gzip replay decoded to %q", decoded)
	}
}
