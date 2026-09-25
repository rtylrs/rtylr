package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/customers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateCampaignViaCustomersRepo_DescriptionIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const want = "Our summer promo campaign"
	id, err := repo.CreateCampaign(ctx, domain.CampaignInput{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		Name:        "Summer Promo",
		Channel:     "email",
		Description: want,
	})
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	c, err := repo.GetCampaignByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetCampaignByID: %v", err)
	}
	if c.Description != want {
		t.Errorf("Description = %q, want %q", c.Description, want)
	}
}

func TestCreateCampaignViaCustomersRepo_ChannelIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCampaign(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "SMS Campaign",
		Channel:    "sms",
	})
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	c, err := repo.GetCampaignByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetCampaignByID: %v", err)
	}
	if c.Channel != "sms" {
		t.Errorf("Channel = %q, want %q", c.Channel, "sms")
	}
}

func TestListCampaignsViaCustomersRepo_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.CreateCampaign(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Owner Filter Check",
		Channel:    "email",
	})

	campaigns, err := repo.ListCampaigns(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("ListCampaigns: %v", err)
	}
	for _, c := range campaigns {
		if c.OwnerID != testutil.OwnerID {
			t.Errorf("campaign %d has owner_id %d, want %d", c.ID, c.OwnerID, testutil.OwnerID)
		}
	}
}

func TestUpdateCampaignViaCustomersRepo_ChangesSubject(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCampaign(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Subject Update Campaign",
		Channel:    "email",
		Subject:    "Original Subject",
	})
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	if err := repo.UpdateCampaign(ctx, domain.CampaignUpdateInput{
		ID:         id,
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Subject Update Campaign",
		Channel:    "email",
		Subject:    "Updated Subject",
	}); err != nil {
		t.Fatalf("UpdateCampaign: %v", err)
	}

	c, err := repo.GetCampaignByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetCampaignByID: %v", err)
	}
	if c.Subject != "Updated Subject" {
		t.Errorf("Subject = %q, want %q", c.Subject, "Updated Subject")
	}
}
