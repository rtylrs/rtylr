package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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
