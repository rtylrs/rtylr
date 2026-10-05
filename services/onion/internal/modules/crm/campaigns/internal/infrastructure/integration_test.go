package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/campaigns/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/campaigns/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestList_Campaigns_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	campaigns, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("campaigns.List: unexpected error: %v", err)
	}

	// Campaigns may be empty for the fixture owner; assert a valid slice is returned.
	if campaigns == nil {
		t.Fatal("campaigns.List: expected non-nil slice, got nil")
	}
}

func TestGetByID_Campaigns_NotFoundReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, testutil.OwnerID, 0)
	if err == nil {
		t.Fatal("campaigns.GetByID(0): expected error for non-existent campaign, got nil")
	}
}

func TestList_Campaigns_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	campaigns, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("List(0): %v", err)
	}
	if len(campaigns) != 0 {
		t.Fatalf("expected 0 campaigns for owner 0, got %d", len(campaigns))
	}
}

func TestCreateCampaign_Persists(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Test Campaign Persists",
		Type:       "email",
		Subject:    "Hello",
		Body:       "Email body content",
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreateCampaign_FieldsAreStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	name := "Stored Fields Campaign"
	subject := "Subject Line"
	body := "Body content here"

	id, err := repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       name,
		Type:       "email",
		Subject:    subject,
		Body:       body,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	c, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if c.Name != name {
		t.Errorf("Name = %q, want %q", c.Name, name)
	}
	if c.Subject != subject {
		t.Errorf("Subject = %q, want %q", c.Subject, subject)
	}
	if c.Body != body {
		t.Errorf("Body = %q, want %q", c.Body, body)
	}
}

func TestGetCampaignByID_NotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, testutil.OwnerID, 999999999)
	if err == nil {
		t.Fatal("expected error for non-existent campaign, got nil")
	}
}

func TestGetCampaignByID_WrongOwner_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Owner Scoped Campaign",
		Type:       "email",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = repo.GetByID(ctx, 999999, id)
	if err == nil {
		t.Fatal("expected error when fetching campaign with wrong owner, got nil")
	}
}

func TestUpdateCampaign_ChangesTitle(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Original Title",
		Type:       "email",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(ctx, domain.CampaignUpdateInput{
		ID:         id,
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Updated Title",
		Type:       "email",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	c, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if c.Name != "Updated Title" {
		t.Errorf("Name = %q, want %q", c.Name, "Updated Title")
	}
}

func TestUpdateCampaign_WrongOwner_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Wrong Owner Update",
		Type:       "email",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = repo.Update(ctx, domain.CampaignUpdateInput{
		ID:      id,
		OwnerID: 999999,
		Name:    "Should Not Change",
		Type:    "email",
	})
	if !errors.Is(err, domain.ErrCampaignNotFound) {
		t.Errorf("expected ErrCampaignNotFound, got %v", err)
	}
}

func TestDeleteCampaign_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Campaign To Delete",
		Type:       "email",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = repo.GetByID(ctx, testutil.OwnerID, id)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestUpdateCampaignStatus_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Status Change Campaign",
		Type:       "email",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.UpdateStatus(ctx, testutil.OwnerID, id, "sent"); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	c, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if c.Status != "sent" {
		t.Errorf("Status = %q, want %q", c.Status, "sent")
	}
}

func TestUpdateCampaignStatus_NotFound_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.UpdateStatus(ctx, testutil.OwnerID, 999999999, "sent")
	if !errors.Is(err, domain.ErrCampaignNotFound) {
		t.Errorf("expected ErrCampaignNotFound, got %v", err)
	}
}

func TestListCampaigns_WrongOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	campaigns, err := repo.List(ctx, 999999)
	if err != nil {
		t.Fatalf("List wrong owner: %v", err)
	}
	if len(campaigns) != 0 {
		t.Fatalf("expected 0 campaigns for unknown owner, got %d", len(campaigns))
	}
}

func TestCreateCampaign_ThenAppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Appears In List Campaign",
		Type:       "email",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	campaigns, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	found := false
	for _, c := range campaigns {
		if c.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("campaign id=%d not found in list", id)
	}
}

func TestListCampaigns_ItemsHaveOwnerID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Owner ID Check Campaign",
		Type:       "email",
	})

	campaigns, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, c := range campaigns {
		if c.OwnerID != testutil.OwnerID {
			t.Errorf("campaigns[%d].OwnerID = %d, want %d", i, c.OwnerID, testutil.OwnerID)
		}
	}
}

func TestDeleteCampaign_WrongOwner_NoEffect(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Wrong Owner Delete",
		Type:       "email",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Delete with wrong owner — should be a noop; campaign should still exist.
	_ = repo.Delete(ctx, 999999, id)

	c, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID after wrong-owner delete: %v", err)
	}
	if c.ID != id {
		t.Errorf("expected campaign id=%d to still exist", id)
	}
}
