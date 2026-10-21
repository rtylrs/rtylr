package businessadapters

import (
	"context"

	"github.com/voxire/onion/internal/modules/hr/hr_core"
	organizationbusiness "github.com/voxire/onion/internal/modules/organization/business"
)

type businessHRLocationAdapter struct {
	hr *hr_core.Module
}

func (a businessHRLocationAdapter) ListPublicLocations(ctx context.Context, businessID int64) ([]organizationbusiness.PublicLocation, error) {
	locations, err := a.hr.ListPublicLocations(ctx, businessID)
	if err != nil {
		return nil, err
	}
	result := make([]organizationbusiness.PublicLocation, 0, len(locations))
	for _, location := range locations {
		result = append(result, organizationbusiness.PublicLocation{
			ID:       location.ID,
			Name:     location.Name,
			Slug:     location.Slug,
			Address:  location.Address,
			Timezone: location.Timezone,
			IsActive: location.IsActive,
		})
	}
	return result, nil
}

func NewHRPublicLocationReader(hr *hr_core.Module) func(context.Context, int64) ([]organizationbusiness.PublicLocation, error) {
	return businessHRLocationAdapter{hr: hr}.ListPublicLocations
}
