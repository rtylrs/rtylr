package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/pos_audit/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_audit/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	posAuditScope = "pos_audit"
	posAuditTTL   = 30 * time.Second
)

type repository interface {
	List(ctx context.Context, filter domain.AuditLogFilter) ([]domain.AuditLog, error)
	Create(ctx context.Context, input domain.AuditLogInput) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func invalidatePosAuditScope(ctx context.Context, businessID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(posAuditScope, businessID))
	}
}

func (s *Service) List(ctx context.Context, filter domain.AuditLogFilter) ([]domain.AuditLog, error) {
	detail := fmt.Sprintf("list|a=%v|e=%v|et=%v|s=%v",
		strOrEmpty(filter.Action), filter.EmployeeID, strOrEmpty(filter.EntityType), strOrEmpty(filter.StartDate))
	key := cacheKeys.Scoped(posAuditScope, filter.BusinessID, detail)
	return cacher.Of[[]domain.AuditLog]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.AuditLog, error) {
			return s.repo.List(ctx, filter)
		},
		cacher.WithExpiry(ctx, posAuditTTL),
	)
}

func (s *Service) Create(ctx context.Context, input domain.AuditLogInput) (int64, error) {
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}
	invalidatePosAuditScope(ctx, input.BusinessID)
	return id, nil
}

func strOrEmpty(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
