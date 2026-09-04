package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/campaigns/internal/domain"
	"github.com/voxire/onion/internal/modules/campaigns/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateCampaign_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const want = "My Stored Campaign"
	id, err := repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       want,
		Type:       "email",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	c, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if c.Name != want {
		t.Errorf("Name = %q, want %q", c.Name, want)
	}
}

func TestCreateCampaign_SubjectIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Subject Campaign",
		Type:       "email",
		Subject:    "Hello World",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	c, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if c.Subject != "Hello World" {
		t.Errorf("Subject = %q, want %q", c.Subject, "Hello World")
	}
}

func TestGetCampaign_WrongOwner_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Wrong Owner Get",
		Type:       "email",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = repo.GetByID(ctx, 9999999, id)
	if err == nil {
		t.Fatal("expected error for wrong owner GetByID, got nil")
	}
	if !errors.Is(err, domain.ErrCampaignNotFound) {
		t.Errorf("expected ErrCampaignNotFound, got %v", err)
	}
}

func TestUpdateCampaign_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Before Name",
		Type:       "email",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(ctx, domain.CampaignUpdateInput{
		ID:         id,
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "After Name",
		Type:       "email",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	c, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if c.Name != "After Name" {
		t.Errorf("Name = %q, want %q", c.Name, "After Name")
	}
}

func TestListCampaigns_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Owner Check Campaign",
		Type:       "email",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	campaigns, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range campaigns {
		if c.OwnerID != testutil.OwnerID {
			t.Errorf("campaign %d has owner_id %d, want %d", c.ID, c.OwnerID, testutil.OwnerID)
		}
	}
}
