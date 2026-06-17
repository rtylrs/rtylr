package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/warehouses/internal/application"
	"github.com/voxire/onion/internal/modules/warehouses/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// --- List Warehouses ---------------------------------------------------------

func ListWarehousesRoutes() []string {
	return []string{"/erp/warehouses"}
}

type ListWarehousesHandler struct {
	Service *application.Service
}

type ListWarehousesRequest struct{}

type WarehousePayload struct {
	ID          int64  `json:"id"`
	OwnerID     int64  `json:"owner_id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	City        string `json:"city"`
	Country     string `json:"country"`
	ManagerName string `json:"manager_name"`
	Capacity    int64  `json:"capacity"`
	IsDefault   bool   `json:"is_default"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ListWarehousesResponse struct {
	Warehouses []WarehousePayload `json:"warehouses"`
}

func newWarehousePayload(w domain.Warehouse) WarehousePayload {
	return WarehousePayload{
		ID:          w.ID,
		OwnerID:     w.OwnerID,
		Name:        w.Name,
		Address:     w.Address,
		City:        w.City,
		Country:     w.Country,
		ManagerName: w.ManagerName,
		Capacity:    w.Capacity,
		IsDefault:   w.IsDefault,
		IsActive:    w.IsActive,
		CreatedAt:   w.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   w.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *ListWarehousesHandler) Serve(ctx context.Context, _ *ListWarehousesRequest) (ListWarehousesResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListWarehousesResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
	warehouses, err := h.Service.ListWarehouses(ctx, ownerID)
	if err != nil {
		return ListWarehousesResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to list warehouses",
		}
	}

	payloads := make([]WarehousePayload, 0, len(warehouses))
	for _, w := range warehouses {
		payloads = append(payloads, newWarehousePayload(w))
	}

	return ListWarehousesResponse{Warehouses: payloads}, nil
}

// --- Create Warehouse --------------------------------------------------------

func CreateWarehouseRoutes() []string {
	return []string{"/erp/warehouses"}
}

type CreateWarehouseHandler struct {
	Service *application.Service
}

type CreateWarehouseRequest struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	City        string `json:"city"`
	Country     string `json:"country"`
	ManagerName string `json:"manager_name"`
	Capacity    int64  `json:"capacity"`
	IsDefault   bool   `json:"is_default"`
	IsActive    bool   `json:"is_active"`
}

type CreateWarehouseResponse struct {
	WarehouseID int64  `json:"warehouse_id"`
	Message     string `json:"message"`
}

func (h *CreateWarehouseHandler) Validate(_ context.Context, req *CreateWarehouseRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Name) == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Warehouse name is required",
		}
	}

	return nil
}

func (h *CreateWarehouseHandler) Serve(ctx context.Context, req *CreateWarehouseRequest) (CreateWarehouseResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateWarehouseResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	input := domain.WarehouseInput{
		Name:        req.Name,
		Address:     req.Address,
		City:        req.City,
		Country:     req.Country,
		ManagerName: req.ManagerName,
		Capacity:    req.Capacity,
		IsDefault:   req.IsDefault,
		IsActive:    req.IsActive,
	}

	ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
	id, err := h.Service.CreateWarehouse(ctx, ownerID, input)
	if err != nil {
		return CreateWarehouseResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to create warehouse",
		}
	}

	return CreateWarehouseResponse{WarehouseID: id, Message: "Warehouse created successfully"}, nil
}

// --- Update Warehouse --------------------------------------------------------

func UpdateWarehouseRoutes() []string {
	return []string{"/erp/warehouses"}
}

type UpdateWarehouseHandler struct {
	Service *application.Service
}

type UpdateWarehouseRequest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	City        string `json:"city"`
	Country     string `json:"country"`
	ManagerName string `json:"manager_name"`
	Capacity    int64  `json:"capacity"`
	IsDefault   bool   `json:"is_default"`
	IsActive    bool   `json:"is_active"`
}

type UpdateWarehouseResponse struct {
	Message string `json:"message"`
}

func (h *UpdateWarehouseHandler) Validate(_ context.Context, req *UpdateWarehouseRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Warehouse ID is required",
		}
	}

	if strings.TrimSpace(req.Name) == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Warehouse name is required",
		}
	}

	return nil
}

func (h *UpdateWarehouseHandler) Serve(ctx context.Context, req *UpdateWarehouseRequest) (UpdateWarehouseResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateWarehouseResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	input := domain.WarehouseInput{
		Name:        req.Name,
		Address:     req.Address,
		City:        req.City,
		Country:     req.Country,
		ManagerName: req.ManagerName,
		Capacity:    req.Capacity,
		IsDefault:   req.IsDefault,
		IsActive:    req.IsActive,
	}

	ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
	if err := h.Service.UpdateWarehouse(ctx, ownerID, req.ID, input); err != nil {
		return UpdateWarehouseResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to update warehouse",
		}
	}

	return UpdateWarehouseResponse{Message: "Warehouse updated successfully"}, nil
}

// --- Delete Warehouse --------------------------------------------------------

func DeleteWarehouseRoutes() []string {
	return []string{"/erp/warehouses"}
}

type DeleteWarehouseHandler struct {
	Service *application.Service
}

type DeleteWarehouseRequest struct {
	ID int64 `json:"id"`
}

type DeleteWarehouseResponse struct {
	Message string `json:"message"`
}

func (h *DeleteWarehouseHandler) Validate(_ context.Context, req *DeleteWarehouseRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Warehouse ID is required",
		}
	}

	return nil
}

func (h *DeleteWarehouseHandler) Serve(ctx context.Context, req *DeleteWarehouseRequest) (DeleteWarehouseResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteWarehouseResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
	if err := h.Service.DeleteWarehouse(ctx, ownerID, req.ID); err != nil {
		return DeleteWarehouseResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to delete warehouse",
		}
	}

	return DeleteWarehouseResponse{Message: "Warehouse deleted successfully"}, nil
}
