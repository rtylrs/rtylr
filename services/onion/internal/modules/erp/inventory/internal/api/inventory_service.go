package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/erp/inventory/internal/application"
	"github.com/voxire/onion/internal/modules/erp/inventory/internal/domain"
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

func (a *InventoryAPI) GetInventoryReport(ctx context.Context, req *erpinventoryv1.GetInventoryReportRequest) (*erpinventoryv1.GetInventoryReportResponse, error) {
	ownerID, err := inventoryOwnerID(ctx, "erp:inventory:view")
	if err != nil {
		return nil, err
	}
	page, listErr := a.Service.GetInventoryReportPage(ctx, ownerID, req.GetLimit(), req.GetOffset())
	if listErr != nil {
		return nil, inventoryStatusError(http.StatusInternalServerError, "Failed to get inventory report")
	}
	out := make([]*erpinventoryv1.InventoryReportItem, 0, len(page.Items))
	for _, item := range page.Items {
		out = append(out, newInventoryReportItemPayload(item))
	}
	return &erpinventoryv1.GetInventoryReportResponse{
		Items:      out,
		TotalValue: page.TotalValue,
		HasMore:    page.HasMore,
		Limit:      page.Limit,
		Offset:     page.Offset,
		Total:      page.Total,
	}, nil
}

func (a *InventoryAPI) GetLowStock(ctx context.Context, req *erpinventoryv1.GetLowStockRequest) (*erpinventoryv1.GetLowStockResponse, error) {
	ownerID, err := inventoryOwnerID(ctx, "erp:inventory:view")
	if err != nil {
		return nil, err
	}
	page, listErr := a.Service.GetLowStockProductsPage(ctx, ownerID, req.GetLimit(), req.GetOffset())
	if listErr != nil {
		return nil, inventoryStatusError(http.StatusInternalServerError, "Failed to get low stock products")
	}
	out := make([]*erpinventoryv1.InventoryReportItem, 0, len(page.Items))
	for _, item := range page.Items {
		out = append(out, newInventoryReportItemPayload(item))
	}
	return &erpinventoryv1.GetLowStockResponse{
		Items:   out,
		Count:   page.Total,
		HasMore: page.HasMore,
		Limit:   page.Limit,
		Offset:  page.Offset,
		Total:   page.Total,
	}, nil
}

func (a *InventoryAPI) GetValuation(ctx context.Context, req *erpinventoryv1.GetValuationRequest) (*erpinventoryv1.GetValuationResponse, error) {
	ownerID, err := inventoryOwnerID(ctx, "erp:inventory:view")
	if err != nil {
		return nil, err
	}
	page, listErr := a.Service.GetValuationPage(ctx, ownerID, req.GetLimit(), req.GetOffset())
	if listErr != nil {
		return nil, inventoryStatusError(http.StatusInternalServerError, "Failed to get inventory valuation")
	}
	out := make([]*erpinventoryv1.ValuationItem, 0, len(page.Items))
	for _, item := range page.Items {
		out = append(out, newValuationItemPayload(item))
	}
	return &erpinventoryv1.GetValuationResponse{
		Items:      out,
		TotalValue: page.TotalValue,
		HasMore:    page.HasMore,
		Limit:      page.Limit,
		Offset:     page.Offset,
		Total:      page.Total,
	}, nil
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
	txType := domain.InventoryTransactionType(req.GetType())
	change := req.GetChange()
	if (txType == domain.InventoryTypeSale || txType == domain.InventoryTypeWaste) && change > 0 {
		change = -change
	}

	adjustErr := a.Service.AdjustInventory(ctx, ownerID, domain.InventoryAdjustInput{
		ProductID: req.GetProductId(),
		Change:    change,
		Type:      txType,
		Notes:     req.GetNotes(),
	})
	if errors.Is(adjustErr, sql.ErrNoRows) {
		return nil, inventoryStatusError(http.StatusNotFound, "Product not found")
	}
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
