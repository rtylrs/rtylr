package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/inventory/internal/application"
	"github.com/voxire/onion/internal/modules/inventory/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func InventoryReportRoutes() []string {
	return []string{"/erp/inventory/report"}
}

type InventoryReportHandler struct {
	Service *application.Service
}

type InventoryReportRequest struct{}

type InventoryReportItemPayload struct {
	ProductID         int64   `json:"product_id"`
	ProductName       string  `json:"product_name"`
	SKU               string  `json:"sku"`
	Unit              string  `json:"unit"`
	StockQuantity     float64 `json:"stock_quantity"`
	LowStockThreshold float64 `json:"low_stock_threshold"`
	ReorderQuantity   float64 `json:"reorder_quantity"`
	CostPrice         float64 `json:"cost_price"`
	SalePrice         float64 `json:"sale_price"`
	StockValue        float64 `json:"stock_value"`
	IsLowStock        bool    `json:"is_low_stock"`
}

type InventoryReportResponse struct {
	Items      []InventoryReportItemPayload `json:"items"`
	TotalValue float64                      `json:"total_value"`
}

func newInventoryReportItemPayload(item domain.InventoryReportItem) InventoryReportItemPayload {
	return InventoryReportItemPayload{
		ProductID:         item.ProductID,
		ProductName:       item.ProductName,
		SKU:               item.SKU,
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

	payloads := make([]InventoryReportItemPayload, 0, len(items))
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
