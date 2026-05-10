package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/inventory/internal/application"
	"github.com/voxire/onion/internal/modules/inventory/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	erpv1 "github.com/voxire/voxie/gen/go/erp/v1"
)

func InventoryReportRoutes() []string {
	return []string{"/erp/inventory/report"}
}

type InventoryReportHandler struct {
	Service *application.Service
}

type InventoryReportRequest struct{}

type InventoryReportResponse struct {
	Items      []*erpv1.InventoryReportItem `json:"items"`
	TotalValue float64                      `json:"total_value"`
}

func newInventoryReportItemPayload(item domain.InventoryReportItem) *erpv1.InventoryReportItem {
	return &erpv1.InventoryReportItem{
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

	items, err := h.Service.GetInventoryReport(ctx, authUser.UserID)
	if err != nil {
		return InventoryReportResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to get inventory report",
		}
	}

	payloads := make([]*erpv1.InventoryReportItem, 0, len(items))
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
