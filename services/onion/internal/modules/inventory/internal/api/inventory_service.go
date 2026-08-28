package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/inventory/internal/application"
	"github.com/voxire/onion/internal/modules/inventory/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	erpinventoryv1 "github.com/voxire/voxie/gen/go/erp/inventory/v1"
)

type InventoryAPI struct {
	Service *application.Service
}

func NewInventoryAPI(service *application.Service) *InventoryAPI {
	return &InventoryAPI{Service: service}
}

func (a *InventoryAPI) GetInventoryReport(ctx context.Context, _ *erpinventoryv1.GetInventoryReportRequest) (*erpinventoryv1.GetInventoryReportResponse, error) {
	ownerID, err := inventoryOwnerID(ctx, "erp:inventory:view")
	if err != nil {
		return nil, err
	}
	items, listErr := a.Service.GetInventoryReport(ctx, ownerID)
	if listErr != nil {
		return nil, inventoryStatusError(http.StatusInternalServerError, "Failed to get inventory report")
	}
	out := make([]*erpinventoryv1.InventoryReportItem, 0, len(items))
	var totalValue float64
	for _, item := range items {
		out = append(out, newInventoryReportItemPayload(item))
		totalValue += item.StockValue
	}
	return &erpinventoryv1.GetInventoryReportResponse{Items: out, TotalValue: totalValue}, nil
}

func (a *InventoryAPI) GetLowStock(ctx context.Context, _ *erpinventoryv1.GetLowStockRequest) (*erpinventoryv1.GetLowStockResponse, error) {
	ownerID, err := inventoryOwnerID(ctx, "erp:inventory:view")
	if err != nil {
		return nil, err
	}
	items, listErr := a.Service.GetLowStockProducts(ctx, ownerID)
	if listErr != nil {
		return nil, inventoryStatusError(http.StatusInternalServerError, "Failed to get low stock products")
	}
	out := make([]*erpinventoryv1.InventoryReportItem, 0, len(items))
	for _, item := range items {
		out = append(out, newInventoryReportItemPayload(item))
	}
	return &erpinventoryv1.GetLowStockResponse{Items: out, Count: int64(len(out))}, nil
}

func (a *InventoryAPI) GetValuation(ctx context.Context, _ *erpinventoryv1.GetValuationRequest) (*erpinventoryv1.GetValuationResponse, error) {
	ownerID, err := inventoryOwnerID(ctx, "erp:inventory:view")
	if err != nil {
		return nil, err
	}
	items, listErr := a.Service.GetValuation(ctx, ownerID)
	if listErr != nil {
		return nil, inventoryStatusError(http.StatusInternalServerError, "Failed to get inventory valuation")
	}
	out := make([]*erpinventoryv1.ValuationItem, 0, len(items))
	var totalValue float64
	for _, item := range items {
		out = append(out, newValuationItemPayload(item))
		totalValue += item.TotalValue
	}
	return &erpinventoryv1.GetValuationResponse{Items: out, TotalValue: totalValue}, nil
}

func (a *InventoryAPI) AdjustInventory(ctx context.Context, req *erpinventoryv1.AdjustInventoryRequest) (*erpinventoryv1.AdjustInventoryResponse, error) {
	ownerID, err := inventoryOwnerID(ctx, "erp:inventory:adjust")
	if err != nil {
		return nil, err
	}
	if req.GetChange() == 0 {
		return nil, inventoryStatusError(http.StatusBadRequest, "Change quantity cannot be zero")
	}
	validTypes := map[string]struct{}{
		string(domain.InventoryTypePurchase):   {},
		string(domain.InventoryTypeSale):       {},
		string(domain.InventoryTypeAdjustment): {},
		string(domain.InventoryTypeReturn):     {},
		string(domain.InventoryTypeWaste):      {},
	}
	if _, ok := validTypes[req.GetType()]; !ok {
		return nil, inventoryStatusError(http.StatusBadRequest, "Invalid type. Must be one of: purchase, sale, adjustment, return, waste")
	}
	adjustErr := a.Service.AdjustInventory(ctx, ownerID, domain.InventoryAdjustInput{
		ProductID: req.GetProductId(),
		Change:    req.GetChange(),
		Type:      domain.InventoryTransactionType(req.GetType()),
		Notes:     req.GetNotes(),
	})
	if adjustErr != nil {
		return nil, inventoryStatusError(http.StatusInternalServerError, "Failed to adjust inventory")
	}
	return &erpinventoryv1.AdjustInventoryResponse{Message: "Inventory adjusted successfully"}, nil
}

func inventoryOwnerID(ctx context.Context, permission string) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || !authPkg.HasAppPermission(authUser, authorization.AppERP, permission) {
		return 0, onekiterrors.PermissionDenied()
	}
	ownerID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return ownerID, nil
}

func inventoryStatusError(status int, message string) error {
	return onekiterrors.New(status, message)
}
