package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_performance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
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
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{
		repo: repo,
	}
}

// ===== Config =====

func (s *Service) GetPerformanceConfig(ctx context.Context, ownerID int64) (*domain.PerformanceConfigRecord, error) {
	return cacher.Of[*domain.PerformanceConfigRecord]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(hrPerformanceScope, ownerID, "config")+"|domain=v2",
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

func (s *Service) CreateReviewCycle(ctx context.Context, businessID int64, params domain.CreateReviewCycleParams) (*domain.ReviewCycleRecord, error) {
	if businessID <= 0 || !performanceUserCanManage(ctx, businessID) {
		return nil, domain.ErrPerformanceAccessDenied
	}
	params.Name = strings.TrimSpace(params.Name)
	if params.CycleType == "" {
		params.CycleType = domain.CycleTypeAnnual
	}
	if err := domain.ValidateReviewCycle(params); err != nil {
		return nil, err
	}
	start, err := time.Parse("2006-01-02", params.PeriodStart)
	if err != nil {
		return nil, fmt.Errorf("invalid period_start: %w", err)
	}
	end, err := time.Parse("2006-01-02", params.PeriodEnd)
	if err != nil {
		return nil, fmt.Errorf("invalid period_end: %w", err)
	}
	if end.Before(start) {
		return nil, errors.New("period_end must not be before period_start")
	}

	rec, err := s.repo.CreateReviewCycle(ctx, businessID, params)
	if err == nil {
		invalidateHRPerformanceScope(ctx, businessID)
	}
	return rec, err
}

func (s *Service) GetReviewCycle(ctx context.Context, ownerID, cycleID int64) (*domain.ReviewCycleRecord, error) {
	key := cacheKeys.Scoped(hrPerformanceScope, ownerID, fmt.Sprintf("cycle=%d", cycleID))
	return cacher.Of[*domain.ReviewCycleRecord]().GetOrComputeSingle(ctx, key+"|domain=v2",
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

func (s *Service) ListReviewCycles(ctx context.Context, businessID int64, params domain.ListCyclesParams) ([]domain.ReviewCycleRecord, int64, error) {
	params.Limit, params.Offset = normalizeListPage(params.Limit, params.Offset)
	type lr struct {
		Items []domain.ReviewCycleRecord
		Total int64
	}
	key := cacheKeys.Scoped(hrPerformanceScope, businessID, fmt.Sprintf("cycles|l=%d|o=%d", params.Limit, params.Offset))
	res, err := cacher.Of[lr]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (lr, error) {
			items, total, err := s.repo.ListReviewCycles(ctx, businessID, params.Limit, params.Offset)
			return lr{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrPerformanceTTL),
	)
	return res.Items, res.Total, err
}

// ===== Goals =====

func (s *Service) CreateGoal(ctx context.Context, businessID int64, params domain.CreateGoalParams) (*domain.ReviewGoalRecord, error) {
	if businessID <= 0 {
		return nil, domain.ErrPerformanceAccessDenied
	}
	if params.EmployeeID <= 0 {
		return nil, domain.ErrEmployeeNotFound
	}
	if params.ReviewCycleID <= 0 {
		return nil, domain.ErrReviewCycleNotFound
	}
	if err := domain.ValidateGoal(params); err != nil {
		return nil, err
	}
	if err := s.authorizeGoalSubject(ctx, businessID, params.EmployeeID); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetReviewCycle(ctx, businessID, params.ReviewCycleID); err != nil {
		return nil, err
	}

	rec, err := s.repo.CreateGoal(ctx, businessID, params)
	if err == nil {
		invalidateHRPerformanceScope(ctx, businessID)
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

func (s *Service) ListGoals(ctx context.Context, businessID int64, params domain.ListGoalsParams) ([]domain.ReviewGoalRecord, int64, error) {
	params.Limit, params.Offset = normalizeListPage(params.Limit, params.Offset)
	if params.EmployeeID < 0 {
		params.EmployeeID = 0
	}
	if params.ReviewCycleID < 0 {
		params.ReviewCycleID = 0
	}
	type lr struct {
		Items []domain.ReviewGoalRecord
		Total int64
	}
	key := cacheKeys.Scoped(hrPerformanceScope, businessID, fmt.Sprintf("goals|e=%d|c=%d|l=%d|o=%d", params.EmployeeID, params.ReviewCycleID, params.Limit, params.Offset))
	res, err := cacher.Of[lr]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (lr, error) {
			items, total, err := s.repo.ListGoals(ctx, businessID, params.EmployeeID, params.ReviewCycleID, params.Limit, params.Offset)
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
	res, err := cacher.Of[lr]().GetOrComputeSingle(ctx, key+"|domain=v2",
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

func normalizeListPage(limit, offset int32) (int32, int32) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func performanceUserCanManage(ctx context.Context, businessID int64) bool {
	device := authPkg.DeviceCtx(ctx)
	user := device.User
	if device.IsAnonymous || user == nil || user.UserID <= 0 || businessID <= 0 {
		return false
	}
	if user.BusinessID > 0 && user.BusinessID != businessID {
		return false
	}
	if user.EffectiveAccess != nil && user.EffectiveAccess.BusinessID > 0 && user.EffectiveAccess.BusinessID != businessID {
		return false
	}
	if !authPkg.HasAnyAppScope(user, authorization.AppHR) {
		return false
	}
	return authPkg.IsBusinessAdmin(user) ||
		authPkg.HasAppPermission(user, authorization.AppHR, "hr:employees:manage")
}

func (s *Service) authorizeGoalSubject(ctx context.Context, businessID, employeeID int64) error {
	belongs, err := s.repo.EmployeeBelongsToBusiness(ctx, businessID, employeeID)
	if err != nil {
		return err
	}
	if !belongs {
		return domain.ErrEmployeeNotFound
	}

	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil || device.User.UserID <= 0 {
		return domain.ErrGoalSubjectNotAllowed
	}
	actorEmployeeID, err := s.repo.GetEmployeeIDByUserID(ctx, businessID, device.User.UserID)
	if err == nil && actorEmployeeID == employeeID {
		return nil
	}
	if err != nil && !errors.Is(err, domain.ErrEmployeeNotFound) {
		return err
	}
	if performanceUserCanManage(ctx, businessID) {
		return nil
	}
	return domain.ErrGoalSubjectNotAllowed
}

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

type repository interface {
	CreateGoal(ctx context.Context, businessID int64, params domain.CreateGoalParams) (*domain.ReviewGoalRecord, error)
	CreateReviewCycle(ctx context.Context, businessID int64, params domain.CreateReviewCycleParams) (*domain.ReviewCycleRecord, error)
	EmployeeBelongsToBusiness(ctx context.Context, businessID, employeeID int64) (bool, error)
	GetEmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error)
	GetGoal(ctx context.Context, businessID, id int64) (*domain.ReviewGoalRecord, error)
	GetPerformanceConfig(ctx context.Context, businessID int64) (*domain.PerformanceConfigRecord, error)
	GetPerformanceReview(ctx context.Context, businessID, id int64) (*domain.PerformanceReviewRecord, error)
	GetReviewCycle(ctx context.Context, businessID, id int64) (*domain.ReviewCycleRecord, error)
	ListGoals(ctx context.Context, businessID, employeeID, cycleID int64, limit, offset int32) ([]domain.ReviewGoalRecord, int64, error)
	ListPerformanceReviews(ctx context.Context, businessID, cycleID int64, limit, offset int32) ([]domain.PerformanceReviewRecord, int64, error)
	ListPerformanceReviewsByEmployee(ctx context.Context, businessID, cycleID, employeeID int64, limit, offset int32) ([]domain.PerformanceReviewRecord, int64, error)
	ListReviewCycles(ctx context.Context, businessID int64, limit, offset int32) ([]domain.ReviewCycleRecord, int64, error)
	OpenReviewCycle(ctx context.Context, businessID, id int64) error
	ShareReview(ctx context.Context, businessID, id int64) error
	UpdateGoal(ctx context.Context, businessID, id int64, params domain.UpdateGoalParams) error
	UpdateReviewStatus(ctx context.Context, businessID, id int64, status string) error
	UpdateReviewSubmission(ctx context.Context, businessID, id int64, rating, narrative string) error
	UpsertPerformanceConfig(ctx context.Context, businessID int64, params domain.UpsertConfigParams) (*domain.PerformanceConfigRecord, error)
}
