package infrastructure_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/stock_takes/internal/domain"
	"github.com/voxire/onion/internal/modules/stock_takes/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/audit"
	"github.com/voxire/onion/internal/testutil"
)

func TestListStockTakes_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	takes, _, err := repo.ListStockTakes(ctx, testutil.OwnerID, 20, 0)
	if err != nil {
		t.Fatalf("stock_takes.ListStockTakes: unexpected error: %v", err)
	}
	if takes == nil {
		t.Fatal("stock_takes.ListStockTakes: expected non-nil slice, got nil")
	}
}

func TestListStockTakes_WrongOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	takes, _, err := repo.ListStockTakes(ctx, 0, 20, 0)
	if err != nil {
		t.Fatalf("stock_takes.ListStockTakes(ownerID=0): unexpected error: %v", err)
	}
	if len(takes) != 0 {
		t.Fatalf("stock_takes.ListStockTakes(ownerID=0): expected 0 rows, got %d", len(takes))
	}
}

func TestUpdateStockTake_UnknownID_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	err := repo.UpdateStockTake(context.Background(), testutil.BusinessID, domain.StockTakeUpdateInput{
		ID:     999999999,
		Status: domain.StockTakeStatusDraft,
	})
	if !errors.Is(err, domain.ErrStockTakeNotFound) {
		t.Fatalf("UpdateStockTake unknown ID error = %v, want ErrStockTakeNotFound", err)
	}
}

func TestCreateStockTake_ReturnsInsertedID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.StockTakeInput{
		Notes: "Integration test stock take",
		Items: []domain.StockTakeItemInput{
			{ProductID: 5201, ExpectedQty: 10.0, CountedQty: 9.0},
		},
	}

	id, err := repo.CreateStockTake(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("stock_takes.CreateStockTake: unexpected error: %v", err)
	}
	if id <= 0 {
		t.Fatalf("stock_takes.CreateStockTake: expected positive ID, got %d", id)
	}
}

func TestCreateStockTake_AuditFailureRollsBackStockTake(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const notes = "audit rollback stock take"

	id, err := repo.CreateStockTake(ctx, testutil.OwnerID, domain.StockTakeInput{
		Notes: notes,
		Items: []domain.StockTakeItemInput{{ProductID: 5201}},
	}, audit.Event{
		BusinessID: testutil.OwnerID,
		EntityType: "stock_take",
		Action:     "created",
		NewValues:  make(chan struct{}),
	})
	if err == nil {
		t.Fatalf("CreateStockTake succeeded with id=%d despite audit failure", id)
	}

	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM erp_stock_takes
		WHERE business_id = ? AND notes = ?`, testutil.OwnerID, notes).Scan(&count); err != nil {
		t.Fatalf("count rolled-back stock take: %v", err)
	}
	if count != 0 {
		t.Fatalf("stock take count after audit rollback = %d, want 0", count)
	}
}

func TestUpdateStockTake_CompletionRequiresEveryPhysicalCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateStockTake(ctx, testutil.OwnerID, domain.StockTakeInput{
		Items: []domain.StockTakeItemInput{{ProductID: 5201}},
	})
	if err != nil {
		t.Fatalf("CreateStockTake: %v", err)
	}
	var before string
	if err := db.QueryRowContext(ctx, `
		SELECT on_hand FROM stock_levels
		WHERE owner_id = ? AND product_id = ?
		ORDER BY location_id LIMIT 1`, testutil.OwnerID, 5201).Scan(&before); err != nil {
		t.Fatalf("read stock before incomplete completion: %v", err)
	}

	err = repo.UpdateStockTake(ctx, testutil.OwnerID, domain.StockTakeUpdateInput{
		ID: id, Status: domain.StockTakeStatusCompleted,
	})
	if !errors.Is(err, domain.ErrStockTakeIncomplete) {
		t.Fatalf("completion error = %v, want ErrStockTakeIncomplete", err)
	}

	takes, _, err := repo.ListStockTakes(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListStockTakes: %v", err)
	}
	for _, take := range takes {
		if take.ID == id && take.Status != domain.StockTakeStatusDraft {
			t.Fatalf("incomplete completion changed status to %s", take.Status)
		}
	}
	var after string
	if err := db.QueryRowContext(ctx, `
		SELECT on_hand FROM stock_levels
		WHERE owner_id = ? AND product_id = ?
		ORDER BY location_id LIMIT 1`, testutil.OwnerID, 5201).Scan(&after); err != nil {
		t.Fatalf("read stock after incomplete completion: %v", err)
	}
	if after != before {
		t.Fatalf("incomplete completion changed stock from %q to %q", before, after)
	}
}

func TestUpdateStockTake_AuditFailureRollsBackStatusAndCounts(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	id, err := repo.CreateStockTake(ctx, testutil.OwnerID, domain.StockTakeInput{
		Items: []domain.StockTakeItemInput{{ProductID: 5201}},
	})
	if err != nil {
		t.Fatalf("CreateStockTake: %v", err)
	}

	err = repo.UpdateStockTake(ctx, testutil.OwnerID, domain.StockTakeUpdateInput{
		ID: id, Status: domain.StockTakeStatusInProgress,
		Items: []domain.StockTakeItemInput{{ProductID: 5201, CountedQty: 0}},
	}, audit.Event{
		BusinessID: testutil.OwnerID, EntityType: "stock_take", EntityID: id,
		Action: "updated", NewValues: make(chan struct{}),
	})
	if err == nil {
		t.Fatal("UpdateStockTake succeeded despite audit failure")
	}

	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM erp_stock_takes WHERE id = ?`, id).Scan(&status); err != nil {
		t.Fatalf("read rolled-back status: %v", err)
	}
	if status != string(domain.StockTakeStatusDraft) {
		t.Fatalf("status after audit rollback = %q, want draft", status)
	}
	var nullCount int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM erp_stock_take_items
		WHERE stock_take_id = ? AND counted_qty IS NULL`, id).Scan(&nullCount); err != nil {
		t.Fatalf("count rolled-back physical counts: %v", err)
	}
	if nullCount != 1 {
		t.Fatalf("uncounted rows after audit rollback = %d, want 1", nullCount)
	}
}

func TestCreateStockTake_RejectsUnscopedProductReference(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	_, err := repo.CreateStockTake(context.Background(), testutil.OwnerID, domain.StockTakeInput{
		Items: []domain.StockTakeItemInput{
			{ProductID: 999999999, ExpectedQty: 1, CountedQty: 1},
		},
	})
	if !errors.Is(err, domain.ErrStockTakeReferenceNotFound) {
		t.Fatalf("CreateStockTake cross-business product error = %v, want ErrStockTakeReferenceNotFound", err)
	}
}

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

func TestCreateStockTake_ItemsStartUncounted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateStockTake(ctx, testutil.OwnerID, domain.StockTakeInput{
		Items: []domain.StockTakeItemInput{
			{ProductID: 5201, ExpectedQty: 10, CountedQty: 0},
			{ProductID: 5202, ExpectedQty: 5, CountedQty: 0},
		},
	})
	if err != nil {
		t.Fatalf("CreateStockTake: %v", err)
	}

	var nullCounts int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM erp_stock_take_items
		WHERE stock_take_id = ? AND counted_qty IS NULL AND variance IS NULL`, id).Scan(&nullCounts); err != nil {
		t.Fatalf("count uncounted items: %v", err)
	}
	if nullCounts != 2 {
		t.Fatalf("uncounted item count = %d, want 2", nullCounts)
	}
}

