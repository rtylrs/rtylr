package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_equipment/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	hrEquipmentScope = "hr_equipment"
	hrEquipmentTTL   = 5 * time.Minute
)

var (
	ErrEquipmentEmployeeNotFound   = errors.New("equipment employee is not in the active business")
	ErrEquipmentAccessDenied       = errors.New("equipment access denied")
	ErrEquipmentApproverNotAllowed = errors.New("equipment approver is not allowed")
	ErrEquipmentSelfApproval       = errors.New("employees cannot approve their own equipment requests")
	ErrEquipmentRequestNotPending  = domain.ErrEquipmentRequestNotPending
	ErrInvalidNeededByDate         = errors.New("needed_by_date must be in YYYY-MM-DD format")
)

type Service struct {
	repo             repository
	employees        EmployeeDirectory
	approvalProvider domain.ApprovalChainProvider
}

type repository interface {
	CreateEquipmentRequest(context.Context, int64, domain.CreateEquipmentRequestParams) (*domain.EquipmentRequestRecord, error)
	GetEquipmentRequest(context.Context, int64, int64) (*domain.EquipmentRequestRecord, error)
	ListEquipmentRequests(context.Context, int64, domain.ListEquipmentRequestsParams) ([]domain.EquipmentRequestRecord, int64, error)
	UpdateEquipmentRequestStatus(context.Context, int64, int64, int32, string, int32, int64, string) error
	CreateAsset(context.Context, int64, domain.CreateAssetParams) (*domain.AssetRecord, error)
	GetAsset(context.Context, int64, int64) (*domain.AssetRecord, error)
	ListAssets(context.Context, int64, domain.ListAssetsParams) ([]domain.AssetRecord, int64, error)
	AssignAssetToEmployee(context.Context, int64, int64, int64) error
	UnassignAsset(context.Context, int64, int64) error
}

// EmployeeDirectory is the HR Core capability this module needs for employee
// membership checks. Equipment owns only equipment state and never reads
// hr_employees directly.
type EmployeeDirectory interface {
	EmployeeBelongsToBusiness(context.Context, int64, int64) (bool, error)
	EmployeeIDByUserID(context.Context, int64, int64) (int64, error)
	EmployeeHasRole(context.Context, int64, int64, int64) (bool, error)
}

func NewService(repo repository, employees EmployeeDirectory, approvalProvider domain.ApprovalChainProvider) *Service {
	return &Service{
		repo:             repo,
		employees:        employees,
		approvalProvider: approvalProvider,
	}
}

func invalidateHREquipmentScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(hrEquipmentScope, businessID))
}

func (s *Service) CreateEquipmentRequest(ctx context.Context, businessID int64, params domain.CreateEquipmentRequestParams) (*domain.EquipmentRequestRecord, error) {
	if _, err := s.authorizeEquipmentSubject(ctx, businessID, params.EmployeeID); err != nil {
		return nil, err
	}

	params.ItemName = strings.TrimSpace(params.ItemName)
	if err := domain.ValidateEquipmentRequest(domain.EquipmentRequestRecord{
		ItemName: params.ItemName,
		Quantity: params.Quantity,
	}); err != nil {
		return nil, err
	}
	if params.NeededByDate != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(params.NeededByDate))
		if err != nil {
			return nil, ErrInvalidNeededByDate
		}
		params.NeededByDate = parsed.Format("2006-01-02")
	}

	rec, err := s.repo.CreateEquipmentRequest(ctx, businessID, params)
	if err != nil {
		return nil, err
	}
	invalidateHREquipmentScope(ctx, businessID)
	return rec, nil
}

// authorizeEquipmentSubject keeps the employee id supplied by the request
// inside the resolved business and only permits an employee to file for
// themselves.  A caller with the existing HR management permission may act for
// another employee; no client-supplied employee id is trusted on its own.
func (s *Service) authorizeEquipmentSubject(ctx context.Context, businessID, requestedEmployeeID int64) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil || device.User.UserID <= 0 || businessID <= 0 {
		return 0, ErrEquipmentAccessDenied
	}
	if requestedEmployeeID <= 0 {
		return 0, ErrEquipmentEmployeeNotFound
	}
	if err := s.employeeBelongsToBusiness(ctx, businessID, requestedEmployeeID); err != nil {
		return 0, err
	}

	actorEmployeeID, actorErr := s.employeeIDForUser(ctx, businessID, device.User.UserID)
	if actorErr == nil && actorEmployeeID == requestedEmployeeID {
		return actorEmployeeID, nil
	}
	if actorErr != nil && !errors.Is(actorErr, ErrEquipmentEmployeeNotFound) {
		return 0, actorErr
	}
	if !equipmentUserCanManage(device.User) {
		return 0, ErrEquipmentAccessDenied
	}
	return 0, nil
}

