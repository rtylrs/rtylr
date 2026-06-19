package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/stock_transfers/internal/application"
	"github.com/voxire/onion/internal/modules/stock_transfers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// --- List Stock Transfers ----------------------------------------------------

func ListStockTransfersRoutes() []string {
	return []string{"/erp/stock-transfers"}
}

type ListStockTransfersHandler struct {
	Service *application.Service
}

type ListStockTransfersRequest struct {
	Page     int64 `json:"page"`
	PageSize int64 `json:"page_size"`
}

type StockTransferItemPayload struct {
	ID         int64   `json:"id"`
	TransferID int64   `json:"transfer_id"`
	ProductID  int64   `json:"product_id"`
	Quantity   float64 `json:"quantity"`
	CreatedAt  string  `json:"created_at"`
}

type StockTransferPayload struct {
	ID              int64                      `json:"id"`
	OwnerID         int64                      `json:"owner_id"`
	FromWarehouseID int64                      `json:"from_warehouse_id"`
	ToWarehouseID   int64                      `json:"to_warehouse_id"`
	Status          string                     `json:"status"`
	Notes           string                     `json:"notes"`
	Items           []StockTransferItemPayload `json:"items"`
	CreatedAt       string                     `json:"created_at"`
	UpdatedAt       string                     `json:"updated_at"`
}

type ListStockTransfersResponse struct {
	StockTransfers []StockTransferPayload `json:"stock_transfers"`
	HasMore        bool                   `json:"has_more"`
	Page           int64                  `json:"page"`
	PageSize       int64                  `json:"page_size"`
}

func newStockTransferPayload(t domain.StockTransfer) StockTransferPayload {
	items := make([]StockTransferItemPayload, 0, len(t.Items))
	for _, item := range t.Items {
		items = append(items, StockTransferItemPayload{
			ID:         item.ID,
			TransferID: item.TransferID,
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			CreatedAt:  item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return StockTransferPayload{
		ID:              t.ID,
		OwnerID:         t.OwnerID,
		FromWarehouseID: t.FromWarehouseID,
		ToWarehouseID:   t.ToWarehouseID,
		Status:          string(t.Status),
		Notes:           t.Notes,
		Items:           items,
		CreatedAt:       t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *ListStockTransfersHandler) Serve(ctx context.Context, req *ListStockTransfersRequest) (ListStockTransfersResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListStockTransfersResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListStockTransfersResponse{}, errorModels.PermissionDeniedServerError()
	}

	transfers, hasMore, err := h.Service.ListStockTransfers(ctx, businessID, req.Page, req.PageSize)
	if err != nil {
		return ListStockTransfersResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to list stock transfers",
		}
	}

	payloads := make([]StockTransferPayload, 0, len(transfers))
	for _, t := range transfers {
		payloads = append(payloads, newStockTransferPayload(t))
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	return ListStockTransfersResponse{StockTransfers: payloads, HasMore: hasMore, Page: page, PageSize: pageSize}, nil
}

// --- Create Stock Transfer ---------------------------------------------------

func CreateStockTransferRoutes() []string {
	return []string{"/erp/stock-transfers"}
}

type CreateStockTransferHandler struct {
	Service *application.Service
}

type CreateStockTransferItemInput struct {
	ProductID int64   `json:"product_id"`
	Quantity  float64 `json:"quantity"`
}

type CreateStockTransferRequest struct {
	FromWarehouseID int64                          `json:"from_warehouse_id"`
	ToWarehouseID   int64                          `json:"to_warehouse_id"`
	Notes           string                         `json:"notes"`
	Items           []CreateStockTransferItemInput `json:"items"`
}

type CreateStockTransferResponse struct {
	TransferID int64  `json:"transfer_id"`
	Message    string `json:"message"`
}

func (h *CreateStockTransferHandler) Validate(_ context.Context, req *CreateStockTransferRequest) *errorModels.ServerError {
	if req.FromWarehouseID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Source warehouse is required",
		}
	}

	if req.ToWarehouseID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Destination warehouse is required",
		}
	}

	return nil
}

func (h *CreateStockTransferHandler) Serve(ctx context.Context, req *CreateStockTransferRequest) (CreateStockTransferResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateStockTransferResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	items := make([]domain.StockTransferItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, domain.StockTransferItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	input := domain.StockTransferInput{
		FromWarehouseID: req.FromWarehouseID,
		ToWarehouseID:   req.ToWarehouseID,
		Notes:           req.Notes,
		Items:           items,
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateStockTransferResponse{}, errorModels.PermissionDeniedServerError()
	}

	id, err := h.Service.CreateStockTransfer(ctx, businessID, input)
	if err != nil {
		return CreateStockTransferResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to create stock transfer",
		}
	}

	return CreateStockTransferResponse{TransferID: id, Message: "Stock transfer created successfully"}, nil
}

// --- Update Stock Transfer Status --------------------------------------------

func UpdateStockTransferStatusRoutes() []string {
	return []string{"/erp/stock-transfers"}
}

type UpdateStockTransferStatusHandler struct {
	Service *application.Service
}

type UpdateStockTransferStatusRequest struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

type UpdateStockTransferStatusResponse struct {
	Message string `json:"message"`
}

func (h *UpdateStockTransferStatusHandler) Validate(_ context.Context, req *UpdateStockTransferStatusRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Transfer ID is required",
		}
	}

	if req.Status == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Status is required",
		}
	}

	validStatuses := map[string]struct{}{
		string(domain.StockTransferStatusDraft):     {},
		string(domain.StockTransferStatusInTransit): {},
		string(domain.StockTransferStatusReceived):  {},
		string(domain.StockTransferStatusCancelled): {},
	}
	if _, ok := validStatuses[req.Status]; !ok {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Invalid status. Must be one of: draft, in_transit, received, cancelled",
		}
	}

	return nil
}

func (h *UpdateStockTransferStatusHandler) Serve(ctx context.Context, req *UpdateStockTransferStatusRequest) (UpdateStockTransferStatusResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateStockTransferStatusResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateStockTransferStatusResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.UpdateStockTransferStatus(ctx, businessID, req.ID, domain.StockTransferStatus(req.Status)); err != nil {
		return UpdateStockTransferStatusResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to update stock transfer status",
		}
	}

	return UpdateStockTransferStatusResponse{Message: "Stock transfer status updated successfully"}, nil
}
