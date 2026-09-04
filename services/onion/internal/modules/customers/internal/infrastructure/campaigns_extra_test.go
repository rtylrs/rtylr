package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/customers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateCampaign_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCampaign(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Spring Sale Campaign",
		Channel:    "email",
		Subject:    "Big Sale!",
		Content:    "Check out our deals",
	})
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreateCampaign_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCampaign(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Name Stored Campaign",
		Channel:    "email",
		Content:    "body",
	})
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	c, err := repo.GetCampaignByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetCampaignByID: %v", err)
	}
	if c.Name != "Name Stored Campaign" {
		t.Errorf("Name = %q, want %q", c.Name, "Name Stored Campaign")
	}
}

func TestGetCampaignByID_NotFound_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetCampaignByID(ctx, testutil.OwnerID, 999999999)
	if err == nil {
		t.Fatal("expected error for non-existent campaign, got nil")
	}
}

func TestGetCampaignByID_WrongOwner_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCampaign(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "WrongOwner Campaign",
		Channel:    "email",
		Content:    "body",
	})
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	_, err = repo.GetCampaignByID(ctx, 999999, id)
	if err == nil {
		t.Fatal("expected error with wrong ownerID, got nil")
	}
}

func TestUpdateCampaign_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCampaign(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Before Campaign",
		Channel:    "email",
		Content:    "body",
	})
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	if err := repo.UpdateCampaign(ctx, domain.CampaignUpdateInput{
		ID:         id,
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "After Campaign",
		Channel:    "email",
		Content:    "body",
	}); err != nil {
		t.Fatalf("UpdateCampaign: %v", err)
	}

	c, err := repo.GetCampaignByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetCampaignByID: %v", err)
	}
	if c.Name != "After Campaign" {
		t.Errorf("Name = %q, want %q", c.Name, "After Campaign")
	}
}

func TestListCampaigns_WrongOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	campaigns, err := repo.ListCampaigns(ctx, 999999, "")
	if err != nil {
		t.Fatalf("ListCampaigns wrong owner: %v", err)
	}
	if len(campaigns) != 0 {
		t.Fatalf("expected 0 campaigns for unknown owner, got %d", len(campaigns))
	}
}

func TestUpdateCampaignStatus_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCampaign(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Status Change Campaign",
		Channel:    "email",
		Content:    "body",
	})
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	if err := repo.UpdateCampaignStatus(ctx, testutil.OwnerID, id, "sent"); err != nil {
		t.Fatalf("UpdateCampaignStatus: %v", err)
	}

	c, err := repo.GetCampaignByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetCampaignByID: %v", err)
	}
	if c.Status != "sent" {
		t.Errorf("Status = %q, want %q", c.Status, "sent")
	}
}
