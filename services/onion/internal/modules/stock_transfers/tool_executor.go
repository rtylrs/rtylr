package stock_transfers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/stock_transfers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	createStockTransferToolName  = "erp.create_stock_transfer"
	receiveStockTransferToolName = "erp.receive_stock_transfer"
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case createStockTransferToolName:
		output, err := m.executeCreateStockTransferTool(ctx, input)
		return output, true, err
	case receiveStockTransferToolName:
		output, err := m.executeReceiveStockTransferTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeCreateStockTransferTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseCreateStockTransferToolInput(input)
	if err != nil {
		return nil, err
	}

	ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
	id, err := m.service.CreateStockTransfer(ctx, ownerID, parsed)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"stock_transfer": map[string]any{
			"id":                id,
			"from_warehouse_id": parsed.FromWarehouseID,
			"to_warehouse_id":   parsed.ToWarehouseID,
			"status":            string(domain.StockTransferStatusDraft),
			"notes":             parsed.Notes,
			"items":             stockTransferToolItemsOutput(parsed.Items),
			"item_count":        len(parsed.Items),
		},
	}, nil
}

func (m *Module) executeReceiveStockTransferTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	transferID, err := parseStockTransferIDToolInput(input)
	if err != nil {
		return nil, err
	}

	ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
	if err := m.service.UpdateStockTransferStatus(ctx, ownerID, transferID, domain.StockTransferStatusReceived); err != nil {
		return nil, err
	}

	return map[string]any{
		"stock_transfer": map[string]any{
			"id":       transferID,
			"status":   string(domain.StockTransferStatusReceived),
			"received": true,
		},
	}, nil
}

func parseCreateStockTransferToolInput(input map[string]any) (domain.StockTransferInput, error) {
	fromWarehouseID, err := stockTransferToolPositiveInt64(firstStockTransferToolValue(input, "from_warehouse_id", "source_warehouse_id", "from"))
	if err != nil {
		return domain.StockTransferInput{}, fmt.Errorf("invalid from_warehouse_id: %w", err)
	}

	toWarehouseID, err := stockTransferToolPositiveInt64(firstStockTransferToolValue(input, "to_warehouse_id", "destination_warehouse_id", "to"))
	if err != nil {
		return domain.StockTransferInput{}, fmt.Errorf("invalid to_warehouse_id: %w", err)
	}
	if fromWarehouseID == toWarehouseID {
		return domain.StockTransferInput{}, errors.New("from_warehouse_id and to_warehouse_id must be different")
	}

	items, err := stockTransferToolItems(firstStockTransferToolValue(input, "items", "line_items"))
	if err != nil {
		return domain.StockTransferInput{}, fmt.Errorf("invalid items: %w", err)
	}
	if len(items) == 0 {
		return domain.StockTransferInput{}, errors.New("items are required")
	}
	if err := validateStockTransferToolItems(items); err != nil {
		return domain.StockTransferInput{}, fmt.Errorf("invalid items: %w", err)
	}

	return domain.StockTransferInput{
		FromWarehouseID: fromWarehouseID,
		ToWarehouseID:   toWarehouseID,
		Notes:           strings.TrimSpace(stockTransferToolString(firstStockTransferToolValue(input, "notes"))),
		Items:           items,
	}, nil
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
	transferID, err := stockTransferToolPositiveInt64(firstStockTransferToolValue(input, "stock_transfer_id", "transfer_id", "id"))
	if err != nil {
		return 0, fmt.Errorf("invalid stock_transfer_id: %w", err)
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
