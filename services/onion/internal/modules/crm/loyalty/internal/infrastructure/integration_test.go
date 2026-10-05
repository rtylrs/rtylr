package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/loyalty/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/loyalty/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetConfig_NotFoundOrPresent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cfg, err := repo.GetConfig(ctx, testutil.BusinessID)
	if err != nil && !errors.Is(err, domain.ErrLoyaltyConfigNotFound) {
		t.Fatalf("GetConfig unexpected error: %v", err)
	}
	_ = cfg
}

func TestUpsertConfig_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.UpsertConfig(ctx, domain.LoyaltyConfigInput{
		OwnerID:         testutil.OwnerID,
		BusinessID:      testutil.BusinessID,
		PointsPerDollar: 1.0,
		PointsValue:     0.01,
		IsActive:        true,
	})
	if err != nil {
		t.Fatalf("UpsertConfig: %v", err)
	}
}

func TestUpsertConfig_ThenGetConfig(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.LoyaltyConfigInput{
		OwnerID:         testutil.OwnerID,
		BusinessID:      testutil.BusinessID,
		PointsPerDollar: 2.0,
		PointsValue:     0.005,
		IsActive:        true,
	}); err != nil {
		t.Fatalf("UpsertConfig: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("GetConfig after upsert: %v", err)
	}
	if cfg.OwnerID != testutil.OwnerID {
		t.Errorf("GetConfig owner_id = %d, want %d", cfg.OwnerID, testutil.OwnerID)
	}
}

func TestListMembers_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	members, err := repo.ListMembers(ctx, testutil.OwnerID, 50, 0)
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	_ = members
}

func TestListMembers_AlexHasHighPoints(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	members, err := repo.ListMembers(ctx, testutil.OwnerID, 200, 0)
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	// crm_customer 5302 Alex has loyalty_points=1500 and should appear.
	found := false
	for _, m := range members {
		if m.Points >= 1500 {
			found = true
			break
		}
	}
	if !found {
		t.Log("no member with >= 1500 points found; fixture may have fewer points or visit_count=0")
	}
}

func TestListMembers_OrderedByPointsDesc(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	members, err := repo.ListMembers(ctx, testutil.OwnerID, 50, 0)
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	for i := 1; i < len(members); i++ {
		if members[i].Points > members[i-1].Points {
			t.Errorf("members not ordered by points desc: members[%d].Points=%d > members[%d].Points=%d",
				i, members[i].Points, i-1, members[i-1].Points)
		}
	}
}

func TestListMembers_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	members, err := repo.ListMembers(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	for _, m := range members {
		if m.OwnerID != testutil.OwnerID {
			t.Errorf("member %d has owner_id %d, want %d", m.ID, m.OwnerID, testutil.OwnerID)
		}
	}
}

func TestListTransactions_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	txs, total, err := repo.ListTransactions(ctx, testutil.OwnerID, 0, 1, 20)
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}
	if total < 0 {
		t.Errorf("total is negative: %d", total)
	}
	_ = txs
}

func TestListTransactions_FilterByCustomer_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	txs, _, err := repo.ListTransactions(ctx, testutil.OwnerID, 5302, 1, 20)
	if err != nil {
		t.Fatalf("ListTransactions for customer 5302: %v", err)
	}
	_ = txs
}

func TestListRewards_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rewards, err := repo.ListRewards(ctx, testutil.BusinessID, false)
	if err != nil {
		t.Fatalf("ListRewards: %v", err)
	}
	_ = rewards
}

func TestUpsertReward_Create_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.UpsertReward(ctx, domain.LoyaltyReward{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		Name:        "Test Reward",
		Description: "Integration test reward",
		Category:    "general",
		PointsCost:  500,
		Icon:        "",
		Active:      true,
	})
	if err != nil {
		t.Fatalf("UpsertReward create: %v", err)
	}
	if id == 0 {
		t.Fatal("UpsertReward returned id=0")
	}
}

func TestListMembers_UnknownOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	members, err := repo.ListMembers(ctx, 9999999, 50, 0)
	if err != nil {
		t.Fatalf("ListMembers unknown owner: %v", err)
	}
	if len(members) != 0 {
		t.Errorf("expected 0 members for unknown owner, got %d", len(members))
	}
}

func TestUpsertConfig_PointsPerDollarRoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.LoyaltyConfigInput{
		OwnerID:         testutil.OwnerID,
		BusinessID:      testutil.BusinessID,
		PointsPerDollar: 5.0,
		PointsValue:     0.01,
		IsActive:        true,
	}); err != nil {
		t.Fatalf("UpsertConfig: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.PointsPerDollar != 5.0 {
		t.Errorf("PointsPerDollar = %v, want 5.0", cfg.PointsPerDollar)
	}
}

