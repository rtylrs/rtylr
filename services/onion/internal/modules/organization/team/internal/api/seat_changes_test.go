package api

import (
	"context"
	"testing"
	"time"
)

type fakeSeatChangeRecorder struct {
	recorded chan recordSeatChangeResult
}

type recordSeatChangeResult struct {
	businessID int64
	err        error
}

func (f *fakeSeatChangeRecorder) RecordSeatChange(ctx context.Context, businessID int64) error {
	f.recorded <- recordSeatChangeResult{
		businessID: businessID,
		err:        ctx.Err(),
	}
	return nil
}

func TestRecordSeatChangeAsyncIgnoresParentCancellation(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	cancelParent()

	recorder := &fakeSeatChangeRecorder{recorded: make(chan recordSeatChangeResult, 1)}
	recordSeatChangeAsync(parent, recorder, 7)

	select {
	case got := <-recorder.recorded:
		if got.err != nil {
			t.Fatalf("expected seat change context to survive parent cancellation, got %v", got.err)
		}
		if got.businessID != 7 {
			t.Fatalf("expected business id 7, got %d", got.businessID)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for async seat change")
	}
}