func (s *Service) employeeBelongsToBusiness(ctx context.Context, businessID, employeeID int64) error {
	if s.employees == nil {
		return errors.New("HR employee directory is unavailable")
	}
	exists, err := s.employees.EmployeeBelongsToBusiness(ctx, businessID, employeeID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrEquipmentEmployeeNotFound
	}
	return nil
}

func (s *Service) employeeIDForUser(ctx context.Context, businessID, userID int64) (int64, error) {
	if s.employees == nil {
		return 0, errors.New("HR employee directory is unavailable")
	}
	employeeID, err := s.employees.EmployeeIDByUserID(ctx, businessID, userID)
	if err != nil {
		return 0, err
	}
	if employeeID <= 0 {
		return 0, ErrEquipmentEmployeeNotFound
	}
	return employeeID, nil
}

func equipmentUserCanManage(user *authPkg.AuthUser) bool {
	if authPkg.IsBusinessAdmin(user) {
		return true
	}
	if !authPkg.HasAnyAppScope(user, authorization.AppHR) {
		return false
	}
	return authPkg.HasAppPermission(user, authorization.AppHR, "hr:employees:manage") ||
		authPkg.HasAppPermission(user, authorization.AppHR, "hr:time_off:approve")
}

type listEquipmentResult struct {
	Records []domain.EquipmentRequestRecord
	Total   int64
}

type listAssetsResult struct {
	Records []domain.AssetRecord
	Total   int64
}

func (s *Service) ListEquipmentRequests(ctx context.Context, businessID int64, params domain.ListEquipmentRequestsParams) ([]domain.EquipmentRequestRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	key := cacheKeys.Scoped(hrEquipmentScope, businessID, fmt.Sprintf("reqs|e=%d|s=%s|l=%d|o=%d",
		params.EmployeeID, params.Status, params.Limit, params.Offset))
	res, err := cacher.Of[listEquipmentResult]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (listEquipmentResult, error) {
			recs, total, err := s.repo.ListEquipmentRequests(ctx, businessID, params)
			return listEquipmentResult{Records: recs, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrEquipmentTTL),
	)
	return res.Records, res.Total, err
}

func (s *Service) ApproveEquipmentRequest(ctx context.Context, businessID, requestID, approverID int64, approve bool, reason string) error {
	req, err := s.repo.GetEquipmentRequest(ctx, businessID, requestID)
	if err != nil {
		return err
	}
	if req.Status != domain.EquipmentStatusPending {
		return ErrEquipmentRequestNotPending
	}

	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil || device.User.UserID <= 0 || device.User.UserID != approverID {
		return ErrEquipmentAccessDenied
	}
	actorEmployeeID, err := s.employeeIDForUser(ctx, businessID, approverID)
	if err != nil {
		return ErrEquipmentAccessDenied
	}
	if actorEmployeeID == req.EmployeeID {
		return ErrEquipmentSelfApproval
	}
	if !equipmentUserCanManage(device.User) {
		return ErrEquipmentApproverNotAllowed
	}
	if s.approvalProvider == nil {
		return domain.ErrApprovalChainNotConfigured
	}

	chain, err := s.approvalProvider.GetApprovalChain(ctx, businessID, "equipment")
	if err != nil {
		return err
	}
	if chain == nil {
		newLevel := req.CurrentApprovalLevel
		status := "approved"
		if !approve {
			status = "rejected"
		}
		if err := s.updateEquipmentRequestStatus(ctx, businessID, requestID, req.CurrentApprovalLevel, newLevel, status, actorEmployeeID, reason); err != nil {
			return err
		}
		invalidateHREquipmentScope(ctx, businessID)
		return nil
	}

	approverMatches, err := s.equipmentApproverMatchesLevel(ctx, businessID, actorEmployeeID, req.CurrentApprovalLevel, chain.Levels)
	if err != nil {
		return err
	}
	if !approverMatches {
		return ErrEquipmentApproverNotAllowed
	}
	if req.CurrentApprovalLevel == int32(^uint32(0)>>1) {
		return ErrEquipmentRequestNotPending
	}
	if !approve {
		if err := s.updateEquipmentRequestStatus(ctx, businessID, requestID, req.CurrentApprovalLevel, req.CurrentApprovalLevel, "rejected", actorEmployeeID, reason); err != nil {
			return err
		}
		invalidateHREquipmentScope(ctx, businessID)
		return nil
	}

	nextLevel := req.CurrentApprovalLevel + 1
	totalLevels := chain.TotalLevels
	if totalLevels <= 0 {
		totalLevels = int32(len(chain.Levels))
	}
	if nextLevel < totalLevels {
		if err := s.updateEquipmentRequestStatus(ctx, businessID, requestID, req.CurrentApprovalLevel, nextLevel, "pending", 0, ""); err != nil {
			return err
		}
	} else if err := s.updateEquipmentRequestStatus(ctx, businessID, requestID, req.CurrentApprovalLevel, nextLevel, "approved", actorEmployeeID, ""); err != nil {
		return err
	}
	invalidateHREquipmentScope(ctx, businessID)
	return nil
}

