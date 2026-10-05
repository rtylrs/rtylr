package api

import (
	"github.com/voxire/onion/internal/modules/erp/categories/internal/domain"
	erpcategoryv1 "github.com/voxire/voxie/gen/go/erp/category/v1"
)

func newCategoryPayload(c domain.Category) *erpcategoryv1.Category {
	return &erpcategoryv1.Category{
		Id:          c.ID,
		OwnerId:     c.OwnerID,
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
