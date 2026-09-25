package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/stock_takes/internal/domain"
	"github.com/voxire/onion/internal/modules/stock_takes/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListStockTakes_ScopedToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.CreateStockTake(ctx, testutil.OwnerID, domain.StockTakeInput{
		Items: []domain.StockTakeItemInput{
			{ProductID: 5201, ExpectedQty: 1.0, CountedQty: 1.0},
		},
	})
	if err != nil {
		t.Fatalf("seed stock take: %v", err)
	}

	other, _, err := repo.ListStockTakes(ctx, testutil.OwnerID+9999, 50, 0)
	if err != nil {
		t.Fatalf("ListStockTakes(other owner): %v", err)
	}
	for _, st := range other {
		if st.OwnerID == testutil.OwnerID {
			t.Errorf("other-owner list contains ownerID=%d record", testutil.OwnerID)
		}
	}
}

func TestCreateStockTake_WithNotes_NotesPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wantNotes := "quarterly count"
	id, err := repo.CreateStockTake(ctx, testutil.OwnerID, domain.StockTakeInput{
		Notes: wantNotes,
		Items: []domain.StockTakeItemInput{
			{ProductID: 5201, ExpectedQty: 10.0, CountedQty: 9.0},
		},
	})
	if err != nil {
		t.Fatalf("CreateStockTake: %v", err)
	}

	takes, _, err := repo.ListStockTakes(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListStockTakes: %v", err)
	}
	for _, st := range takes {
		if st.ID == id {
			if st.Notes != wantNotes {
				t.Errorf("Notes mismatch: got %q, want %q", st.Notes, wantNotes)
			}
			return
		}
	}
	t.Fatalf("stock take id=%d not found", id)
}

func TestUpdateStockTake_StatusToCompleted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateStockTake(ctx, testutil.OwnerID, domain.StockTakeInput{
		Items: []domain.StockTakeItemInput{
			{ProductID: 5201, ExpectedQty: 5.0, CountedQty: 5.0},
		},
	})
	if err != nil {
		t.Fatalf("seed stock take: %v", err)
	}

	err = repo.UpdateStockTake(ctx, testutil.OwnerID, domain.StockTakeUpdateInput{
		ID:     id,
		Status: domain.StockTakeStatusCompleted,
		Notes:  "completed run",
		Items: []domain.StockTakeItemInput{
			{ProductID: 5201, CountedQty: 5},
		},
	})
	if err != nil {
		t.Fatalf("UpdateStockTake to completed: %v", err)
	}

	takes, _, err := repo.ListStockTakes(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListStockTakes: %v", err)
	}
	for _, st := range takes {
		if st.ID == id {
			if st.Status != domain.StockTakeStatusCompleted {
				t.Errorf("expected status completed, got %s", st.Status)
			}
			return
		}
	}
	t.Fatalf("stock take id=%d not found after update", id)
}

