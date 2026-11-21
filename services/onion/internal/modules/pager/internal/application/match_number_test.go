package application

import (
	"strconv"
	"testing"

	"github.com/voxire/onion/internal/modules/pager/internal/domain"
)

func TestFreeMatchNumberSkipsTakenNumbers(t *testing.T) {
	used := map[string]bool{}
	for n := domain.MinMatchNumber; n <= domain.MaxMatchNumber; n++ {
		if n != 512 {
			used[strconv.Itoa(n)] = true
		}
	}
	code, err := freeMatchNumber(used)
	if err != nil || code != "512" {
		t.Fatalf("freeMatchNumber with only 512 free = %q, %v", code, err)
	}
}

func TestFreeMatchNumberFallsBackToFourDigitsWhenEveryThreeDigitNumberIsTaken(t *testing.T) {
	used := map[string]bool{}
	for n := domain.MinMatchNumber; n <= domain.MaxMatchNumber; n++ {
		used[strconv.Itoa(n)] = true
	}
	// Old codes in the same set never block a number.
	used["panda2"], used["A7"] = true, true
	for range 50 {
		code, err := freeMatchNumber(used)
		if err != nil {
			t.Fatal(err)
		}
		n, err := strconv.Atoi(code)
		if err != nil || len(code) != 4 || n < domain.MinFallbackMatchNumber || n > domain.MaxFallbackMatchNumber {
			t.Fatalf("fallback code %q is not four digits", code)
		}
	}
}
