package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/support/knowledge/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	supportKnowledgeScope = "support_knowledge"
	supportKnowledgeTTL   = 10 * time.Minute
)

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{
		repo: repo,
	}
}

func invalidateSupportKnowledgeScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(supportKnowledgeScope, businessID))
}

type knowledgeListResult struct {
	Items []domain.Article
	Total int
}

func (s *Service) List(ctx context.Context, businessID int64, page, pageSize int) ([]domain.Article, int, error) {
	key := cacheKeys.Scoped(supportKnowledgeScope, businessID, fmt.Sprintf("list|p=%d|ps=%d", page, pageSize))
	res, err := cacher.Of[knowledgeListResult]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (knowledgeListResult, error) {
			items, total, err := s.repo.List(ctx, businessID, page, pageSize)
			return knowledgeListResult{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, supportKnowledgeTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) Get(ctx context.Context, businessID, articleID int64, slug, title string) (domain.Article, error) {
	slug = strings.TrimSpace(slug)
	title = strings.TrimSpace(title)
	key := cacheKeys.Scoped(supportKnowledgeScope, businessID, fmt.Sprintf("get|id=%d|slug=%s|title=%s", articleID, slug, title))
	return cacher.Of[domain.Article]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (domain.Article, error) {
			return s.repo.Get(ctx, businessID, articleID, slug, title)
		},
		cacher.WithExpiry(ctx, supportKnowledgeTTL),
	)
}

func (s *Service) Create(ctx context.Context, input domain.ArticleInput) (int64, error) {
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidateSupportKnowledgeScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.ArticleInput) error {
	if err := s.repo.Update(ctx, businessID, id, input); err != nil {
		return err
	}
	invalidateSupportKnowledgeScope(ctx, businessID)
	return nil
}

func (s *Service) Delete(ctx context.Context, businessID, id int64) error {
	if err := s.repo.Delete(ctx, businessID, id); err != nil {
		return err
	}
	invalidateSupportKnowledgeScope(ctx, businessID)
	return nil
}

type repository interface {
	Create(ctx context.Context, input domain.ArticleInput) (int64, error)
	Delete(ctx context.Context, businessID, id int64) error
	Get(ctx context.Context, businessID, articleID int64, slug, title string) (domain.Article, error)
	List(ctx context.Context, businessID int64, page, pageSize int) ([]domain.Article, int, error)
	Update(ctx context.Context, businessID, id int64, input domain.ArticleInput) error
}
