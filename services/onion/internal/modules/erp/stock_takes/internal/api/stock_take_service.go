package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/erp/stock_takes/internal/application"
	"github.com/voxire/onion/internal/modules/erp/stock_takes/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	erpstocktakev1 "github.com/voxire/voxie/gen/go/erp/stock_take/v1"
)

type StockTakeAPI struct {
	Service *application.Service
}

func NewStockTakeAPI(service *application.Service) *StockTakeAPI {
	return &StockTakeAPI{Service: service}
}

func (a *StockTakeAPI) ListStockTakes(ctx context.Context, req *erpstocktakev1.ListStockTakesRequest) (*erpstocktakev1.ListStockTakesResponse, error) {
	ownerID, err := stockTakeOwnerID(ctx, "erp:inventory:view")
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
	stockTakes, hasMore, listErr := a.Service.ListStockTakes(ctx, ownerID, page, pageSize)
	if listErr != nil {
		return nil, stockTakeStatusError(http.StatusInternalServerError, "Failed to list stock takes")
	}
	out := make([]*erpstocktakev1.StockTake, 0, len(stockTakes))
	for _, s := range stockTakes {
		out = append(out, newStockTakeProto(s))
	}
	return &erpstocktakev1.ListStockTakesResponse{
		StockTakes: out,
		HasMore:    hasMore,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (a *StockTakeAPI) CreateStockTake(ctx context.Context, req *erpstocktakev1.CreateStockTakeRequest) (*erpstocktakev1.CreateStockTakeResponse, error) {
	ownerID, err := stockTakeOwnerID(ctx, "erp:inventory:adjust")
	if err != nil {
		return nil, err
	}
	items := make([]domain.StockTakeItemInput, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		items = append(items, domain.StockTakeItemInput{
			ProductID:   item.GetProductId(),
			ExpectedQty: item.GetExpectedQty(),
			CountedQty:  item.GetCountedQty(),
		})
	}
	id, createErr := a.Service.CreateStockTake(ctx, ownerID, domain.StockTakeInput{
		WarehouseID: req.GetWarehouseId(),
		Notes:       req.GetNotes(),
		Items:       items,
	})
	if createErr != nil {
		if errors.Is(createErr, domain.ErrStockTakeReferenceNotFound) {
			return nil, stockTakeStatusError(http.StatusNotFound, createErr.Error())
		}
		return nil, stockTakeStatusError(http.StatusInternalServerError, "Failed to create stock take")
	}
	return &erpstocktakev1.CreateStockTakeResponse{
		StockTakeId: id,
		Message:     "Stock take created successfully",
	}, nil
}

func (a *StockTakeAPI) UpdateStockTake(ctx context.Context, req *erpstocktakev1.UpdateStockTakeRequest) (*erpstocktakev1.UpdateStockTakeResponse, error) {
	ownerID, err := stockTakeOwnerID(ctx, "erp:inventory:approve")
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil {
		return nil, stockTakeStatusError(http.StatusBadRequest, "Stock take body is required")
	}
	status := body.GetStatus()
	validStatuses := map[string]struct{}{
		string(domain.StockTakeStatusDraft):      {},
		string(domain.StockTakeStatusInProgress): {},
		string(domain.StockTakeStatusCompleted):  {},
	}
	if _, ok := validStatuses[status]; !ok {
		return nil, stockTakeStatusError(http.StatusBadRequest, "Invalid status. Must be one of: draft, in_progress, completed")
	}
	items := make([]domain.StockTakeItemInput, 0, len(body.GetItems()))
	for _, item := range body.GetItems() {
		items = append(items, domain.StockTakeItemInput{
			ProductID:   item.GetProductId(),
			ExpectedQty: item.GetExpectedQty(),
			CountedQty:  item.GetCountedQty(),
		})
	}
	if updateErr := a.Service.UpdateStockTake(ctx, ownerID, domain.StockTakeUpdateInput{
		ID:     req.GetId(),
		Status: domain.StockTakeStatus(status),
		Notes:  body.GetNotes(),
		Items:  items,
	}); updateErr != nil {
		return nil, stockTakeMutationError(updateErr)
	}
	return &erpstocktakev1.UpdateStockTakeResponse{Message: "Stock take updated successfully"}, nil
}

func newStockTakeProto(s domain.StockTake) *erpstocktakev1.StockTake {
	items := make([]*erpstocktakev1.StockTakeItem, 0, len(s.Items))
	for _, item := range s.Items {
		items = append(items, &erpstocktakev1.StockTakeItem{
			Id:          item.ID,
			StockTakeId: item.StockTakeID,
			ProductId:   item.ProductID,
			ExpectedQty: item.ExpectedQty,
			CountedQty:  item.CountedQty,
			Variance:    item.Variance,
			CreatedAt:   item.CreatedAt,
		})
	}
	return &erpstocktakev1.StockTake{
		Id:          s.ID,
		OwnerId:     s.OwnerID,
		WarehouseId: s.WarehouseID,
		Status:      string(s.Status),
		Notes:       s.Notes,
		Items:       items,
		StartedAt:   s.StartedAt,
		CompletedAt: s.CompletedAt,
		CreatedAt:   s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func stockTakeOwnerID(ctx context.Context, permission string) (int64, error) {
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

func stockTakeStatusError(status int, message string) error {
	return onekiterrors.New(status, message)
}
