package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm_sales/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateContract_SignedByNameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "SignedBy Contract Corp",
		Stage:       domain.LeadStageContracted,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	contract, err := repo.CreateContract(ctx, testutil.OwnerID, domain.CreateContractRequest{
		LeadID:       lead.ID,
		Plan:         "pro",
		MRR:          money.Amount(50000),
		Seats:        1,
		SignedByName: "CEO Jane Smith",
	})
	if err != nil {
		t.Fatalf("CreateContract: %v", err)
	}

	got, err := repo.GetContract(ctx, contract.ID, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetContract: %v", err)
	}
	if got.SignedByName != "CEO Jane Smith" {
		t.Errorf("SignedByName = %q, want %q", got.SignedByName, "CEO Jane Smith")
	}
}

func TestCreateContract_NotesIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Notes Contract Corp",
		Stage:       domain.LeadStageContracted,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	const want = "Annual plan with quarterly review"
	contract, err := repo.CreateContract(ctx, testutil.OwnerID, domain.CreateContractRequest{
		LeadID: lead.ID,
		Plan:   "enterprise",
		Notes:  want,
	})
	if err != nil {
		t.Fatalf("CreateContract: %v", err)
	}

	got, err := repo.GetContract(ctx, contract.ID, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetContract: %v", err)
	}
	if got.Notes != want {
		t.Errorf("Notes = %q, want %q", got.Notes, want)
	}
}

func TestGetContract_NotFound_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetContract(ctx, 999999999, testutil.OwnerID)
	if err == nil {
		t.Fatal("expected error for non-existent contract, got nil")
	}
	if !errors.Is(err, domain.ErrContractNotFound) {
		t.Errorf("expected ErrContractNotFound, got %v", err)
	}
}

func TestListContracts_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lead, _ := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Owner Check Contract Corp",
		Stage:       domain.LeadStageContracted,
	})
	_, _ = repo.CreateContract(ctx, testutil.OwnerID, domain.CreateContractRequest{
		LeadID: lead.ID,
		Plan:   "starter",
	})

	contracts, err := repo.ListContracts(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListContracts: %v", err)
	}
	for _, c := range contracts {
		if c.OwnerID != testutil.OwnerID {
			t.Errorf("contract %d has owner_id %d, want %d", c.ID, c.OwnerID, testutil.OwnerID)
		}
	}
}

func TestUpdateContract_ChangesMRR(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "MRR Update Corp",
		Stage:       domain.LeadStageContracted,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	contract, err := repo.CreateContract(ctx, testutil.OwnerID, domain.CreateContractRequest{
		LeadID: lead.ID,
		Plan:   "pro",
		MRR:    money.Amount(100000),
	})
	if err != nil {
		t.Fatalf("CreateContract: %v", err)
	}

	_, err = repo.UpdateContract(ctx, contract.ID, testutil.OwnerID, domain.UpdateContractRequest{
		Plan: "pro",
		MRR:  money.Amount(150000),
	})
	if err != nil {
		t.Fatalf("UpdateContract: %v", err)
	}

	got, err := repo.GetContract(ctx, contract.ID, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetContract: %v", err)
	}
	if got.MRR != money.Amount(150000) {
		t.Errorf("MRR = %s, want 1500.0", got.MRR.String())
	}
}
