package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/erp/stock_transfers/internal/application"
	"github.com/voxire/onion/internal/modules/erp/stock_transfers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	erpstocktransferv1 "github.com/voxire/voxie/gen/go/erp/stock_transfer/v1"
)

type StockTransferAPI struct {
	Service *application.Service
}

func NewStockTransferAPI(service *application.Service) *StockTransferAPI {
	return &StockTransferAPI{Service: service}
}

func (a *StockTransferAPI) ListStockTransfers(ctx context.Context, req *erpstocktransferv1.ListStockTransfersRequest) (*erpstocktransferv1.ListStockTransfersResponse, error) {
	ownerID, err := stockTransferOwnerID(ctx, "erp:inventory:view")
	if err != nil {
		return nil, err
	}
	page := req.GetPage()
	if page < 1 {
		page = 1
	}
	pageSize := req.GetPageSize()
	if pageSize < 1 {
		pageSize = 50
	}
	transfers, hasMore, listErr := a.Service.ListStockTransfers(ctx, ownerID, page, pageSize)
	if listErr != nil {
		return nil, stockTransferStatusError(http.StatusInternalServerError, "Failed to list stock transfers")
	}
	out := make([]*erpstocktransferv1.StockTransfer, 0, len(transfers))
	for _, t := range transfers {
		out = append(out, newStockTransferProto(t))
	}
	return &erpstocktransferv1.ListStockTransfersResponse{
		StockTransfers: out,
		HasMore:        hasMore,
		Page:           page,
		PageSize:       pageSize,
	}, nil
}

func (a *StockTransferAPI) CreateStockTransfer(ctx context.Context, req *erpstocktransferv1.CreateStockTransferRequest) (*erpstocktransferv1.CreateStockTransferResponse, error) {
	ownerID, err := stockTransferOwnerID(ctx, "erp:inventory:transfer")
	if err != nil {
		return nil, err
	}
	items := make([]domain.StockTransferItemInput, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		items = append(items, domain.StockTransferItemInput{
			ProductID: item.GetProductId(),
			Quantity:  item.GetQuantity(),
		})
	}
	id, createErr := a.Service.CreateStockTransfer(ctx, ownerID, domain.StockTransferInput{
		FromWarehouseID: req.GetFromWarehouseId(),
		ToWarehouseID:   req.GetToWarehouseId(),
		Notes:           req.GetNotes(),
		Items:           items,
	})
	if createErr != nil {
		if errors.Is(createErr, domain.ErrInvalidStockTransfer) {
			return nil, stockTransferStatusError(http.StatusBadRequest, createErr.Error())
		}
		if errors.Is(createErr, domain.ErrStockTransferReferenceNotFound) {
			return nil, stockTransferStatusError(http.StatusNotFound, createErr.Error())
		}
		return nil, stockTransferStatusError(http.StatusInternalServerError, "Failed to create stock transfer")
	}
	return &erpstocktransferv1.CreateStockTransferResponse{
		TransferId: id,
		Message:    "Stock transfer created successfully",
	}, nil
}

func (a *StockTransferAPI) UpdateStockTransferStatus(ctx context.Context, req *erpstocktransferv1.UpdateStockTransferStatusRequest) (*erpstocktransferv1.UpdateStockTransferStatusResponse, error) {
	ownerID, err := stockTransferOwnerID(ctx, "erp:inventory:transfer")
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil {
		return nil, stockTransferStatusError(http.StatusBadRequest, "Status body is required")
	}
	status := body.GetStatus()
	validStatuses := map[string]struct{}{
		string(domain.StockTransferStatusDraft):     {},
		string(domain.StockTransferStatusInTransit): {},
		string(domain.StockTransferStatusReceived):  {},
		string(domain.StockTransferStatusCancelled): {},
	}
	if _, ok := validStatuses[status]; !ok {
		return nil, stockTransferStatusError(http.StatusBadRequest, "Invalid status. Must be one of: draft, in_transit, received, cancelled")
	}
	if updateErr := a.Service.UpdateStockTransferStatus(ctx, ownerID, req.GetId(), domain.StockTransferStatus(status)); updateErr != nil {
		return nil, stockTransferMutationError(updateErr)
	}
	return &erpstocktransferv1.UpdateStockTransferStatusResponse{Message: "Stock transfer status updated successfully"}, nil
}

func newStockTransferProto(t domain.StockTransfer) *erpstocktransferv1.StockTransfer {
	items := make([]*erpstocktransferv1.StockTransferItem, 0, len(t.Items))
	for _, item := range t.Items {
		items = append(items, &erpstocktransferv1.StockTransferItem{
			Id:         item.ID,
			TransferId: item.TransferID,
			ProductId:  item.ProductID,
			Quantity:   item.Quantity,
			CreatedAt:  item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return &erpstocktransferv1.StockTransfer{
		Id:              t.ID,
		OwnerId:         t.OwnerID,
		FromWarehouseId: t.FromWarehouseID,
		ToWarehouseId:   t.ToWarehouseID,
		Status:          string(t.Status),
		Notes:           t.Notes,
		Items:           items,
		CreatedAt:       t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func stockTransferOwnerID(ctx context.Context, permission string) (int64, error) {
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

func stockTransferStatusError(status int, message string) error {
	return onekiterrors.New(status, message)
}
