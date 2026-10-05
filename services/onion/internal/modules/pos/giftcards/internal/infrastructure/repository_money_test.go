package infrastructure

import (
	"math"
	"testing"

	"github.com/voxire/onion/internal/platform/money"
)

func TestGiftCardRedemption_ComposesCentsWithoutFloatDrift(t *testing.T) {
	balance, err := decimalToCents("0.30")
	if err != nil {
		t.Fatal(err)
	}
	amount, err := dollarsToCents(0.10)
	if err != nil {
		t.Fatal(err)
	}

	if got := balance - amount; got != money.Amount(20) {
		t.Fatalf("remaining balance = %d cents, want 20", got)
	}
}

func TestDollarsToCents_RejectsNonFiniteValues(t *testing.T) {
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := dollarsToCents(value); err == nil {
			t.Fatalf("dollarsToCents(%v) accepted a non-finite value", value)
		}
	}
}
