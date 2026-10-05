package cdn

import "testing"

func TestCacheKey_Deterministic(t *testing.T) {
	a := cacheKey("file", "256", "256", "cover", "webp", "85")
	b := cacheKey("file", "256", "256", "cover", "webp", "85")
	if a != b {
		t.Errorf("cacheKey not deterministic: %q != %q", a, b)
	}
}

func TestCacheKey_Length(t *testing.T) {
	if got := len(cacheKey("x")); got != 32 {
		t.Errorf("cacheKey length = %d, want 32", got)
	}
}

// The null separator must make part boundaries unambiguous: ("1","23") and
// ("12","3") concatenate to the same string but must hash differently.
func TestCacheKey_SeparatorPreventsCollision(t *testing.T) {
	if cacheKey("1", "23") == cacheKey("12", "3") {
		t.Error("cacheKey collided across different part boundaries")
	}
}

func TestCacheKey_DistinctInputsDistinctOutputs(t *testing.T) {
	if cacheKey("a") == cacheKey("b") {
		t.Error("different inputs produced the same cache key")
	}
}
