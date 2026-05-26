package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListLeaveRequests_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	requests, _, err := repo.ListLeaveRequests(ctx, testutil.OwnerID, domain.ListLeaveRequestsParams{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListLeaveRequests: %v", err)
	}
	// Result may be empty -- fixture has no leave requests; just assert no error.
	_ = requests
}

func TestListLeaveTypes_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	types, _, err := repo.ListLeaveTypes(ctx, testutil.OwnerID, domain.ListLeaveTypesParams{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListLeaveTypes: %v", err)
	}
	_ = types
}

func TestListAllActiveLeaveTypes_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	types, err := repo.ListAllActiveLeaveTypesForOwner(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListAllActiveLeaveTypesForOwner: %v", err)
	}
	_ = types
}
