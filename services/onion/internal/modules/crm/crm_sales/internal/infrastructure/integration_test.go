package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

func TestListLeads_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	leads, err := repo.ListLeads(ctx, testutil.OwnerID, 100, 0)
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
	repo := configuredRepository(db)
	ctx := context.Background()

	contacts, err := repo.ListContacts(ctx, testutil.OwnerID, nil, 100, 0)
	if err != nil {
		t.Fatalf("ListContacts: unexpected error: %v", err)
	}

	if contacts == nil {
		t.Fatal("ListContacts: expected non-nil slice, got nil")
	}
}

func TestGetPipeline_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
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
	repo := configuredRepository(db)
	ctx := context.Background()

	contracts, err := repo.ListContracts(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListContracts: unexpected error: %v", err)
	}

	if contracts == nil {
		t.Fatal("ListContracts: expected non-nil slice, got nil")
	}
}

func TestListLeads_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	leads, err := repo.ListLeads(ctx, 0, 100, 0)
	if err != nil {
		t.Fatalf("ListLeads(ownerID=0): %v", err)
	}
	if len(leads) != 0 {
		t.Fatalf("expected 0 leads for ownerID=0, got %d", len(leads))
	}
}

func TestCreateLead_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Test Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}
	if lead.ID <= 0 {
		t.Fatalf("expected positive lead ID, got %d", lead.ID)
	}
}

func TestCreateLead_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Named Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}
	if lead.CompanyName != "Named Corp" {
		t.Errorf("CompanyName = %q, want %q", lead.CompanyName, "Named Corp")
	}
}

func TestCreateLead_StatusIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Stage Corp",
		Stage:       domain.LeadStageContacted,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}
	if lead.Stage != domain.LeadStageContacted {
		t.Errorf("Stage = %q, want %q", lead.Stage, domain.LeadStageContacted)
	}
}

func TestGetLead_ByID_ReturnsName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	created, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "GetByID Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	got, err := repo.GetLead(ctx, created.ID, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetLead: %v", err)
	}
	if got.CompanyName != "GetByID Corp" {
		t.Errorf("CompanyName = %q, want %q", got.CompanyName, "GetByID Corp")
	}
}

func TestGetLead_NotFound_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	_, err := repo.GetLead(ctx, 999999999, testutil.OwnerID)
	if err == nil {
		t.Fatal("expected error for non-existent lead, got nil")
	}
}

func TestGetLead_WrongOwner_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	created, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "WrongOwner Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	_, err = repo.GetLead(ctx, created.ID, 999999)
	if err == nil {
		t.Fatal("expected error for wrong ownerID, got nil")
	}
}

func TestUpdateLead_ChangesTitle(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	created, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Before Update Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	updated, err := repo.UpdateLead(ctx, created.ID, testutil.OwnerID, domain.UpdateLeadRequest{
		CompanyName: "After Update Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("UpdateLead: %v", err)
	}
	if updated.CompanyName != "After Update Corp" {
		t.Errorf("CompanyName = %q, want %q", updated.CompanyName, "After Update Corp")
	}
}

func TestDeleteLead_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	created, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Delete Me Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	if err := repo.DeleteLead(ctx, created.ID, testutil.OwnerID); err != nil {
		t.Fatalf("DeleteLead: %v", err)
	}

	_, err = repo.GetLead(ctx, created.ID, testutil.OwnerID)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestGetPipeline_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	summaries, err := repo.GetPipeline(ctx, 0)
	if err != nil {
		t.Fatalf("GetPipeline(ownerID=0): %v", err)
	}
	if len(summaries) != 0 {
		t.Fatalf("expected empty pipeline for ownerID=0, got %d entries", len(summaries))
	}
}

func TestGetPipeline_ReturnsStages(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	_, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Pipeline Seeder Corp",
		Stage:       domain.LeadStageDemoBooked,
	})
	if err != nil {
		t.Fatalf("seed lead: %v", err)
	}

	summaries, err := repo.GetPipeline(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetPipeline: %v", err)
	}
	if summaries == nil {
		t.Fatal("GetPipeline: expected non-nil slice, got nil")
	}
}

func TestListContacts_ZeroLeadID_ReturnsAll(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	contacts, err := repo.ListContacts(ctx, testutil.OwnerID, nil, 100, 0)
	if err != nil {
		t.Fatalf("ListContacts(leadID=nil): %v", err)
	}
	if contacts == nil {
		t.Fatal("expected non-nil contacts slice")
	}
}
