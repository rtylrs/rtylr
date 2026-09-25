package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm_sales/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateContract_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Contract Lead Corp",
		Stage:       domain.LeadStageContracted,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	contract, err := repo.CreateContract(ctx, testutil.OwnerID, domain.CreateContractRequest{
		LeadID: lead.ID,
		Plan:   "pro",
		MRR:    money.Amount(99900),
		Seats:  5,
	})
	if err != nil {
		t.Fatalf("CreateContract: %v", err)
	}
	if contract.ID <= 0 {
		t.Fatalf("expected positive contract ID, got %d", contract.ID)
	}
}

func TestCreateContract_PlanIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Plan Contract Corp",
		Stage:       domain.LeadStageContracted,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	contract, err := repo.CreateContract(ctx, testutil.OwnerID, domain.CreateContractRequest{
		LeadID: lead.ID,
		Plan:   "enterprise",
		MRR:    money.Amount(500000),
		Seats:  20,
	})
	if err != nil {
		t.Fatalf("CreateContract: %v", err)
	}
	if contract.Plan != "enterprise" {
		t.Errorf("Plan = %q, want %q", contract.Plan, "enterprise")
	}
}

func TestCreateContract_MRRIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "MRR Contract Corp",
		Stage:       domain.LeadStageContracted,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	contract, err := repo.CreateContract(ctx, testutil.OwnerID, domain.CreateContractRequest{
		LeadID: lead.ID,
		Plan:   "pro",
		MRR:    money.Amount(123456),
		Seats:  3,
	})
	if err != nil {
		t.Fatalf("CreateContract: %v", err)
	}
	if contract.MRR != money.Amount(123456) {
		t.Errorf("MRR = %s, want 1234.56", contract.MRR.String())
	}
}

func TestGetContract_NotFound_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetContract(ctx, 999999999, testutil.OwnerID)
	if err == nil {
		t.Fatal("expected error for non-existent contract, got nil")
	}
}

func TestGetContract_WrongOwner_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "WrongOwner Contract Corp",
		Stage:       domain.LeadStageContracted,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	contract, err := repo.CreateContract(ctx, testutil.OwnerID, domain.CreateContractRequest{
		LeadID: lead.ID,
		Plan:   "pro",
		MRR:    money.Amount(50000),
		Seats:  1,
	})
	if err != nil {
		t.Fatalf("CreateContract: %v", err)
	}

	_, err = repo.GetContract(ctx, contract.ID, 999999)
	if err == nil {
		t.Fatal("expected error with wrong ownerID, got nil")
	}
}

func TestUpdateContract_ChangesPlan(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Update Contract Corp",
		Stage:       domain.LeadStageContracted,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	contract, err := repo.CreateContract(ctx, testutil.OwnerID, domain.CreateContractRequest{
		LeadID: lead.ID,
		Plan:   "starter",
		MRR:    money.Amount(9900),
		Seats:  1,
	})
	if err != nil {
		t.Fatalf("CreateContract: %v", err)
	}

	updated, err := repo.UpdateContract(ctx, contract.ID, testutil.OwnerID, domain.UpdateContractRequest{
		Plan:  "pro",
		MRR:   money.Amount(29900),
		Seats: 3,
	})
	if err != nil {
		t.Fatalf("UpdateContract: %v", err)
	}
	if updated.Plan != "pro" {
		t.Errorf("Plan = %q, want %q", updated.Plan, "pro")
	}
}

func TestListContracts_AfterCreate_ContainsNew(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "List Contract Corp",
		Stage:       domain.LeadStageContracted,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	contract, err := repo.CreateContract(ctx, testutil.OwnerID, domain.CreateContractRequest{
		LeadID: lead.ID,
		Plan:   "pro",
		MRR:    money.Amount(50000),
		Seats:  2,
	})
	if err != nil {
		t.Fatalf("CreateContract: %v", err)
	}

	contracts, err := repo.ListContracts(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListContracts: %v", err)
	}

	found := false
	for _, c := range contracts {
		if c.ID == contract.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("contract id=%d not found in list", contract.ID)
	}
}

func TestListContracts_WrongOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	contracts, err := repo.ListContracts(ctx, 999999, 100, 0)
	if err != nil {
		t.Fatalf("ListContracts wrong owner: %v", err)
	}
	if len(contracts) != 0 {
		t.Fatalf("expected 0 contracts for unknown owner, got %d", len(contracts))
	}
}
