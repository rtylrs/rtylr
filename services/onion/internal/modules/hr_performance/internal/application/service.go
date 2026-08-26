package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	hrPerformanceScope = "hr_performance"
	hrPerformanceTTL   = 5 * time.Minute
)

func invalidateHRPerformanceScope(ctx context.Context, ownerID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(hrPerformanceScope, ownerID))
}

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

// ===== Config =====

func (s *Service) GetPerformanceConfig(ctx context.Context, ownerID int64) (*domain.PerformanceConfigRecord, error) {
	return cacher.Of[*domain.PerformanceConfigRecord]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(hrPerformanceScope, ownerID, "config"),
		func(ctx context.Context) (*domain.PerformanceConfigRecord, error) {
			cfg, err := s.repo.GetPerformanceConfig(ctx, ownerID)
			if errors.Is(err, domain.ErrPerformanceConfigNotFound) {
				logger.Ctx(ctx).Debug("no performance config, using defaults", logger.Any("owner_id", ownerID))
				return domain.DefaultPerformanceConfig(), nil
			}
			if err != nil {
				return nil, err
			}
			return cfg, nil
		},
		cacher.WithExpiry(ctx, hrPerformanceTTL),
	)
}

func (s *Service) UpsertPerformanceConfig(ctx context.Context, ownerID int64, params domain.UpsertConfigParams) (*domain.PerformanceConfigRecord, error) {
	if params.RatingScaleType == "" {
		params.RatingScaleType = "numeric"
	}
	if params.RatingMin == 0 && params.RatingMax == 0 {
		params.RatingMin = 1
		params.RatingMax = 5
	}
	rec, err := s.repo.UpsertPerformanceConfig(ctx, ownerID, params)
	if err == nil {
		invalidateHRPerformanceScope(ctx, ownerID)
	}
	return rec, err
}

// ===== Review Cycles =====

func (s *Service) CreateReviewCycle(ctx context.Context, ownerID int64, params domain.CreateReviewCycleParams) (*domain.ReviewCycleRecord, error) {
	if params.CycleType == "" {
		params.CycleType = "annual"
	}
	rec, err := s.repo.CreateReviewCycle(ctx, ownerID, params)
	if err == nil {
		invalidateHRPerformanceScope(ctx, ownerID)
	}
	return rec, err
}

func (s *Service) GetReviewCycle(ctx context.Context, ownerID, cycleID int64) (*domain.ReviewCycleRecord, error) {
	key := cacheKeys.Scoped(hrPerformanceScope, ownerID, fmt.Sprintf("cycle=%d", cycleID))
	return cacher.Of[*domain.ReviewCycleRecord]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (*domain.ReviewCycleRecord, error) {
			return s.repo.GetReviewCycle(ctx, ownerID, cycleID)
		},
		cacher.WithExpiry(ctx, hrPerformanceTTL),
	)
}

func (s *Service) OpenReviewCycle(ctx context.Context, ownerID, cycleID int64) error {
	cycle, err := s.repo.GetReviewCycle(ctx, ownerID, cycleID)
	if err != nil {
		return err
	}
	if cycle.Status == "open" {
		return domain.ErrCycleAlreadyOpen
	}
	if cycle.Status != "draft" {
		return domain.ErrCycleNotDraft
	}
	if err := s.repo.OpenReviewCycle(ctx, ownerID, cycleID); err != nil {
		return err
	}
	invalidateHRPerformanceScope(ctx, ownerID)
	return nil
}

func (s *Service) ListReviewCycles(ctx context.Context, ownerID int64, params domain.ListCyclesParams) ([]domain.ReviewCycleRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	type lr struct {
		Items []domain.ReviewCycleRecord
		Total int64
	}
	key := cacheKeys.Scoped(hrPerformanceScope, ownerID, fmt.Sprintf("cycles|l=%d|o=%d", params.Limit, params.Offset))
	res, err := cacher.Of[lr]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (lr, error) {
			items, total, err := s.repo.ListReviewCycles(ctx, ownerID, params.Limit, params.Offset)
			return lr{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrPerformanceTTL),
	)
	return res.Items, res.Total, err
}

// ===== Goals =====

func (s *Service) CreateGoal(ctx context.Context, ownerID int64, params domain.CreateGoalParams) (*domain.ReviewGoalRecord, error) {
	rec, err := s.repo.CreateGoal(ctx, ownerID, params)
	if err == nil {
		invalidateHRPerformanceScope(ctx, ownerID)
	}
	return rec, err
}

func (s *Service) UpdateGoal(ctx context.Context, ownerID, goalID int64, params domain.UpdateGoalParams) error {
	_, err := s.repo.GetGoal(ctx, ownerID, goalID)
	if err != nil {
		return err
	}
	if params.Status == "" {
		params.Status = "active"
	}
	if err := s.repo.UpdateGoal(ctx, ownerID, goalID, params); err != nil {
		return err
	}
	invalidateHRPerformanceScope(ctx, ownerID)
	return nil
}

