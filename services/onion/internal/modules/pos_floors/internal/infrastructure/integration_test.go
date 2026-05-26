package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_floors/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListFloors_RunsWithoutError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// businesses.id=1, branch/location id=3001 per fixture data.
	const businessID int64 = 1
	const branchID int64 = 3001

	_, err := repo.List(ctx, businessID, branchID)
	if err != nil {
		t.Fatalf("unexpected error listing floors for business %d branch %d: %v", businessID, branchID, err)
	}
}
