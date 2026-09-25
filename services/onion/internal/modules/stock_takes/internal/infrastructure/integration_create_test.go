package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/stock_takes/internal/domain"
	"github.com/voxire/onion/internal/modules/stock_takes/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateStockTake_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.StockTakeInput{
		Notes: "Appears In List",
		Items: []domain.StockTakeItemInput{
			{ProductID: 5202, ExpectedQty: 5.0, CountedQty: 5.0},
		},
	}

	id, err := repo.CreateStockTake(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("stock_takes.CreateStockTake: %v", err)
	}

	takes, _, err := repo.ListStockTakes(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("stock_takes.ListStockTakes: %v", err)
	}

	var found bool
	for _, st := range takes {
		if st.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created stock take id=%d not found in list", id)
	}
}

func TestCreateStockTake_ItemsLoaded(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.StockTakeInput{
		Items: []domain.StockTakeItemInput{
			{ProductID: 5201, ExpectedQty: 10.0, CountedQty: 8.0},
			{ProductID: 5203, ExpectedQty: 4.0, CountedQty: 4.0},
		},
	}

	id, err := repo.CreateStockTake(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("stock_takes.CreateStockTake: %v", err)
	}

	takes, _, err := repo.ListStockTakes(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("stock_takes.ListStockTakes: %v", err)
	}

	for _, st := range takes {
		if st.ID == id {
			if len(st.Items) != 2 {
				t.Errorf("expected 2 items, got %d", len(st.Items))
			}
			return
		}
	}
	t.Fatalf("created stock take id=%d not found in list", id)
}

func TestListStockTakes_HasMoreFalseWhenBelowLimit(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, hasMore, err := repo.ListStockTakes(ctx, testutil.OwnerID, 1000, 0)
	if err != nil {
		t.Fatalf("stock_takes.ListStockTakes: %v", err)
	}
	if hasMore {
		t.Error("expected hasMore=false when limit > total rows")
	}
}

func TestListStockTakes_HasMoreTrueWhenOverLimit(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Seed at least 2 stock takes
	for i := 0; i < 2; i++ {
		_, err := repo.CreateStockTake(ctx, testutil.OwnerID, domain.StockTakeInput{
			Items: []domain.StockTakeItemInput{
				{ProductID: 5201, ExpectedQty: 1.0, CountedQty: 1.0},
			},
		})
		if err != nil {
			t.Fatalf("seed stock take[%d]: %v", i, err)
		}
	}

	_, hasMore, err := repo.ListStockTakes(ctx, testutil.OwnerID, 1, 0)
	if err != nil {
		t.Fatalf("stock_takes.ListStockTakes(limit=1): %v", err)
	}
	if !hasMore {
		t.Error("expected hasMore=true when total rows > limit=1")
	}
}

func TestCreateStockTake_OwnerIDSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateStockTake(ctx, testutil.OwnerID, domain.StockTakeInput{
		Items: []domain.StockTakeItemInput{
			{ProductID: 5201, ExpectedQty: 2.0, CountedQty: 2.0},
		},
	})
	if err != nil {
		t.Fatalf("stock_takes.CreateStockTake: %v", err)
	}

	takes, _, err := repo.ListStockTakes(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("stock_takes.ListStockTakes: %v", err)
	}

	for _, st := range takes {
		if st.ID == id {
			if st.OwnerID != testutil.OwnerID {
				t.Errorf("OwnerID mismatch: got %d, want %d", st.OwnerID, testutil.OwnerID)
			}
			return
		}
	}
	t.Fatalf("created stock take id=%d not found", id)
}

func TestUpdateStockTake_StatusChanges(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateStockTake(ctx, testutil.OwnerID, domain.StockTakeInput{
		Items: []domain.StockTakeItemInput{
			{ProductID: 5201, ExpectedQty: 3.0, CountedQty: 3.0},
		},
	})
	if err != nil {
		t.Fatalf("seed stock take: %v", err)
	}

	err = repo.UpdateStockTake(ctx, testutil.OwnerID, domain.StockTakeUpdateInput{
		ID:     id,
		Status: domain.StockTakeStatusInProgress,
		Notes:  "in progress now",
	})
	if err != nil {
		t.Fatalf("stock_takes.UpdateStockTake: %v", err)
	}

	takes, _, err := repo.ListStockTakes(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("stock_takes.ListStockTakes after update: %v", err)
	}

	for _, st := range takes {
		if st.ID == id {
			if st.Status != domain.StockTakeStatusInProgress {
				t.Errorf("expected status=%s, got %s", domain.StockTakeStatusInProgress, st.Status)
			}
			return
		}
	}
	t.Fatalf("updated stock take id=%d not found", id)
}

func TestListStockTakes_PaginationOffset(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, err := repo.CreateStockTake(ctx, testutil.OwnerID, domain.StockTakeInput{
			Items: []domain.StockTakeItemInput{
				{ProductID: 5201, ExpectedQty: 1.0, CountedQty: 1.0},
			},
		})
		if err != nil {
			t.Fatalf("seed stock take[%d]: %v", i, err)
		}
	}

	page0, _, err := repo.ListStockTakes(ctx, testutil.OwnerID, 2, 0)
	if err != nil {
		t.Fatalf("List page 0: %v", err)
	}
	page1, _, err := repo.ListStockTakes(ctx, testutil.OwnerID, 2, 2)
	if err != nil {
		t.Fatalf("List page 1: %v", err)
	}

	if len(page0) > 0 && len(page1) > 0 {
		if page0[0].ID == page1[0].ID {
			t.Error("pagination offset not working: first items on both pages are identical")
		}
	}
}
