package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/hr_overtime/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_overtime/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/logger"
)

type Service struct {
	repo             *infrastructure.Repository
	approvalProvider domain.ApprovalChainProvider
}

func NewService(ctx context.Context, approvalProvider domain.ApprovalChainProvider) *Service {
	return &Service{
		repo:             infrastructure.NewRepository(ctx),
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
	return s.repo.ListOvertimeRequests(ctx, ownerID, params)
}

func (s *Service) CreateOvertimeRequest(ctx context.Context, ownerID int64, req domain.OvertimeRequestRecord) (*domain.OvertimeRequestRecord, error) {
	req.Status = "pending"
	req.CurrentApprovalLevel = 0
	return s.repo.CreateOvertimeRequest(ctx, ownerID, req)
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

	return s.repo.ApproveOvertimeRequest(ctx, ownerID, requestID, approverID, multiplier)
}

func (s *Service) RejectOvertime(ctx context.Context, ownerID, requestID int64) error {
	_, err := s.repo.GetOvertimeRequest(ctx, ownerID, requestID)
	if err != nil {
		return err
	}
	return s.repo.RejectOvertimeRequest(ctx, ownerID, requestID)
}

func (s *Service) GetOvertimeConfig(ctx context.Context, ownerID int64) (*domain.OvertimeConfigRecord, error) {
	cfg, err := s.repo.GetOvertimeConfig(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Debug("no overtime config found, auto-detection disabled",
			logger.Any("owner_id", ownerID))
		return nil, domain.ErrOvertimeConfigNotFound
	}
	return cfg, nil
}

func (s *Service) UpsertOvertimeConfig(ctx context.Context, ownerID int64, cfg domain.OvertimeConfigRecord) error {
	return s.repo.UpsertOvertimeConfig(ctx, ownerID, cfg)
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
