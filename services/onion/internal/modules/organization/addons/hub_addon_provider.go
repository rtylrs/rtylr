package addons

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/organization/addons/internal/application"
	"github.com/voxire/onion/internal/modules/organization/addons/internal/domain"
	"github.com/voxire/onion/internal/platform/hubapp"
)

type hubAddonProvider struct {
	service *application.Service
}

func NewHubAddonProvider(service *application.Service) hubapp.AddonProvider {
	return &hubAddonProvider{service: service}
}

func (p *hubAddonProvider) ListForBusiness(ctx context.Context, businessID int64) ([]hubapp.BusinessAddon, error) {
	addons, err := p.service.ListForBusiness(ctx, businessID)
	if err != nil {
		return nil, err
	}

	out := make([]hubapp.BusinessAddon, 0, len(addons))
	for _, a := range addons {
		out = append(out, hubapp.BusinessAddon{
			AddonID:     a.AddonID,
			Name:        a.Name,
			Description: a.Description,
			Icon:        a.Icon,
			Price:       a.Price.Float64(),
			IsFree:      a.IsFree,
			IsEnabled:   a.IsEnabled,
			EnabledAt:   a.EnabledAt,
			DisabledAt:  a.DisabledAt,
		})
	}
	return out, nil
}

func (p *hubAddonProvider) Enable(ctx context.Context, businessID int64, addonID string) error {
	err := p.service.Enable(ctx, businessID, addonID)
	if errors.Is(err, domain.ErrAddonNotFound) {
		return hubapp.ErrAddonNotFound
	}
	if errors.Is(err, domain.ErrAddonNotEntitled) {
		return hubapp.ErrAddonNotEntitled
	}
	return err
}

func (p *hubAddonProvider) Disable(ctx context.Context, businessID int64, addonID string) error {
	return p.service.Disable(ctx, businessID, addonID)
}

func (p *hubAddonProvider) IsEnabled(ctx context.Context, businessID int64, addonID string) bool {
	return p.service.IsEnabled(ctx, businessID, addonID)
}
