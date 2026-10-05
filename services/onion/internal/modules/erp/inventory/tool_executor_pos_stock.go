package inventory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	posstockdomain "github.com/voxire/onion/internal/modules/erp/inventory/internal/pos_stock/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listStockMovementsToolName       = "pos.list_stock_movements"
	listBranchInventoryToolName      = "pos.list_branch_inventory"
	listLowStockItemsToolName        = "pos.list_low_stock_items"
	defaultPOSStockToolLimit         = int64(20)
	maxPOSStockToolLimit             = int64(50)
	posStockToolTimestampDate        = "2006-01-02T15:04:05Z"
	posStockMovementTypeReceive      = "receive"
	posStockMovementTypeAdjust       = "adjust"
	posStockMovementTypeWaste        = "waste"
	posStockMovementTypeTransferIn   = "transfer_in"
	posStockMovementTypeTransferOut  = "transfer_out"
	posStockMovementTypeSale         = "sale"
	posStockMovementTypeReturn       = "return"
	posStockMovementTypeReservation  = "reservation"
	posStockMovementTypeCancellation = "cancellation"
)

type listStockMovementsToolInput struct {
	BranchID   int64
	ProductID  int64
	Type       string
	EmployeeID int64
	Page       int64
	Limit      int64
}

type listBranchInventoryToolInput struct {
	BranchID int64
	Page     int64
	Limit    int64
}

func (m *Module) executePOSStockTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listStockMovementsToolName:
		output, err := m.executeListStockMovementsTool(ctx, input)
		return output, true, err
	case listBranchInventoryToolName:
		output, err := m.executeListBranchInventoryTool(ctx, input)
		return output, true, err
	case listLowStockItemsToolName:
		output, err := m.executeListLowStockItemsTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListStockMovementsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ownerID, err := posStockToolAuth(ctx)
	if err != nil {
		return nil, err
	}

	parsed, err := parseListStockMovementsToolInput(input)
	if err != nil {
		return nil, err
	}

	movements, err := m.posStockService.ListMovements(ctx, authUser.BusinessID, ownerID, parsed.BranchID, parsed.ProductID)
	if err != nil {
		return nil, err
	}
	movements = filterStockMovementsForTool(movements, parsed)

	pageItems, hasMore := paginateStockMovementsForTool(movements, parsed.Page, parsed.Limit)
	payloads := make([]map[string]any, 0, len(pageItems))
	for _, movement := range pageItems {
		payloads = append(payloads, stockMovementToolPayload(movement))
	}

	return map[string]any{
		"stock_movements": payloads,
		"total":           len(movements),
		"returned":        len(payloads),
		"page":            parsed.Page,
		"page_size":       parsed.Limit,
		"has_more":        hasMore,
		"filters": map[string]any{
			"branch_id":   parsed.BranchID,
			"product_id":  parsed.ProductID,
			"type":        parsed.Type,
			"employee_id": parsed.EmployeeID,
		},
	}, nil
}

func (m *Module) executeListBranchInventoryTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ownerID, err := posStockToolAuth(ctx)
	if err != nil {
		return nil, err
	}

	parsed, err := parseListBranchInventoryToolInput(input)
	if err != nil {
		return nil, err
	}

	items, err := m.posStockService.ListBranchInventory(ctx, authUser.BusinessID, ownerID, parsed.BranchID)
	if err != nil {
		return nil, err
	}
	pageItems, hasMore := paginateBranchInventoryForTool(items, parsed.Page, parsed.Limit)

	payloads := make([]map[string]any, 0, len(pageItems))
	for _, item := range pageItems {
		payloads = append(payloads, branchInventoryToolPayload(item))
	}

	return map[string]any{
		"inventory": payloads,
		"total":     len(items),
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  hasMore,
		"filters": map[string]any{
			"branch_id": parsed.BranchID,
		},
	}, nil
}

