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
	hrOvertimeScope           = "hr_overtime"
	hrOvertimeTTL             = 2 * time.Minute
	permissionApproveOvertime = "approve_overtime"
)

func invalidateHROvertimeScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(hrOvertimeScope, businessID))
}

type repository interface {
	ListOvertimeRequests(ctx context.Context, businessID int64, params domain.ListOvertimeRequestsParams) ([]domain.OvertimeRequestRecord, int64, error)
	GetOvertimeRequest(ctx context.Context, businessID, requestID int64) (*domain.OvertimeRequestRecord, error)
	CreateOvertimeRequest(ctx context.Context, businessID int64, req domain.OvertimeRequestRecord) (*domain.OvertimeRequestRecord, error)
	ApproveOvertimeRequest(ctx context.Context, businessID, requestID, approverID int64, multiplier float64) error
	AdvanceOvertimeApprovalLevel(ctx context.Context, businessID, requestID int64, currentLevel int32) error
	RejectOvertimeRequest(ctx context.Context, businessID, requestID int64) error
	CreateFlaggedOvertimeRequest(ctx context.Context, businessID, employeeID int64, date string, extraMinutes int32) error
	ListFlaggedForDate(ctx context.Context, businessID int64, date string) ([]domain.OvertimeRequestRecord, error)
	ListOvertimeCandidates(ctx context.Context, businessID int64, date string, thresholdMinutes int32) ([]domain.OvertimeDetectionCandidate, error)
	GetOvertimeConfig(ctx context.Context, businessID int64) (*domain.OvertimeConfigRecord, error)
	UpsertOvertimeConfig(ctx context.Context, businessID int64, cfg domain.OvertimeConfigRecord) error
	ListAllConfiguredOwnerIDs(ctx context.Context, limit, offset int32) ([]int64, error)
}

type Service struct {
	repo             repository
	approvalProvider domain.ApprovalChainProvider
	permissions      domain.PermissionProvider
}

func NewService(_ context.Context, db *sql.DB, approvalProvider domain.ApprovalChainProvider, permissions domain.PermissionProvider) *Service {
	return &Service{
		repo:             infrastructure.NewRepository(db),
		approvalProvider: approvalProvider,
		permissions:      permissions,
	}
}

