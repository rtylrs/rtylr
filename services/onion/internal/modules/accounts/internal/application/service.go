package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/accounts/internal/domain"
	"github.com/voxire/onion/internal/modules/accounts/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	accountsScope = "accounts"
	accountsTTL   = 10 * time.Minute
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.Account, error)
	Create(ctx context.Context, ownerID int64, input domain.AccountInput) (int64, error)
	Update(ctx context.Context, ownerID, id int64, input domain.AccountInput) error
	Delete(ctx context.Context, ownerID, id int64) error
	ListForTree(ctx context.Context, ownerID int64) ([]domain.Account, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateAccountsScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(accountsScope, ownerID))
	}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Account, error) {
	return cacher.Of[[]domain.Account]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(accountsScope, ownerID, "list"),
		func(ctx context.Context) ([]domain.Account, error) {
			return s.repo.List(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, accountsTTL),
	)
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.AccountInput) (int64, error) {
	id, err := s.repo.Create(ctx, ownerID, input)
	if err != nil {
		return 0, err
	}
	invalidateAccountsScope(ctx, ownerID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, ownerID, id int64, input domain.AccountInput) error {
	if err := s.repo.Update(ctx, ownerID, id, input); err != nil {
		return err
	}
	invalidateAccountsScope(ctx, ownerID)
	return nil
}

func (s *Service) Delete(ctx context.Context, ownerID, id int64) error {
	if err := s.repo.Delete(ctx, ownerID, id); err != nil {
		return err
	}
	invalidateAccountsScope(ctx, ownerID)
	return nil
}

func (s *Service) ChartOfAccounts(ctx context.Context, ownerID int64) ([]*domain.AccountTreeNode, error) {
	return cacher.Of[[]*domain.AccountTreeNode]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(accountsScope, ownerID, "tree"),
		func(ctx context.Context) ([]*domain.AccountTreeNode, error) {
			return s.buildChartOfAccounts(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, accountsTTL),
	)
}

func (s *Service) buildChartOfAccounts(ctx context.Context, ownerID int64) ([]*domain.AccountTreeNode, error) {
	accounts, err := s.repo.ListForTree(ctx, ownerID)
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
