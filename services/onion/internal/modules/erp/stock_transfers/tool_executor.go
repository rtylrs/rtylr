package stocktransfers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/erp/stock_transfers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listStockTransfersToolName            = "erp.list_stock_transfers"
	getStockTransferToolName              = "erp.get_stock_transfer"
	createStockTransferToolName           = "erp.create_stock_transfer"
	receiveStockTransferToolName          = "erp.receive_stock_transfer"
	defaultListStockTransfersToolPageSize = int64(20)
	maxListStockTransfersToolPageSize     = int64(50)
)

type listStockTransfersToolInput struct {
	Status   domain.StockTransferStatus
	Page     int64
	PageSize int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	required, dispatched := stockTransferToolPermissions[name]
	if !dispatched {
		return nil, false, nil
	}
	if err := requireStockTransferToolPermission(ctx, required); err != nil {
		return nil, true, err
	}
	switch name {
	case listStockTransfersToolName:
		output, err := m.executeListStockTransfersTool(ctx, input)
		return output, true, err
	case getStockTransferToolName:
		output, err := m.executeGetStockTransferTool(ctx, input)
		return output, true, err
	case createStockTransferToolName:
		output, err := m.executeCreateStockTransferTool(ctx, input)
		return output, true, err
	case receiveStockTransferToolName:
		output, err := m.executeReceiveStockTransferTool(ctx, input)
		return output, true, err
	default:
		return nil, true, fmt.Errorf("stock transfer tool %q has no executor", name)
	}
}

func (m *Module) executeGetStockTransferTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	transferID, err := parseStockTransferIDToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return nil, err
	}

	transfer, err := m.service.GetByID(ctx, businessID, transferID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"stock_transfer": stockTransferToolPayload(transfer),
	}, nil
}

func (m *Module) executeListStockTransfersTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseListStockTransfersToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return nil, err
	}

	transfers, hasMore, err := m.service.ListStockTransfers(ctx, businessID, parsed.Page, parsed.PageSize)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(transfers))
	for _, transfer := range transfers {
		if parsed.Status != "" && transfer.Status != parsed.Status {
			continue
		}
		payloads = append(payloads, stockTransferToolPayload(transfer))
	}

	return map[string]any{
		"stock_transfers": payloads,
		"returned":        len(payloads),
		"page":            parsed.Page,
		"page_size":       parsed.PageSize,
		"has_more":        hasMore,
		"filters": map[string]any{
			"status": string(parsed.Status),
		},
	}, nil
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

	businessID, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return nil, err
	}

	id, err := m.service.CreateStockTransfer(ctx, businessID, parsed)
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

	businessID, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return nil, err
	}

	if err := m.service.UpdateStockTransferStatus(ctx, businessID, transferID, domain.StockTransferStatusReceived); err != nil {
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