func TestUpdateStockTake_CompletedUsesSelectedWarehouse(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	warehouseName := fmt.Sprintf("Stock take warehouse %d", time.Now().UnixNano())
	result, err := db.ExecContext(ctx, `
		INSERT INTO erp_warehouses (owner_id, business_id, name, capacity, is_default, is_active)
		VALUES (?, ?, ?, 0, FALSE, TRUE)
	`, testutil.OwnerID, testutil.OwnerID, warehouseName)
	if err != nil {
		t.Fatalf("create warehouse: %v", err)
	}
	warehouseID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("get warehouse id: %v", err)
	}

	var stockTakeID int64
	t.Cleanup(func() {
		if stockTakeID != 0 {
			_, _ = db.ExecContext(ctx, "DELETE FROM erp_inventory_transactions WHERE reference_type = 'stock_take' AND reference_id = ?", stockTakeID)
			_, _ = db.ExecContext(ctx, "DELETE FROM stock_ledger WHERE reference_type = 'stock_take' AND reference_id = ?", stockTakeID)
			_, _ = db.ExecContext(ctx, "DELETE FROM erp_stock_take_items WHERE stock_take_id = ?", stockTakeID)
			_, _ = db.ExecContext(ctx, "DELETE FROM erp_stock_takes WHERE id = ? AND business_id = ?", stockTakeID, testutil.OwnerID)
		}
		_, _ = db.ExecContext(ctx, `
			DELETE sl FROM stock_levels sl
			JOIN stock_locations l ON l.id = sl.location_id
			WHERE l.source_table = 'erp_warehouses' AND l.source_id = ? AND l.owner_id = ?
		`, warehouseID, testutil.OwnerID)
		_, _ = db.ExecContext(ctx, "DELETE FROM stock_locations WHERE source_table = 'erp_warehouses' AND source_id = ? AND owner_id = ?", warehouseID, testutil.OwnerID)
		_, _ = db.ExecContext(ctx, "DELETE FROM erp_warehouses WHERE id = ? AND business_id = ?", warehouseID, testutil.OwnerID)
	})

	stockTakeID, err = repo.CreateStockTake(ctx, testutil.OwnerID, domain.StockTakeInput{
		WarehouseID: warehouseID,
		Items: []domain.StockTakeItemInput{
			{ProductID: 5201, ExpectedQty: 999, CountedQty: 0},
		},
	})
	if err != nil {
		t.Fatalf("create warehouse stock take: %v", err)
	}
	var snapshottedExpected string
	if err := db.QueryRowContext(ctx, `
		SELECT expected_qty FROM erp_stock_take_items
		WHERE stock_take_id = ? AND product_id = ?`, stockTakeID, 5201).Scan(&snapshottedExpected); err != nil {
		t.Fatalf("read warehouse-scoped expected quantity: %v", err)
	}
	if snapshottedExpected != "0" && snapshottedExpected != "0.000" {
		t.Fatalf("warehouse-scoped expected quantity = %q, want zero", snapshottedExpected)
	}

	if err := repo.UpdateStockTake(ctx, testutil.OwnerID, domain.StockTakeUpdateInput{
		ID:     stockTakeID,
		Status: domain.StockTakeStatusCompleted,
		Items: []domain.StockTakeItemInput{
			{ProductID: 5201, CountedQty: 0},
		},
	}); err != nil {
		t.Fatalf("complete warehouse stock take: %v", err)
	}

	var onHand string
	err = db.QueryRowContext(ctx, `
		SELECT sl.on_hand
		FROM stock_levels sl
		JOIN stock_locations l ON l.id = sl.location_id
		WHERE sl.owner_id = ? AND sl.product_id = ?
		  AND l.owner_id = ? AND l.source_table = 'erp_warehouses' AND l.source_id = ?
	`, testutil.OwnerID, 5201, testutil.OwnerID, warehouseID).Scan(&onHand)
	if err != nil {
		t.Fatalf("selected warehouse stock level: %v", err)
	}
	if onHand != "0" && onHand != "0.000" {
		t.Fatalf("selected warehouse on-hand = %q, want zero", onHand)
	}
}

func TestUpdateStockTake_NotesPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateStockTake(ctx, testutil.OwnerID, domain.StockTakeInput{
		Notes: "initial note",
		Items: []domain.StockTakeItemInput{
			{ProductID: 5201, ExpectedQty: 2.0, CountedQty: 2.0},
		},
	})
	if err != nil {
		t.Fatalf("seed stock take: %v", err)
	}

	wantNotes := "updated note"
	err = repo.UpdateStockTake(ctx, testutil.OwnerID, domain.StockTakeUpdateInput{
		ID:     id,
		Status: domain.StockTakeStatusInProgress,
		Notes:  wantNotes,
	})
	if err != nil {
		t.Fatalf("UpdateStockTake: %v", err)
	}

	takes, _, err := repo.ListStockTakes(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListStockTakes: %v", err)
	}
	for _, st := range takes {
		if st.ID == id {
			if st.Notes != wantNotes {
				t.Errorf("Notes mismatch: got %q, want %q", st.Notes, wantNotes)
			}
			return
		}
	}
	t.Fatalf("stock take id=%d not found", id)
}

func TestListStockTakes_LimitOnePaginates(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		_, err := repo.CreateStockTake(ctx, testutil.OwnerID, domain.StockTakeInput{
			Items: []domain.StockTakeItemInput{
				{ProductID: 5201, ExpectedQty: 1.0, CountedQty: 1.0},
			},
		})
		if err != nil {
			t.Fatalf("seed[%d]: %v", i, err)
		}
	}

	takes, _, err := repo.ListStockTakes(ctx, testutil.OwnerID, 1, 0)
	if err != nil {
		t.Fatalf("ListStockTakes limit=1: %v", err)
	}
	if len(takes) > 1 {
		t.Errorf("expected at most 1 result with limit=1, got %d", len(takes))
	}
}

func TestListStockTakes_HasMoreFlagWithLimitOne(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		_, err := repo.CreateStockTake(ctx, testutil.OwnerID, domain.StockTakeInput{
			Items: []domain.StockTakeItemInput{
				{ProductID: 5201, ExpectedQty: 1.0, CountedQty: 1.0},
			},
		})
		if err != nil {
			t.Fatalf("seed[%d]: %v", i, err)
		}
	}

	_, hasMore, err := repo.ListStockTakes(ctx, testutil.OwnerID, 1, 0)
	if err != nil {
		t.Fatalf("ListStockTakes: %v", err)
	}
	if !hasMore {
		t.Error("expected hasMore=true when more than 1 record exists and limit=1")
	}
}
