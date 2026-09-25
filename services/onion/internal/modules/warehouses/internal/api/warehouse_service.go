package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/warehouses/internal/application"
	"github.com/voxire/onion/internal/modules/warehouses/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	erpwarehousev1 "github.com/voxire/voxie/gen/go/erp/warehouse/v1"
)

type WarehouseAPI struct {
	Service *application.Service
}

func NewWarehouseAPI(service *application.Service) *WarehouseAPI {
	return &WarehouseAPI{Service: service}
}

func (a *WarehouseAPI) ListWarehouses(ctx context.Context, _ *erpwarehousev1.ListWarehousesRequest) (*erpwarehousev1.ListWarehousesResponse, error) {
	ownerID, err := warehouseOwnerID(ctx, "erp:warehouses:manage")
	if err != nil {
		return nil, err
	}
	warehouses, listErr := a.Service.ListWarehouses(ctx, ownerID)
	if listErr != nil {
		return nil, warehouseStatusError(http.StatusInternalServerError, "Failed to list warehouses")
	}
	out := make([]*erpwarehousev1.Warehouse, 0, len(warehouses))
	for _, w := range warehouses {
		out = append(out, newWarehousePayload(w))
	}
	return &erpwarehousev1.ListWarehousesResponse{Warehouses: out}, nil
}

func (a *WarehouseAPI) CreateWarehouse(ctx context.Context, req *erpwarehousev1.CreateWarehouseRequest) (*erpwarehousev1.CreateWarehouseResponse, error) {
	ownerID, err := warehouseOwnerID(ctx, "erp:warehouses:manage")
	if err != nil {
		return nil, err
	}
	id, createErr := a.Service.CreateWarehouse(ctx, ownerID, domain.WarehouseInput{
		Name:        strings.TrimSpace(req.GetName()),
		Address:     req.GetAddress(),
		City:        req.GetCity(),
		Country:     req.GetCountry(),
		ManagerName: req.GetManagerName(),
		Capacity:    req.GetCapacity(),
		IsDefault:   req.GetIsDefault(),
		IsActive:    req.GetIsActive(),
	})
	if createErr != nil {
		return nil, warehouseStatusError(http.StatusInternalServerError, "Failed to create warehouse")
	}
	return &erpwarehousev1.CreateWarehouseResponse{WarehouseId: id, Message: "Warehouse created successfully"}, nil
}

func (a *WarehouseAPI) UpdateWarehouse(ctx context.Context, req *erpwarehousev1.UpdateWarehouseRequest) (*erpwarehousev1.UpdateWarehouseResponse, error) {
	ownerID, err := warehouseOwnerID(ctx, "erp:warehouses:manage")
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil {
		return nil, warehouseStatusError(http.StatusBadRequest, "Warehouse body is required")
	}
	if updateErr := a.Service.UpdateWarehouse(ctx, ownerID, req.GetId(), domain.WarehouseInput{
		Name:        strings.TrimSpace(body.GetName()),
		Address:     body.GetAddress(),
		City:        body.GetCity(),
		Country:     body.GetCountry(),
		ManagerName: body.GetManagerName(),
		Capacity:    body.GetCapacity(),
		IsDefault:   body.GetIsDefault(),
		IsActive:    body.GetIsActive(),
	}); updateErr != nil {
		return nil, warehouseStatusError(http.StatusInternalServerError, "Failed to update warehouse")
	}
	return &erpwarehousev1.UpdateWarehouseResponse{Message: "Warehouse updated successfully"}, nil
}

func (a *WarehouseAPI) DeleteWarehouse(ctx context.Context, req *erpwarehousev1.DeleteWarehouseRequest) (*erpwarehousev1.DeleteWarehouseResponse, error) {
	ownerID, err := warehouseOwnerID(ctx, "erp:warehouses:manage")
	if err != nil {
		return nil, err
	}
	if deleteErr := a.Service.DeleteWarehouse(ctx, ownerID, req.GetId()); deleteErr != nil {
		return nil, warehouseStatusError(http.StatusInternalServerError, "Failed to delete warehouse")
	}
	return &erpwarehousev1.DeleteWarehouseResponse{Message: "Warehouse deleted successfully"}, nil
}

func warehouseOwnerID(ctx context.Context, permission string) (int64, error) {
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

func warehouseStatusError(status int, message string) error {
	return onekiterrors.New(status, message)
}
