package api

import (
	"context"
	"testing"
	"time"
)

type fakeSeatChangeRecorder struct {
	ownerID    int64
	recorded   chan recordSeatChangeResult
	actorID    int64
	businessID int64
}

type recordSeatChangeResult struct {
	ownerID int64
	err     error
}

func (f *fakeSeatChangeRecorder) BusinessOwnerID(ctx context.Context, fallbackOwnerID, businessID int64) (int64, error) {
	f.actorID = fallbackOwnerID
	f.businessID = businessID
	return f.ownerID, ctx.Err()
}

func (f *fakeSeatChangeRecorder) RecordSeatChange(ctx context.Context, ownerID int64) error {
	f.recorded <- recordSeatChangeResult{
		ownerID: ownerID,
		err:     ctx.Err(),
	}
	return nil
}

func TestRecordSeatChangeAsyncIgnoresParentCancellation(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	cancelParent()

	recorder := &fakeSeatChangeRecorder{
		ownerID:  99,
		recorded: make(chan recordSeatChangeResult, 1),
	}
	recordSeatChangeAsync(parent, recorder, 42, 7)

	select {
	case got := <-recorder.recorded:
		if got.err != nil {
			t.Fatalf("expected seat change context to survive parent cancellation, got %v", got.err)
		}
		if got.ownerID != 99 {
			t.Fatalf("expected owner id 99, got %d", got.ownerID)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for async seat change")
	}

	if recorder.actorID != 42 {
		t.Fatalf("expected actor id 42, got %d", recorder.actorID)
	}
	if recorder.businessID != 7 {
		t.Fatalf("expected business id 7, got %d", recorder.businessID)
	}
}
