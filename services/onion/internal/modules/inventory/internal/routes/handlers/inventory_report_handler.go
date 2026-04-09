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
	ProductID         int64   `json:"productId"`
	ProductName       string  `json:"productName"`
	SKU               string  `json:"sku"`
	Unit              string  `json:"unit"`
	StockQuantity     float64 `json:"stockQuantity"`
	LowStockThreshold float64 `json:"lowStockThreshold"`
	ReorderQuantity   float64 `json:"reorderQuantity"`
	CostPrice         float64 `json:"costPrice"`
	SalePrice         float64 `json:"salePrice"`
	StockValue        float64 `json:"stockValue"`
	IsLowStock        bool    `json:"isLowStock"`
}

type InventoryReportResponse struct {
	Items      []InventoryReportItemPayload `json:"items"`
	TotalValue float64                      `json:"totalValue"`
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
