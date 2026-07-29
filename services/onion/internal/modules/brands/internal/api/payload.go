package api

import (
	"github.com/voxire/onion/internal/modules/brands/internal/domain"
	erpbrandv1 "github.com/voxire/voxie/gen/go/erp/brand/v1"
)

func newBrandPayload(b domain.Brand) *erpbrandv1.Brand {
	return &erpbrandv1.Brand{
		Id:           b.ID,
		OwnerId:      b.OwnerID,
		Name:         b.Name,
		LogoUrl:      b.LogoURL,
		Description:  b.Description,
		ProductCount: b.ProductCount,
		CreatedAt:    b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
