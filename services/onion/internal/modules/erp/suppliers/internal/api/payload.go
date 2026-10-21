package api

import (
	"github.com/voxire/onion/internal/modules/erp/suppliers/internal/domain"
	erpsupplierv1 "github.com/voxire/voxie/gen/go/erp/supplier/v1"
)

func newSupplierPayload(s domain.Supplier) *erpsupplierv1.Supplier {
	return &erpsupplierv1.Supplier{
		Id:           s.ID,
		OwnerId:      s.OwnerID,
		Name:         s.Name,
		Email:        s.Email,
		Phone:        s.Phone,
		Address:      s.Address,
		City:         s.City,
		Country:      s.Country,
		TaxId:        s.TaxID,
		PaymentTerms: s.PaymentTerms,
		Notes:        s.Notes,
		IsActive:     s.IsActive,
		CreatedAt:    s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
