package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm_sales/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListLeads_WrongOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	leads, err := repo.ListLeads(ctx, 999999, 100, 0)
	if err != nil {
		t.Fatalf("ListLeads wrong owner: %v", err)
	}
	if len(leads) != 0 {
		t.Fatalf("expected 0 leads for unknown owner, got %d", len(leads))
	}
}

func TestCreateLead_SourceIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Source Corp",
		Source:      "referral",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}
	if lead.Source != "referral" {
		t.Errorf("Source = %q, want %q", lead.Source, "referral")
	}
}

func TestCreateLead_OwnerIDIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "OwnerID Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}
	if lead.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", lead.OwnerID, testutil.OwnerID)
	}
}

func TestUpdateLead_StageChanges(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Stage Change Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	updated, err := repo.UpdateLead(ctx, created.ID, testutil.OwnerID, domain.UpdateLeadRequest{
		CompanyName: "Stage Change Corp",
		Stage:       domain.LeadStageDemoBooked,
	})
	if err != nil {
		t.Fatalf("UpdateLead: %v", err)
	}
	if updated.Stage != domain.LeadStageDemoBooked {
		t.Errorf("Stage = %q, want %q", updated.Stage, domain.LeadStageDemoBooked)
	}
}

func TestDeleteLead_WrongOwner_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "DeleteWrong Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	err = repo.DeleteLead(ctx, created.ID, 999999)
	if err == nil {
		t.Fatal("expected error deleting lead with wrong owner, got nil")
	}
}

func TestListLeads_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.ListLeads(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListLeads before: %v", err)
	}

	_, err = repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Count Inc Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	after, err := repo.ListLeads(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListLeads after: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Errorf("expected count +1: before=%d after=%d", len(before), len(after))
	}
}

func TestCreateLead_IDsAreUnique(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	a, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{CompanyName: "Unique A", Stage: domain.LeadStageCold})
	if err != nil {
		t.Fatalf("CreateLead A: %v", err)
	}
	b, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{CompanyName: "Unique B", Stage: domain.LeadStageCold})
	if err != nil {
		t.Fatalf("CreateLead B: %v", err)
	}
	if a.ID == b.ID {
		t.Errorf("expected unique IDs, both got %d", a.ID)
	}
}
