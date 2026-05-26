package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_stock_movements/internal/application"
	"github.com/voxire/onion/internal/modules/pos_stock_movements/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	posv1 "github.com/voxire/voxie/gen/go/pos/v1"
)

// ─── List Movements ──────────────────────────────────────────────────────────

func ListMovementsRoutes() []string { return []string{"/pos/stock/movements"} }

type (
	ListMovementsHandler struct{ Service *application.Service }
	ListMovementsRequest struct {
		BranchID  int64 `json:"branch_id"`
		ProductID int64 `json:"product_id"`
	}
)

type ListMovementsResponse struct {
	Movements []*posv1.StockMovement `json:"movements"`
}

func (h *ListMovementsHandler) Serve(ctx context.Context, req *ListMovementsRequest) (ListMovementsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListMovementsResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.BranchID == 0 || req.ProductID == 0 {
		return ListMovementsResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "branch_id and product_id are required"}
	}
	movements, err := h.Service.ListMovements(ctx, device.User.BusinessID, req.BranchID, req.ProductID)
	if err != nil {
		return ListMovementsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list stock movements"}
	}
	payloads := make([]*posv1.StockMovement, len(movements))
	for i, m := range movements {
		payloads[i] = &posv1.StockMovement{
			Id:            m.ID,
			BusinessId:    m.BusinessID,
			BranchId:      m.BranchID,
			ProductId:     m.ProductID,
			Type:          m.Type,
			Quantity:      m.Quantity,
			ReferenceId:   m.ReferenceID,
			ReferenceType: m.ReferenceType,
			Note:          m.Note,
			EmployeeId:    m.EmployeeID,
			CreatedAt:     m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}
	return ListMovementsResponse{Movements: payloads}, nil
}

// ─── Receive Stock ───────────────────────────────────────────────────────────

func ReceiveStockRoutes() []string { return []string{"/pos/stock/receive"} }

type ReceiveStockHandler struct{ Service *application.Service }

type ReceiveStockRequest struct {
	BranchID   int64   `json:"branch_id"`
	ProductID  int64   `json:"product_id"`
	Quantity   float64 `json:"quantity"`
	Note       *string `json:"note"`
	EmployeeID *int64  `json:"employee_id"`
}

type ReceiveStockResponse struct {
	MovementID int64  `json:"movement_id"`
	Message    string `json:"message"`
}

func (h *ReceiveStockHandler) Serve(ctx context.Context, req *ReceiveStockRequest) (ReceiveStockResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ReceiveStockResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.BranchID == 0 || req.ProductID == 0 || req.Quantity <= 0 {
		return ReceiveStockResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "branchId, productId, and positive quantity are required"}
	}
	id, err := h.Service.ReceiveStock(ctx, domain.ReceiveStockInput{
		BusinessID: device.User.BusinessID, BranchID: req.BranchID, ProductID: req.ProductID,
		Quantity: req.Quantity, Note: req.Note, EmployeeID: req.EmployeeID,
	})
	if err != nil {
		return ReceiveStockResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to receive stock"}
	}
	return ReceiveStockResponse{MovementID: id, Message: "Stock received successfully"}, nil
}

// ─── Adjust Stock ────────────────────────────────────────────────────────────

func AdjustStockRoutes() []string { return []string{"/pos/stock/adjust"} }

type AdjustStockHandler struct{ Service *application.Service }

type AdjustStockRequest struct {
	BranchID   int64   `json:"branch_id"`
	ProductID  int64   `json:"product_id"`
	Quantity   float64 `json:"quantity"`
	Note       *string `json:"note"`
	EmployeeID *int64  `json:"employee_id"`
}

type AdjustStockResponse struct {
	MovementID int64  `json:"movement_id"`
	Message    string `json:"message"`
}

func (h *AdjustStockHandler) Validate(_ context.Context, req *AdjustStockRequest) *errorModels.ServerError {
	if req.BranchID == 0 || req.ProductID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "branchId and productId are required"}
	}
	if req.Quantity == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "quantity must be non-zero"}
	}
	return nil
}

func (h *AdjustStockHandler) Serve(ctx context.Context, req *AdjustStockRequest) (AdjustStockResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return AdjustStockResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.AdjustStock(ctx, domain.AdjustStockInput{
		BusinessID: device.User.BusinessID, BranchID: req.BranchID, ProductID: req.ProductID,
		Quantity: req.Quantity, Note: req.Note, EmployeeID: req.EmployeeID,
	})
	if err != nil {
		return AdjustStockResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to adjust stock"}
	}
	return AdjustStockResponse{MovementID: id, Message: "Stock adjusted successfully"}, nil
}

// ─── Record Waste ────────────────────────────────────────────────────────────

func WasteStockRoutes() []string { return []string{"/pos/stock/waste"} }

type WasteStockHandler struct{ Service *application.Service }

type WasteStockRequest struct {
	BranchID   int64   `json:"branch_id"`
	ProductID  int64   `json:"product_id"`
	Quantity   float64 `json:"quantity"`
	Note       *string `json:"note"`
	EmployeeID *int64  `json:"employee_id"`
}

