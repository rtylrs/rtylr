package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/erp/stock_takes/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/stock_takes/internal/infrastructure"
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
