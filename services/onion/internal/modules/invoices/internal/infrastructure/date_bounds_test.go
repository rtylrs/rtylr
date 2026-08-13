package infrastructure

import (
	"testing"
	"time"
)

func TestInvoiceSequenceYearBounds(t *testing.T) {
	start, end := invoiceSequenceYearBounds(2024)

	if want := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC); !start.Equal(want) {
		t.Fatalf("start = %s, want %s", start, want)
	}
	if want := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC); !end.Equal(want) {
		t.Fatalf("end = %s, want %s", end, want)
	}
	if !start.AddDate(1, 0, 0).Equal(end) {
		t.Fatal("exclusive end must be the following year")
	}
}
