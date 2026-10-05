package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/voxire/onion/internal/modules/erp/internal/outbox"
	"math"
	"strconv"
	"strings"
	"time"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	"github.com/voxire/onion/internal/modules/erp/internal/stock"
	"github.com/voxire/onion/internal/modules/erp/onlineorders/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/onlineorders/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{queries: sqlc.New(pool), pool: pool}
}

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.StoreOrder, error) {
	rows, err := r.queries.ListStoreOrders(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list store orders", logger.Error(err))
		return nil, err
	}
	items := make([]domain.StoreOrder, 0, len(rows))
	for _, row := range rows {
		items = append(items, orderFromRow(ctx, row))
	}
	return items, nil
}

// ListPaginated returns a single page of store orders for the business,
// ordered by creation date. Limit and offset are applied as-is; callers clamp
// them.
func (r *Repository) ListPaginated(ctx context.Context, businessID int64, limit, offset int32) ([]domain.StoreOrder, error) {
	rows, err := r.queries.ListStoreOrdersPaginated(ctx, sqlc.ListStoreOrdersPaginatedParams{
		BusinessID: businessID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list store orders page", logger.Error(err))
		return nil, err
	}
	items := make([]domain.StoreOrder, 0, len(rows))
	for _, row := range rows {
		items = append(items, orderFromRow(ctx, row))
	}
	return items, nil
}

// Count returns the total number of store orders for the business.
func (r *Repository) Count(ctx context.Context, businessID int64) (int64, error) {
	count, err := r.queries.CountStoreOrders(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count store orders", logger.Error(err))
		return 0, err
	}
	return count, nil
}

func (r *Repository) GetDetail(ctx context.Context, businessID, orderID int64) (domain.StoreOrderDetail, error) {
	order, err := r.queries.GetStoreOrder(ctx, sqlc.GetStoreOrderParams{BusinessID: businessID, ID: orderID})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get store order", logger.Error(err))
		return domain.StoreOrderDetail{}, err
	}
	itemRows, err := r.queries.ListStoreOrderItems(ctx, sqlc.ListStoreOrderItemsParams{BusinessID: businessID, OrderID: orderID})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list store order items", logger.Error(err))
		return domain.StoreOrderDetail{}, err
	}
	items := make([]domain.StoreOrderItem, 0, len(itemRows))
	for _, row := range itemRows {
		items = append(items, itemFromRow(ctx, row))
	}
	return domain.StoreOrderDetail{Order: orderFromRow(ctx, order), Items: items}, nil
}