type WasteStockResponse struct {
	MovementID int64  `json:"movement_id"`
	Message    string `json:"message"`
}

func (h *WasteStockHandler) Serve(ctx context.Context, req *WasteStockRequest) (WasteStockResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return WasteStockResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.BranchID == 0 || req.ProductID == 0 || req.Quantity <= 0 {
		return WasteStockResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "branchId, productId, and positive quantity are required"}
	}
	id, err := h.Service.RecordWaste(ctx, domain.WasteStockInput{
		BusinessID: device.User.BusinessID, BranchID: req.BranchID, ProductID: req.ProductID,
		Quantity: req.Quantity, Note: req.Note, EmployeeID: req.EmployeeID,
	})
	if err != nil {
		return WasteStockResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to record waste"}
	}
	return WasteStockResponse{MovementID: id, Message: "Waste recorded successfully"}, nil
}

// ─── Transfer Stock ──────────────────────────────────────────────────────────

func TransferStockRoutes() []string { return []string{"/pos/stock/transfer"} }

type TransferStockHandler struct{ Service *application.Service }

type TransferStockRequest struct {
	FromBranchID int64   `json:"from_branch_id"`
	ToBranchID   int64   `json:"to_branch_id"`
	ProductID    int64   `json:"product_id"`
	Quantity     float64 `json:"quantity"`
	Note         *string `json:"note"`
	EmployeeID   *int64  `json:"employee_id"`
}

type TransferStockResponse struct {
	MovementID int64  `json:"movement_id"`
	Message    string `json:"message"`
}

func (h *TransferStockHandler) Serve(ctx context.Context, req *TransferStockRequest) (TransferStockResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return TransferStockResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.FromBranchID == 0 || req.ToBranchID == 0 || req.ProductID == 0 || req.Quantity <= 0 {
		return TransferStockResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "fromBranchId, toBranchId, productId, and positive quantity are required"}
	}
	id, err := h.Service.TransferStock(ctx, domain.TransferStockInput{
		BusinessID: device.User.BusinessID, FromBranchID: req.FromBranchID, ToBranchID: req.ToBranchID,
		ProductID: req.ProductID, Quantity: req.Quantity, Note: req.Note, EmployeeID: req.EmployeeID,
	})
	if err != nil {
		return TransferStockResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to transfer stock"}
	}
	return TransferStockResponse{MovementID: id, Message: "Stock transferred successfully"}, nil
}

// ─── List Branch Inventory ───────────────────────────────────────────────────

func ListInventoryRoutes() []string { return []string{"/pos/stock/inventory"} }

type (
	ListInventoryHandler struct{ Service *application.Service }
	ListInventoryRequest struct {
		BranchID int64 `json:"branch_id"`
	}
)

type ListInventoryResponse struct {
	Inventory []*posv1.BranchInventory `json:"inventory"`
}

func toProtoBranchInventory(item domain.BranchInventory) *posv1.BranchInventory {
	p := &posv1.BranchInventory{
		Id:         item.ID,
		BusinessId: item.BusinessID,
		BranchId:   item.BranchID,
		ProductId:  item.ProductID,
		Quantity:   item.Quantity,
		UpdatedAt:  item.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if item.ReorderPoint != nil {
		p.ReorderPoint = *item.ReorderPoint
	}
	if item.ReorderQuantity != nil {
		p.ReorderQuantity = *item.ReorderQuantity
	}
	return p
}

func (h *ListInventoryHandler) Serve(ctx context.Context, req *ListInventoryRequest) (ListInventoryResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListInventoryResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.BranchID == 0 {
		return ListInventoryResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "branch_id is required"}
	}
	items, err := h.Service.ListBranchInventory(ctx, device.User.BusinessID, req.BranchID)
	if err != nil {
		return ListInventoryResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list inventory"}
	}
	payloads := make([]*posv1.BranchInventory, len(items))
	for i, item := range items {
		payloads[i] = toProtoBranchInventory(item)
	}
	return ListInventoryResponse{Inventory: payloads}, nil
}

// ─── Low Stock Items ─────────────────────────────────────────────────────────

func LowStockRoutes() []string { return []string{"/pos/stock/low-stock"} }

type (
	LowStockHandler struct{ Service *application.Service }
	LowStockRequest struct {
		BranchID int64 `json:"branch_id"`
	}
)

type LowStockResponse struct {
	Items []*posv1.BranchInventory `json:"items"`
}

func (h *LowStockHandler) Serve(ctx context.Context, req *LowStockRequest) (LowStockResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return LowStockResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.BranchID == 0 {
		return LowStockResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "branch_id is required"}
	}
	items, err := h.Service.ListLowStockItems(ctx, device.User.BusinessID, req.BranchID)
	if err != nil {
		return LowStockResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list low stock items"}
	}
	payloads := make([]*posv1.BranchInventory, len(items))
	for i, item := range items {
		payloads[i] = toProtoBranchInventory(item)
	}
	return LowStockResponse{Items: payloads}, nil
}
