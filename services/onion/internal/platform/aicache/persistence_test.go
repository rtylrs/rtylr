package aicache

import (
	"testing"
	"time"
)

func TestRecordUpdatesInProcessCollectorWithoutWaitingForDurablePersistence(t *testing.T) {
	ResetForTest()
	started := time.Now()
	Record(Event{Outcome: OutcomeWrite, Mode: ModeAsk, ResponseChars: 12})
	snapshot := Snapshot()
	if snapshot.Totals.Writes != 1 {
		t.Fatalf("writes = %d, want 1", snapshot.Totals.Writes)
	}
	if time.Since(started) > 100*time.Millisecond {
		t.Fatal("Record should not wait on durable analytics backends")
	}
}
