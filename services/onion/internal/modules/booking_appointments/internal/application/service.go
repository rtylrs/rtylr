package application

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/booking_appointments/internal/domain"
	"github.com/voxire/onion/internal/modules/booking_appointments/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

const cacheTTL = 5 * time.Minute

type Service struct{ repo *infrastructure.Repository }

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Appointment, error) {
	cacheKey := "booking_appointments:" + strconv.FormatInt(businessID, 10)
	return cacher.Of[[]domain.Appointment]().GetOrComputeSingle(ctx, cacheKey, func(ctx context.Context) ([]domain.Appointment, error) {
		return s.repo.List(ctx, businessID)
	}, cacher.WithExpiry(ctx, cacheTTL))
}

func (s *Service) ListToday(ctx context.Context, businessID int64) ([]domain.Appointment, error) {
	return s.repo.ListToday(ctx, businessID)
}

func (s *Service) GetByID(ctx context.Context, businessID, id int64) (domain.Appointment, error) {
	return s.repo.GetByID(ctx, businessID, id)
}

func (s *Service) Create(ctx context.Context, businessID int64, input domain.AppointmentInput) (int64, error) {
	id, err := s.repo.Create(ctx, businessID, input)
	if err != nil {
		return 0, err
	}
	s.invalidateCache(ctx, businessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.AppointmentInput) error {
	if err := s.repo.Update(ctx, businessID, id, input); err != nil {
		return err
	}
	s.invalidateCache(ctx, businessID)
	return nil
}

func (s *Service) Delete(ctx context.Context, businessID, id int64) error {
	if err := s.repo.Delete(ctx, businessID, id); err != nil {
		return err
	}
	s.invalidateCache(ctx, businessID)
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	cacher.Of[[]domain.Appointment]().Invalidate(ctx, []string{"booking_appointments:" + strconv.FormatInt(businessID, 10)})
}