func (s *Service) equipmentApproverMatchesLevel(ctx context.Context, businessID, employeeID int64, currentLevel int32, levels []domain.ApprovalLevelInfo) (bool, error) {
	if len(levels) == 0 {
		return false, nil
	}
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
	if level.ApproverRoleID <= 0 || s.employees == nil {
		return false, nil
	}
	return s.employees.EmployeeHasRole(ctx, businessID, employeeID, level.ApproverRoleID)
}

func (s *Service) updateEquipmentRequestStatus(
	ctx context.Context,
	businessID, requestID int64,
	expectedLevel, newLevel int32,
	status string,
	approvedBy int64,
	reason string,
) error {
	return s.repo.UpdateEquipmentRequestStatus(ctx, businessID, requestID, expectedLevel, status, newLevel, approvedBy, reason)
}

func (s *Service) CreateAsset(ctx context.Context, businessID int64, params domain.CreateAssetParams) (*domain.AssetRecord, error) {
	if err := requireEquipmentManager(ctx); err != nil {
		return nil, err
	}
	params.Condition = strings.TrimSpace(params.Condition)
	if params.Condition == "" {
		params.Condition = domain.AssetConditionGood
	}
	if !domain.IsValidAssetCondition(params.Condition) {
		return nil, domain.ErrInvalidAssetCondition
	}
	rec, err := s.repo.CreateAsset(ctx, businessID, params)
	if err != nil {
		return nil, err
	}
	invalidateHREquipmentScope(ctx, businessID)
	return rec, nil
}

func (s *Service) ListAssets(ctx context.Context, businessID int64, params domain.ListAssetsParams) ([]domain.AssetRecord, int64, error) {
	if params.Condition != nil {
		condition := strings.TrimSpace(*params.Condition)
		if condition == "" {
			params.Condition = nil
		} else {
			if !domain.IsValidAssetCondition(condition) {
				return nil, 0, domain.ErrInvalidAssetCondition
			}
			params.Condition = &condition
		}
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	loc := int64(0)
	if params.LocationID != nil {
		loc = *params.LocationID
	}
	status := ""
	if params.Status != nil {
		status = *params.Status
	}
	condition := ""
	if params.Condition != nil {
		condition = strings.TrimSpace(*params.Condition)
	}
	assignedTo := int64(0)
	if params.AssignedTo != nil {
		assignedTo = *params.AssignedTo
	}
	key := cacheKeys.Scoped(hrEquipmentScope, businessID, fmt.Sprintf("assets|loc=%d|s=%s|cond=%s|a=%d|l=%d|o=%d", loc, status, condition, assignedTo, params.Limit, params.Offset))
	res, err := cacher.Of[listAssetsResult]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (listAssetsResult, error) {
			records, total, err := s.repo.ListAssets(ctx, businessID, params)
			return listAssetsResult{Records: records, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrEquipmentTTL),
	)
	return res.Records, res.Total, err
}

func (s *Service) AssignAsset(ctx context.Context, businessID, assetID, employeeID int64) error {
	if err := requireEquipmentManager(ctx); err != nil {
		return err
	}
	asset, err := s.repo.GetAsset(ctx, businessID, assetID)
	if err != nil {
		return err
	}
	if asset.Status == "assigned" {
		return domain.ErrAssetAlreadyAssigned
	}
	if err := s.repo.AssignAssetToEmployee(ctx, businessID, assetID, employeeID); err != nil {
		return err
	}
	invalidateHREquipmentScope(ctx, businessID)
	return nil
}

func (s *Service) UnassignAsset(ctx context.Context, businessID, assetID int64) error {
	if err := requireEquipmentManager(ctx); err != nil {
		return err
	}
	if _, err := s.repo.GetAsset(ctx, businessID, assetID); err != nil {
		return err
	}
	if err := s.repo.UnassignAsset(ctx, businessID, assetID); err != nil {
		return err
	}
	invalidateHREquipmentScope(ctx, businessID)
	return nil
}

// requireEquipmentManager guards the asset mutators, which change which
// employee holds company property. They previously authenticated the caller but
// applied no permission check, so any member with HR app access could reassign
// or unassign assets across the tenant.
func requireEquipmentManager(ctx context.Context) error {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !equipmentUserCanManage(device.User) {
		return ErrEquipmentAccessDenied
	}
	return nil
}
