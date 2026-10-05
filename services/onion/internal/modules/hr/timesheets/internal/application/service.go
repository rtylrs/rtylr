package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/hr/timesheets/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	timesheetsScope = "timesheets"
	timesheetsTTL   = 2 * time.Minute
)

type repository interface {
	GetTimesheets(ctx context.Context, businessID int64, weekStart string) ([]domain.TimesheetEntry, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetTimesheets(ctx context.Context, businessID int64, weekStart string) ([]domain.TimesheetEntry, error) {
	key := cacheKeys.Scoped(timesheetsScope, businessID, fmt.Sprintf("ws=%s", weekStart))
	return cacher.Of[[]domain.TimesheetEntry]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) ([]domain.TimesheetEntry, error) {
			return s.repo.GetTimesheets(ctx, businessID, weekStart)
		},
		cacher.WithExpiry(ctx, timesheetsTTL),
	)
}
