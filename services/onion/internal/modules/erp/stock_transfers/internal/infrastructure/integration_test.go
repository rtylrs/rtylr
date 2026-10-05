package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/erp/stock_transfers/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/stock_transfers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListStockTransfers_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	transfers, _, err := repo.ListStockTransfers(ctx, testutil.OwnerID, 20, 0)
	if err != nil {
		t.Fatalf("stock_transfers.ListStockTransfers: unexpected error: %v", err)
	}
	if transfers == nil {
		t.Fatal("stock_transfers.ListStockTransfers: expected non-nil slice, got nil")
	}
}

func TestListStockTransfers_WrongOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	transfers, _, err := repo.ListStockTransfers(ctx, 0, 20, 0)
	if err != nil {
		t.Fatalf("stock_transfers.ListStockTransfers(ownerID=0): unexpected error: %v", err)
	}
	if len(transfers) != 0 {
		t.Fatalf("stock_transfers.ListStockTransfers(ownerID=0): expected 0 rows, got %d", len(transfers))
	}
}

func TestCreateStockTransfer_ReturnsInsertedID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.StockTransferInput{
		FromWarehouseID: 1,
		ToWarehouseID:   2,
		Notes:           "Integration test transfer",
		Items: []domain.StockTransferItemInput{
			{ProductID: 5201, Quantity: 3.0},
		},
	}

	id, err := repo.CreateStockTransfer(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("stock_transfers.CreateStockTransfer: unexpected error: %v", err)
	}
	if id <= 0 {
		t.Fatalf("stock_transfers.CreateStockTransfer: expected positive ID, got %d", id)
	}
}

func TestCreateStockTransfer_AuditFailureRollsBackTransfer(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const notes = "audit rollback transfer"

	id, err := repo.CreateStockTransfer(ctx, testutil.OwnerID, domain.StockTransferInput{
		FromWarehouseID: 1,
		ToWarehouseID:   2,
		Notes:           notes,
		Items:           []domain.StockTransferItemInput{{ProductID: 5201, Quantity: 1}},
	}, audit.Event{
		BusinessID: testutil.OwnerID,
		EntityType: "stock_transfer",
		Action:     "created",
		NewValues:  make(chan struct{}),
	})
	if err == nil {
		t.Fatalf("CreateStockTransfer succeeded with id=%d despite audit failure", id)
	}

	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM erp_stock_transfers
		WHERE business_id = ? AND notes = ?`, testutil.OwnerID, notes).Scan(&count); err != nil {
		t.Fatalf("count rolled-back transfer: %v", err)
	}
	if count != 0 {
		t.Fatalf("transfer count after audit rollback = %d, want 0", count)
	}
}

func TestListStockTransfers_ReturnsEveryItemWithoutGlobalCap(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	items := make([]domain.StockTransferItemInput, 201)
	for i := range items {
		items[i] = domain.StockTransferItemInput{ProductID: 5201, Quantity: 1}
	}

	id, err := repo.CreateStockTransfer(ctx, testutil.OwnerID, domain.StockTransferInput{
		FromWarehouseID: 1,
		ToWarehouseID:   2,
		Notes:           "more than global item cap",
		Items:           items,
	})
	if err != nil {
		t.Fatalf("CreateStockTransfer: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM erp_stock_transfer_items WHERE transfer_id = ?", id)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM erp_stock_transfers WHERE id = ? AND business_id = ?", id, testutil.OwnerID)
	})

	transfers, _, err := repo.ListStockTransfers(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListStockTransfers: %v", err)
	}
	for _, transfer := range transfers {
		if transfer.ID == id {
			if len(transfer.Items) != len(items) {
				t.Fatalf("transfer item count = %d, want %d", len(transfer.Items), len(items))
			}
			return
		}
	}
	t.Fatalf("created transfer id=%d not found", id)
}

func TestUpdateStockTransferStatus_AuditFailureRollsBackStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	id, err := repo.CreateStockTransfer(ctx, testutil.OwnerID, domain.StockTransferInput{
		FromWarehouseID: 1,
		ToWarehouseID:   2,
		Items:           []domain.StockTransferItemInput{{ProductID: 5201, Quantity: 1}},
	})
	if err != nil {
		t.Fatalf("CreateStockTransfer: %v", err)
	}

	err = repo.UpdateStockTransferStatus(ctx, testutil.OwnerID, id, domain.StockTransferStatusInTransit, audit.Event{
		BusinessID: testutil.OwnerID, EntityType: "stock_transfer", EntityID: id,
		Action: "status_changed", NewValues: make(chan struct{}),
	})
	if err == nil {
		t.Fatal("UpdateStockTransferStatus succeeded despite audit failure")
	}

	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM erp_stock_transfers WHERE id = ?`, id).Scan(&status); err != nil {
		t.Fatalf("read rolled-back status: %v", err)
	}
	if status != string(domain.StockTransferStatusDraft) {
		t.Fatalf("status after audit rollback = %q, want draft", status)
	}
}

