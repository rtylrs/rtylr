package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/erp/inventory/internal/pos_stock/application"
	"github.com/voxire/onion/internal/modules/erp/inventory/internal/pos_stock/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	posstockv1 "github.com/voxire/voxie/gen/go/pos/stock/v1"
)

type StockAPI struct {
	Service *application.Service
}

func NewStockAPI(service *application.Service) *StockAPI {
	return &StockAPI{Service: service}
}

func (a *StockAPI) ListStockMovements(ctx context.Context, req *posstockv1.ListStockMovementsRequest) (*posstockv1.ListStockMovementsResponse, error) {
	businessID, _, err := requireStockPermission(ctx, "pos:reports:view")
	if err != nil {
		return nil, err
	}
	if req.GetBranchId() == 0 || req.GetProductId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "branch_id and product_id are required")
	}

	movements, listErr := a.Service.ListMovements(ctx, businessID, businessID, req.GetBranchId(), req.GetProductId())
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list stock movements")
	}

	payloads := make([]*posstockv1.StockMovement, len(movements))
	for i, m := range movements {
		payloads[i] = toProtoStockMovement(m)
	}
	return &posstockv1.ListStockMovementsResponse{Movements: payloads}, nil
}

func (a *StockAPI) ReceiveStock(ctx context.Context, req *posstockv1.ReceiveStockRequest) (*posstockv1.ReceiveStockResponse, error) {
	businessID, device, err := requireStockPermission(ctx, "pos:employees:manage")
	if err != nil {
		return nil, err
	}
	if req.GetBranchId() == 0 || req.GetProductId() == 0 || req.GetQuantity() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "branch_id, product_id, and positive quantity are required")
	}

	id, receiveErr := a.Service.ReceiveStock(ctx, domain.ReceiveStockInput{
		BusinessID: businessID,
		OwnerID:    device.User.UserID,
		BranchID:   req.GetBranchId(),
		ProductID:  req.GetProductId(),
		Quantity:   req.GetQuantity(),
		Note:       optionalString(req.GetNote()),
		EmployeeID: optionalInt64(req.GetEmployeeId()),
	})
	if receiveErr != nil {
		return nil, stockMovementError("Failed to receive stock", receiveErr)
	}
	return &posstockv1.ReceiveStockResponse{MovementId: id, Message: "Stock received successfully"}, nil
}

func (a *StockAPI) AdjustStock(ctx context.Context, req *posstockv1.AdjustStockRequest) (*posstockv1.AdjustStockResponse, error) {
	businessID, device, err := requireStockPermission(ctx, "pos:employees:manage")
	if err != nil {
		return nil, err
	}
	if req.GetBranchId() == 0 || req.GetProductId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "branch_id and product_id are required")
	}
	if req.GetQuantity() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "quantity must be non-zero")
	}

	id, adjustErr := a.Service.AdjustStock(ctx, domain.AdjustStockInput{
		BusinessID: businessID,
		OwnerID:    device.User.UserID,
		BranchID:   req.GetBranchId(),
		ProductID:  req.GetProductId(),
		Quantity:   req.GetQuantity(),
		Note:       optionalString(req.GetNote()),
		EmployeeID: optionalInt64(req.GetEmployeeId()),
	})
	if adjustErr != nil {
		return nil, stockMovementError("Failed to adjust stock", adjustErr)
	}
	return &posstockv1.AdjustStockResponse{MovementId: id, Message: "Stock adjusted successfully"}, nil
}

func (a *StockAPI) WasteStock(ctx context.Context, req *posstockv1.WasteStockRequest) (*posstockv1.WasteStockResponse, error) {
	businessID, device, err := requireStockPermission(ctx, "pos:employees:manage")
	if err != nil {
		return nil, err
	}
	if req.GetBranchId() == 0 || req.GetProductId() == 0 || req.GetQuantity() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "branch_id, product_id, and positive quantity are required")
	}

	id, wasteErr := a.Service.RecordWaste(ctx, domain.WasteStockInput{
		BusinessID: businessID,
		OwnerID:    device.User.UserID,
		BranchID:   req.GetBranchId(),
		ProductID:  req.GetProductId(),
		Quantity:   req.GetQuantity(),
		Note:       optionalString(req.GetNote()),
		EmployeeID: optionalInt64(req.GetEmployeeId()),
	})
	if wasteErr != nil {
		return nil, stockMovementError("Failed to record waste", wasteErr)
	}
	return &posstockv1.WasteStockResponse{MovementId: id, Message: "Waste recorded successfully"}, nil
}

