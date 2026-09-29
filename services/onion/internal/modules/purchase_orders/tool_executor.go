package purchase_orders

import (
	"context"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/modules/purchase_orders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listPurchaseOrdersToolName            = "erp.list_purchase_orders"
	getPurchaseOrderToolName              = "erp.get_purchase_order"
	createPurchaseOrderToolName           = "erp.create_purchase_order"
	approvePurchaseOrderToolName          = "erp.approve_purchase_order"
	receivePurchaseOrderToolName          = "erp.receive_purchase_order"
	defaultListPurchaseOrdersToolPageSize = int64(20)
	maxListPurchaseOrdersToolPageSize     = int64(50)
)

type listPurchaseOrdersToolInput struct {
	Status   string
	Page     int64
	PageSize int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	required, dispatched := purchaseOrderToolPermissions[name]
	if !dispatched {
		return nil, false, nil
	}
	if err := requirePurchaseOrderToolPermission(ctx, required); err != nil {
		return nil, true, err
	}
	switch name {
	case listPurchaseOrdersToolName:
		output, err := m.executeListPurchaseOrdersTool(ctx, input)
		return output, true, err
	case getPurchaseOrderToolName:
		output, err := m.executeGetPurchaseOrderTool(ctx, input)
		return output, true, err
	case createPurchaseOrderToolName:
		output, err := m.executeCreatePurchaseOrderTool(ctx, input)
		return output, true, err
	case approvePurchaseOrderToolName:
		output, err := m.executeApprovePurchaseOrderTool(ctx, input)
		return output, true, err
	case receivePurchaseOrderToolName:
		output, err := m.executeReceivePurchaseOrderTool(ctx, input)
		return output, true, err
	default:
		return nil, true, fmt.Errorf("purchase order tool %q has no executor", name)
	}
}

func (m *Module) executeGetPurchaseOrderTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	poID, err := parsePurchaseOrderIDToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}
	purchaseOrder, err := m.service.GetByID(ctx, businessID, poID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"purchase_order": purchaseOrderToolPayload(purchaseOrder),
	}, nil
}

func (m *Module) executeListPurchaseOrdersTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseListPurchaseOrdersToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}
	orders, hasMore, err := m.service.ListPurchaseOrders(ctx, businessID, parsed.Status, parsed.Page, parsed.PageSize)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(orders))
	for _, order := range orders {
		payloads = append(payloads, purchaseOrderToolPayload(order))
	}

	return map[string]any{
		"purchase_orders": payloads,
		"returned":        len(payloads),
		"page":            parsed.Page,
		"page_size":       parsed.PageSize,
		"has_more":        hasMore,
		"filters": map[string]any{
			"status": parsed.Status,
		},
	}, nil
}

func (m *Module) executeCreatePurchaseOrderTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseCreatePurchaseOrderToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}
	id, err := m.service.CreatePurchaseOrder(ctx, businessID, parsed, authUser.UserID)
	if err != nil {
		return nil, err
	}

	subtotal := purchaseOrderToolSubtotal(parsed.Items)
	return map[string]any{
		"purchase_order": map[string]any{
			"id":                id,
			"supplier_id":       parsed.SupplierID,
			"po_number":         parsed.PONumber,
			"status":            string(domain.POStatusDraft),
			"subtotal":          subtotal,
			"tax_amount":        parsed.TaxAmount,
			"total":             subtotal + parsed.TaxAmount,
			"expected_delivery": parsed.ExpectedDelivery,
			"notes":             parsed.Notes,
			"items":             purchaseOrderToolItemsOutput(parsed.Items),
			"item_count":        len(parsed.Items),
		},
	}, nil
}

func (m *Module) executeApprovePurchaseOrderTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	poID, err := parseApprovePurchaseOrderToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}
	if err := m.service.ApprovePurchaseOrder(ctx, businessID, poID, authUser.UserID, authUser.Role); err != nil {
		return nil, err
	}

	return map[string]any{
		"purchase_order": map[string]any{
			"id":       poID,
			"status":   string(domain.POStatusApproved),
			"approved": true,
		},
	}, nil
}

func (m *Module) executeReceivePurchaseOrderTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	poID, err := parsePurchaseOrderIDToolInput(input)
	if err != nil {
		return nil, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, bizErr
	}
	if err := m.service.ReceivePurchaseOrder(ctx, businessID, poID, authUser.UserID, nil); err != nil {
		return nil, err
	}

	return map[string]any{
		"purchase_order": map[string]any{
			"id":       poID,
			"status":   string(domain.POStatusReceived),
			"received": true,
		},
	}, nil
}
