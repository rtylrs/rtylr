package application

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/logger"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

// ===== Config =====

func (s *Service) GetPerformanceConfig(ctx context.Context, ownerID int64) (*domain.PerformanceConfigRecord, error) {
	cfg, err := s.repo.GetPerformanceConfig(ctx, ownerID)
	if errors.Is(err, domain.ErrPerformanceConfigNotFound) {
		logger.Ctx(ctx).Debug("no performance config, using defaults", logger.Any("owner_id", ownerID))
		return domain.DefaultPerformanceConfig(), nil
	}
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *Service) UpsertPerformanceConfig(ctx context.Context, ownerID int64, params domain.UpsertConfigParams) (*domain.PerformanceConfigRecord, error) {
	if params.RatingScaleType == "" {
		params.RatingScaleType = "numeric"
	}
	if params.RatingMin == 0 && params.RatingMax == 0 {
		params.RatingMin = 1
		params.RatingMax = 5
	}
	return s.repo.UpsertPerformanceConfig(ctx, ownerID, params)
}

// ===== Review Cycles =====

func (s *Service) CreateReviewCycle(ctx context.Context, ownerID int64, params domain.CreateReviewCycleParams) (*domain.ReviewCycleRecord, error) {
	if params.CycleType == "" {
		params.CycleType = "annual"
	}
	return s.repo.CreateReviewCycle(ctx, ownerID, params)
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
	return s.repo.OpenReviewCycle(ctx, ownerID, cycleID)
}

func (s *Service) ListReviewCycles(ctx context.Context, ownerID int64, params domain.ListCyclesParams) ([]domain.ReviewCycleRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	return s.repo.ListReviewCycles(ctx, ownerID, params.Limit, params.Offset)
}

// ===== Goals =====

func (s *Service) CreateGoal(ctx context.Context, ownerID int64, params domain.CreateGoalParams) (*domain.ReviewGoalRecord, error) {
	return s.repo.CreateGoal(ctx, ownerID, params)
}

func (s *Service) UpdateGoal(ctx context.Context, ownerID, goalID int64, params domain.UpdateGoalParams) error {
	_, err := s.repo.GetGoal(ctx, ownerID, goalID)
	if err != nil {
		return err
	}
	if params.Status == "" {
		params.Status = "active"
	}
	return s.repo.UpdateGoal(ctx, ownerID, goalID, params)
}

func (s *Service) ListGoals(ctx context.Context, ownerID int64, params domain.ListGoalsParams) ([]domain.ReviewGoalRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	return s.repo.ListGoals(ctx, ownerID, params.EmployeeID, params.ReviewCycleID, params.Limit, params.Offset)
}

// ===== Reviews =====

func (s *Service) ListReviews(ctx context.Context, ownerID int64, params domain.ListReviewsParams) ([]domain.PerformanceReviewRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.EmployeeID > 0 {
		return s.repo.ListPerformanceReviewsByEmployee(ctx, ownerID, params.CycleID, params.EmployeeID, params.Limit, params.Offset)
	}
	return s.repo.ListPerformanceReviews(ctx, ownerID, params.CycleID, params.Limit, params.Offset)
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

	return s.repo.UpdateReviewSubmission(ctx, ownerID, reviewID, rating, narrative)
}

func (s *Service) CompleteReview(ctx context.Context, ownerID, reviewID int64) error {
	review, err := s.repo.GetPerformanceReview(ctx, ownerID, reviewID)
	if err != nil {
		return err
	}
	if review.Status != "submitted" {
		return domain.ErrReviewNotSubmitted
	}
	return s.repo.UpdateReviewStatus(ctx, ownerID, reviewID, "completed")
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
	return s.repo.ShareReview(ctx, ownerID, reviewID)
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
