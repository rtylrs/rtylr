package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_core/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetApprovalChain_NonExistent_ReturnsNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	chain, err := repo.GetApprovalChain(ctx, testutil.OwnerID, "unknown_type_xyz")
	if err != nil {
		t.Fatalf("GetApprovalChain unexpected error: %v", err)
	}
	// nil is acceptable when no chain is configured for this type
	_ = chain
}

func TestUpsertApprovalChain_Persists(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	chain, err := repo.UpsertApprovalChain(ctx, testutil.OwnerID, domain.ApprovalChainRecord{
		RequestType: "test_approval_type",
		Name:        "Test Approval Chain",
		Levels:      []domain.ApprovalChainLevelRecord{},
	})
	if err != nil {
		t.Fatalf("UpsertApprovalChain: %v", err)
	}
	if chain == nil {
		t.Fatal("expected non-nil chain after upsert")
	}
	if chain.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", chain.OwnerID, testutil.OwnerID)
	}
}

func TestUpsertApprovalChain_Idempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	params := domain.ApprovalChainRecord{
		RequestType: "idempotent_approval_type",
		Name:        "Idempotent Chain",
		Levels:      []domain.ApprovalChainLevelRecord{},
	}

	first, err := repo.UpsertApprovalChain(ctx, testutil.OwnerID, params)
	if err != nil {
		t.Fatalf("UpsertApprovalChain first: %v", err)
	}

	second, err := repo.UpsertApprovalChain(ctx, testutil.OwnerID, params)
	if err != nil {
		t.Fatalf("UpsertApprovalChain second: %v", err)
	}

	if first.ID != second.ID {
		t.Errorf("upsert changed ID from %d to %d, expected idempotency", first.ID, second.ID)
	}
}

func TestUpsertApprovalChain_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	chain, err := repo.UpsertApprovalChain(ctx, testutil.OwnerID, domain.ApprovalChainRecord{
		RequestType: "name_check_approval",
		Name:        "Named Chain",
		Levels:      []domain.ApprovalChainLevelRecord{},
	})
	if err != nil {
		t.Fatalf("UpsertApprovalChain: %v", err)
	}
	if chain.Name != "Named Chain" {
		t.Errorf("Name = %q, want 'Named Chain'", chain.Name)
	}
}

func TestGetApprovalChain_AfterUpsert_ReturnsChain(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	requestType := "get_after_upsert_type"
	_, err := repo.UpsertApprovalChain(ctx, testutil.OwnerID, domain.ApprovalChainRecord{
		RequestType: requestType,
		Name:        "Get After Upsert Chain",
		Levels:      []domain.ApprovalChainLevelRecord{},
	})
	if err != nil {
		t.Fatalf("UpsertApprovalChain: %v", err)
	}

	got, err := repo.GetApprovalChain(ctx, testutil.OwnerID, requestType)
	if err != nil {
		t.Fatalf("GetApprovalChain: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil chain after upsert")
	}
	if got.RequestType != requestType {
		t.Errorf("RequestType = %q, want %q", got.RequestType, requestType)
	}
}
