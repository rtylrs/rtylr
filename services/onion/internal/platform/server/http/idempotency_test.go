package http

import (
	"context"
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