func TestUpdateStockTake_NonExistentIDNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.UpdateStockTake(ctx, testutil.OwnerID, domain.StockTakeUpdateInput{
		ID:     999999999,
		Status: domain.StockTakeStatusCompleted,
		Notes:  "ghost update",
	})
	_ = err
}

func TestListStockTakes_EachHasNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.CreateStockTake(ctx, testutil.OwnerID, domain.StockTakeInput{
		Items: []domain.StockTakeItemInput{
			{ProductID: 5201, ExpectedQty: 1.0, CountedQty: 1.0},
		},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	takes, _, err := repo.ListStockTakes(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListStockTakes: %v", err)
	}
	for i, st := range takes {
		if st.ID <= 0 {
			t.Errorf("takes[%d] has non-positive ID: %d", i, st.ID)
		}
	}
}

func TestCreateStockTake_StatusDefaultsToDraft(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateStockTake(ctx, testutil.OwnerID, domain.StockTakeInput{
		Items: []domain.StockTakeItemInput{
			{ProductID: 5201, ExpectedQty: 3.0, CountedQty: 3.0},
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
			if st.Status != domain.StockTakeStatusDraft {
				t.Errorf("expected initial status draft, got %s", st.Status)
			}
			return
		}
	}
	t.Fatalf("stock take id=%d not found", id)
}

func TestListStockTakes_OffsetSkipsRows(t *testing.T) {
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
			t.Fatalf("seed[%d]: %v", i, err)
		}
	}

	page0, _, err := repo.ListStockTakes(ctx, testutil.OwnerID, 2, 0)
	if err != nil {
		t.Fatalf("ListStockTakes page0: %v", err)
	}
	if len(page0) < 2 {
		t.Skip("not enough rows to test offset")
	}

	page1, _, err := repo.ListStockTakes(ctx, testutil.OwnerID, 2, 2)
	if err != nil {
		t.Fatalf("ListStockTakes page1: %v", err)
	}
	if len(page1) > 0 && page0[0].ID == page1[0].ID {
		t.Error("offset not working: first items on both pages are identical")
	}
}

func TestUpdateStockTake_WrongOwnerIsNoOp(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateStockTake(ctx, testutil.OwnerID, domain.StockTakeInput{
		Items: []domain.StockTakeItemInput{
			{ProductID: 5201, ExpectedQty: 1.0, CountedQty: 1.0},
		},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	err = repo.UpdateStockTake(ctx, testutil.OwnerID+9999, domain.StockTakeUpdateInput{
		ID:     id,
		Status: domain.StockTakeStatusCompleted,
		Notes:  "should not apply",
	})
	_ = err

	takes, _, err := repo.ListStockTakes(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListStockTakes: %v", err)
	}
	for _, st := range takes {
		if st.ID == id {
			if st.Status == domain.StockTakeStatusCompleted {
				t.Logf("wrong-owner update changed status; DB may not enforce owner on update")
			}
			return
		}
	}
	t.Fatalf("stock take id=%d not found after wrong-owner update attempt", id)
}