func (s *Service) ListGoals(ctx context.Context, ownerID int64, params domain.ListGoalsParams) ([]domain.ReviewGoalRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	type lr struct {
		Items []domain.ReviewGoalRecord
		Total int64
	}
	key := cacheKeys.Scoped(hrPerformanceScope, ownerID, fmt.Sprintf("goals|e=%d|c=%d|l=%d|o=%d", params.EmployeeID, params.ReviewCycleID, params.Limit, params.Offset))
	res, err := cacher.Of[lr]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (lr, error) {
			items, total, err := s.repo.ListGoals(ctx, ownerID, params.EmployeeID, params.ReviewCycleID, params.Limit, params.Offset)
			return lr{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrPerformanceTTL),
	)
	return res.Items, res.Total, err
}

// ===== Reviews =====

func (s *Service) ListReviews(ctx context.Context, ownerID int64, params domain.ListReviewsParams) ([]domain.PerformanceReviewRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	type lr struct {
		Items []domain.PerformanceReviewRecord
		Total int64
	}
	key := cacheKeys.Scoped(hrPerformanceScope, ownerID, fmt.Sprintf("reviews|c=%d|e=%d|l=%d|o=%d", params.CycleID, params.EmployeeID, params.Limit, params.Offset))
	res, err := cacher.Of[lr]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (lr, error) {
			if params.EmployeeID > 0 {
				items, total, err := s.repo.ListPerformanceReviewsByEmployee(ctx, ownerID, params.CycleID, params.EmployeeID, params.Limit, params.Offset)
				return lr{Items: items, Total: total}, err
			}
			items, total, err := s.repo.ListPerformanceReviews(ctx, ownerID, params.CycleID, params.Limit, params.Offset)
			return lr{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrPerformanceTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) SubmitReview(ctx context.Context, ownerID, reviewID int64, rating, narrative string) error {
	review, err := s.repo.GetPerformanceReview(ctx, ownerID, reviewID)
	if err != nil {
		return err
	}
	if review.Status == "submitted" || review.Status == "completed" || review.Status == "shared" {
		return domain.ErrReviewAlreadySubmitted
	}

	cfg, err := s.GetPerformanceConfig(ctx, ownerID)
	if err != nil {
		return err
	}

	if err := validateRating(cfg, rating); err != nil {
		return err
	}

	if err := s.repo.UpdateReviewSubmission(ctx, ownerID, reviewID, rating, narrative); err != nil {
		return err
	}
	invalidateHRPerformanceScope(ctx, ownerID)
	return nil
}

func (s *Service) CompleteReview(ctx context.Context, ownerID, reviewID int64) error {
	review, err := s.repo.GetPerformanceReview(ctx, ownerID, reviewID)
	if err != nil {
		return err
	}
	if review.Status != "submitted" {
		return domain.ErrReviewNotSubmitted
	}
	if err := s.repo.UpdateReviewStatus(ctx, ownerID, reviewID, "completed"); err != nil {
		return err
	}
	invalidateHRPerformanceScope(ctx, ownerID)
	return nil
}

func (s *Service) ShareReviewWithEmployee(ctx context.Context, ownerID, reviewID int64) error {
	review, err := s.repo.GetPerformanceReview(ctx, ownerID, reviewID)
	if err != nil {
		return err
	}
	if !review.SharedWithEmployeeAt.IsZero() {
		return nil // already shared — idempotent
	}
	if review.Status != "submitted" && review.Status != "completed" {
		return domain.ErrReviewNotSubmitted
	}
	if err := s.repo.ShareReview(ctx, ownerID, reviewID); err != nil {
		return err
	}
	invalidateHRPerformanceScope(ctx, ownerID)
	return nil
}

// ===== Helpers =====

func validateRating(cfg *domain.PerformanceConfigRecord, rating string) error {
	switch cfg.RatingScaleType {
	case "numeric":
		num, err := strconv.Atoi(rating)
		if err != nil {
			return domain.ErrInvalidRating
		}
		if num < int(cfg.RatingMin) || num > int(cfg.RatingMax) {
			return domain.ErrInvalidRating
		}
	case "descriptive":
		if cfg.RatingLabels == "" {
			return domain.ErrInvalidRating
		}
		var labels []string
		if err := json.Unmarshal([]byte(cfg.RatingLabels), &labels); err != nil {
			return domain.ErrInvalidRating
		}
		found := false
		for _, l := range labels {
			if l == rating {
				found = true
				break
			}
		}
		if !found {
			return domain.ErrInvalidRating
		}
	default:
		return domain.ErrInvalidRating
	}
	return nil
}
