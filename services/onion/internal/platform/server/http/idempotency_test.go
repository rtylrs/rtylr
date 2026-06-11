package http

import (
	"context"
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
