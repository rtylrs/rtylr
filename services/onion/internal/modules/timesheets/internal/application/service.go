package application

import (
	"context"
	"database/sql"
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
	GetTimesheets(ctx context.Context, businessID int64, weekStart string) ([]domain.TimesheetEntry, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func (s *Service) GetTimesheets(ctx context.Context, businessID int64, weekStart string) ([]domain.TimesheetEntry, error) {
	key := cacheKeys.Scoped(timesheetsScope, businessID, fmt.Sprintf("ws=%s", weekStart))
	return cacher.Of[[]domain.TimesheetEntry]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.TimesheetEntry, error) {
			return s.repo.GetTimesheets(ctx, businessID, weekStart)
		},
		cacher.WithExpiry(ctx, timesheetsTTL),
	)
}
