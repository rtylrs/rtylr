package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/erp/internal/outbox"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/erp/purchase_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/purchase_orders/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

type purchaseOrderReceiptLine struct {
	item  sqlc.ErpPurchaseOrderItem
	delta float64
}

type purchaseOrderInventoryLedgerEntry struct {
	ProductID      int64
	QuantityChange string
	QuantityBefore string
	QuantityAfter  string
	OwnerID        int64
	BusinessID     int64
	ReferenceID    int64
	CreatedBy      int64
	Notes          string
}

func NewRepository(db *sql.DB) *Repository {
	pool := db

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

type purchaseOrderItemReceivedUpdate struct {
	ID       int64
	Quantity string
}

func updatePurchaseOrderItemsReceived(ctx context.Context, tx *sql.Tx, purchaseOrderID int64, updates []purchaseOrderItemReceivedUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	caseExpr := sq.Case()
	ids := make([]int64, 0, len(updates))
	for _, update := range updates {
		caseExpr = caseExpr.When(sq.Expr("id = ?", update.ID), update.Quantity)
		ids = append(ids, update.ID)
	}
	query, args, err := sq.Update("erp_purchase_order_items").
		Set("quantity_received", caseExpr).
		Where(sq.Eq{"purchase_order_id": purchaseOrderID, "id": ids}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, query, args...)
	return err
}

func insertPurchaseOrderInventoryLedger(ctx context.Context, tx *sql.Tx, entries []purchaseOrderInventoryLedgerEntry) error {
	if len(entries) == 0 {
		return nil
	}

	insert := sq.Insert("erp_inventory_transactions").Columns(
		"owner_id", "business_id", "product_id", "type", "quantity_change", "quantity_before", "quantity_after",
		"reference_id", "reference_type", "created_by", "notes",
	)
	for _, entry := range entries {
		insert = insert.Values(
			entry.OwnerID,
			entry.BusinessID,
			entry.ProductID,
			sqlc.ErpInventoryTransactionsType(domain.InventoryTypePurchase),
			entry.QuantityChange,
			entry.QuantityBefore,
			entry.QuantityAfter,
			sqlutil.NullInt64(entry.ReferenceID),
			sqlutil.NullString("purchase_order"),
			sqlutil.NullInt64(entry.CreatedBy),
			sqlutil.NullString(entry.Notes),
		)
	}
	query, args, err := insert.ToSql()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) ListPurchaseOrders(ctx context.Context, businessID int64, status string, limit, offset int32) ([]domain.PurchaseOrder, error) {
	var rows []sqlc.ErpPurchaseOrder
	var err error

	if status != "" {
		rows, err = r.queries.ListPurchaseOrdersByStatus(ctx, sqlc.ListPurchaseOrdersByStatusParams{
			BusinessID: businessID,
			Status:     sqlc.NullErpPurchaseOrdersStatus{ErpPurchaseOrdersStatus: sqlc.ErpPurchaseOrdersStatus(status), Valid: true},
			Limit:      limit,
			Offset:     offset,
		})
	} else {
		rows, err = r.queries.ListPurchaseOrders(ctx, sqlc.ListPurchaseOrdersParams{
			BusinessID: businessID,
			Limit:      limit,
			Offset:     offset,
		})
	}

	if err != nil {
		logger.Ctx(ctx).Error("Failed to list purchase orders", logger.Error(err))
		return nil, err
	}

	orders := make([]domain.PurchaseOrder, 0, len(rows))
	for _, row := range rows {
		orders = append(orders, purchaseOrderFromRow(row))
	}

	return orders, nil
}

func (r *Repository) GetByID(ctx context.Context, businessID, id int64) (domain.PurchaseOrder, error) {
	row, err := r.queries.GetPurchaseOrderByID(ctx, sqlc.GetPurchaseOrderByIDParams{
		ID:         id,
		BusinessID: businessID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PurchaseOrder{}, domain.ErrPurchaseOrderNotFound
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get purchase order by ID", logger.Error(err))
		return domain.PurchaseOrder{}, err
	}
	po := purchaseOrderFromRow(row)
	items, itemErr := r.GetPurchaseOrderItems(ctx, po.ID)
	if itemErr != nil {
		logger.Ctx(ctx).Error("Failed to get purchase order items", logger.Error(itemErr))
		return domain.PurchaseOrder{}, itemErr
	}
	po.Items = items
	return po, nil
}

func (r *Repository) GetPurchaseOrderItems(ctx context.Context, purchaseOrderID int64) ([]domain.PurchaseOrderItem, error) {
	rows, err := r.queries.GetPurchaseOrderItems(ctx, purchaseOrderID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get purchase order items", logger.Error(err))
		return nil, err
	}

	items := make([]domain.PurchaseOrderItem, 0, len(rows))
	for _, row := range rows {
		var item domain.PurchaseOrderItem
		item.ID = row.ID
		item.PurchaseOrderID = row.PurchaseOrderID
		item.ProductID = row.ProductID
		item.QuantityOrdered = sqlutil.ParseDecimal(row.QuantityOrdered)
		item.UnitCost = sqlutil.ParseDecimal(row.UnitCost)
		item.TotalCost = sqlutil.ParseDecimal(row.TotalCost)

		if row.QuantityReceived.Valid {
			item.QuantityReceived = sqlutil.ParseDecimal(row.QuantityReceived.String)
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *Repository) CreatePurchaseOrder(ctx context.Context, businessID int64, input domain.PurchaseOrderInput, createdBy int64) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin transaction", logger.Error(err))
		return 0, err
	}

	// Always roll back when the function exits; after a successful commit the
	// rollback is a no-op (sql.ErrTxDone), and every early return is covered
	// even if it does not assign the outer err. The previous conditional defer
	// keyed on the outer err, which the reference-validation path below never
	// assigned, so a validation failure leaked the transaction's connection.
	defer func() { _ = tx.Rollback() }()

	qtx := r.queries.WithTx(tx)
	if err := validatePurchaseOrderReferences(ctx, qtx, businessID, input); err != nil {
		return 0, err
	}

	var subtotal float64
	for _, item := range input.Items {
		subtotal += item.QuantityOrdered * item.UnitCost
	}

	total := subtotal + input.TaxAmount

	var expectedDelivery sql.NullTime
	if input.ExpectedDelivery != "" {
		t, parseErr := time.Parse("2006-01-02", input.ExpectedDelivery)
		if parseErr == nil {
			expectedDelivery = sql.NullTime{Time: t, Valid: true}
		}
	}

	res, err := qtx.InsertPurchaseOrder(ctx, sqlc.InsertPurchaseOrderParams{
		OwnerID:          createdBy,
		BusinessID:       businessID,
		SupplierID:       input.SupplierID,
		PoNumber:         input.PONumber,
		Subtotal:         sql.NullString{String: strconv.FormatFloat(subtotal, 'f', 2, 64), Valid: true},
		TaxAmount:        sql.NullString{String: strconv.FormatFloat(input.TaxAmount, 'f', 2, 64), Valid: true},
		Total:            sql.NullString{String: strconv.FormatFloat(total, 'f', 2, 64), Valid: true},
		ExpectedDelivery: expectedDelivery,
		Notes:            sqlutil.NullString(input.Notes),
		CreatedBy:        sqlutil.NullInt64(createdBy),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert purchase order", logger.Error(err))
		return 0, err
	}

	poID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, item := range input.Items {
		itemTotal := item.QuantityOrdered * item.UnitCost

		_, err = qtx.InsertPurchaseOrderItem(ctx, sqlc.InsertPurchaseOrderItemParams{
			PurchaseOrderID: poID,
			ProductID:       item.ProductID,
			QuantityOrdered: strconv.FormatFloat(item.QuantityOrdered, 'f', -1, 64),
			UnitCost:        strconv.FormatFloat(item.UnitCost, 'f', -1, 64),
			TotalCost:       strconv.FormatFloat(itemTotal, 'f', -1, 64),
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to insert purchase order item", logger.Error(err))
			return 0, err
		}
	}

	if err := outbox.AppendAuditTx(ctx, tx, audit.Event{BusinessID: businessID, ActorID: createdBy, EntityType: "purchase_order", EntityID: poID, Action: "created", NewValues: purchaseOrderInputSnapshotOf(input)}); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit purchase order transaction", logger.Error(err))
		return 0, err
	}

	return poID, nil
}

func (r *Repository) Exists(ctx context.Context, businessID, id int64) (bool, error) {
	present, err := r.queries.PurchaseOrderExists(ctx, sqlc.PurchaseOrderExistsParams{BusinessID: businessID, ID: id})
	return present, err
}