func TestReceiveStockTransfer_AuditFailureRollsBackStockAndStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	id, err := repo.CreateStockTransfer(ctx, testutil.OwnerID, domain.StockTransferInput{
		FromWarehouseID: 1,
		ToWarehouseID:   2,
		Items:           []domain.StockTransferItemInput{{ProductID: 5201, Quantity: 1}},
	})
	if err != nil {
		t.Fatalf("CreateStockTransfer: %v", err)
	}

	err = repo.UpdateStockTransferStatus(ctx, testutil.OwnerID, id, domain.StockTransferStatusReceived, audit.Event{
		BusinessID: testutil.OwnerID, EntityType: "stock_transfer", EntityID: id,
		Action: "status_changed", NewValues: make(chan struct{}),
	})
	if err == nil {
		t.Fatal("receive succeeded despite audit failure")
	}
	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM erp_stock_transfers WHERE id = ?`, id).Scan(&status); err != nil {
		t.Fatalf("read rolled-back receive status: %v", err)
	}
	if status != string(domain.StockTransferStatusDraft) {
		t.Fatalf("status after receive audit rollback = %q, want draft", status)
	}
	var ledgerCount int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM stock_ledger
		WHERE owner_id = ? AND reference_type = 'stock_transfer' AND reference_id = ?`,
		testutil.OwnerID, id).Scan(&ledgerCount); err != nil {
		t.Fatalf("count rolled-back receive ledger entries: %v", err)
	}
	if ledgerCount != 0 {
		t.Fatalf("ledger entries after receive audit rollback = %d, want 0", ledgerCount)
	}
}

func TestCreateStockTransfer_RejectsUnscopedWarehouseReference(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	_, err := repo.CreateStockTransfer(context.Background(), testutil.OwnerID, domain.StockTransferInput{
		FromWarehouseID: 999999999,
		ToWarehouseID:   2,
		Items: []domain.StockTransferItemInput{
			{ProductID: 5201, Quantity: 1},
		},
	})
	if !errors.Is(err, domain.ErrStockTransferReferenceNotFound) {
		t.Fatalf("CreateStockTransfer cross-business warehouse error = %v, want ErrStockTransferReferenceNotFound", err)
	}
}

func TestCreateStockTransfer_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.StockTransferInput{
		FromWarehouseID: 1,
		ToWarehouseID:   2,
		Notes:           "Appears In List",
		Items: []domain.StockTransferItemInput{
			{ProductID: 5202, Quantity: 2.0},
		},
	}

	id, err := repo.CreateStockTransfer(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("stock_transfers.CreateStockTransfer: %v", err)
	}

	transfers, _, err := repo.ListStockTransfers(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("stock_transfers.ListStockTransfers: %v", err)
	}

	var found bool
	for _, tr := range transfers {
		if tr.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created stock transfer id=%d not found in list", id)
	}
}

func TestCreateStockTransfer_OwnerIDSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateStockTransfer(ctx, testutil.OwnerID, domain.StockTransferInput{
		FromWarehouseID: 1,
		ToWarehouseID:   2,
		Items: []domain.StockTransferItemInput{
			{ProductID: 5201, Quantity: 1.0},
		},
	})
	if err != nil {
		t.Fatalf("stock_transfers.CreateStockTransfer: %v", err)
	}

	transfers, _, err := repo.ListStockTransfers(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("stock_transfers.ListStockTransfers: %v", err)
	}

	for _, tr := range transfers {
		if tr.ID == id {
			if tr.OwnerID != testutil.OwnerID {
				t.Errorf("OwnerID mismatch: got %d, want %d", tr.OwnerID, testutil.OwnerID)
			}
			return
		}
	}
	t.Fatalf("created stock transfer id=%d not found", id)
}

