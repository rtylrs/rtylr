package application

import (
	"context"
	"fmt"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

func (s *Service) SearchSalesContacts(ctx context.Context, businessID int64, query string) ([]domain.SalesContact, error) {
	key := fmt.Sprintf("%s|pos_search=%s|contact=v1", cacheKeys.CustomersList(businessID), query)
	return cacher.Of[[]domain.SalesContact]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) ([]domain.SalesContact, error) {
			return s.repo.SearchSalesContacts(ctx, businessID, query)
		},
		cacher.WithExpiry(ctx, customerCacheTTL))
}
