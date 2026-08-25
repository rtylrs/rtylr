package money

import (
	"encoding/json"
	"testing"
)

func TestFromStringUsesMinorUnits(t *testing.T) {
	got, err := FromString("0.1")
	if err != nil {
		t.Fatal(err)
	}
	other, err := FromString("0.2")
	if err != nil {
		t.Fatal(err)
	}
	if got+other != Amount(30) {
		t.Fatalf("got %d cents, want 30", got+other)
	}
}

func TestDollarsToCentsRoundsAtTheBoundary(t *testing.T) {
	if got := DollarsToCents(10.13); got != 1013 {
		t.Fatalf("got %d cents, want 1013", got)
	}
	if got := DollarsToCents(15.19); got != 1519 {
		t.Fatalf("got %d cents, want 1519", got)
	}
}

func TestJSONRoundTripKeepsDecimalShape(t *testing.T) {
	value := Amount(1234)
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "12.34" {
		t.Fatalf("got %s, want 12.34", data)
	}

	var decoded Amount
	if err := json.Unmarshal([]byte("12.34"), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != value {
		t.Fatalf("got %d cents, want %d", decoded, value)
	}
}

func TestScanDecimalString(t *testing.T) {
	var value Amount
	if err := value.Scan([]byte("42.50")); err != nil {
		t.Fatal(err)
	}
	if value != Amount(4250) {
		t.Fatalf("got %d cents, want 4250", value)
	}
}
