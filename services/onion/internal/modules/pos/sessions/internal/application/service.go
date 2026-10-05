package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/pos/sessions/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	posSessionsScope = "pos_sessions"
	posSessionsTTL   = 30 * time.Second
)

type repository interface {
	List(ctx context.Context, ownerID int64) ([]domain.Session, error)
	Get(ctx context.Context, ownerID, id int64) (*domain.Session, error)
	Open(ctx context.Context, input domain.OpenSessionInput) (int64, error)
	Close(ctx context.Context, input domain.CloseSessionInput) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func invalidatePosSessionsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(posSessionsScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Session, error) {
	return cacher.Of[[]domain.Session]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(posSessionsScope, businessID, "list|domain=v2"),
		func(ctx context.Context) ([]domain.Session, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, posSessionsTTL),
	)
}

func (s *Service) Get(ctx context.Context, businessID, id int64) (*domain.Session, error) {
	key := cacheKeys.Scoped(posSessionsScope, businessID, fmt.Sprintf("id=%d", id))
	return cacher.Of[*domain.Session]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (*domain.Session, error) {
			return s.repo.Get(ctx, businessID, id)
		},
		cacher.WithExpiry(ctx, posSessionsTTL),
	)
}

func (s *Service) Open(ctx context.Context, input domain.OpenSessionInput) (int64, error) {
	if err := domain.ValidateOpenSessionInput(input); err != nil {
		return 0, err
	}
	id, err := s.repo.Open(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidatePosSessionsScope(ctx, input.BusinessID)
	return id, nil
}

func (s *Service) Close(ctx context.Context, input domain.CloseSessionInput) error {
	if err := domain.ValidateCloseSessionInput(input); err != nil {
		return err
	}
	if err := s.repo.Close(ctx, input); err != nil {
		return err
	}
	invalidatePosSessionsScope(ctx, input.BusinessID)
	return nil
}
