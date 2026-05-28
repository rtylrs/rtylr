package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/loyalty/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListMembers_NegativeOffset_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Negative offset should be treated like 0; just ensure no crash
	members, err := repo.ListMembers(ctx, testutil.OwnerID, 10, -1)
	if err != nil {
		t.Fatalf("ListMembers(offset=-1): %v", err)
	}
	_ = members
}

func TestListMembers_LargeLimit_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	members, err := repo.ListMembers(ctx, testutil.OwnerID, 10000, 0)
	if err != nil {
		t.Fatalf("ListMembers(limit=10000): %v", err)
	}
	_ = members
}

func TestListMembers_AllHavePositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	members, err := repo.ListMembers(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	for i, m := range members {
		if m.ID <= 0 {
			t.Errorf("members[%d].ID = %d, want > 0", i, m.ID)
		}
	}
}

func TestListMembers_AllHaveNonEmptyName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	members, err := repo.ListMembers(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	for i, m := range members {
		if m.Name == "" {
			t.Errorf("members[%d].Name is empty", i)
		}
	}
}

func TestListMembers_AllHaveNonNegativePoints(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	members, err := repo.ListMembers(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	for i, m := range members {
		if m.Points < 0 {
			t.Errorf("members[%d].Points = %d, want non-negative", i, m.Points)
		}
	}
}
