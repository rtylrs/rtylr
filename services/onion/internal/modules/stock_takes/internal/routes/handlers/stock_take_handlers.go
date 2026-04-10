package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/stock_takes/internal/application"
	"github.com/voxire/onion/internal/modules/stock_takes/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// --- List Stock Takes --------------------------------------------------------

func ListStockTakesRoutes() []string {
	return []string{"/erp/stock-takes"}
}

type ListStockTakesHandler struct {
	Service *application.Service
}

type ListStockTakesRequest struct{}

type StockTakeItemPayload struct {
	ID          int64   `json:"id"`
	StockTakeID int64   `json:"stockTakeId"`
	ProductID   int64   `json:"productId"`
	ExpectedQty float64 `json:"expectedQty"`
	CountedQty  float64 `json:"countedQty"`
	Variance    float64 `json:"variance"`
	CreatedAt   string  `json:"createdAt"`
}

type StockTakePayload struct {
	ID          int64                  `json:"id"`
	OwnerID     int64                  `json:"ownerId"`
	WarehouseID int64                  `json:"warehouseId"`
	Status      string                 `json:"status"`
	Notes       string                 `json:"notes"`
	Items       []StockTakeItemPayload `json:"items"`
	StartedAt   string                 `json:"startedAt"`
	CompletedAt string                 `json:"completedAt"`
	CreatedAt   string                 `json:"createdAt"`
}

type ListStockTakesResponse struct {
	StockTakes []StockTakePayload `json:"stockTakes"`
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

func (h *ListStockTakesHandler) Serve(ctx context.Context, _ *ListStockTakesRequest) (ListStockTakesResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListStockTakesResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	stockTakes, err := h.Service.ListStockTakes(ctx, authUser.UserID)
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

	return ListStockTakesResponse{StockTakes: payloads}, nil
}

// --- Create Stock Take -------------------------------------------------------

func CreateStockTakeRoutes() []string {
	return []string{"/erp/stock-takes"}
}

type CreateStockTakeHandler struct {
	Service *application.Service
}

type CreateStockTakeItemInput struct {
	ProductID   int64   `json:"productId"`
	ExpectedQty float64 `json:"expectedQty"`
	CountedQty  float64 `json:"countedQty"`
}

type CreateStockTakeRequest struct {
	WarehouseID int64                      `json:"warehouseId"`
	Notes       string                     `json:"notes"`
	Items       []CreateStockTakeItemInput `json:"items"`
}

type CreateStockTakeResponse struct {
	StockTakeID int64  `json:"stockTakeId"`
	Message     string `json:"message"`
}

func (h *CreateStockTakeHandler) Validate(_ context.Context, _ *CreateStockTakeRequest) *errorModels.ServerError {
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

	id, err := h.Service.CreateStockTake(ctx, authUser.UserID, input)
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
	ProductID   int64   `json:"productId"`
	ExpectedQty float64 `json:"expectedQty"`
	CountedQty  float64 `json:"countedQty"`
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

func (h *UpdateStockTakeHandler) Validate(_ context.Context, req *UpdateStockTakeRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Stock take ID is required",
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

	if err := h.Service.UpdateStockTake(ctx, authUser.UserID, input); err != nil {
		return UpdateStockTakeResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to update stock take",
		}
	}

	return UpdateStockTakeResponse{Message: "Stock take updated successfully"}, nil
}
