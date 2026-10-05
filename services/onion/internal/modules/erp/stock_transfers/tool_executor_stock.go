package stocktransfers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/erp/stock_transfers/internal/domain"
)

func parseListStockTransfersToolInput(input map[string]any) (listStockTransfersToolInput, error) {
	status := normalizeStockTransferStatus(firstStockTransferToolValue(input, "status"))
	if strings.TrimSpace(stockTransferToolString(firstStockTransferToolValue(input, "status"))) != "" && status == "" {
		return listStockTransfersToolInput{}, errors.New("status must be one of draft, in_transit, received, or cancelled")
	}

	page, err := stockTransferToolInt64(firstStockTransferToolValue(input, "page"))
	if err != nil {
		return listStockTransfersToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page < 0 {
		return listStockTransfersToolInput{}, errors.New("page must be non-negative")
	}
	if page == 0 {
		page = 1
	}

	pageSize, err := stockTransferToolInt64(firstStockTransferToolValue(input, "page_size", "limit"))
	if err != nil {
		return listStockTransfersToolInput{}, fmt.Errorf("invalid page_size: %w", err)
	}
	if pageSize < 0 {
		return listStockTransfersToolInput{}, errors.New("page_size must be non-negative")
	}
	if pageSize == 0 {
		pageSize = defaultListStockTransfersToolPageSize
	}
	if pageSize > maxListStockTransfersToolPageSize {
		pageSize = maxListStockTransfersToolPageSize
	}

	return listStockTransfersToolInput{
		Status:   status,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func normalizeStockTransferStatus(value any) domain.StockTransferStatus {
	status := strings.ToLower(strings.TrimSpace(stockTransferToolString(value)))
	if status == "" {
		return ""
	}
	status = strings.ReplaceAll(status, "-", "_")
	status = strings.ReplaceAll(status, " ", "_")
	switch domain.StockTransferStatus(status) {
	case domain.StockTransferStatusDraft,
		domain.StockTransferStatusInTransit,
		domain.StockTransferStatusReceived,
		domain.StockTransferStatusCancelled:
		return domain.StockTransferStatus(status)
	default:
		return ""
	}
}

func validateStockTransferToolItems(items []domain.StockTransferItemInput) error {
	for idx, item := range items {
		if item.ProductID <= 0 {
			return fmt.Errorf("item %d product_id must be positive", idx+1)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("item %d quantity must be positive", idx+1)
		}
	}
	return nil
}

func parseStockTransferIDToolInput(input map[string]any) (int64, error) {
	transferID, err := stockTransferIDToolValue(firstStockTransferToolValue(input, "stock_transfer_id", "transfer_id", "stock_transfer", "transfer", "id"))
	if err != nil {
		return 0, fmt.Errorf("invalid stock_transfer_id: %w", err)
	}
	if transferID <= 0 {
		return 0, errors.New("stock_transfer_id is required")
	}
	return transferID, nil
}

func firstStockTransferToolValue(input map[string]any, keys ...string) any {
	for _, key := range keys {
		value, ok := input[key]
		if !ok || value == nil {
			continue
		}
		if s, ok := value.(string); ok && strings.TrimSpace(s) == "" {
			continue
		}
		return value
	}
	return nil
}

func stockTransferToolItems(value any) ([]domain.StockTransferItemInput, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case []domain.StockTransferItemInput:
		return v, nil
	case []map[string]any:
		items := make([]domain.StockTransferItemInput, 0, len(v))
		for idx, item := range v {
			parsed, err := stockTransferToolItem(item)
			if err != nil {
				return nil, fmt.Errorf("item %d: %w", idx+1, err)
			}
			items = append(items, parsed)
		}
		return items, nil
	case []any:
		items := make([]domain.StockTransferItemInput, 0, len(v))
		for idx, item := range v {
			itemMap, ok := item.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("item %d must be an object", idx+1)
			}
			parsed, err := stockTransferToolItem(itemMap)
			if err != nil {
				return nil, fmt.Errorf("item %d: %w", idx+1, err)
			}
			items = append(items, parsed)
		}
		return items, nil
	default:
		return nil, fmt.Errorf("must be an array, got %T", value)
	}
}

func stockTransferToolItem(input map[string]any) (domain.StockTransferItemInput, error) {
	productID, err := stockTransferToolPositiveInt64(firstStockTransferToolValue(input, "product_id"))
	if err != nil {
		return domain.StockTransferItemInput{}, fmt.Errorf("invalid product_id: %w", err)
	}

	quantity, err := stockTransferToolPositiveFloat64(firstStockTransferToolValue(input, "quantity", "qty", "units"))
	if err != nil {
		return domain.StockTransferItemInput{}, fmt.Errorf("invalid quantity: %w", err)
	}

	return domain.StockTransferItemInput{
		ProductID: productID,
		Quantity:  quantity,
	}, nil
}

func stockTransferIDToolValue(value any) (int64, error) {
	switch v := value.(type) {
	case string:
		raw := strings.TrimSpace(v)
		raw = strings.TrimPrefix(raw, "#")
		raw = strings.ToLower(raw)
		for _, prefix := range []string{"stock-transfer-", "stock_transfer_", "stocktransfer-", "stocktransfer_", "transfer-", "transfer_", "st-", "st_"} {
			raw = strings.TrimPrefix(raw, prefix)
		}
		return stockTransferToolInt64(raw)
	default:
		return stockTransferToolInt64(value)
	}
}

func stockTransferToolPositiveInt64(value any) (int64, error) {
	out, err := stockTransferToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out <= 0 {
		return 0, errors.New("must be positive")
	}
	return out, nil
}

func stockTransferToolInt64(value any) (int64, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, errors.New("must be finite")
		}
		if math.Trunc(v) != v {
			return 0, errors.New("must be an integer")
		}
		return int64(v), nil
	case json.Number:
		return v.Int64()
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return 0, nil
		}
		return strconv.ParseInt(raw, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported value %T", value)
	}
}