func (r *Repository) Create(ctx context.Context, ownerID, businessID int64, input domain.StoreOrderInput) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	input, err = r.priceOrderInput(ctx, qtx, ownerID, input)
	if err != nil {
		return 0, err
	}
	result, err := qtx.CreateStoreOrder(ctx, sqlc.CreateStoreOrderParams{
		OwnerID:         ownerID,
		BusinessID:      businessID,
		CustomerEmail:   sql.NullString{String: input.CustomerEmail, Valid: input.CustomerEmail != ""},
		CustomerName:    sql.NullString{String: input.CustomerName, Valid: input.CustomerName != ""},
		CustomerPhone:   sql.NullString{String: input.CustomerPhone, Valid: input.CustomerPhone != ""},
		Status:          sqlc.NullStoreOrdersStatus{StoreOrdersStatus: sqlc.StoreOrdersStatus("pending"), Valid: true},
		Subtotal:        strconv.FormatFloat(input.Subtotal, 'f', 2, 64),
		Tax:             strconv.FormatFloat(input.Tax, 'f', 2, 64),
		ShippingCost:    "0.00",
		Total:           strconv.FormatFloat(input.Total, 'f', 2, 64),
		ShippingAddress: nullJSON(input.ShippingAddress),
		Notes:           sql.NullString{String: input.Notes, Valid: input.Notes != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create store order", logger.Error(err))
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read store order insert id", logger.Error(err))
		return 0, err
	}
	for _, item := range input.Items {
		_, err := qtx.CreateStoreOrderItem(ctx, sqlc.CreateStoreOrderItemParams{
			OrderID:   id,
			ProductID: sql.NullInt64{Int64: item.ProductID, Valid: item.ProductID != 0},
			Name:      item.Name,
			Quantity:  int32(item.Quantity), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			Price:     strconv.FormatFloat(item.Price, 'f', 2, 64),
			Total:     strconv.FormatFloat(item.Total, 'f', 2, 64),
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to create store order item", logger.Error(err))
			return 0, err
		}
	}
	if businessID > 0 && (strings.TrimSpace(input.CustomerEmail) != "" || strings.TrimSpace(input.CustomerPhone) != "") {
		contact := crmcontract.CustomerContactCaptured{
			Name:  input.CustomerName,
			Email: input.CustomerEmail,
			Phone: input.CustomerPhone,
		}
		if err = outbox.Publish(ctx, tx, eventbus.Event{
			BusinessID:  businessID,
			Type:        crmcontract.CustomerContactCapturedType,
			SubjectType: "store_order",
			SubjectID:   id,
			Payload:     contact,
			DedupeKey:   crmcontract.CustomerContactCaptureDedupeKey("store_order", id, contact),
		}); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) applyStoreOrderItemStockTx(ctx context.Context, tx *sql.Tx, qtx *sqlc.Queries, ownerID, locationID, orderID, productID int64, quantity float64, entryType string) error {
	tracked, err := r.productTracksInventory(ctx, qtx, ownerID, productID)
	if err != nil {
		return err
	}
	if !tracked {
		return nil
	}

	movement := stock.Movement{
		OwnerID:       ownerID,
		LocationID:    locationID,
		ProductID:     productID,
		EntryType:     stock.EntrySale,
		Change:        -quantity,
		ReferenceID:   int64Ptr(orderID),
		ReferenceType: "store_order",
		// stock_ledger.channel is an ENUM('pos','erp'); a store order is an
		// ERP-side sale, and any other value aborts the confirmation on the
		// ledger insert.
		Channel: "erp",
		Note:    "store order confirmation",
	}
	if entryType == "return" {
		movement.EntryType = stock.EntryReturn
		movement.Change = quantity
		movement.Note = "store order cancellation/refund"
	}
	_, _, err = stock.ApplyMovementTx(ctx, tx, movement)
	if err != nil {
		return fmt.Errorf("apply store order stock movement: %w", err)
	}
	return nil
}

func (r *Repository) priceOrderInput(ctx context.Context, qtx *sqlc.Queries, ownerID int64, input domain.StoreOrderInput) (domain.StoreOrderInput, error) {
	if len(input.Items) == 0 {
		// A zero-value order is retained for existing administrative workflows;
		// the application layer rejects non-zero client totals without catalog
		// items before this repository method is reached.
		input.Status = "pending"
		return input, nil
	}
	var subtotalCents, taxCents int64
	pricedItems := make([]domain.OrderItemInput, 0, len(input.Items))
	for _, requested := range input.Items {
		if requested.ProductID <= 0 {
			// Keep direct repository support for legacy manual line items. Public
			// service calls reject these because their price cannot be verified.
			pricedItems = append(pricedItems, requested)
			continue
		}
		if requested.Quantity <= 0 {
			return domain.StoreOrderInput{}, domain.ErrInvalidAmount
		}
		product, err := qtx.GetStoreProductPricing(ctx, sqlc.GetStoreProductPricingParams{OwnerID: ownerID, ID: requested.ProductID})
		if errors.Is(err, sql.ErrNoRows) {
			return domain.StoreOrderInput{}, fmt.Errorf("store order product %d is unavailable", requested.ProductID)
		}
		if err != nil {
			return domain.StoreOrderInput{}, err
		}
		unitCents, err := decimalCents(product.SalePrice)
		if err != nil || unitCents < 0 {
			return domain.StoreOrderInput{}, fmt.Errorf("invalid catalog price for product %d", requested.ProductID)
		}
		taxRate, err := strconv.ParseFloat(product.TaxRate, 64)
		if err != nil || taxRate < 0 {
			return domain.StoreOrderInput{}, fmt.Errorf("invalid catalog tax for product %d", requested.ProductID)
		}
		lineCents := unitCents * int64(requested.Quantity)
		subtotalCents += lineCents
		taxCents += int64(math.Round(float64(lineCents) * taxRate / 100))
		pricedItems = append(pricedItems, domain.OrderItemInput{
			ProductID: requested.ProductID,
			Name:      product.Name,
			Quantity:  requested.Quantity,
			Price:     domain.CentsToFloat(unitCents),
			Total:     domain.CentsToFloat(lineCents),
		})
	}
	input.Items = pricedItems
	input.Subtotal = domain.CentsToFloat(subtotalCents)
	input.Tax = domain.CentsToFloat(taxCents)
	input.ShippingCost = 0
	input.Total = domain.CentsToFloat(subtotalCents + taxCents)
	input.Status = "pending"
	return input, nil
}

func decimalCents(value string) (int64, error) {
	amount, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0, errors.New("invalid decimal amount")
	}
	return int64(math.Round(amount * 100)), nil
}

func (r *Repository) productTracksInventory(ctx context.Context, qtx *sqlc.Queries, ownerID, productID int64) (bool, error) {
	pricing, err := qtx.GetStoreProductPricing(ctx, sqlc.GetStoreProductPricingParams{
		OwnerID: ownerID,
		ID:      productID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("store order product %d not found for owner %d", productID, ownerID)
	}
	if err != nil {
		return false, err
	}
	return !pricing.TrackInventory.Valid || pricing.TrackInventory.Bool, nil
}

func int64Ptr(v int64) *int64 {
	return &v
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.StoreOrderInput) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	current, err := qtx.GetStoreOrderStatus(ctx, sqlc.GetStoreOrderStatusParams{BusinessID: businessID, ID: id})
	if err != nil {
		return err
	}
	currentStatus := string(current.Status.StoreOrdersStatus)
	nextStatus := strings.ToLower(strings.TrimSpace(input.Status))
	if nextStatus == "" {
		nextStatus = currentStatus
	}
	if err := validateOrderStatusTransition(currentStatus, nextStatus); err != nil {
		return err
	}

	entryType, moveStock := stockMovementForTransition(currentStatus, nextStatus)
	if moveStock {
		locationID, locationErr := stock.DefaultLocationID(ctx, tx, current.OwnerID)
		if locationErr != nil {
			return locationErr
		}
		items, listErr := qtx.ListStoreOrderItemsForUpdate(ctx, sqlc.ListStoreOrderItemsForUpdateParams{BusinessID: businessID, OrderID: id})
		if listErr != nil {
			return listErr
		}
		for _, item := range items {
			if !item.ProductID.Valid || item.Quantity <= 0 {
				continue
			}
			if moveErr := r.applyStoreOrderItemStockTx(ctx, tx, qtx, current.OwnerID, locationID, id, item.ProductID.Int64, float64(item.Quantity), entryType); moveErr != nil {
				return moveErr
			}
		}
	}

	err = qtx.UpdateStoreOrder(ctx, sqlc.UpdateStoreOrderParams{
		CustomerEmail:   sql.NullString{String: input.CustomerEmail, Valid: input.CustomerEmail != ""},
		CustomerName:    sql.NullString{String: input.CustomerName, Valid: input.CustomerName != ""},
		CustomerPhone:   sql.NullString{String: input.CustomerPhone, Valid: input.CustomerPhone != ""},
		Status:          sqlc.NullStoreOrdersStatus{StoreOrdersStatus: sqlc.StoreOrdersStatus(nextStatus), Valid: true},
		ShippingAddress: nullJSON(input.ShippingAddress),
		Notes:           sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		BusinessID:      businessID,
		ID:              id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update store order", logger.Error(err))
		return err
	}
	updatedAt := time.Now().UTC().Format(time.RFC3339Nano)
	if businessID > 0 && (strings.TrimSpace(input.CustomerEmail) != "" || strings.TrimSpace(input.CustomerPhone) != "") {
		contact := crmcontract.CustomerContactCaptured{
			Name:  input.CustomerName,
			Email: input.CustomerEmail,
			Phone: input.CustomerPhone,
		}
		if err = outbox.Publish(ctx, tx, eventbus.Event{
			BusinessID:  businessID,
			Type:        crmcontract.CustomerContactCapturedType,
			SubjectType: "store_order_contact",
			SubjectID:   id,
			Payload:     contact,
			DedupeKey:   "store_order_contact:" + strconv.FormatInt(id, 10) + ":" + updatedAt,
		}); err != nil {
			return err
		}
	}
	if nextStatus == "confirmed" && currentStatus != "confirmed" && businessID > 0 &&
		(strings.TrimSpace(input.CustomerEmail) != "" || strings.TrimSpace(input.CustomerPhone) != "") {
		total, parseErr := strconv.ParseFloat(current.Total, 64)
		if parseErr != nil {
			return fmt.Errorf("parse store order total %q: %w", current.Total, parseErr)
		}
		purchase := crmcontract.CustomerContactCaptured{
			Name:       input.CustomerName,
			Email:      input.CustomerEmail,
			Phone:      input.CustomerPhone,
			TotalSpent: total,
			CountVisit: true,
		}
		if err = outbox.Publish(ctx, tx, eventbus.Event{
			BusinessID:  businessID,
			Type:        crmcontract.CustomerContactCapturedType,
			SubjectType: "store_order_purchase",
			SubjectID:   id,
			Payload:     purchase,
			DedupeKey:   "store_order_purchase:" + strconv.FormatInt(id, 10) + ":" + updatedAt,
		}); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func validateOrderStatusTransition(current, next string) error {
	valid := map[string]bool{
		"pending": true, "confirmed": true, "processing": true, "shipped": true,
		"delivered": true, "cancelled": true, "refunded": true, "completed": true,
	}
	if !valid[next] {
		return domain.ErrInvalidStatus
	}
	if current == next || next == "cancelled" || next == "refunded" {
		return nil
	}
	if current == "pending" && next != "confirmed" {
		return domain.ErrInvalidStatus
	}
	if (current == "cancelled" || current == "refunded") && next != "confirmed" {
		return domain.ErrInvalidStatus
	}
	// Confirming an order deducts stock and consumes cost layers; the reverse
	// transition applies no reversal, so an order that went back to pending
	// would deduct the same quantity again on the next confirmation. A
	// deducted order can only be reopened through cancellation, which returns
	// the stock (cancelled/refunded -> confirmed re-applies a single sale).
	if next == "pending" {
		return domain.ErrInvalidStatus
	}
	return nil
}

func stockMovementForTransition(current, next string) (string, bool) {
	if next == "confirmed" && current != "confirmed" && current != "processing" && current != "shipped" && current != "delivered" && current != "completed" {
		return "sale", true
	}
	if (next == "cancelled" || next == "refunded") && current != "pending" && current != "cancelled" && current != "refunded" {
		return "return", true
	}
	return "", false
}

func parseMonetary(ctx context.Context, s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Ctx(ctx).Warn("Failed to parse monetary value", logger.String("value", s), logger.Error(err))
		return 0
	}
	return v
}

func orderFromRow(ctx context.Context, row sqlc.StoreOrder) domain.StoreOrder {
	o := domain.StoreOrder{ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID}
	if row.CustomerEmail.Valid {
		o.CustomerEmail = row.CustomerEmail.String
	}
	if row.CustomerName.Valid {
		o.CustomerName = row.CustomerName.String
	}
	if row.CustomerPhone.Valid {
		o.CustomerPhone = row.CustomerPhone.String
	}
	if row.Status.Valid {
		o.Status = string(row.Status.StoreOrdersStatus)
	}
	o.Subtotal = parseMonetary(ctx, row.Subtotal)
	o.Tax = parseMonetary(ctx, row.Tax)
	o.ShippingCost = parseMonetary(ctx, row.ShippingCost)
	o.Total = parseMonetary(ctx, row.Total)
	if row.ShippingAddress != nil {
		o.ShippingAddress = string(row.ShippingAddress)
	}
	if row.Notes.Valid {
		o.Notes = row.Notes.String
	}
	if row.CreatedAt.Valid {
		o.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		o.UpdatedAt = row.UpdatedAt.Time
	}
	return o
}

func itemFromRow(ctx context.Context, row sqlc.StoreOrderItem) domain.StoreOrderItem {
	i := domain.StoreOrderItem{ID: row.ID, OrderID: row.OrderID, Name: row.Name, Quantity: int(row.Quantity)}
	if row.ProductID.Valid {
		i.ProductID = row.ProductID.Int64
	}
	i.Price = parseMonetary(ctx, row.Price)
	i.Total = parseMonetary(ctx, row.Total)
	if row.CreatedAt.Valid {
		i.CreatedAt = row.CreatedAt.Time
	}
	return i
}

func nullJSON(s string) []byte {
	if s == "" {
		return []byte("{}")
	}
	return []byte(s)
}
