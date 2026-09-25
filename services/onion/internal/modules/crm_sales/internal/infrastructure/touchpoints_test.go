package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm_sales/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateTouchpoint_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Touchpoint Lead Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	tp, err := repo.CreateTouchpoint(ctx, testutil.OwnerID, testutil.OwnerID, domain.CreateTouchpointRequest{
		LeadID:  lead.ID,
		Type:    "call",
		Outcome: "positive",
	})
	if err != nil {
		t.Fatalf("CreateTouchpoint: %v", err)
	}
	if tp.ID <= 0 {
		t.Fatalf("expected positive touchpoint ID, got %d", tp.ID)
	}
}

func TestCreateTouchpoint_TypeIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Type Touchpoint Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	tp, err := repo.CreateTouchpoint(ctx, testutil.OwnerID, testutil.OwnerID, domain.CreateTouchpointRequest{
		LeadID:  lead.ID,
		Type:    "email",
		Outcome: "no_response",
	})
	if err != nil {
		t.Fatalf("CreateTouchpoint: %v", err)
	}
	if tp.Type != "email" {
		t.Errorf("Type = %q, want %q", tp.Type, "email")
	}
}

func TestListTouchpoints_AfterCreate_ContainsNew(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "List Touchpoint Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	tp, err := repo.CreateTouchpoint(ctx, testutil.OwnerID, testutil.OwnerID, domain.CreateTouchpointRequest{
		LeadID:  lead.ID,
		Type:    "meeting",
		Outcome: "demo",
	})
	if err != nil {
		t.Fatalf("CreateTouchpoint: %v", err)
	}

	tps, err := repo.ListTouchpoints(ctx, testutil.OwnerID, lead.ID)
	if err != nil {
		t.Fatalf("ListTouchpoints: %v", err)
	}

	found := false
	for _, t2 := range tps {
		if t2.ID == tp.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("touchpoint id=%d not found in list", tp.ID)
	}
}

func TestGetTouchpoint_NotFound_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetTouchpoint(ctx, 999999999, testutil.OwnerID)
	if err == nil {
		t.Fatal("expected error for non-existent touchpoint, got nil")
	}
}

func TestDeleteTouchpoint_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Delete Touchpoint Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	tp, err := repo.CreateTouchpoint(ctx, testutil.OwnerID, testutil.OwnerID, domain.CreateTouchpointRequest{
		LeadID:  lead.ID,
		Type:    "call",
		Outcome: "other",
	})
	if err != nil {
		t.Fatalf("CreateTouchpoint: %v", err)
	}

	if err := repo.DeleteTouchpoint(ctx, tp.ID, testutil.OwnerID); err != nil {
		t.Fatalf("DeleteTouchpoint: %v", err)
	}

	_, err = repo.GetTouchpoint(ctx, tp.ID, testutil.OwnerID)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestDeleteTouchpoint_WrongOwner_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Wrong Delete Touchpoint Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	tp, err := repo.CreateTouchpoint(ctx, testutil.OwnerID, testutil.OwnerID, domain.CreateTouchpointRequest{
		LeadID:  lead.ID,
		Type:    "call",
		Outcome: "other",
	})
	if err != nil {
		t.Fatalf("CreateTouchpoint: %v", err)
	}

	err = repo.DeleteTouchpoint(ctx, tp.ID, 999999)
	if err == nil {
		t.Fatal("expected error deleting touchpoint with wrong owner, got nil")
	}
}
