package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/segments/internal/domain"
	"github.com/voxire/onion/internal/modules/segments/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateSegment_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const want = "Premium Buyers"
	id, err := repo.Create(ctx, domain.SegmentInput{
		OwnerID: testutil.OwnerID,
		Name:    want,
		Filters: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	seg, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if seg.Name != want {
		t.Errorf("Name = %q, want %q", seg.Name, want)
	}
}

func TestCreateSegment_PositiveIDReturned(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.SegmentInput{
		OwnerID: testutil.OwnerID,
		Name:    "Positive ID Seg",
		Filters: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive id, got %d", id)
	}
}

func TestGetByID_WrongOwner_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.SegmentInput{
		OwnerID: testutil.OwnerID,
		Name:    "WrongOwner Get Seg",
		Filters: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = repo.GetByID(ctx, 999999, id)
	if err == nil {
		t.Fatal("expected error for wrong owner GetByID, got nil")
	}
	if !errors.Is(err, domain.ErrSegmentNotFound) {
		t.Errorf("expected ErrSegmentNotFound, got %v", err)
	}
}

func TestDeleteSegment_ThenNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.SegmentInput{
		OwnerID: testutil.OwnerID,
		Name:    "Delete Then Get Seg",
		Filters: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = repo.GetByID(ctx, testutil.OwnerID, id)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestUpdateSegment_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.SegmentInput{
		OwnerID: testutil.OwnerID,
		Name:    "Before Seg Name",
		Filters: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(ctx, domain.SegmentUpdateInput{
		ID:      id,
		OwnerID: testutil.OwnerID,
		Name:    "After Seg Name",
		Filters: []byte(`{}`),
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	seg, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if seg.Name != "After Seg Name" {
		t.Errorf("Name = %q, want %q", seg.Name, "After Seg Name")
	}
}

func TestListSegments_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	segments, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("List(ownerID=0): %v", err)
	}
	if len(segments) != 0 {
		t.Errorf("expected 0 segments for ownerID=0, got %d", len(segments))
	}
}
