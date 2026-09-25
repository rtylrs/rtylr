package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/loyalty/internal/domain"
	"github.com/voxire/onion/internal/modules/loyalty/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestUpsertConfig_MinRedemptionPoints_IsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.LoyaltyConfigInput{
		OwnerID:             testutil.OwnerID,
		BusinessID:          testutil.BusinessID,
		MinRedemptionPoints: 250,
		PointsPerDollar:     1.0,
		IsActive:            true,
	}); err != nil {
		t.Fatalf("UpsertConfig: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.MinRedemptionPoints != 250 {
		t.Errorf("MinRedemptionPoints = %d, want 250", cfg.MinRedemptionPoints)
	}
}

func TestUpsertConfig_PointsValue_IsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.LoyaltyConfigInput{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		PointsValue: 0.025,
		IsActive:    true,
	}); err != nil {
		t.Fatalf("UpsertConfig: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.PointsValue != 0.025 {
		t.Errorf("PointsValue = %v, want 0.025", cfg.PointsValue)
	}
}

func TestUpsertConfig_Idempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.LoyaltyConfigInput{
		OwnerID:         testutil.OwnerID,
		BusinessID:      testutil.BusinessID,
		PointsPerDollar: 3.0,
		IsActive:        true,
	}

	if err := repo.UpsertConfig(ctx, input); err != nil {
		t.Fatalf("UpsertConfig first: %v", err)
	}
	if err := repo.UpsertConfig(ctx, input); err != nil {
		t.Fatalf("UpsertConfig second (idempotent): %v", err)
	}
}

func TestUpsertConfig_UnknownOwner_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.UpsertConfig(ctx, domain.LoyaltyConfigInput{
		OwnerID:  7777777,
		IsActive: false,
	})
	if err != nil {
		t.Fatalf("UpsertConfig for new owner: %v", err)
	}
}
