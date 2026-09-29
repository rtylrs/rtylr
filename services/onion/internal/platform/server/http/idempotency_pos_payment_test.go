package http

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// The POS payment route the generated oneKit client calls
// (voxie/gen/ts/pos/orders/payments/v1/client.ts). The client supplies
// Idempotency-Key per user action; see clients/pos/src/features/payment/ for
// the client-side half, which is covered by wire-level tests there.
const generatedPosPaymentBody = `{"order_id":42,"payment_method":"card","amount_paid":50}`

func generatedPosPaymentRequest(t *testing.T) *http.Request {
	t.Helper()
	r := httptest.NewRequest(
		http.MethodPost,
		"/pos/orders/payments/v1/orders/42/payment",
		bytes.NewReader([]byte(generatedPosPaymentBody)),
	)
	r.Header.Set("Content-Type", "application/json")
	return r
}

func countingPosPaymentHandler(t *testing.T, executed *int) http.Handler {
	t.Helper()
	return IdempotencyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*executed++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"order_id":42,"status":"paid"}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
}

// TestPosPaymentRequiresIdempotencyKey pins the money-path contract from the
// server side: a payment that arrives without a key is refused before the
// handler runs. This is the property that must not be relaxed — a payment
// admitted without a dedup key can be charged twice by a blind retry.
func TestPosPaymentRequiresIdempotencyKey(t *testing.T) {
	executed := 0
	rec := httptest.NewRecorder()
	countingPosPaymentHandler(t, &executed).ServeHTTP(rec, generatedPosPaymentRequest(t))

	if rec.Code != http.StatusPreconditionRequired {
		t.Fatalf("payment without Idempotency-Key returned %d (%s), want 428",
			rec.Code, rec.Body.String())
	}
	if executed != 0 {
		t.Fatalf("handler executed %d times on a refused request, want 0", executed)
	}
}

// TestPosPaymentWithIdempotencyKeyReachesHandler is the other half: the same
// request carrying a key is admitted and runs exactly once.
func TestPosPaymentWithIdempotencyKeyReachesHandler(t *testing.T) {
	executed := 0
	req := generatedPosPaymentRequest(t)
	req.Header.Set("Idempotency-Key", fmt.Sprintf("pos-admitted-%d", time.Now().UnixNano()))

	rec := httptest.NewRecorder()
	countingPosPaymentHandler(t, &executed).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("payment with Idempotency-Key returned %d (%s), want 200",
			rec.Code, rec.Body.String())
	}
	if executed != 1 {
		t.Fatalf("handler executed %d times, want 1", executed)
	}
}

// TestPosGeneratedClientPaymentReplaysOnRetry proves a retried payment executes
// once. The client reuses the key it generated for the original user action.
func TestPosGeneratedClientPaymentReplaysOnRetry(t *testing.T) {
	// The middleware dedups against the process-local cache when Redis is off,
	// and that cache is a no-op until it is initialized.
	if err := cacher.InitializeLocalCache(context.Background()); err != nil {
		t.Fatalf("InitializeLocalCache: %v", err)
	}
	key := fmt.Sprintf("pos-checkout-%d", time.Now().UnixNano())

	executed := 0
	handler := IdempotencyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		executed++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"order_id":42,"status":"paid"}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))

	first := httptest.NewRecorder()
	firstReq := generatedPosPaymentRequest(t)
	firstReq.Header.Set("Idempotency-Key", key)
	handler.ServeHTTP(first, firstReq)

	second := httptest.NewRecorder()
	secondReq := generatedPosPaymentRequest(t)
	secondReq.Header.Set("Idempotency-Key", key)
	handler.ServeHTTP(second, secondReq)

	if first.Code != http.StatusOK || second.Code != http.StatusOK {
		t.Fatalf("statuses = %d/%d, want 200/200 (%s)", first.Code, second.Code, second.Body.String())
	}
	if executed != 1 {
		t.Fatalf("retried payment executed handler %d times, want 1", executed)
	}
	if second.Header().Get("Idempotent-Replayed") != "true" {
		t.Fatal("retry must be answered from the stored response")
	}
	if first.Body.String() != second.Body.String() {
		t.Fatalf("replay body %q differs from original %q", second.Body.String(), first.Body.String())
	}
}
