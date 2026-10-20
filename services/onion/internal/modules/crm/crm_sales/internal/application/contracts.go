package application

import (
	"context"
	"fmt"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

// ContractFilter selects contracts by signature: nil for all.
type ContractFilter struct {
	Signed *bool
}

func normalizeOffsetPage(limit, offset int32) (int32, int32) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// ListContracts pages contracts, newest first. Finance reads signed ones to
// prefill invoices, so pages are cached until a contract changes.
func (s *Service) ListContracts(ctx context.Context, businessID int64, f ContractFilter, limit, offset int32) (domain.ContractPage, error) {
	limit, offset = normalizeOffsetPage(limit, offset)
	state := "all"
	if f.Signed != nil {
		state = fmt.Sprintf("signed=%t", *f.Signed)
	}
	key := cacheKeys.Scoped(crmSalesScope, businessID, fmt.Sprintf("contracts|%s|n=%d|o=%d", state, limit, offset))
	return cacher.Of[domain.ContractPage]().GetOrComputeSingle(ctx, key+"|domain=v3",
		func(ctx context.Context) (domain.ContractPage, error) {
			var (
				contracts []domain.Contract
				total     int64
				err       error
			)
			switch {
			case f.Signed == nil:
				contracts, total, err = s.repo.ListContracts(ctx, businessID, limit, offset)
			case *f.Signed:
				contracts, total, err = s.repo.ListSignedContracts(ctx, businessID, limit, offset)
			default:
				contracts, total, err = s.repo.ListUnsignedContracts(ctx, businessID, limit, offset)
			}
			if err != nil {
				return domain.ContractPage{}, err
			}
			return domain.ContractPage{Contracts: contracts, Total: total, Limit: limit, Offset: offset}, nil
		},
		cacher.WithExpiry(ctx, crmSalesTTL),
	)
}

func (s *Service) GetContract(ctx context.Context, businessID, id int64) (domain.Contract, error) {
	key := cacheKeys.Scoped(crmSalesScope, businessID, fmt.Sprintf("contract=%d", id))
	return cacher.Of[domain.Contract]().GetOrComputeSingle(ctx, key+"|domain=v3",
		func(ctx context.Context) (domain.Contract, error) {
			return s.repo.GetContract(ctx, businessID, id)
		},
		cacher.WithExpiry(ctx, crmSalesTTL),
	)
}

// CreateContract drafts a contract, usually for a won deal.
func (s *Service) CreateContract(ctx context.Context, businessID, actorID int64, in domain.ContractInput) (domain.Contract, error) {
	if err := in.Normalize(); err != nil {
		return domain.Contract{}, err
	}
	contract, err := s.repo.CreateContract(ctx, businessID, actorID, in)
	if err != nil {
		return domain.Contract{}, err
	}
	s.invalidate(ctx, businessID)
	return contract, nil
}

func (s *Service) UpdateContract(ctx context.Context, businessID, id int64, in domain.ContractInput) (domain.Contract, error) {
	if err := in.Normalize(); err != nil {
		return domain.Contract{}, err
	}
	contract, err := s.repo.UpdateContract(ctx, businessID, id, in)
	if err != nil {
		return domain.Contract{}, err
	}
	s.invalidate(ctx, businessID)
	return contract, nil
}