func (m *Module) executeListLowStockItemsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ownerID, err := posStockToolAuth(ctx)
	if err != nil {
		return nil, err
	}

	parsed, err := parseListBranchInventoryToolInput(input)
	if err != nil {
		return nil, err
	}

	items, err := m.posStockService.ListLowStockItems(ctx, authUser.BusinessID, ownerID, parsed.BranchID)
	if err != nil {
		return nil, err
	}
	pageItems, hasMore := paginateBranchInventoryForTool(items, parsed.Page, parsed.Limit)

	payloads := make([]map[string]any, 0, len(pageItems))
	for _, item := range pageItems {
		payloads = append(payloads, branchInventoryToolPayload(item))
	}

	return map[string]any{
		"low_stock_items": payloads,
		"total":           len(items),
		"returned":        len(payloads),
		"page":            parsed.Page,
		"page_size":       parsed.Limit,
		"has_more":        hasMore,
		"filters": map[string]any{
			"branch_id": parsed.BranchID,
		},
	}, nil
}

func parseListStockMovementsToolInput(input map[string]any) (listStockMovementsToolInput, error) {
	branchID, err := posStockToolInt64(firstPOSStockToolValue(input, "branch_id", "location_id", "branch"))
	if err != nil {
		return listStockMovementsToolInput{}, fmt.Errorf("invalid branch_id: %w", err)
	}
	if branchID <= 0 {
		return listStockMovementsToolInput{}, errors.New("branch_id is required")
	}

	productID, err := posStockToolInt64(firstPOSStockToolValue(input, "product_id", "item_id", "product"))
	if err != nil {
		return listStockMovementsToolInput{}, fmt.Errorf("invalid product_id: %w", err)
	}
	if productID <= 0 {
		return listStockMovementsToolInput{}, errors.New("product_id is required")
	}

	movementType := normalizeStockMovementToolType(input["type"])
	if strings.TrimSpace(posStockToolString(input["type"])) != "" && movementType == "" {
		return listStockMovementsToolInput{}, errors.New("type must be one of receive, adjust, waste, transfer_in, transfer_out, sale, return, reservation, or cancellation")
	}

	employeeID, err := posStockToolNonNegativeInt64(firstPOSStockToolValue(input, "employee_id", "staff_id", "employee"))
	if err != nil {
		return listStockMovementsToolInput{}, fmt.Errorf("invalid employee_id: %w", err)
	}

	page, limit, err := parsePOSStockToolPage(input)
	if err != nil {
		return listStockMovementsToolInput{}, err
	}

	return listStockMovementsToolInput{
		BranchID:   branchID,
		ProductID:  productID,
		Type:       movementType,
		EmployeeID: employeeID,
		Page:       page,
		Limit:      limit,
	}, nil
}

func parseListBranchInventoryToolInput(input map[string]any) (listBranchInventoryToolInput, error) {
	branchID, err := posStockToolInt64(firstPOSStockToolValue(input, "branch_id", "location_id", "branch"))
	if err != nil {
		return listBranchInventoryToolInput{}, fmt.Errorf("invalid branch_id: %w", err)
	}
	if branchID <= 0 {
		return listBranchInventoryToolInput{}, errors.New("branch_id is required")
	}

	page, limit, err := parsePOSStockToolPage(input)
	if err != nil {
		return listBranchInventoryToolInput{}, err
	}

	return listBranchInventoryToolInput{
		BranchID: branchID,
		Page:     page,
		Limit:    limit,
	}, nil
}

func parsePOSStockToolPage(input map[string]any) (int64, int64, error) {
	page, err := posStockToolNonNegativeInt64(firstPOSStockToolValue(input, "page"))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := posStockToolNonNegativeInt64(firstPOSStockToolValue(input, "limit", "page_size"))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultPOSStockToolLimit
	}
	if limit > maxPOSStockToolLimit {
		limit = maxPOSStockToolLimit
	}
	return page, limit, nil
}

func filterStockMovementsForTool(movements []posstockdomain.StockMovement, input listStockMovementsToolInput) []posstockdomain.StockMovement {
	if input.Type == "" && input.EmployeeID == 0 {
		return movements
	}

	out := make([]posstockdomain.StockMovement, 0, len(movements))
	for _, movement := range movements {
		if input.Type != "" && !strings.EqualFold(movement.Type, input.Type) {
			continue
		}
		if input.EmployeeID > 0 && (movement.EmployeeID == nil || *movement.EmployeeID != input.EmployeeID) {
			continue
		}
		out = append(out, movement)
	}
	return out
}

