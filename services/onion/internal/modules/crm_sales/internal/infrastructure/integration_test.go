package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/crm_sales/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListLeads_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	leads, err := repo.ListLeads(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListLeads: unexpected error: %v", err)
	}

	// Leads may be empty for the fixture owner; just assert a valid slice is returned.
	if leads == nil {
		t.Fatal("ListLeads: expected non-nil slice, got nil")
	}
}

func TestListContacts_NoLeadFilter_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	contacts, err := repo.ListContacts(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("ListContacts: unexpected error: %v", err)
	}

	if contacts == nil {
		t.Fatal("ListContacts: expected non-nil slice, got nil")
	}
}

func TestGetPipeline_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	summaries, err := repo.GetPipeline(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetPipeline: unexpected error: %v", err)
	}

	if summaries == nil {
		t.Fatal("GetPipeline: expected non-nil slice, got nil")
	}
}

func TestListContracts_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	contracts, err := repo.ListContracts(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListContracts: unexpected error: %v", err)
	}

	if contracts == nil {
		t.Fatal("ListContracts: expected non-nil slice, got nil")
	}
}
