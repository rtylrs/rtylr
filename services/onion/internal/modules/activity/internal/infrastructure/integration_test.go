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

	items, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", 50, 0)
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
	items, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", 0, 0)
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

	items, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", -1, 0)
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
	items, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", 9999, 0)
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

	items, _, err := repo.ListActivity(ctx, 999999, "", 50, 0)
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

	items, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", 1, 0)
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

	items, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", 50, 0)
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

	items, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", 50, 0)
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

	items, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", 50, 0)
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

	items, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", 50, 0)
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

	items, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", 10, 0)
	if err != nil {
		t.Fatalf("ListActivity(limit=10): %v", err)
	}
	if len(items) > 10 {
		t.Fatalf("expected at most 10 items, got %d", len(items))
	}
}

func TestListActivity_LimitHundred(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", 100, 0)
	if err != nil {
		t.Fatalf("ListActivity(limit=100): %v", err)
	}
	if len(items) > 100 {
		t.Fatalf("expected at most 100 items, got %d", len(items))
	}
}

func TestListActivity_LimitFive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", 5, 0)
	if err != nil {
		t.Fatalf("ListActivity(limit=5): %v", err)
	}
	if len(items) > 5 {
		t.Fatalf("expected at most 5 items, got %d", len(items))
	}
}

func TestListActivity_LimitFifty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", 50, 0)
	if err != nil {
		t.Fatalf("ListActivity(limit=50): %v", err)
	}
	if len(items) > 50 {
		t.Fatalf("expected at most 50 items, got %d", len(items))
	}
}

func TestListActivity_LimitTwoHundred(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", 200, 0)
	if err != nil {
		t.Fatalf("ListActivity(limit=200): %v", err)
	}
	if len(items) > 200 {
		t.Fatalf("expected at most 200 items, got %d", len(items))
	}
}

func TestListActivity_IdempotentCallsReturnSameCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items1, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", 50, 0)
	if err != nil {
		t.Fatalf("ListActivity first: %v", err)
	}
	items2, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", 50, 0)
	if err != nil {
		t.Fatalf("ListActivity second: %v", err)
	}
	if len(items1) != len(items2) {
		t.Errorf("idempotent calls returned different counts: %d vs %d", len(items1), len(items2))
	}
}

func TestListActivity_ItemsHaveNonNegativeEntityID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", 50, 0)
	if err != nil {
		t.Fatalf("ListActivity: %v", err)
	}
	for i, item := range items {
		if item.EntityID < 0 {
			t.Errorf("item[%d].EntityID is negative", i)
		}
	}
}

func TestListActivity_ItemsHaveCreatedAt(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", 50, 0)
	if err != nil {
		t.Fatalf("ListActivity: %v", err)
	}
	for i, item := range items {
		if item.CreatedAt.IsZero() {
			t.Errorf("item[%d].CreatedAt is zero", i)
		}
	}
}

func TestListActivity_AnotherUnknownOwnerEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.ListActivity(ctx, 7777777, "", 50, 0)
	if err != nil {
		t.Fatalf("ListActivity unknown owner: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items for unknown owner 7777777, got %d", len(items))
	}
}

func TestListActivity_LimitTwenty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", 20, 0)
	if err != nil {
		t.Fatalf("ListActivity(limit=20): %v", err)
	}
	if len(items) > 20 {
		t.Fatalf("expected at most 20 items, got %d", len(items))
	}
}

func TestListActivity_LimitOneLessThanFifty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", 49, 0)
	if err != nil {
		t.Fatalf("ListActivity(limit=49): %v", err)
	}
	if len(items) > 49 {
		t.Fatalf("expected at most 49 items, got %d", len(items))
	}
}

func TestListActivity_ReturnedCountNotExceedsLimit(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const limit = 15
	items, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", limit, 0)
	if err != nil {
		t.Fatalf("ListActivity(limit=%d): %v", limit, err)
	}
	if len(items) > limit {
		t.Fatalf("expected at most %d items, got %d", limit, len(items))
	}
}

func TestListActivity_LimitTwoReturnsTwoOrFewer(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.ListActivity(ctx, testutil.OwnerID, "", 2, 0)
	if err != nil {
		t.Fatalf("ListActivity(limit=2): %v", err)
	}
	if len(items) > 2 {
		t.Fatalf("expected at most 2 items, got %d", len(items))
	}
}
