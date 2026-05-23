package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/hr_overtime/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_overtime/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	hrOvertimeScope = "hr_overtime"
	hrOvertimeTTL   = 2 * time.Minute
)

func invalidateHROvertimeScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(hrOvertimeScope, ownerID))
	}
}

type Service struct {
	repo             *infrastructure.Repository
	approvalProvider domain.ApprovalChainProvider
}

func NewService(_ context.Context, db *sql.DB, approvalProvider domain.ApprovalChainProvider) *Service {
	return &Service{
		repo:             infrastructure.NewRepository(db),
		approvalProvider: approvalProvider,
	}
}

func (s *Service) ListOvertimeRequests(ctx context.Context, ownerID int64, params domain.ListOvertimeRequestsParams) ([]domain.OvertimeRequestRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	key := cacheKeys.Scoped(hrOvertimeScope, ownerID, fmt.Sprintf("reqs|l=%d|o=%d", params.Limit, params.Offset))
	type listResult struct {
		Items []domain.OvertimeRequestRecord
		Total int64
	}
	res, err := cacher.Of[listResult]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (listResult, error) {
			items, total, err := s.repo.ListOvertimeRequests(ctx, ownerID, params)
			return listResult{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrOvertimeTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) CreateOvertimeRequest(ctx context.Context, ownerID int64, req domain.OvertimeRequestRecord) (*domain.OvertimeRequestRecord, error) {
	req.Status = "pending"
	req.CurrentApprovalLevel = 0
	rec, err := s.repo.CreateOvertimeRequest(ctx, ownerID, req)
	if err == nil {
		invalidateHROvertimeScope(ctx, ownerID)
	}
	return rec, err
}

func (s *Service) ApproveOvertime(ctx context.Context, ownerID, requestID, approverID int64, multiplier float64) error {
	req, err := s.repo.GetOvertimeRequest(ctx, ownerID, requestID)
	if err != nil {
		return err
	}

	chain, chainErr := s.approvalProvider.GetApprovalChain(ctx, ownerID, "overtime")
	if chainErr != nil {
		return chainErr
	}
	if chain == nil {
		return domain.ErrApprovalChainNotConfigured
	}

	if req.CurrentApprovalLevel+1 < chain.TotalLevels {
		return s.repo.AdvanceOvertimeApprovalLevel(ctx, ownerID, requestID)
	}

	if multiplier <= 0 {
		cfg, cfgErr := s.repo.GetOvertimeConfig(ctx, ownerID)
		if cfgErr == nil && cfg != nil {
			multiplier = cfg.DefaultMultiplier
		} else {
			multiplier = 1.5
		}
	}

	if err := s.repo.ApproveOvertimeRequest(ctx, ownerID, requestID, approverID, multiplier); err != nil {
		return err
	}
	invalidateHROvertimeScope(ctx, ownerID)
	return nil
}

func (s *Service) RejectOvertime(ctx context.Context, ownerID, requestID int64) error {
	if _, err := s.repo.GetOvertimeRequest(ctx, ownerID, requestID); err != nil {
		return err
	}
	if err := s.repo.RejectOvertimeRequest(ctx, ownerID, requestID); err != nil {
		return err
	}
	invalidateHROvertimeScope(ctx, ownerID)
	return nil
}

func (s *Service) GetOvertimeConfig(ctx context.Context, ownerID int64) (*domain.OvertimeConfigRecord, error) {
	return cacher.Of[*domain.OvertimeConfigRecord]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(hrOvertimeScope, ownerID, "config"),
		func(ctx context.Context) (*domain.OvertimeConfigRecord, error) {
			cfg, err := s.repo.GetOvertimeConfig(ctx, ownerID)
			if err != nil {
				logger.Ctx(ctx).Debug("no overtime config found, auto-detection disabled",
					logger.Any("owner_id", ownerID))
				return nil, domain.ErrOvertimeConfigNotFound
			}
			return cfg, nil
		},
		cacher.WithExpiry(ctx, 10*time.Minute),
	)
}

func (s *Service) UpsertOvertimeConfig(ctx context.Context, ownerID int64, cfg domain.OvertimeConfigRecord) error {
	if err := s.repo.UpsertOvertimeConfig(ctx, ownerID, cfg); err != nil {
		return err
	}
	invalidateHROvertimeScope(ctx, ownerID)
	return nil
}

func (s *Service) DetectAndFlagOvertime(ctx context.Context, ownerID int64, date string) error {
	cfg, err := s.repo.GetOvertimeConfig(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Debug("overtime auto-detect: no config, skipping",
			logger.Any("owner_id", ownerID))
		return nil
	}

	if !cfg.AutoDetectEnabled {
		return nil
	}

	existing, err := s.repo.ListFlaggedForDate(ctx, ownerID, date)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		logger.Ctx(ctx).Debug("overtime auto-detect: already flagged for date, skipping",
			logger.Any("owner_id", ownerID), logger.Any("date", date))
		return nil
	}

	logger.Ctx(ctx).Info("overtime auto-detect: would scan timesheets for date",
		logger.Any("owner_id", ownerID), logger.Any("date", date),
		logger.Any("threshold_minutes", cfg.DailyThresholdMinutes))

	return nil
}

func (s *Service) RunDailyDetection(ctx context.Context) error {
	ownerIDs, err := s.repo.ListAllConfiguredOwnerIDs(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("overtime daily detection: failed to list owner IDs", logger.Error(err))
		return err
	}

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	for _, ownerID := range ownerIDs {
		if detectErr := s.DetectAndFlagOvertime(ctx, ownerID, yesterday); detectErr != nil {
			logger.Ctx(ctx).Error("overtime daily detection: failed for owner",
				logger.Any("owner_id", ownerID), logger.Error(detectErr))
		}
	}

	return nil
}
