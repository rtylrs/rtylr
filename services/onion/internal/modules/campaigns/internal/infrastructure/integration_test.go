package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/campaigns/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestList_Campaigns_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	campaigns, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("campaigns.List: unexpected error: %v", err)
	}

	// Campaigns may be empty for the fixture owner; assert a valid slice is returned.
	if campaigns == nil {
		t.Fatal("campaigns.List: expected non-nil slice, got nil")
	}
}

func TestGetByID_Campaigns_NotFoundReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, testutil.OwnerID, 0)
	if err == nil {
		t.Fatal("campaigns.GetByID(0): expected error for non-existent campaign, got nil")
	}
}
