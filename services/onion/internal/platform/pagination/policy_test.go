package pagination

import "testing"

func TestPolicyNormalizeDefaultsAndCapsPageSize(t *testing.T) {
	policy := Policy{DefaultPageSize: 50, MaxPageSize: 500, MaxOffset: 1<<31 - 1}

	got := policy.Normalize(0, 10000)
	if got.Number != 1 || got.Size != 500 || got.Offset != 0 {
		t.Fatalf("normalized page = %+v, want number=1 size=500 offset=0", got)
	}
}

func TestPolicyNormalizeClampsDeepOffset(t *testing.T) {
	policy := Policy{DefaultPageSize: 20, MaxPageSize: 100, MaxOffset: 1000}

	got := policy.Normalize(100, 100)
	if got.Number != 100 || got.Size != 100 || got.Offset != 1000 {
		t.Fatalf("normalized page = %+v, want number=100 size=100 offset=1000", got)
	}
}

func TestPolicyNormalizeAvoidsOffsetOverflow(t *testing.T) {
	policy := Policy{DefaultPageSize: 20, MaxPageSize: 100, MaxOffset: 1000}

	got := policy.Normalize(1<<62, 100)
	if got.Offset != 1000 {
		t.Fatalf("normalized offset = %d, want 1000", got.Offset)
	}
}

func TestPolicyNormalizeAvoidsOverflowWithoutConfiguredOffsetCap(t *testing.T) {
	policy := Policy{DefaultPageSize: 20, MaxPageSize: 100}

	got := policy.Normalize(1<<62, 100)
	if got.Offset != maxInt64 {
		t.Fatalf("normalized offset = %d, want %d", got.Offset, maxInt64)
	}
}

func TestSliceStartSaturatesInsteadOfOverflowing(t *testing.T) {
	for _, tc := range []struct {
		page, size int64
		length     int
		want       int
	}{
		{1, 20, 100, 0},
		{2, 20, 100, 20},
		{6, 20, 100, 100},
		{1 << 62, 20, 100, 100},
		{1 << 62, 1<<62 - 1, 100, 100},
	} {
		if got := SliceStart(tc.page, tc.size, tc.length); got != tc.want {
			t.Fatalf("SliceStart(%d, %d, %d) = %d, want %d", tc.page, tc.size, tc.length, got, tc.want)
		}
	}
}
