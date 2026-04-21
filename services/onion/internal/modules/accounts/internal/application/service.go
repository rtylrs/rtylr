package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/accounts/internal/domain"
	"github.com/voxire/onion/internal/modules/accounts/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) List(ctx context.Context, ownerID int64) ([]domain.Account, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *Service) Create(ctx context.Context, ownerID int64, input domain.AccountInput) (int64, error) {
	return s.repo.Create(ctx, ownerID, input)
}

func (s *Service) Update(ctx context.Context, ownerID, id int64, input domain.AccountInput) error {
	return s.repo.Update(ctx, ownerID, id, input)
}

func (s *Service) Delete(ctx context.Context, ownerID, id int64) error {
	return s.repo.Delete(ctx, ownerID, id)
}

// ChartOfAccounts returns the owner's accounts arranged as a tree keyed by
// parent_id. Rows with NULL parent_id (or whose parent is missing/belongs to
// another owner) are returned as top-level nodes, so orphaned entries stay
// visible instead of being silently dropped.
func (s *Service) ChartOfAccounts(ctx context.Context, ownerID int64) ([]*domain.AccountTreeNode, error) {
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

	roots := make([]*domain.AccountTreeNode, 0)
	for _, a := range accounts {
		node := nodes[a.ID]
		if a.ParentID == 0 {
			roots = append(roots, node)
			continue
		}
		parent, ok := nodes[a.ParentID]
		if !ok {
			// Orphaned: parent missing (deleted or cross-owner). Surface as
			// a root so the frontend doesn't silently lose the account.
			roots = append(roots, node)
			continue
		}
		parent.Children = append(parent.Children, node)
	}

	return roots, nil
}
