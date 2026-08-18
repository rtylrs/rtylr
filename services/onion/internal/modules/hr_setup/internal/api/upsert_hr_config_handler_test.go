package api

import (
	"encoding/json"
	"testing"
)

func TestUpsertHRConfigRequestAcceptsSetupWizardPayload(t *testing.T) {
	var req UpsertHRConfigRequest
	if err := json.Unmarshal([]byte(`{"setupCompletedAt":"1755439200"}`), &req); err != nil {
		t.Fatalf("unmarshal setup wizard payload: %v", err)
	}
	if req.SetupCompletedAt == nil {
		t.Fatal("expected setup_completed_at to be populated")
	}
	if got, want := *req.SetupCompletedAt, int64(1755439200); got != want {
		t.Fatalf("setup_completed_at = %d, want %d", got, want)
	}
}

func TestUpsertHRConfigRequestAcceptsLegacyPayload(t *testing.T) {
	var req UpsertHRConfigRequest
	if err := json.Unmarshal([]byte(`{"setup_completed_at":1755439200}`), &req); err != nil {
		t.Fatalf("unmarshal legacy payload: %v", err)
	}
	if req.SetupCompletedAt == nil {
		t.Fatal("expected setup_completed_at to be populated")
	}
	if got, want := *req.SetupCompletedAt, int64(1755439200); got != want {
		t.Fatalf("setup_completed_at = %d, want %d", got, want)
	}
}
