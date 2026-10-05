package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/segments/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/segments/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateSegment_DescriptionIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.SegmentInput{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		Name:        "Desc Stored Seg",
		Description: "a test description",
		Filters:     []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	seg, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if seg.Description != "a test description" {
		t.Errorf("Description = %q, want %q", seg.Description, "a test description")
	}
}

func TestCreateSegment_OwnerIDIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.SegmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "OwnerID Stored Seg",
		Filters:    []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	seg, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if seg.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", seg.OwnerID, testutil.OwnerID)
	}
}

func TestCreateSegment_IDsAreUnique(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.Create(ctx, domain.SegmentInput{OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "UniqueA", Filters: []byte(`{}`)})
	if err != nil {
		t.Fatalf("Create A: %v", err)
	}
	id2, err := repo.Create(ctx, domain.SegmentInput{OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "UniqueB", Filters: []byte(`{}`)})
	if err != nil {
		t.Fatalf("Create B: %v", err)
	}
	if id1 == id2 {
		t.Errorf("expected unique IDs, both got %d", id1)
	}
}

func TestUpdateSegment_ChangesDescription(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.SegmentInput{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		Name:        "Desc Update Seg",
		Description: "before desc",
		Filters:     []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(ctx, domain.SegmentUpdateInput{
		ID:          id,
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		Name:        "Desc Update Seg",
		Description: "after desc",
		Filters:     []byte(`{}`),
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	seg, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if seg.Description != "after desc" {
		t.Errorf("Description = %q, want %q", seg.Description, "after desc")
	}
}

func TestListSegments_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	_, err = repo.Create(ctx, domain.SegmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Count Check Seg",
		Filters:    []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	after, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Errorf("expected count +1: before=%d after=%d", len(before), len(after))
	}
}

func TestListSegments_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, domain.SegmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Owner Check Seg",
		Filters:    []byte(`{}`),
	})

	segments, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, s := range segments {
		if s.OwnerID != testutil.OwnerID {
			t.Errorf("segment %d has owner_id %d, want %d", s.ID, s.OwnerID, testutil.OwnerID)
		}
	}
}

func TestUpdateSegment_WrongOwner_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.SegmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "WrongOwner Update Seg",
		Filters:    []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = repo.Update(ctx, domain.SegmentUpdateInput{
		ID:      id,
		OwnerID: 999999,
		Name:    "Sneaky Update",
		Filters: []byte(`{}`),
	})
	if err == nil {
		t.Fatal("expected error updating segment with wrong owner, got nil")
	}
}
