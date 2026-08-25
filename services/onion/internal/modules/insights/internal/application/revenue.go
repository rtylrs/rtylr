package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// DefaultLookbackDays is the default `from` offset (in days) applied when a
// caller omits the `from` query parameter. 30 days matches the product's
// default dashboard window.
const DefaultLookbackDays = 30

// ResolveDateRange normalises (from, to) strings coming off the wire. Empty
// strings default to [today-30d, today]. Invalid strings surface an error
// back to the handler so the caller sees a 400, not a silent wrong range.
func ResolveDateRange(fromStr, toStr string) (time.Time, time.Time, error) {
	now := time.Now()
	to := now
	from := now.AddDate(0, 0, -DefaultLookbackDays)
	if toStr != "" {
		t, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid to date: %w", err)
		}
		to = t
	}
	if fromStr != "" {
		t, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid from date: %w", err)
		}
		from = t
	}
	if from.After(to) {
		return time.Time{}, time.Time{}, errors.New("from date is after to date")
	}
	return from, to, nil
}

// GetRevenueSeries returns the daily paid-revenue series for the window.
// The cache key includes the resolved range so different callers with
// different windows don't trample each other's entries.
func (s *Service) GetRevenueSeries(ctx context.Context, businessID int64, from, to time.Time) ([]domain.RevenuePoint, error) {
	key, err := s.insightsCacheKey(ctx, businessID, fmt.Sprintf("revenue:%s:%s", from.Format("2006-01-02"), to.Format("2006-01-02")))
	if err != nil {
		return nil, err
	}
	type result struct {
		Points []domain.RevenuePoint
	}
	cached, err := cacher.Of[result]().GetOrComputeSingle(ctx, key, func(ctx context.Context) (result, error) {
		points, rerr := s.repo.GetRevenueSeries(ctx, businessID, from, to)
		if rerr != nil {
			return result{}, rerr
		}
		return result{Points: points}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return nil, err
	}
	return cached.Points, nil
}
