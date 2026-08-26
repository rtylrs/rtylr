package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/pos_audit/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_audit/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/cursor"
)

const (
	posAuditScope         = "pos_audit"
	posAuditTTL           = 30 * time.Second
	defaultPageSize int32 = 100
	maxPageSize     int32 = 100
)

type repository interface {
	List(ctx context.Context, filter domain.AuditLogFilter) ([]domain.AuditLog, error)
	Create(ctx context.Context, input domain.AuditLogInput) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidatePosAuditScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(posAuditScope, businessID))
}

func (s *Service) List(ctx context.Context, filter domain.AuditLogFilter) ([]domain.AuditLog, error) {
	page, err := s.ListPage(ctx, filter)
	return page.Logs, err
}

func (s *Service) ListPage(ctx context.Context, filter domain.AuditLogFilter) (domain.AuditLogPage, error) {
	if filter.Limit <= 0 {
		filter.Limit = defaultPageSize
	}
	if filter.Limit > maxPageSize {
		filter.Limit = maxPageSize
	}
	pageSize := filter.Limit

	detail := fmt.Sprintf("list|a=%v|e=%v|et=%v|s=%v|end=%v|limit=%d|cursor=%v",
		strOrEmpty(filter.Action), filter.EmployeeID, strOrEmpty(filter.EntityType), strOrEmpty(filter.StartDate), strOrEmpty(filter.EndDate), filter.Limit, filter.CursorID)
	key := cacheKeys.Scoped(posAuditScope, filter.BusinessID, detail)
	return cacher.Of[domain.AuditLogPage]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.AuditLogPage, error) {
			filter.Limit = pageSize + 1
			logs, err := s.repo.List(ctx, filter)
			if err != nil {
				return domain.AuditLogPage{}, err
			}

			page := domain.AuditLogPage{Logs: logs}
			if int32(len(logs)) > pageSize {
				last := logs[pageSize-1]
				page.Logs = logs[:pageSize]
				page.NextCursor = cursor.EncodeCreatedAtID(cursor.CreatedAtID{CreatedAt: last.CreatedAt, ID: last.ID})
			}

			return page, nil
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