func (s *Service) ListOvertimeRequests(ctx context.Context, businessID int64, params domain.ListOvertimeRequestsParams) ([]domain.OvertimeRequestRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	key := cacheKeys.Scoped(hrOvertimeScope, businessID, fmt.Sprintf("reqs|l=%d|o=%d", params.Limit, params.Offset))
	type listResult struct {
		Items []domain.OvertimeRequestRecord
		Total int64
	}
	res, err := cacher.Of[listResult]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (listResult, error) {
			items, total, err := s.repo.ListOvertimeRequests(ctx, businessID, params)
			return listResult{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrOvertimeTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) CreateOvertimeRequest(ctx context.Context, businessID int64, req domain.OvertimeRequestRecord) (*domain.OvertimeRequestRecord, error) {
	req.Status = "pending"
	req.CurrentApprovalLevel = 0
	rec, err := s.repo.CreateOvertimeRequest(ctx, businessID, req)
	if err == nil {
		invalidateHROvertimeScope(ctx, businessID)
	}
	return rec, err
}

func (s *Service) ApproveOvertime(ctx context.Context, businessID, requestID, approverID int64, multiplier float64) error {
	req, approverEmployeeID, chain, err := s.authorizedOvertimeDecision(ctx, businessID, requestID, approverID)
	if err != nil {
		return err
	}

	if req.CurrentApprovalLevel+1 < chain.TotalLevels {
		return s.repo.AdvanceOvertimeApprovalLevel(ctx, businessID, requestID, req.CurrentApprovalLevel)
	}

	if multiplier <= 0 {
		cfg, cfgErr := s.repo.GetOvertimeConfig(ctx, businessID)
		if cfgErr != nil {
			return cfgErr
		}
		if cfg == nil || cfg.DefaultMultiplier <= 0 {
			return domain.ErrOvertimeConfigNotFound
		}
		multiplier = cfg.DefaultMultiplier
	}

	if err := s.repo.ApproveOvertimeRequest(ctx, businessID, requestID, approverEmployeeID, multiplier); err != nil {
		return err
	}
	invalidateHROvertimeScope(ctx, businessID)
	return nil
}

func (s *Service) RejectOvertime(ctx context.Context, businessID, requestID, approverID int64) error {
	if _, _, _, err := s.authorizedOvertimeDecision(ctx, businessID, requestID, approverID); err != nil {
		return err
	}
	if err := s.repo.RejectOvertimeRequest(ctx, businessID, requestID); err != nil {
		return err
	}
	invalidateHROvertimeScope(ctx, businessID)
	return nil
}

func (s *Service) GetOvertimeConfig(ctx context.Context, businessID int64) (*domain.OvertimeConfigRecord, error) {
	return cacher.Of[*domain.OvertimeConfigRecord]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(hrOvertimeScope, businessID, "config"),
		func(ctx context.Context) (*domain.OvertimeConfigRecord, error) {
			cfg, err := s.repo.GetOvertimeConfig(ctx, businessID)
			if err != nil {
				logger.Ctx(ctx).Debug("no overtime config found, auto-detection disabled",
					logger.Any("business_id", businessID))
				return nil, domain.ErrOvertimeConfigNotFound
			}
			return cfg, nil
		},
		cacher.WithExpiry(ctx, 10*time.Minute),
	)
}

func (s *Service) UpsertOvertimeConfig(ctx context.Context, businessID int64, cfg domain.OvertimeConfigRecord) error {
	if err := s.repo.UpsertOvertimeConfig(ctx, businessID, cfg); err != nil {
		return err
	}
	invalidateHROvertimeScope(ctx, businessID)
	return nil
}

func (s *Service) DetectAndFlagOvertime(ctx context.Context, businessID int64, date string) error {
	cfg, err := s.repo.GetOvertimeConfig(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Debug("overtime auto-detect: no config, skipping",
			logger.Any("business_id", businessID))
		return nil
	}

	if !cfg.AutoDetectEnabled {
		return nil
	}

	existing, err := s.repo.ListFlaggedForDate(ctx, businessID, date)
	if err != nil {
		return err
	}
	existingByEmployee := make(map[int64]struct{}, len(existing))
	for _, request := range existing {
		existingByEmployee[request.EmployeeID] = struct{}{}
	}

	candidates, err := s.repo.ListOvertimeCandidates(ctx, businessID, date, cfg.DailyThresholdMinutes)
	if err != nil {
		return err
	}
	for _, candidate := range candidates {
		if _, alreadyExists := existingByEmployee[candidate.EmployeeID]; alreadyExists {
			continue
		}
		if err := s.repo.CreateFlaggedOvertimeRequest(ctx, businessID, candidate.EmployeeID, candidate.Date, candidate.ExtraMinutes); err != nil {
			return err
		}
	}
	if len(candidates) > 0 {
		invalidateHROvertimeScope(ctx, businessID)
	}

	return nil
}

func (s *Service) authorizedOvertimeDecision(ctx context.Context, businessID, requestID, approverUserID int64) (*domain.OvertimeRequestRecord, int64, *domain.ApprovalChainInfo, error) {
	if s.permissions == nil || approverUserID <= 0 {
		return nil, 0, nil, domain.ErrOvertimeAccessDenied
	}
	approverEmployeeID, err := s.permissions.GetEmployeeIDByUserID(ctx, businessID, approverUserID)
	if err != nil || approverEmployeeID <= 0 {
		return nil, 0, nil, domain.ErrOvertimeAccessDenied
	}
	allowed, err := s.permissions.HasPermission(ctx, businessID, approverEmployeeID, permissionApproveOvertime)
	if err != nil {
		return nil, 0, nil, err
	}
	if !allowed {
		return nil, 0, nil, domain.ErrOvertimeApproverNotAllowed
	}
	req, err := s.repo.GetOvertimeRequest(ctx, businessID, requestID)
	if err != nil {
		return nil, 0, nil, err
	}
	if req.EmployeeID == approverEmployeeID {
		return nil, 0, nil, domain.ErrOvertimeSelfApproval
	}
	if s.approvalProvider == nil {
		return nil, 0, nil, domain.ErrApprovalChainNotConfigured
	}
	chain, err := s.approvalProvider.GetApprovalChain(ctx, businessID, "overtime")
	if err != nil {
		return nil, 0, nil, err
	}
	if chain == nil || chain.TotalLevels <= 0 || len(chain.Levels) == 0 {
		return nil, 0, nil, domain.ErrApprovalChainNotConfigured
	}
	matched, err := s.approverMatchesChain(ctx, businessID, approverEmployeeID, req.CurrentApprovalLevel, chain.Levels)
	if err != nil {
		return nil, 0, nil, err
	}
	if !matched {
		return nil, 0, nil, domain.ErrOvertimeApproverNotAllowed
	}
	return req, approverEmployeeID, chain, nil
}

func (s *Service) approverMatchesChain(ctx context.Context, businessID, employeeID int64, currentLevel int32, levels []domain.ApprovalLevelInfo) (bool, error) {
	targetLevel := currentLevel + 1
	levelIndex := -1
	for index, level := range levels {
		if level.LevelOrder == targetLevel {
			levelIndex = index
			break
		}
	}
	if levelIndex < 0 && currentLevel >= 0 && int(currentLevel) < len(levels) {
		levelIndex = int(currentLevel)
	}
	if levelIndex < 0 {
		return false, nil
	}
	level := levels[levelIndex]
	if level.ApproverEmployeeID > 0 {
		return level.ApproverEmployeeID == employeeID, nil
	}
	if level.ApproverRoleID <= 0 {
		return false, nil
	}
	return s.permissions.HasRole(ctx, businessID, employeeID, level.ApproverRoleID)
}

func (s *Service) RunDailyDetection(ctx context.Context) error {
	// Paginate internally against ListAllConfiguredOwnerIDs' own row cap so the
	// job still visits every configured business in one run rather than
	// reimplementing paging at the call site. maxPages is a runaway-loop
	// safety bound, not a real product limit.
	const pageSize = int32(100)
	const maxPages = 100

	var businessIDs []int64
	for page := int32(0); page < maxPages; page++ {
		ids, err := s.repo.ListAllConfiguredOwnerIDs(ctx, pageSize, page*pageSize)
		if err != nil {
			logger.Ctx(ctx).Error("overtime daily detection: failed to list owner IDs", logger.Error(err))
			return err
		}
		businessIDs = append(businessIDs, ids...)
		if int32(len(ids)) < pageSize { // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			break
		}
	}

	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")

	for _, businessID := range businessIDs {
		if detectErr := s.DetectAndFlagOvertime(ctx, businessID, yesterday); detectErr != nil {
			logger.Ctx(ctx).Error("overtime daily detection: failed for owner",
				logger.Any("business_id", businessID), logger.Error(detectErr))
		}
	}

	return nil
}
