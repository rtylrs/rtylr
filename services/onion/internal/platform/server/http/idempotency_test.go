package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIdempotencyHeadersForReplayKeepsContentTypeOnly(t *testing.T) {
	headers := http.Header{
		"Content-Type":     []string{"application/json"},
		"X-Request-Id":     []string{"abc"},
		"Content-Encoding": []string{"gzip"},
	}
	got := idempotencyHeadersForReplay(headers)
	if got.Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type = %q", got.Get("Content-Type"))
	}
	if got.Get("Content-Encoding") != "gzip" {
		t.Fatalf("Content-Encoding = %q", got.Get("Content-Encoding"))
	}
	if got.Get("X-Request-Id") != "" {
		t.Fatalf("unexpected X-Request-Id header: %q", got.Get("X-Request-Id"))
	}
}

func TestIdempotencyCacheContextIgnoresParentCancellation(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	cancelParent()

	cacheCtx, cancelCache := idempotencyCacheContext(parent)
	defer cancelCache()

	select {
	case <-cacheCtx.Done():
		t.Fatal("cache context should survive parent cancellation")
	default:
	}

	deadline, ok := cacheCtx.Deadline()
	if !ok {
		t.Fatal("expected cache context to have a deadline")
	}
	if timeout := time.Until(deadline); timeout <= 0 || timeout > idempotencyCacheWriteTimeout {
		t.Fatalf("unexpected cache context timeout: %s", timeout)
	}
}

func TestResponseRecorderKeepsFirstStatus(t *testing.T) {
	underlying := httptest.NewRecorder()
	rec := &responseRecorder{ResponseWriter: underlying, status: http.StatusOK}

	rec.WriteHeader(http.StatusCreated)
	rec.WriteHeader(http.StatusInternalServerError)

	if rec.status != http.StatusCreated {
		t.Fatalf("expected first status to be kept, got %d", rec.status)
	}
	if underlying.Code != http.StatusCreated {
		t.Fatalf("expected underlying response to keep first status, got %d", underlying.Code)
	}
}

