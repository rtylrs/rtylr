package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"

	"github.com/voxire/onion/internal/modules/erp/internal/stock"
	"github.com/voxire/onion/internal/modules/erp/inventory/internal/pos_stock/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) ApplyRecordedOrderInventory(ctx context.Context, input domain.RecordedOrderInventory) (err error) {
	if err = r.validateBranch(ctx, input.BusinessID, input.OwnerID, input.LocationID); err != nil {
		return err
	}
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollbackOnError(tx, &err)

	var locationID int64
	switch input.LocationKind {
	case "pos_branch":
		locationID, err = stock.POSBranchLocationID(ctx, tx, input.OwnerID, input.BusinessID, input.LocationID)
	case "default":
		locationID, err = stock.DefaultLocationID(ctx, tx, input.OwnerID)
	default:
		return fmt.Errorf("unsupported inventory location kind %q", input.LocationKind)
	}
	if err != nil {
		return fmt.Errorf("resolve inventory location: %w", err)
	}

	quantities := make(map[int64]float64, len(input.Items))
	for _, item := range input.Items {
		quantities[item.ProductID] += item.Quantity
	}
	items := make([]domain.OrderInventoryItem, 0, len(quantities))
	for productID, quantity := range quantities {
		items = append(items, domain.OrderInventoryItem{ProductID: productID, Quantity: quantity})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ProductID < items[j].ProductID })
	for _, item := range items {
		var tracksInventory bool
		if err = tx.QueryRowContext(ctx, `
			SELECT track_inventory
			FROM erp_products
			WHERE owner_id = ? AND business_id = ? AND id = ?
			FOR UPDATE`, input.OwnerID, input.BusinessID, item.ProductID).Scan(&tracksInventory); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				if input.SourceType == "pos_order" {
					continue
				}
				return fmt.Errorf("inventory product %d is unavailable for business %d: %w", item.ProductID, input.BusinessID, err)
			}
			return fmt.Errorf("read inventory tracking for product %d: %w", item.ProductID, err)
		}
		if !tracksInventory {
			continue
		}

		entryType := stock.EntrySale
		change := -item.Quantity
		note := "order confirmation"
		if input.Return {
			entryType = stock.EntryReturn
			change = item.Quantity
			note = "order cancellation or refund"
		}
		referenceID := input.SourceID
		_, _, err = stock.ApplyMovementTx(ctx, tx, stock.Movement{
			OwnerID:       input.OwnerID,
			BusinessID:    input.BusinessID,
			LocationID:    locationID,
			ProductID:     item.ProductID,
			EntryType:     entryType,
			Change:        change,
			ReferenceID:   &referenceID,
			ReferenceType: input.SourceType,
			Channel:       input.Channel,
			Note:          note,
			AllowNegative: input.Channel == "pos" && !input.Return,
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to apply inventory movement from order fact",
				logger.Any("business_id", input.BusinessID),
				logger.Any("source_type", input.SourceType),
				logger.Any("source_id", input.SourceID),
				logger.Any("product_id", item.ProductID),
				logger.Error(err))
			if errors.Is(err, stock.ErrInsufficientStock) {
				return domain.ErrInsufficientStock
			}
			return fmt.Errorf("apply inventory movement for product %d: %w", item.ProductID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit recorded order inventory: %w", err)
	}
	return nil
}
