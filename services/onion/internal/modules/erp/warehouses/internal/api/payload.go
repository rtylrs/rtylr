package api

import (
	"github.com/voxire/onion/internal/modules/erp/warehouses/internal/domain"
	erpwarehousev1 "github.com/voxire/voxie/gen/go/erp/warehouse/v1"
)

func newWarehousePayload(w domain.Warehouse) *erpwarehousev1.Warehouse {
	return &erpwarehousev1.Warehouse{
		Id:          w.ID,
		OwnerId:     w.OwnerID,
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