func stockTransferToolPositiveFloat64(value any) (float64, error) {
	out, err := stockTransferToolFloat64(value)
	if err != nil {
		return 0, err
	}
	if out <= 0 {
		return 0, errors.New("must be positive")
	}
	return out, nil
}

func stockTransferToolFloat64(value any) (float64, error) {
	var out float64
	switch v := value.(type) {
	case nil:
		return 0, nil
	case float64:
		out = v
	case float32:
		out = float64(v)
	case int:
		out = float64(v)
	case int64:
		out = float64(v)
	case int32:
		out = float64(v)
	case json.Number:
		parsed, err := v.Float64()
		if err != nil {
			return 0, err
		}
		out = parsed
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return 0, nil
		}
		parsed, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return 0, err
		}
		out = parsed
	default:
		return 0, fmt.Errorf("unsupported value %T", value)
	}
	if math.IsNaN(out) || math.IsInf(out, 0) {
		return 0, errors.New("must be finite")
	}
	return out, nil
}

func stockTransferToolString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		return v.String()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

func stockTransferToolPayload(transfer domain.StockTransfer) map[string]any {
	return map[string]any{
		"id":                transfer.ID,
		"from_warehouse_id": transfer.FromWarehouseID,
		"to_warehouse_id":   transfer.ToWarehouseID,
		"status":            string(transfer.Status),
		"notes":             transfer.Notes,
		"items":             stockTransferDomainItemsOutput(transfer.Items),
		"item_count":        len(transfer.Items),
		"created_at":        stockTransferToolTime(transfer.CreatedAt),
		"updated_at":        stockTransferToolTime(transfer.UpdatedAt),
	}
}

func stockTransferDomainItemsOutput(items []domain.StockTransferItem) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]any{
			"id":          item.ID,
			"transfer_id": item.TransferID,
			"product_id":  item.ProductID,
			"quantity":    item.Quantity,
			"created_at":  stockTransferToolTime(item.CreatedAt),
		})
	}
	return out
}

func stockTransferToolItemsOutput(items []domain.StockTransferItemInput) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]any{
			"product_id": item.ProductID,
			"quantity":   item.Quantity,
		})
	}
	return out
}

func stockTransferToolTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
