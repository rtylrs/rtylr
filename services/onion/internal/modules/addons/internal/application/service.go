package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/addons/internal/domain"
	"github.com/voxire/onion/internal/modules/addons/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

// ListAll returns all available addons in the catalog.
func (s *Service) ListAll(ctx context.Context) ([]domain.Addon, error) {
	return s.repo.ListAll(ctx)
}

// ListForBusiness returns addons with their enabled/disabled status for a business.
// Merges the catalog with what the business has enabled.
func (s *Service) ListForBusiness(ctx context.Context, businessID int64) ([]domain.BusinessAddon, error) {
	all, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	enabled, err := s.repo.ListForBusiness(ctx, businessID)
	if err != nil {
		return nil, err
	}

	enabledMap := make(map[string]domain.BusinessAddon)
	for _, ba := range enabled {
		enabledMap[ba.AddonID] = ba
	}

	result := make([]domain.BusinessAddon, 0, len(all))
	for _, addon := range all {
		if ba, ok := enabledMap[addon.ID]; ok {
			result = append(result, ba)
		} else {
			result = append(result, domain.BusinessAddon{
				AddonID:     addon.ID,
				Name:        addon.Name,
				Description: addon.Description,
				Icon:        addon.Icon,
				Price:       addon.Price,
				IsFree:      addon.IsFree,
				IsEnabled:   false,
			})
		}
	}

	return result, nil
}

func (s *Service) Enable(ctx context.Context, businessID int64, addonID string) error {
	return s.repo.Enable(ctx, businessID, addonID)
}

func (s *Service) Disable(ctx context.Context, businessID int64, addonID string) error {
	return s.repo.Disable(ctx, businessID, addonID)
}

// IsEnabled checks if a specific addon is enabled for a business.
func (s *Service) IsEnabled(ctx context.Context, businessID int64, addonID string) bool {
	addons, err := s.repo.ListForBusiness(ctx, businessID)
	if err != nil {
		return false
	}
	for _, a := range addons {
		if a.AddonID == addonID && a.IsEnabled {
			return true
		}
	}
	return false
}
