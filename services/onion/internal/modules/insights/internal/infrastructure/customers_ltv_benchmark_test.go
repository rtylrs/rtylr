package infrastructure_test

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/voxire/onion/internal/testutil"
)

// BenchmarkGetLTVDistribution measures the database-side, fixed-row LTV
// histogram. Set DB_HOST and LTV_BENCHMARK_BUSINESS_ID to a representative
// large tenant before running:
//
// go test -run '^$' -bench BenchmarkGetLTVDistribution -benchtime=10x ./internal/modules/insights/internal/infrastructure
func BenchmarkGetLTVDistribution(b *testing.B) {
	if os.Getenv("DB_HOST") == "" {
		b.Skip("DB_HOST not set; benchmark requires an integration database")
	}

	businessID := testutil.BusinessID
	if raw := os.Getenv("LTV_BENCHMARK_BUSINESS_ID"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			b.Fatalf("invalid LTV_BENCHMARK_BUSINESS_ID %q", raw)
		}
		businessID = parsed
	}

	db := testutil.OpenDB(b)
	repo := newRepository(db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, err := repo.GetLTVDistribution(ctx, businessID)
		if err != nil {
			b.Fatal(err)
		}
		if len(rows) != 6 {
			b.Fatalf("LTV bucket count = %d, want 6", len(rows))
		}
	}
}
