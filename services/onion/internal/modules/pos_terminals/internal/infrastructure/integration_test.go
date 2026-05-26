package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_terminals/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListTerminals_RunsWithoutError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// businesses.id=1 is owned by owner_id=1001 per fixture data.
	const businessID int64 = 1

	_, err := repo.List(ctx, businessID)
	if err != nil {
		t.Fatalf("unexpected error listing terminals for business %d: %v", businessID, err)
	}
}
