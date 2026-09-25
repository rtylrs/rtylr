package api

import (
	"encoding/json"
	"testing"
	"time"
)

// A status-only task update must not look like an explicit "clear the due
// date": only a due_at that was actually sent can change the stored value.
func TestAdminTaskRequestDueAtPresence(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantSet  bool
		wantNil  bool
		wantDate string
		wantErr  bool
	}{
		{name: "omitted", body: `{"status":"done"}`, wantSet: false, wantNil: true},
		{name: "empty string clears", body: `{"status":"done","due_at":""}`, wantSet: true, wantNil: true},
		{name: "null clears", body: `{"status":"done","due_at":null}`, wantSet: true, wantNil: true},
		{name: "date", body: `{"status":"done","due_at":"2026-10-01"}`, wantSet: true, wantDate: "2026-10-01"},
		{name: "rfc3339", body: `{"status":"done","due_at":"2026-10-01T09:30:00Z"}`, wantSet: true, wantDate: "2026-10-01T09:30:00Z"},
		{name: "invalid", body: `{"status":"done","due_at":"not-a-date"}`, wantSet: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req adminTaskRequest
			if err := json.Unmarshal([]byte(tt.body), &req); err != nil {
				t.Fatalf("unmarshal %s: %v", tt.body, err)
			}
			in, err := req.taskInput(9)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an invalid due date error")
				}
				return
			}
			if err != nil {
				t.Fatalf("taskInput: %v", err)
			}
			if in.ID != 9 || in.Status != "done" {
				t.Fatalf("taskInput = %#v", in)
			}
			if in.DueAtSet != tt.wantSet {
				t.Fatalf("DueAtSet = %v, want %v", in.DueAtSet, tt.wantSet)
			}
			if tt.wantNil {
				if in.DueAt != nil {
					t.Fatalf("DueAt = %v, want nil", in.DueAt)
				}
				return
			}
			want, err := time.Parse(time.RFC3339, tt.wantDate)
			if err != nil {
				want, err = time.Parse("2006-01-02", tt.wantDate)
			}
			if err != nil {
				t.Fatalf("parse want date %q: %v", tt.wantDate, err)
			}
			if in.DueAt == nil || !in.DueAt.Equal(want) {
				t.Fatalf("DueAt = %v, want %v", in.DueAt, want)
			}
		})
	}
}