func TestCreateStockTransfer_DefaultStatusIsPending(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateStockTransfer(ctx, testutil.OwnerID, domain.StockTransferInput{
		FromWarehouseID: 1,
		ToWarehouseID:   2,
		Items: []domain.StockTransferItemInput{
			{ProductID: 5201, Quantity: 1.0},
		},
	})
	if err != nil {
		t.Fatalf("stock_transfers.CreateStockTransfer: %v", err)
	}

	transfers, _, err := repo.ListStockTransfers(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("stock_transfers.ListStockTransfers: %v", err)
	}

	for _, tr := range transfers {
		if tr.ID == id {
			if tr.Status != "draft" {
				t.Errorf("expected default status=draft, got %q", tr.Status)
			}
			return
		}
	}
	t.Fatalf("created stock transfer id=%d not found", id)
}

func TestUpdateStockTransferStatus_ChangesToCompleted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateStockTransfer(ctx, testutil.OwnerID, domain.StockTransferInput{
		FromWarehouseID: 1,
		ToWarehouseID:   2,
		Items: []domain.StockTransferItemInput{
			{ProductID: 5201, Quantity: 1.0},
		},
	})
	if err != nil {
		t.Fatalf("stock_transfers.CreateStockTransfer: %v", err)
	}

	err = repo.UpdateStockTransferStatus(ctx, testutil.OwnerID, id, "received")
	if err != nil {
		t.Fatalf("stock_transfers.UpdateStockTransferStatus: %v", err)
	}

	transfers, _, err := repo.ListStockTransfers(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("stock_transfers.ListStockTransfers after update: %v", err)
	}

	for _, tr := range transfers {
		if tr.ID == id {
			if tr.Status != "received" {
				t.Errorf("expected status=received, got %q", tr.Status)
			}
			return
		}
	}
	t.Fatalf("updated stock transfer id=%d not found", id)
}

func TestUpdateStockTransferStatus_UnknownID_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	err := repo.UpdateStockTransferStatus(context.Background(), testutil.BusinessID, 999999999, domain.StockTransferStatusDraft)
	if !errors.Is(err, domain.ErrStockTransferNotFound) {
		t.Fatalf("UpdateStockTransferStatus unknown ID error = %v, want ErrStockTransferNotFound", err)
	}
}

func TestReceivedStockTransfer_CannotBeReopenedOrMovedAgain(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateStockTransfer(ctx, testutil.OwnerID, domain.StockTransferInput{
		FromWarehouseID: 1,
		ToWarehouseID:   2,
		Items: []domain.StockTransferItemInput{
			{ProductID: 5201, Quantity: 1.0},
		},
	})
	if err != nil {
		t.Fatalf("stock_transfers.CreateStockTransfer: %v", err)
	}

	if err := repo.UpdateStockTransferStatus(ctx, testutil.OwnerID, id, domain.StockTransferStatusReceived); err != nil {
		t.Fatalf("first receive: %v", err)
	}

	var initialLedgerCount int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM stock_ledger
		WHERE owner_id = ? AND reference_type = 'stock_transfer' AND reference_id = ?`,
		testutil.OwnerID, id).Scan(&initialLedgerCount); err != nil {
		t.Fatalf("count initial stock-transfer ledger entries: %v", err)
	}
	if initialLedgerCount != 2 {
		t.Fatalf("expected two ledger entries after first receive, got %d", initialLedgerCount)
	}

	for _, status := range []domain.StockTransferStatus{
		domain.StockTransferStatusDraft,
		domain.StockTransferStatusInTransit,
		domain.StockTransferStatusCancelled,
	} {
		if err := repo.UpdateStockTransferStatus(ctx, testutil.OwnerID, id, status); err != nil {
			t.Fatalf("reopen received transfer as %q: %v", status, err)
		}

		transfer, err := repo.GetByID(ctx, testutil.OwnerID, id)
		if err != nil {
			t.Fatalf("get transfer after attempted %q reopen: %v", status, err)
		}
		if transfer.Status != domain.StockTransferStatusReceived {
			t.Fatalf("received transfer changed to %q after attempted %q reopen", transfer.Status, status)
		}
	}

	if err := repo.UpdateStockTransferStatus(ctx, testutil.OwnerID, id, domain.StockTransferStatusReceived); err != nil {
		t.Fatalf("second receive: %v", err)
	}

	var finalLedgerCount int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM stock_ledger
		WHERE owner_id = ? AND reference_type = 'stock_transfer' AND reference_id = ?`,
		testutil.OwnerID, id).Scan(&finalLedgerCount); err != nil {
		t.Fatalf("count final stock-transfer ledger entries: %v", err)
	}
	if finalLedgerCount != initialLedgerCount {
		t.Fatalf("expected receive to remain idempotent with %d ledger entries, got %d", initialLedgerCount, finalLedgerCount)
	}
}