func (a *StockAPI) TransferStock(ctx context.Context, req *posstockv1.TransferStockRequest) (*posstockv1.TransferStockResponse, error) {
	businessID, device, err := requireStockPermission(ctx, "pos:employees:manage")
	if err != nil {
		return nil, err
	}
	if req.GetFromBranchId() == 0 || req.GetToBranchId() == 0 || req.GetProductId() == 0 || req.GetQuantity() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "from_branch_id, to_branch_id, product_id, and positive quantity are required")
	}

	id, transferErr := a.Service.TransferStock(ctx, domain.TransferStockInput{
		BusinessID:   businessID,
		OwnerID:      device.User.UserID,
		FromBranchID: req.GetFromBranchId(),
		ToBranchID:   req.GetToBranchId(),
		ProductID:    req.GetProductId(),
		Quantity:     req.GetQuantity(),
		Note:         optionalString(req.GetNote()),
		EmployeeID:   optionalInt64(req.GetEmployeeId()),
	})
	if transferErr != nil {
		return nil, stockMovementError("Failed to transfer stock", transferErr)
	}
	return &posstockv1.TransferStockResponse{MovementId: id, Message: "Stock transferred successfully"}, nil
}

func (a *StockAPI) ListInventory(ctx context.Context, req *posstockv1.ListInventoryRequest) (*posstockv1.ListInventoryResponse, error) {
	businessID, _, err := requireStockPermission(ctx, "pos:reports:view")
	if err != nil {
		return nil, err
	}
	if req.GetBranchId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "branch_id is required")
	}

	items, listErr := a.Service.ListBranchInventory(ctx, businessID, businessID, req.GetBranchId())
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list inventory")
	}

	payloads := make([]*posstockv1.BranchInventory, len(items))
	for i, item := range items {
		payloads[i] = toProtoBranchInventory(item)
	}
	return &posstockv1.ListInventoryResponse{Inventory: payloads}, nil
}

func (a *StockAPI) LowStock(ctx context.Context, req *posstockv1.LowStockRequest) (*posstockv1.LowStockResponse, error) {
	businessID, _, err := requireStockPermission(ctx, "pos:reports:view")
	if err != nil {
		return nil, err
	}
	if req.GetBranchId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "branch_id is required")
	}

	items, listErr := a.Service.ListLowStockItems(ctx, businessID, businessID, req.GetBranchId())
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list low stock items")
	}

	payloads := make([]*posstockv1.BranchInventory, len(items))
	for i, item := range items {
		payloads[i] = toProtoBranchInventory(item)
	}
	return &posstockv1.LowStockResponse{Items: payloads}, nil
}

func requireStockPermission(ctx context.Context, permission string) (int64, authPkg.Device, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !authPkg.HasAppPermission(device.User, authorization.AppPOS, permission) {
		return 0, device, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return 0, device, onekiterrors.PermissionDenied()
	}
	return businessID, device, nil
}

func stockMovementError(fallback string, err error) error {
	if errors.Is(err, domain.ErrInsufficientStock) {
		return onekiterrors.New(http.StatusConflict, "Insufficient stock for this operation")
	}
	if errors.Is(err, domain.ErrInventoryNotFound) {
		return onekiterrors.New(http.StatusNotFound, "Inventory product was not found")
	}
	return onekiterrors.New(http.StatusInternalServerError, fallback)
}

func toProtoStockMovement(m domain.StockMovement) *posstockv1.StockMovement {
	out := &posstockv1.StockMovement{
		Id:         m.ID,
		BusinessId: m.BusinessID,
		BranchId:   m.BranchID,
		ProductId:  m.ProductID,
		Type:       m.Type,
		Quantity:   m.Quantity,
		CreatedAt:  m.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if m.ReferenceID != nil {
		out.ReferenceId = m.ReferenceID
	}
	if m.ReferenceType != nil {
		out.ReferenceType = m.ReferenceType
	}
	if m.Note != nil {
		out.Note = m.Note
	}
	if m.EmployeeID != nil {
		out.EmployeeId = m.EmployeeID
	}
	return out
}

func toProtoBranchInventory(item domain.BranchInventory) *posstockv1.BranchInventory {
	out := &posstockv1.BranchInventory{
		Id:         item.ID,
		BusinessId: item.BusinessID,
		BranchId:   item.BranchID,
		ProductId:  item.ProductID,
		ItemId:     item.ProductID,
		Name:       item.Name,
		Sku:        item.SKU,
		Unit:       item.Unit,
		Quantity:   item.Quantity,
		CurrentQty: item.Quantity,
		UpdatedAt:  item.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if item.LowStockThreshold != nil {
		out.LowStockThreshold = *item.LowStockThreshold
	}
	if item.ReorderPoint != nil {
		out.ReorderPoint = *item.ReorderPoint
	}
	if item.ReorderQuantity != nil {
		out.ReorderQuantity = *item.ReorderQuantity
	}
	return out
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func optionalInt64(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}
