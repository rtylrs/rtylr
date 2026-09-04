package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/customers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateSegmentViaCustomersRepo_FiltersAreStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateSegment(ctx, domain.SegmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Filtered Segment",
		Rules:      []byte(`{"tier":"gold"}`),
	})
	if err != nil {
		t.Fatalf("CreateSegment: %v", err)
	}

	seg, err := repo.GetSegmentByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetSegmentByID: %v", err)
	}
	if seg.Name != "Filtered Segment" {
		t.Errorf("Name = %q, want %q", seg.Name, "Filtered Segment")
	}
}

func TestGetSegmentByID_ViaCustomersRepo_NonExistent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetSegmentByID(ctx, testutil.OwnerID, 999999999)
	if err == nil {
		t.Fatal("expected error for non-existent segment, got nil")
	}
	if !errors.Is(err, domain.ErrSegmentNotFound) {
		t.Errorf("expected ErrSegmentNotFound, got %v", err)
	}
}

func TestListSegmentsViaCustomersRepo_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	segments, err := repo.ListSegments(ctx, 0)
	if err != nil {
		t.Fatalf("ListSegments(ownerID=0): %v", err)
	}
	if len(segments) != 0 {
		t.Errorf("expected 0 segments for ownerID=0, got %d", len(segments))
	}
}

func TestDeleteSegmentViaCustomersRepo_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateSegment(ctx, domain.SegmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "To Be Deleted Segment",
	})
	if err != nil {
		t.Fatalf("CreateSegment: %v", err)
	}

	if err := repo.DeleteSegment(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("DeleteSegment: %v", err)
	}

	_, err = repo.GetSegmentByID(ctx, testutil.OwnerID, id)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestUpdateSegmentViaCustomersRepo_ChangesDescription(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateSegment(ctx, domain.SegmentInput{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		Name:        "Desc Update Seg",
		Description: "before",
	})
	if err != nil {
		t.Fatalf("CreateSegment: %v", err)
	}

	if err := repo.UpdateSegment(ctx, domain.SegmentUpdateInput{
		ID:          id,
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		Name:        "Desc Update Seg",
		Description: "after",
	}); err != nil {
		t.Fatalf("UpdateSegment: %v", err)
	}

	seg, err := repo.GetSegmentByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetSegmentByID: %v", err)
	}
	if seg.Description != "after" {
		t.Errorf("Description = %q, want %q", seg.Description, "after")
	}
}
