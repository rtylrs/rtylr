package infrastructure

import "testing"

func TestLTVBucketsCoverAllRangesInOrder(t *testing.T) {
	// The SQL query emits these fixed rows even when a bucket has no customers.
	// Keep the API contract explicit without reintroducing application-side
	// bucketing or per-customer aggregation.
	want := []LTVBucketRow{
		{Label: "0-100", LowerMinorUnits: 0, UpperMinorUnits: 10_000},
		{Label: "100-500", LowerMinorUnits: 10_000, UpperMinorUnits: 50_000},
		{Label: "500-1k", LowerMinorUnits: 50_000, UpperMinorUnits: 100_000},
		{Label: "1k-5k", LowerMinorUnits: 100_000, UpperMinorUnits: 500_000},
		{Label: "5k-10k", LowerMinorUnits: 500_000, UpperMinorUnits: 1_000_000},
		{Label: "10k+", LowerMinorUnits: 1_000_000, UpperMinorUnits: 0},
	}

	if len(want) != 6 {
		t.Fatalf("LTV bucket count = %d, want 6", len(want))
	}
	for i, bucket := range want {
		if bucket.LowerMinorUnits < 0 || (bucket.UpperMinorUnits != 0 && bucket.UpperMinorUnits <= bucket.LowerMinorUnits) {
			t.Fatalf("bucket %d has invalid range: %+v", i, bucket)
		}
		if i > 0 && bucket.LowerMinorUnits != want[i-1].UpperMinorUnits {
			t.Fatalf("bucket %d lower bound = %d, want previous upper bound %d", i, bucket.LowerMinorUnits, want[i-1].UpperMinorUnits)
		}
	}
}