// posStockToolPageStart converts a 1-based page number into a slice offset,
// saturating at itemCount. Multiplying (page-1)*limit directly can overflow
// int64 for very large page numbers (JSON numbers decode to float64, so
// page=2^62 is accepted by parsePOSStockToolPage), wrapping to a negative
// offset and panicking on the slice below.
func posStockToolPageStart(page, limit int64, itemCount int) int {
	if page <= 1 || limit <= 0 {
		return 0
	}
	if page-1 > int64(itemCount)/limit {
		return itemCount
	}
	return int((page - 1) * limit)
}

func paginateStockMovementsForTool(items []posstockdomain.StockMovement, page, limit int64) ([]posstockdomain.StockMovement, bool) {
	start := posStockToolPageStart(page, limit, len(items))
	end := start + int(limit)
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], end < len(items)
}

func paginateBranchInventoryForTool(items []posstockdomain.BranchInventory, page, limit int64) ([]posstockdomain.BranchInventory, bool) {
	start := posStockToolPageStart(page, limit, len(items))
	end := start + int(limit)
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], end < len(items)
}

func stockMovementToolPayload(movement posstockdomain.StockMovement) map[string]any {
	return map[string]any{
		"id":             movement.ID,
		"business_id":    movement.BusinessID,
		"branch_id":      movement.BranchID,
		"product_id":     movement.ProductID,
		"type":           movement.Type,
		"quantity":       movement.Quantity,
		"reference_id":   posStockToolOptionalID(movement.ReferenceID),
		"reference_type": posStockToolOptionalString(movement.ReferenceType),
		"note":           posStockToolOptionalString(movement.Note),
		"employee_id":    posStockToolOptionalID(movement.EmployeeID),
		"created_at":     formatPOSStockToolTime(movement.CreatedAt),
	}
}

func branchInventoryToolPayload(item posstockdomain.BranchInventory) map[string]any {
	return map[string]any{
		"id":                  item.ID,
		"business_id":         item.BusinessID,
		"branch_id":           item.BranchID,
		"product_id":          item.ProductID,
		"item_id":             item.ProductID,
		"name":                item.Name,
		"sku":                 item.SKU,
		"unit":                item.Unit,
		"quantity":            item.Quantity,
		"current_qty":         item.Quantity,
		"low_stock_threshold": posStockToolOptionalFloat64(item.LowStockThreshold),
		"reorder_point":       posStockToolOptionalFloat64(item.ReorderPoint),
		"reorder_quantity":    posStockToolOptionalFloat64(item.ReorderQuantity),
		"updated_at":          formatPOSStockToolTime(item.UpdatedAt),
	}
}

func normalizeStockMovementToolType(value any) string {
	movementType := strings.ToLower(strings.TrimSpace(posStockToolString(value)))
	movementType = strings.ReplaceAll(movementType, "-", "_")
	movementType = strings.ReplaceAll(movementType, " ", "_")
	switch movementType {
	case "", "all":
		return ""
	case posStockMovementTypeReceive, posStockMovementTypeAdjust, posStockMovementTypeWaste,
		posStockMovementTypeTransferIn, posStockMovementTypeTransferOut, posStockMovementTypeSale,
		posStockMovementTypeReturn, posStockMovementTypeReservation, posStockMovementTypeCancellation:
		return movementType
	default:
		return ""
	}
}

func firstPOSStockToolValue(input map[string]any, keys ...string) any {
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

func posStockToolAuth(ctx context.Context) (*authPkg.AuthUser, int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, 0, errors.New("authentication is required")
	}
	if authUser.BusinessID <= 0 {
		return nil, 0, errors.New("business context is required")
	}
	businessID, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return nil, 0, errors.New("business context is required")
	}
	return authUser, businessID, nil
}

func posStockToolNonNegativeInt64(value any) (int64, error) {
	out, err := posStockToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func posStockToolInt64(value any) (int64, error) {
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

func posStockToolString(value any) string {
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

func posStockToolOptionalID(value *int64) any {
	if value == nil || *value <= 0 {
		return nil
	}
	return *value
}

func posStockToolOptionalString(value *string) any {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	return *value
}

func posStockToolOptionalFloat64(value *float64) any {
	if value == nil {
		return nil
	}
	return *value
}

func formatPOSStockToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(posStockToolTimestampDate)
}
