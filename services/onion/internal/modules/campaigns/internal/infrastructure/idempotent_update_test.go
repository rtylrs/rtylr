package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/campaigns/internal/domain"
	"github.com/voxire/onion/internal/modules/campaigns/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// MySQL reports zero changed rows when an UPDATE matches a row whose values are
// already identical. A re-sent payload must be an idempotent success; only a
// campaign that does not exist for the business is a not-found.
func TestUpdateCampaign_UnchangedValues_Succeeds(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Idempotent Campaign",
		Type:       "email",
		Subject:    "Hello",
		Body:       "Body content",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	input := domain.CampaignUpdateInput{
		ID:         id,
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Idempotent Campaign",
		Type:       "email",
		Subject:    "Hello",
		Body:       "Body content",
	}
	if err := repo.Update(ctx, input); err != nil {
		t.Fatalf("first Update: %v", err)
	}
	if err := repo.Update(ctx, input); err != nil {
		t.Fatalf("re-sent Update with unchanged values: %v", err)
	}
}

func TestUpdateCampaignStatus_UnchangedStatus_Succeeds(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CampaignInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Idempotent Status Campaign",
		Type:       "email",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.UpdateStatus(ctx, testutil.BusinessID, id, "paused"); err != nil {
		t.Fatalf("first UpdateStatus: %v", err)
	}
	if err := repo.UpdateStatus(ctx, testutil.BusinessID, id, "paused"); err != nil {
		t.Fatalf("re-sent UpdateStatus with the unchanged status: %v", err)
	}
}

func TestUpdateCampaign_MissingCampaign_StillNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.Update(ctx, domain.CampaignUpdateInput{
		ID:         999999999,
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Ghost Campaign",
		Type:       "email",
	})
	if !errors.Is(err, domain.ErrCampaignNotFound) {
		t.Fatalf("Update on a missing campaign: err = %v, want ErrCampaignNotFound", err)
	}

	err = repo.UpdateStatus(ctx, testutil.BusinessID, 999999999, "paused")
	if !errors.Is(err, domain.ErrCampaignNotFound) {
		t.Fatalf("UpdateStatus on a missing campaign: err = %v, want ErrCampaignNotFound", err)
	}
}
