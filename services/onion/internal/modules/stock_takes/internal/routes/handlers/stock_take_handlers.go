package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/stock_takes/internal/application"
	"github.com/voxire/onion/internal/modules/stock_takes/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// --- List Stock Takes --------------------------------------------------------

func ListStockTakesRoutes() []string {
	return []string{"/erp/stock-takes"}
}

type ListStockTakesHandler struct {
	Service *application.Service
}

type ListStockTakesRequest struct {
	Page     int64 `json:"page"`
	PageSize int64 `json:"page_size"`
}

type StockTakeItemPayload struct {
	ID          int64   `json:"id"`
	StockTakeID int64   `json:"stock_take_id"`
	ProductID   int64   `json:"product_id"`
	ExpectedQty float64 `json:"expected_qty"`
	CountedQty  float64 `json:"counted_qty"`
	Variance    float64 `json:"variance"`
	CreatedAt   string  `json:"created_at"`
}

type StockTakePayload struct {
	ID          int64                  `json:"id"`
	OwnerID     int64                  `json:"owner_id"`
	WarehouseID int64                  `json:"warehouse_id"`
	Status      string                 `json:"status"`
	Notes       string                 `json:"notes"`
	Items       []StockTakeItemPayload `json:"items"`
	StartedAt   string                 `json:"started_at"`
	CompletedAt string                 `json:"completed_at"`
	CreatedAt   string                 `json:"created_at"`
}

type ListStockTakesResponse struct {
	StockTakes []StockTakePayload `json:"stock_takes"`
	HasMore    bool               `json:"has_more"`
	Page       int64              `json:"page"`
	PageSize   int64              `json:"page_size"`
}

func newStockTakePayload(s domain.StockTake) StockTakePayload {
	items := make([]StockTakeItemPayload, 0, len(s.Items))
	for _, item := range s.Items {
		items = append(items, StockTakeItemPayload{
			ID:          item.ID,
			StockTakeID: item.StockTakeID,
			ProductID:   item.ProductID,
			ExpectedQty: item.ExpectedQty,
			CountedQty:  item.CountedQty,
			Variance:    item.Variance,
			CreatedAt:   item.CreatedAt,
		})
	}

	return StockTakePayload{
		ID:          s.ID,
		OwnerID:     s.OwnerID,
		WarehouseID: s.WarehouseID,
		Status:      string(s.Status),
		Notes:       s.Notes,
		Items:       items,
		StartedAt:   s.StartedAt,
		CompletedAt: s.CompletedAt,
		CreatedAt:   s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *ListStockTakesHandler) Serve(ctx context.Context, req *ListStockTakesRequest) (ListStockTakesResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListStockTakesResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListStockTakesResponse{}, errorModels.PermissionDeniedServerError()
	}

	stockTakes, hasMore, err := h.Service.ListStockTakes(ctx, businessID, req.Page, req.PageSize)
	if err != nil {
		return ListStockTakesResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to list stock takes",
		}
	}

	payloads := make([]StockTakePayload, 0, len(stockTakes))
	for _, s := range stockTakes {
		payloads = append(payloads, newStockTakePayload(s))
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	return ListStockTakesResponse{StockTakes: payloads, HasMore: hasMore, Page: page, PageSize: pageSize}, nil
}

// --- Create Stock Take -------------------------------------------------------

func CreateStockTakeRoutes() []string {
	return []string{"/erp/stock-takes"}
}

type CreateStockTakeHandler struct {
	Service *application.Service
}

type CreateStockTakeItemInput struct {
	ProductID   int64   `json:"product_id"`
	ExpectedQty float64 `json:"expected_qty"`
	CountedQty  float64 `json:"counted_qty"`
}

type CreateStockTakeRequest struct {
	WarehouseID int64                      `json:"warehouse_id"`
	Notes       string                     `json:"notes"`
	Items       []CreateStockTakeItemInput `json:"items"`
}

type CreateStockTakeResponse struct {
	StockTakeID int64  `json:"stock_take_id"`
	Message     string `json:"message"`
}

func (h *CreateStockTakeHandler) Validate(_ context.Context, req *CreateStockTakeRequest) *errorModels.ServerError {
	if req.WarehouseID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Warehouse ID is required",
		}
	}

	return nil
}

func (h *CreateStockTakeHandler) Serve(ctx context.Context, req *CreateStockTakeRequest) (CreateStockTakeResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateStockTakeResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	items := make([]domain.StockTakeItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, domain.StockTakeItemInput{
			ProductID:   item.ProductID,
			ExpectedQty: item.ExpectedQty,
			CountedQty:  item.CountedQty,
		})
	}

	input := domain.StockTakeInput{
		WarehouseID: req.WarehouseID,
		Notes:       req.Notes,
		Items:       items,
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateStockTakeResponse{}, errorModels.PermissionDeniedServerError()
	}

	id, err := h.Service.CreateStockTake(ctx, businessID, input)
	if err != nil {
		return CreateStockTakeResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to create stock take",
		}
	}

	return CreateStockTakeResponse{StockTakeID: id, Message: "Stock take created successfully"}, nil
}

// --- Update Stock Take -------------------------------------------------------

func UpdateStockTakeRoutes() []string {
	return []string{"/erp/stock-takes"}
}

type UpdateStockTakeHandler struct {
	Service *application.Service
}

type UpdateStockTakeItemInput struct {
	ProductID   int64   `json:"product_id"`
	ExpectedQty float64 `json:"expected_qty"`
	CountedQty  float64 `json:"counted_qty"`
}

type UpdateStockTakeRequest struct {
	ID     int64                      `json:"id"`
	Status string                     `json:"status"`
	Notes  string                     `json:"notes"`
	Items  []UpdateStockTakeItemInput `json:"items"`
}

type UpdateStockTakeResponse struct {
	Message string `json:"message"`
}

var validStockTakeStatuses = map[string]struct{}{
	string(domain.StockTakeStatusDraft):      {},
	string(domain.StockTakeStatusInProgress): {},
	string(domain.StockTakeStatusCompleted):  {},
}

func (h *UpdateStockTakeHandler) Validate(_ context.Context, req *UpdateStockTakeRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Stock take ID is required",
		}
	}

	if req.Status == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Status is required",
		}
	}

	if _, ok := validStockTakeStatuses[req.Status]; !ok {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Invalid status. Must be one of: draft, in_progress, completed",
		}
	}

	return nil
}

func (h *UpdateStockTakeHandler) Serve(ctx context.Context, req *UpdateStockTakeRequest) (UpdateStockTakeResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateStockTakeResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	items := make([]domain.StockTakeItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, domain.StockTakeItemInput{
			ProductID:   item.ProductID,
			ExpectedQty: item.ExpectedQty,
			CountedQty:  item.CountedQty,
		})
	}

	input := domain.StockTakeUpdateInput{
		ID:     req.ID,
		Status: domain.StockTakeStatus(req.Status),
		Notes:  req.Notes,
		Items:  items,
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateStockTakeResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.UpdateStockTake(ctx, businessID, input); err != nil {
		return UpdateStockTakeResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to update stock take",
		}
	}

	return UpdateStockTakeResponse{Message: "Stock take updated successfully"}, nil
}
