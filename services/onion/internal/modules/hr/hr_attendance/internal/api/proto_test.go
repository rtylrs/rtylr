package api

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_attendance/internal/domain"
)

func TestAttendanceToProtoIncludesZeroDurationForClosedRecord(t *testing.T) {
	clockOut := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	payload := attendanceToProto(domain.AttendanceRecordRecord{
		ClockInAt:       clockOut.Add(-30 * time.Second),
		ClockOutAt:      clockOut,
		DurationMinutes: 0,
	})

	if payload.DurationMinutes == nil {
		t.Fatal("closed record must expose duration_minutes even when it is zero")
	}
	if *payload.DurationMinutes != 0 {
		t.Fatalf("duration = %d, want 0", *payload.DurationMinutes)
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if !strings.Contains(string(encoded), `"duration_minutes":0`) {
		t.Fatalf("encoded payload omitted zero duration: %s", encoded)
	}
}

func TestAttendanceToProtoOmitsDurationForOpenRecord(t *testing.T) {
	payload := attendanceToProto(domain.AttendanceRecordRecord{
		ClockInAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	})

	if payload.DurationMinutes != nil {
		t.Fatal("open record must omit duration_minutes")
	}
}
