package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/segments/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/segments/internal/infrastructure"
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

func TestListSegments_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	segments, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(segments) != 0 {
		t.Errorf("expected empty slice for owner_id=0, got %d rows", len(segments))
	}
}

func TestCreateSegment_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.SegmentInput{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		Name:        "VIP Customers",
		Description: "High value segment",
		Filters:     []byte(`{}`),
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestCreateSegment_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.SegmentInput{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		Name:        "StoredSegment",
		Description: "",
		Filters:     []byte(`{}`),
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	seg, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID after create: %v", err)
	}
	if seg.Name != "StoredSegment" {
		t.Errorf("expected name='StoredSegment', got %q", seg.Name)
	}
}

func TestGetSegmentByID_Exists(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.SegmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "ExistsSegment",
		Filters:    []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	seg, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if seg.ID != id {
		t.Errorf("expected ID=%d, got %d", id, seg.ID)
	}
}

func TestGetSegmentByID_NotFound_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, testutil.OwnerID, 999999999)
	if err == nil {
		t.Fatal("expected error for non-existent segment id, got nil")
	}
}

func TestGetSegmentByID_WrongOwner_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.SegmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "WrongOwnerSegment",
		Filters:    []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	_, err = repo.GetByID(ctx, 0, id)
	if err == nil {
		t.Errorf("expected error fetching segment id=%d with wrong owner, got nil", id)
	}
}

func TestUpdateSegment_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.SegmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "BeforeUpdateSeg",
		Filters:    []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	if err := repo.Update(ctx, domain.SegmentUpdateInput{
		ID:         id,
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "AfterUpdateSeg",
		Filters:    []byte(`{}`),
	}); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	seg, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if seg.Name != "AfterUpdateSeg" {
		t.Errorf("expected name='AfterUpdateSeg', got %q", seg.Name)
	}
}

func TestDeleteSegment_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.SegmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "DeleteMeSeg",
		Filters:    []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	if err := repo.Delete(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	_, err = repo.GetByID(ctx, testutil.OwnerID, id)
	if err == nil {
		t.Errorf("expected error fetching deleted segment id=%d, got nil", id)
	}
}

func TestDeleteSegment_NonExistentIsNoop(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.Delete(ctx, testutil.OwnerID, 999999999); err != nil {
		t.Errorf("expected no error deleting non-existent segment, got: %v", err)
	}
}

func TestListSegments_AfterCreate_ContainsNew(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.SegmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "NewlyCreatedSeg",
		Filters:    []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	segments, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, s := range segments {
		if s.ID == id {
			return
		}
	}
	t.Fatalf("newly created segment id=%d not found in list", id)
}

func TestCreate_ThenGet_Roundtrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.SegmentInput{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		Name:        "RoundtripSeg",
		Description: "roundtrip desc",
		Filters:     []byte(`{"key":"value"}`),
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	seg, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if seg.Name != "RoundtripSeg" {
		t.Errorf("expected name='RoundtripSeg', got %q", seg.Name)
	}
	if seg.Description != "roundtrip desc" {
		t.Errorf("expected description='roundtrip desc', got %q", seg.Description)
	}
}