func TestCreateStockTransfer_ItemsLoaded(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.StockTransferInput{
		FromWarehouseID: 1,
		ToWarehouseID:   2,
		Items: []domain.StockTransferItemInput{
			{ProductID: 5201, Quantity: 2.0},
			{ProductID: 5202, Quantity: 5.0},
		},
	}

	id, err := repo.CreateStockTransfer(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("stock_transfers.CreateStockTransfer: %v", err)
	}

	transfers, _, err := repo.ListStockTransfers(ctx, testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("stock_transfers.ListStockTransfers: %v", err)
	}

	for _, tr := range transfers {
		if tr.ID == id {
			if len(tr.Items) != 2 {
				t.Errorf("expected 2 transfer items, got %d", len(tr.Items))
			}
			return
		}
	}
	t.Fatalf("created stock transfer id=%d not found", id)
}

func TestListStockTransfers_HasMoreFalseWhenBelowLimit(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, hasMore, err := repo.ListStockTransfers(ctx, testutil.OwnerID, 1000, 0)
	if err != nil {
		t.Fatalf("stock_transfers.ListStockTransfers: %v", err)
	}
	if hasMore {
		t.Error("expected hasMore=false when limit > total rows")
	}
}

func TestListStockTransfers_HasMoreTrueWhenOverLimit(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Seed at least 2
	for i := 0; i < 2; i++ {
		_, err := repo.CreateStockTransfer(ctx, testutil.OwnerID, domain.StockTransferInput{
			FromWarehouseID: 1,
			ToWarehouseID:   2,
			Items: []domain.StockTransferItemInput{
				{ProductID: 5201, Quantity: 1.0},
			},
		})
		if err != nil {
			t.Fatalf("seed transfer[%d]: %v", i, err)
		}
	}

	_, hasMore, err := repo.ListStockTransfers(ctx, testutil.OwnerID, 1, 0)
	if err != nil {
		t.Fatalf("stock_transfers.ListStockTransfers(limit=1): %v", err)
	}
	if !hasMore {
		t.Error("expected hasMore=true when total rows > limit=1")
	}
}

func TestListStockTransfers_PaginationOffset(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, err := repo.CreateStockTransfer(ctx, testutil.OwnerID, domain.StockTransferInput{
			FromWarehouseID: 1,
			ToWarehouseID:   2,
			Items: []domain.StockTransferItemInput{
				{ProductID: 5201, Quantity: 1.0},
			},
		})
		if err != nil {
			t.Fatalf("seed transfer[%d]: %v", i, err)
		}
	}

	page0, _, err := repo.ListStockTransfers(ctx, testutil.OwnerID, 2, 0)
	if err != nil {
		t.Fatalf("List page 0: %v", err)
	}
	page1, _, err := repo.ListStockTransfers(ctx, testutil.OwnerID, 2, 2)
	if err != nil {
		t.Fatalf("List page 1: %v", err)
	}

	if len(page0) > 0 && len(page1) > 0 {
		if page0[0].ID == page1[0].ID {
			t.Error("pagination offset not working: first items on both pages are identical")
		}
	}
}

func TestListStockTransfers_ScopedToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.CreateStockTransfer(ctx, testutil.OwnerID, domain.StockTransferInput{
		FromWarehouseID: 1,
		ToWarehouseID:   2,
		Items: []domain.StockTransferItemInput{
			{ProductID: 5201, Quantity: 1.0},
		},
	})
	if err != nil {
		t.Fatalf("seed transfer: %v", err)
	}

	other, _, err := repo.ListStockTransfers(ctx, testutil.OwnerID+9999, 50, 0)
	if err != nil {
		t.Fatalf("ListStockTransfers(other owner): %v", err)
	}
	for _, tr := range other {
		if tr.OwnerID == testutil.OwnerID {
			t.Errorf("other-owner list contains ownerID=%d record", testutil.OwnerID)
		}
	}
}
