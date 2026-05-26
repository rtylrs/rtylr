package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/activity/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListActivity_ReturnsSliceForOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListActivity(ctx, testutil.OwnerID, 50)
	if err != nil {
		t.Fatalf("ListActivity: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
}

func TestListActivity_DefaultLimitApplied(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// limit=0 should use default of 50
	items, err := repo.ListActivity(ctx, testutil.OwnerID, 0)
	if err != nil {
		t.Fatalf("ListActivity(limit=0): %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestListActivity_NegativeLimitApplied(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListActivity(ctx, testutil.OwnerID, -1)
	if err != nil {
		t.Fatalf("ListActivity(limit=-1): %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestListActivity_MaxLimitCapped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// limit > 200 should be capped at 200
	items, err := repo.ListActivity(ctx, testutil.OwnerID, 9999)
	if err != nil {
		t.Fatalf("ListActivity(limit=9999): %v", err)
	}
	if len(items) > 200 {
		t.Fatalf("expected at most 200 items, got %d", len(items))
	}
}

func TestListActivity_WrongOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListActivity(ctx, 999999, 50)
	if err != nil {
		t.Fatalf("ListActivity wrong owner: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items for unknown owner, got %d", len(items))
	}
}

func TestListActivity_LimitOne(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListActivity(ctx, testutil.OwnerID, 1)
	if err != nil {
		t.Fatalf("ListActivity(limit=1): %v", err)
	}
	if len(items) > 1 {
		t.Fatalf("expected at most 1 item, got %d", len(items))
	}
}

func TestListActivity_ItemsHaveOwnerID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListActivity(ctx, testutil.OwnerID, 50)
	if err != nil {
		t.Fatalf("ListActivity: %v", err)
	}
	for i, item := range items {
		if item.OwnerID != testutil.OwnerID {
			t.Errorf("item[%d].OwnerID = %d, want %d", i, item.OwnerID, testutil.OwnerID)
		}
	}
}

func TestListActivity_ItemsHaveNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListActivity(ctx, testutil.OwnerID, 50)
	if err != nil {
		t.Fatalf("ListActivity: %v", err)
	}
	for i, item := range items {
		if item.ID == 0 {
			t.Errorf("item[%d].ID is zero", i)
		}
	}
}

func TestListActivity_ItemsHaveAppSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListActivity(ctx, testutil.OwnerID, 50)
	if err != nil {
		t.Fatalf("ListActivity: %v", err)
	}
	for i, item := range items {
		if item.App == "" {
			t.Errorf("item[%d].App is empty", i)
		}
	}
}

func TestListActivity_ItemsHaveAction(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListActivity(ctx, testutil.OwnerID, 50)
	if err != nil {
		t.Fatalf("ListActivity: %v", err)
	}
	for i, item := range items {
		if item.Action == "" {
			t.Errorf("item[%d].Action is empty", i)
		}
	}
}

func TestListActivity_LimitTen(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.ListActivity(ctx, testutil.OwnerID, 10)
	if err != nil {
		t.Fatalf("ListActivity(limit=10): %v", err)
	}
	if len(items) > 10 {
		t.Fatalf("expected at most 10 items, got %d", len(items))
	}
}
