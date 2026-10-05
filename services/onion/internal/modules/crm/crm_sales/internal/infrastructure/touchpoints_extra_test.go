package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateTouchpoint_OutcomeIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Outcome TP Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	tp, err := repo.CreateTouchpoint(ctx, testutil.OwnerID, testutil.OwnerID, domain.CreateTouchpointRequest{
		LeadID:  lead.ID,
		Type:    "meeting",
		Outcome: "very_positive",
	})
	if err != nil {
		t.Fatalf("CreateTouchpoint: %v", err)
	}
	if tp.Outcome != "very_positive" {
		t.Errorf("Outcome = %q, want %q", tp.Outcome, "very_positive")
	}
}

func TestCreateTouchpoint_NotesIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Notes TP Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	const want = "Discussed pricing options"
	tp, err := repo.CreateTouchpoint(ctx, testutil.OwnerID, testutil.OwnerID, domain.CreateTouchpointRequest{
		LeadID:  lead.ID,
		Type:    "call",
		Outcome: "neutral",
		Notes:   want,
	})
	if err != nil {
		t.Fatalf("CreateTouchpoint: %v", err)
	}
	if tp.Notes != want {
		t.Errorf("Notes = %q, want %q", tp.Notes, want)
	}
}

func TestListTouchpoints_AfterCreate_IncludesNew(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "List TP Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	tp, err := repo.CreateTouchpoint(ctx, testutil.OwnerID, testutil.OwnerID, domain.CreateTouchpointRequest{
		LeadID:  lead.ID,
		Type:    "demo",
		Outcome: "positive",
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
		t.Fatalf("newly created touchpoint id=%d not found in list", tp.ID)
	}
}

func TestListTouchpoints_WrongLeadID_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	tps, err := repo.ListTouchpoints(ctx, testutil.OwnerID, 9999999)
	if err != nil {
		t.Fatalf("ListTouchpoints(leadID=9999999): %v", err)
	}
	if len(tps) != 0 {
		t.Errorf("expected 0 touchpoints for non-existent lead, got %d", len(tps))
	}
}

func TestCreateTouchpoint_OwnerIDIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "OwnerID TP Corp",
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
	if tp.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", tp.OwnerID, testutil.OwnerID)
	}
}
