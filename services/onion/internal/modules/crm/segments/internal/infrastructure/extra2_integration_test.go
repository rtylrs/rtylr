package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/segments/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/segments/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateSegment_PositiveIDReturned(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.SegmentInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Positive ID Seg",
		Filters:    []byte(`{}`),
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
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "WrongOwner Get Seg",
		Filters:    []byte(`{}`),
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
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Delete Then Get Seg",
		Filters:    []byte(`{}`),
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