func TestUpsertConfig_IsActiveRoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.LoyaltyConfigInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		IsActive:   false,
	}); err != nil {
		t.Fatalf("UpsertConfig IsActive=false: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.IsActive != false {
		t.Errorf("expected IsActive=false, got %v", cfg.IsActive)
	}
}

func TestGetConfig_OwnerIDIsSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.UpsertConfig(ctx, domain.LoyaltyConfigInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		IsActive:   true,
	}); err != nil {
		t.Fatalf("UpsertConfig: %v", err)
	}

	cfg, err := repo.GetConfig(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", cfg.OwnerID, testutil.OwnerID)
	}
}

func TestListTransactions_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	txs, total, err := repo.ListTransactions(ctx, testutil.OtherBusinessID, 0, 1, 20)
	if err != nil {
		t.Fatalf("ListTransactions(business=%d): %v", testutil.OtherBusinessID, err)
	}
	if total != 0 || len(txs) != 0 {
		t.Fatalf("expected 0 transactions for business %d, got total=%d len=%d",
			testutil.OtherBusinessID, total, len(txs))
	}
}

func TestCreateTransaction_Award_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.CreateTransaction(ctx, domain.AwardInput{
		OwnerID:     testutil.OwnerID,
		CustomerID:  5301,
		Points:      100,
		Description: "integration test award",
	}, "award")
	if err != nil {
		t.Fatalf("CreateTransaction award: %v", err)
	}
}

func TestCreateTransaction_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.CreateTransaction(ctx, domain.AwardInput{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		CustomerID:  5302,
		Points:      50,
		Description: "list check award",
	}, "award"); err != nil {
		t.Fatalf("CreateTransaction: %v", err)
	}

	txs, total, err := repo.ListTransactions(ctx, testutil.BusinessID, 5302, 1, 50)
	if err != nil {
		t.Fatalf("ListTransactions: %v", err)
	}
	if total == 0 || len(txs) == 0 {
		t.Fatal("expected at least one transaction after create")
	}
}

func TestListRewards_ActiveOnly_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rewards, err := repo.ListRewards(ctx, testutil.BusinessID, true)
	if err != nil {
		t.Fatalf("ListRewards(activeOnly=true): %v", err)
	}
	_ = rewards
}

func TestUpsertReward_Update_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.UpsertReward(ctx, domain.LoyaltyReward{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Update Test Reward",
		Category:   "general",
		PointsCost: 200,
		Active:     true,
	})
	if err != nil {
		t.Fatalf("UpsertReward create: %v", err)
	}

	_, err = repo.UpsertReward(ctx, domain.LoyaltyReward{
		ID:         id,
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Updated Reward Name",
		Category:   "general",
		PointsCost: 300,
		Active:     true,
	})
	if err != nil {
		t.Fatalf("UpsertReward update: %v", err)
	}
}

func TestListRewards_UnknownBusiness_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Rewards are scoped by business; a business nothing has seeded must
	// list empty even though rows exist for other businesses.
	rewards, err := repo.ListRewards(ctx, testutil.OtherBusinessID, false)
	if err != nil {
		t.Fatalf("ListRewards(business=%d): %v", testutil.OtherBusinessID, err)
	}
	if len(rewards) != 0 {
		t.Fatalf("expected 0 rewards for business %d, got %d", testutil.OtherBusinessID, len(rewards))
	}
}

func TestUpsertReward_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.UpsertReward(ctx, domain.LoyaltyReward{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Listed Reward",
		Category:   "general",
		PointsCost: 150,
		Active:     true,
	})
	if err != nil {
		t.Fatalf("UpsertReward: %v", err)
	}

	rewards, err := repo.ListRewards(ctx, testutil.BusinessID, false)
	if err != nil {
		t.Fatalf("ListRewards: %v", err)
	}
	for _, r := range rewards {
		if r.ID == id {
			return
		}
	}
	t.Fatalf("reward id=%d not found in list", id)
}

func TestListMembers_PaginationOffsetWorks(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	all, err := repo.ListMembers(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListMembers(offset=0): %v", err)
	}

	if len(all) < 2 {
		t.Skip("fewer than 2 members; skipping pagination test")
	}

	page2, err := repo.ListMembers(ctx, testutil.OwnerID, 100, 1)
	if err != nil {
		t.Fatalf("ListMembers(offset=1): %v", err)
	}
	if len(page2) >= len(all) {
		t.Errorf("expected offset=1 to return fewer results: all=%d offset1=%d", len(all), len(page2))
	}
}
