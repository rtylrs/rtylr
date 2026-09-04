package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/campaigns/internal/domain"
	"github.com/voxire/onion/internal/modules/campaigns/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateCampaign_TypeIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Type Check Campaign",
		Type:       "sms",
		Subject:    "",
		Body:       "SMS body",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	c, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if c.Type != "sms" {
		t.Errorf("Type = %q, want %q", c.Type, "sms")
	}
}

func TestCreateCampaign_OwnerIDIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "OwnerID Check Campaign",
		Type:       "email",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	c, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if c.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", c.OwnerID, testutil.OwnerID)
	}
}

func TestCreateCampaign_IDsAreUnique(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.Create(ctx, domain.CampaignInput{OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "Unique A", Type: "email"})
	if err != nil {
		t.Fatalf("Create A: %v", err)
	}
	id2, err := repo.Create(ctx, domain.CampaignInput{OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "Unique B", Type: "email"})
	if err != nil {
		t.Fatalf("Create B: %v", err)
	}
	if id1 == id2 {
		t.Errorf("expected unique IDs, both got %d", id1)
	}
}

func TestUpdateCampaign_UpdatesBody(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Body Update Campaign",
		Type:       "email",
		Body:       "original body",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(ctx, domain.CampaignUpdateInput{
		ID:         id,
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Body Update Campaign",
		Type:       "email",
		Body:       "updated body",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	c, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if c.Body != "updated body" {
		t.Errorf("Body = %q, want %q", c.Body, "updated body")
	}
}

func TestListCampaigns_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	_, err = repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Count Increase Campaign",
		Type:       "email",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	after, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Errorf("expected count +1: before=%d after=%d", len(before), len(after))
	}
}

func TestDeleteCampaign_NonExistent_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Deleting a non-existent campaign with the correct owner should be a noop
	err := repo.Delete(ctx, testutil.OwnerID, 9999999)
	if err != nil {
		t.Fatalf("unexpected error deleting non-existent campaign: %v", err)
	}
}

func TestUpdateCampaignStatus_WrongOwner_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Status Wrong Owner",
		Type:       "email",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = repo.UpdateStatus(ctx, 999999, id, "sent")
	if err == nil {
		t.Fatal("expected error for wrong owner UpdateStatus, got nil")
	}
}
