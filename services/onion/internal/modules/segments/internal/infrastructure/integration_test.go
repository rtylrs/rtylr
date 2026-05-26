package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/segments/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestList_Segments_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	segments, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("segments.List: unexpected error: %v", err)
	}

	// Segments may be empty for the fixture owner; just assert a valid slice is returned.
	if segments == nil {
		t.Fatal("segments.List: expected non-nil slice, got nil")
	}
}

func TestGetByID_Segments_NotFoundReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, testutil.OwnerID, 0)
	if err == nil {
		t.Fatal("segments.GetByID(0): expected error for non-existent segment, got nil")
	}
}
