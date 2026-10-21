package infrastructure

import (
	"testing"
	"time"
)

func TestAdvanceDate(t *testing.T) {
	date := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		frequency string
		want      time.Time
	}{
		{frequency: "weekly", want: time.Date(2026, time.February, 7, 0, 0, 0, 0, time.UTC)},
		{frequency: "monthly", want: time.Date(2026, time.February, 28, 0, 0, 0, 0, time.UTC)},
		{frequency: "quarterly", want: time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)},
		{frequency: "yearly", want: time.Date(2027, time.January, 31, 0, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		got, err := advanceDate(date, tt.frequency)
		if err != nil {
			t.Fatalf("advanceDate(%q): %v", tt.frequency, err)
		}
		if !got.Equal(tt.want) {
			t.Errorf("advanceDate(%q) = %s, want %s", tt.frequency, got.Format("2006-01-02"), tt.want.Format("2006-01-02"))
		}
	}
}

func TestAdvanceDateRejectsUnknownFrequency(t *testing.T) {
	if _, err := advanceDate(time.Now(), "biweekly"); err == nil {
		t.Fatal("expected unknown frequency error")
	}
}