func TestResponseRecorderWriteLocksStatusOK(t *testing.T) {
	underlying := httptest.NewRecorder()
	rec := &responseRecorder{ResponseWriter: underlying, status: http.StatusOK}

	if _, err := rec.Write([]byte("ok")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	rec.WriteHeader(http.StatusInternalServerError)

	if rec.status != http.StatusOK {
		t.Fatalf("expected implicit OK status to be kept, got %d", rec.status)
	}
	if underlying.Code != http.StatusOK {
		t.Fatalf("expected underlying response to keep OK, got %d", underlying.Code)
	}
}

func TestResponseRecorderPreservesFlusher(t *testing.T) {
	underlying := httptest.NewRecorder()
	rec := &responseRecorder{ResponseWriter: underlying, status: http.StatusOK}

	flusher, ok := any(rec).(http.Flusher)
	if !ok {
		t.Fatal("responseRecorder should preserve http.Flusher")
	}
	flusher.Flush()
	if !underlying.Flushed {
		t.Fatal("Flush did not reach the underlying ResponseWriter")
	}
}

func TestReserveIdempotencyKeyRetriesAfterTransientReservationRace(t *testing.T) {
	store := &idempotencyTestStore{setResults: []bool{false, true}}
	reserved, cached := reserveIdempotencyKey(context.Background(), store, "key", []byte(`{"status":0}`))
	if !reserved || cached != nil {
		t.Fatalf("reserved=%t cached=%q, want successful retry", reserved, cached)
	}
	if store.setCalls != 2 {
		t.Fatalf("SetNx calls = %d, want 2", store.setCalls)
	}
}

func TestReserveIdempotencyKeyDoesNotProceedWithoutReservation(t *testing.T) {
	store := &idempotencyTestStore{setResults: []bool{false, false}}
	reserved, cached := reserveIdempotencyKey(context.Background(), store, "key", []byte(`{"status":0}`))
	if reserved || cached != nil {
		t.Fatalf("reserved=%t cached=%q, want no reservation", reserved, cached)
	}
}

func TestAnonymousIdempotencyPrincipalIgnoresEphemeralPort(t *testing.T) {
	first := anonymousIdempotencyPrincipal("192.0.2.10:41000")
	second := anonymousIdempotencyPrincipal("192.0.2.10:42000")
	if first != second || first != "192.0.2.10" {
		t.Fatalf("principals = %q, %q; want stable IP-only principal", first, second)
	}

	if got := anonymousIdempotencyPrincipal("[2001:db8::10]:41000"); got != "2001:db8::10" {
		t.Fatalf("IPv6 principal = %q", got)
	}
}

func TestLocalIdempotencyReservationSerializesSameKey(t *testing.T) {
	first := acquireLocalIdempotencyReservation("test-key")
	secondAcquired := make(chan struct{})
	secondReleased := make(chan struct{})
	go func() {
		second := acquireLocalIdempotencyReservation("test-key")
		close(secondAcquired)
		second.release()
		close(secondReleased)
	}()

	select {
	case <-secondAcquired:
		t.Fatal("same-key reservation acquired concurrently")
	case <-time.After(20 * time.Millisecond):
	}

	first.release()
	select {
	case <-secondReleased:
	case <-time.After(time.Second):
		t.Fatal("same-key reservation did not release")
	}
}

func TestStoreIdempotencyRecordUsesTombstoneForOversizedBody(t *testing.T) {
	underlying := httptest.NewRecorder()
	rec := &responseRecorder{ResponseWriter: underlying, status: http.StatusCreated}
	largeBody := bytes.Repeat([]byte("x"), idempotencyMaxStoredBodyBytes+1)
	if _, err := rec.Write(largeBody); err != nil {
		t.Fatalf("Write: %v", err)
	}

	var payload []byte
	storeIdempotencyRecord(context.Background(), "key", rec, func(_ context.Context, _ string, got []byte) {
		payload = got
	})
	if len(payload) == 0 {
		t.Fatal("expected oversized response tombstone to be persisted")
	}

	var entry idempotencyRecord
	if err := json.Unmarshal(payload, &entry); err != nil {
		t.Fatalf("unmarshal tombstone: %v", err)
	}
	if entry.Status != http.StatusCreated || len(entry.Body) != 0 {
		t.Fatalf("tombstone = status %d, body %d bytes; want 201 and empty body", entry.Status, len(entry.Body))
	}
}

func TestRequiresIdempotencyKeyForCriticalMutations(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		required bool
	}{
		{name: "order payment", method: http.MethodPost, path: "/pos/orders/123/payment", required: true},
		{name: "paddle webhook", method: http.MethodPost, path: "/pos/payments/webhook", required: false},
		{name: "invoice send", method: http.MethodPost, path: "/finance/invoices/v1/invoices/123/send", required: true},
		{name: "stock transfer", method: http.MethodPost, path: "/erp/stock-transfer/v1/stock-transfers", required: true},
		{name: "ordinary update", method: http.MethodPatch, path: "/finance/invoices/v1/invoices/123/status", required: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.path, nil)
			if got := requiresIdempotencyKey(r); got != tt.required {
				t.Fatalf("requiresIdempotencyKey = %t, want %t", got, tt.required)
			}
		})
	}
}

func TestIdempotencyMiddlewareRejectsCriticalMutationWithoutKey(t *testing.T) {
	nextCalled := false
	handler := IdempotencyMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/finance/invoices/v1/invoices/123/send", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusPreconditionRequired {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusPreconditionRequired)
	}
	if nextCalled {
		t.Fatal("critical mutation reached handler without an idempotency key")
	}
}

type idempotencyTestStore struct {
	setResults []bool
	setCalls   int
	cached     []byte
}

func (s *idempotencyTestStore) SetNx(_ context.Context, _ string, _ []byte, _ time.Duration) bool {
	result := s.setResults[s.setCalls]
	s.setCalls++
	return result
}

func (s *idempotencyTestStore) Get(_ context.Context, _ string) []byte                     { return s.cached }
func (s *idempotencyTestStore) Put(_ context.Context, _ string, _ []byte, _ time.Duration) {}
func (s *idempotencyTestStore) Delete(_ context.Context, _ []string)                       {}
