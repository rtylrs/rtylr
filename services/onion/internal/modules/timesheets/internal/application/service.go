package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/timesheets/internal/domain"
	"github.com/voxire/onion/internal/modules/timesheets/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	timesheetsScope = "timesheets"
	timesheetsTTL   = 2 * time.Minute
)

type repository interface {
	GetTimesheets(ctx context.Context, ownerID int64, weekStart string) ([]domain.TimesheetEntry, error)
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) GetTimesheets(ctx context.Context, ownerID int64, weekStart string) ([]domain.TimesheetEntry, error) {
	key := cacheKeys.Scoped(timesheetsScope, ownerID, fmt.Sprintf("ws=%s", weekStart))
	return cacher.Of[[]domain.TimesheetEntry]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.TimesheetEntry, error) {
			return s.repo.GetTimesheets(ctx, ownerID, weekStart)
		},
		cacher.WithExpiry(ctx, timesheetsTTL),
	)
}
