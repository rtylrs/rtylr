package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/inventory/internal/application"
	"github.com/voxire/onion/internal/modules/inventory/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	erpinventoryv1 "github.com/voxire/voxie/gen/go/erp/inventory/v1"
)

func InventoryReportRoutes() []string {
	return []string{"/erp/inventory/report"}
}

type InventoryReportHandler struct {
	Service *application.Service
}

type InventoryReportRequest struct{}

type InventoryReportResponse struct {
	Items      []*erpinventoryv1.InventoryReportItem `json:"items"`
	TotalValue float64                               `json:"total_value"`
}

func newInventoryReportItemPayload(item domain.InventoryReportItem) *erpinventoryv1.InventoryReportItem {
	return &erpinventoryv1.InventoryReportItem{
		ProductId:         item.ProductID,
		ProductName:       item.ProductName,
		Sku:               item.SKU,
		Unit:              item.Unit,
		StockQuantity:     item.StockQuantity,
		LowStockThreshold: item.LowStockThreshold,
		ReorderQuantity:   item.ReorderQuantity,
		CostPrice:         item.CostPrice,
		SalePrice:         item.SalePrice,
		StockValue:        item.StockValue,
		IsLowStock:        item.IsLowStock,
	}
}

func (h *InventoryReportHandler) Serve(ctx context.Context, _ *InventoryReportRequest) (InventoryReportResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return InventoryReportResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
	items, err := h.Service.GetInventoryReport(ctx, ownerID)
	if err != nil {
		return InventoryReportResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to get inventory report",
		}
	}

	payloads := make([]*erpinventoryv1.InventoryReportItem, 0, len(items))
	var totalValue float64

	for _, item := range items {
		payloads = append(payloads, newInventoryReportItemPayload(item))
		totalValue += item.StockValue
	}

	return InventoryReportResponse{
		Items:      payloads,
		TotalValue: totalValue,
	}, nil
}
