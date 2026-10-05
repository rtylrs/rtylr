package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/finance/accounts/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/financeaccounts"
)

const (
	accountsScope = "accounts"
	accountsTTL   = 10 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64, limit, offset int) ([]domain.Account, error)
	Create(ctx context.Context, businessID int64, input domain.AccountInput) (int64, error)
	Update(ctx context.Context, businessID, id int64, input domain.AccountInput) error
	Delete(ctx context.Context, businessID, id int64) error
	ListForTree(ctx context.Context, businessID int64) ([]domain.Account, error)
	GetPayrollAccountMapping(ctx context.Context, businessID int64) (financeaccounts.PayrollAccountMapping, error)
}

func (s *Service) GetPayrollAccountMapping(ctx context.Context, businessID int64) (financeaccounts.PayrollAccountMapping, error) {
	return s.repo.GetPayrollAccountMapping(ctx, businessID)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func invalidateAccountsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(accountsScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID, page, pageSize int64) ([]domain.Account, bool, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}
	limit := int(pageSize) + 1
	offset := int((page - 1) * pageSize)
	items, err := s.repo.List(ctx, businessID, limit, offset)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(items) > int(pageSize)
	if hasMore {
		items = items[:pageSize]
	}
	return items, hasMore, nil
}

func (s *Service) Create(ctx context.Context, businessID int64, input domain.AccountInput) (int64, error) {
	if input.Balance < 0 {
		return 0, domain.ErrNegativeBalance
	}
	id, err := s.repo.Create(ctx, businessID, input)
	if err != nil {
		return 0, err
	}
	invalidateAccountsScope(ctx, businessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.AccountInput) error {
	if input.Balance < 0 {
		return domain.ErrNegativeBalance
	}
	if err := s.repo.Update(ctx, businessID, id, input); err != nil {
		return err
	}
	invalidateAccountsScope(ctx, businessID)
	return nil
}

func (s *Service) Delete(ctx context.Context, businessID, id int64) error {
	if err := s.repo.Delete(ctx, businessID, id); err != nil {
		return err
	}
	invalidateAccountsScope(ctx, businessID)
	return nil
}

func (s *Service) ChartOfAccounts(ctx context.Context, businessID int64) ([]*domain.AccountTreeNode, error) {
	return cacher.Of[[]*domain.AccountTreeNode]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(accountsScope, businessID, "tree")+"|domain=v2",
		func(ctx context.Context) ([]*domain.AccountTreeNode, error) {
			return s.buildChartOfAccounts(ctx, businessID)
		},
		cacher.WithExpiry(ctx, accountsTTL),
	)
}

func (s *Service) buildChartOfAccounts(ctx context.Context, businessID int64) ([]*domain.AccountTreeNode, error) {
	accounts, err := s.repo.ListForTree(ctx, businessID)
	if err != nil {
		return nil, err
	}

	nodes := make(map[int64]*domain.AccountTreeNode, len(accounts))
	for _, a := range accounts {
		nodes[a.ID] = &domain.AccountTreeNode{
			ID:       a.ID,
			Name:     a.Name,
			Code:     a.Code,
			Type:     a.Type,
			Children: []*domain.AccountTreeNode{},
		}
	}

	roots := make([]*domain.AccountTreeNode, 0, len(accounts))
	for _, a := range accounts {
		node := nodes[a.ID]
		if a.ParentID == 0 {
			roots = append(roots, node)
			continue
		}
		parent, ok := nodes[a.ParentID]
		if !ok {
			// Orphaned: parent missing or cross-owner. Surface as root so
			// the frontend doesn't silently lose it.
			roots = append(roots, node)
			continue
		}
		parent.Children = append(parent.Children, node)
	}

	return roots, nil
}
