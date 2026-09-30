package warehouses

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/warehouses/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	createWarehouseToolName = "erp.create_warehouse"
	getWarehouseToolName    = "erp.get_warehouse"
	listWarehousesToolName  = "erp.list_warehouses"
)

type getWarehouseToolInput struct {
	WarehouseID int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	required, dispatched := warehouseToolPermissions[name]
	if !dispatched {
		return nil, false, nil
	}
	if err := requireWarehouseToolPermission(ctx, required); err != nil {
		return nil, true, err
	}
	switch name {
	case getWarehouseToolName:
		output, err := m.executeGetWarehouseTool(ctx, input)
		return output, true, err
	case listWarehousesToolName:
		output, err := m.executeListWarehousesTool(ctx)
		return output, true, err
	case createWarehouseToolName:
		output, err := m.executeCreateWarehouseTool(ctx, input)
		return output, true, err
	default:
		return nil, true, fmt.Errorf("warehouse tool %q has no executor", name)
	}
}

func (m *Module) executeGetWarehouseTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseGetWarehouseToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}

	warehouse, err := m.service.GetWarehouseByID(ctx, businessID, parsed.WarehouseID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"warehouse": warehouseToolPayload(warehouse),
	}, nil
}

func (m *Module) executeListWarehousesTool(ctx context.Context) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}

	warehouses, err := m.service.ListWarehouses(ctx, businessID)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(warehouses))
	for _, warehouse := range warehouses {
		payloads = append(payloads, warehouseToolPayload(warehouse))
	}

	return map[string]any{
		"warehouses": payloads,
		"returned":   len(payloads),
	}, nil
}

func (m *Module) executeCreateWarehouseTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseCreateWarehouseToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}

	id, err := m.service.CreateWarehouse(ctx, businessID, parsed)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"warehouse": map[string]any{
			"id":           id,
			"name":         parsed.Name,
			"address":      parsed.Address,
			"city":         parsed.City,
			"country":      parsed.Country,
			"manager_name": parsed.ManagerName,
			"capacity":     parsed.Capacity,
			"is_default":   parsed.IsDefault,
			"is_active":    parsed.IsActive,
		},
	}, nil
}

func parseGetWarehouseToolInput(input map[string]any) (getWarehouseToolInput, error) {
	warehouseID, err := warehouseIDToolValue(firstWarehouseToolValue(input, "warehouse_id", "warehouse", "id"))
	if err != nil {
		return getWarehouseToolInput{}, fmt.Errorf("invalid warehouse_id: %w", err)
	}
	if warehouseID <= 0 {
		return getWarehouseToolInput{}, errors.New("warehouse_id is required")
	}
	return getWarehouseToolInput{WarehouseID: warehouseID}, nil
}

func parseCreateWarehouseToolInput(input map[string]any) (domain.WarehouseInput, error) {
	capacity, err := warehouseToolNonNegativeInt64(firstWarehouseToolValue(input, "capacity"))
	if err != nil {
		return domain.WarehouseInput{}, fmt.Errorf("invalid capacity: %w", err)
	}

	isDefault, err := warehouseToolBool(firstWarehouseToolValue(input, "is_default", "default"), false)
	if err != nil {
		return domain.WarehouseInput{}, fmt.Errorf("invalid is_default: %w", err)
	}

	isActive, err := warehouseToolBool(firstWarehouseToolValue(input, "is_active", "active"), true)
	if err != nil {
		return domain.WarehouseInput{}, fmt.Errorf("invalid is_active: %w", err)
	}

	parsed := domain.WarehouseInput{
		Name:        strings.TrimSpace(warehouseToolString(firstWarehouseToolValue(input, "name", "warehouse_name"))),
		Address:     strings.TrimSpace(warehouseToolString(firstWarehouseToolValue(input, "address"))),
		City:        strings.TrimSpace(warehouseToolString(firstWarehouseToolValue(input, "city"))),
		Country:     strings.TrimSpace(warehouseToolString(firstWarehouseToolValue(input, "country"))),
		ManagerName: strings.TrimSpace(warehouseToolString(firstWarehouseToolValue(input, "manager_name", "manager"))),
		Capacity:    capacity,
		IsDefault:   isDefault,
		IsActive:    isActive,
	}
	if err := parsed.Validate(); err != nil {
		return domain.WarehouseInput{}, err
	}
	return parsed, nil
}

func firstWarehouseToolValue(input map[string]any, keys ...string) any {
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

func warehouseIDToolValue(value any) (int64, error) {
	switch v := value.(type) {
	case string:
		raw := strings.TrimSpace(v)
		raw = strings.TrimPrefix(raw, "#")
		raw = strings.ToLower(raw)
		for _, prefix := range []string{"warehouse-", "warehouse_", "wh-", "wh_"} {
			raw = strings.TrimPrefix(raw, prefix)
		}
		return warehouseToolInt64(raw)
	default:
		return warehouseToolInt64(value)
	}
}

func warehouseToolPayload(warehouse domain.Warehouse) map[string]any {
	return map[string]any{
		"id":           warehouse.ID,
		"name":         warehouse.Name,
		"address":      warehouse.Address,
		"city":         warehouse.City,
		"country":      warehouse.Country,
		"manager_name": warehouse.ManagerName,
		"capacity":     warehouse.Capacity,
		"is_default":   warehouse.IsDefault,
		"is_active":    warehouse.IsActive,
		"created_at":   warehouseToolTime(warehouse.CreatedAt),
		"updated_at":   warehouseToolTime(warehouse.UpdatedAt),
	}
}

func warehouseToolNonNegativeInt64(value any) (int64, error) {
	out, err := warehouseToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func warehouseToolInt64(value any) (int64, error) {
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

func warehouseToolBool(value any, defaultValue bool) (bool, error) {
	switch v := value.(type) {
	case nil:
		return defaultValue, nil
	case bool:
		return v, nil
	case json.Number:
		n, err := strconv.ParseInt(v.String(), 10, 64)
		if err != nil {
			return false, err
		}
		switch n {
		case 0:
			return false, nil
		case 1:
			return true, nil
		default:
			return false, errors.New("must be true or false")
		}
	case string:
		raw := strings.ToLower(strings.TrimSpace(v))
		if raw == "" {
			return defaultValue, nil
		}
		switch raw {
		case "true", "t", "yes", "y", "1", "active", "default":
			return true, nil
		case "false", "f", "no", "n", "0", "inactive":
			return false, nil
		default:
			return false, errors.New("must be true or false")
		}
	default:
		return false, fmt.Errorf("unsupported value %T", value)
	}
}

func warehouseToolString(value any) string {
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

func warehouseToolTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
