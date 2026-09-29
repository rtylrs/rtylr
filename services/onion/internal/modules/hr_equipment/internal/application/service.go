package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr_equipment/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/infrastructure"
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
	ErrEquipmentRequestNotPending  = errors.New("equipment request is no longer pending")
	ErrInvalidNeededByDate         = errors.New("needed_by_date must be in YYYY-MM-DD format")
)

type Service struct {
	repo             *infrastructure.Repository
	db               *sql.DB
	approvalProvider domain.ApprovalChainProvider
}

func NewService(_ context.Context, db *sql.DB, approvalProvider domain.ApprovalChainProvider) *Service {
	return &Service{
		repo:             infrastructure.NewRepository(db),
		db:               db,
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
	if s.db == nil {
		return errors.New("equipment repository database is unavailable")
	}
	var exists int
	err := s.db.QueryRowContext(ctx, `
		SELECT 1
		FROM hr_employees
		WHERE id = ? AND business_id = ?
		LIMIT 1`, employeeID, businessID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrEquipmentEmployeeNotFound
	}
	return err
}

func (s *Service) employeeIDForUser(ctx context.Context, businessID, userID int64) (int64, error) {
	if s.db == nil {
		return 0, errors.New("equipment repository database is unavailable")
	}
	var employeeID int64
	err := s.db.QueryRowContext(ctx, `
		SELECT id
		FROM hr_employees
		WHERE business_id = ? AND user_id = ?
		LIMIT 1`, businessID, userID).Scan(&employeeID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrEquipmentEmployeeNotFound
	}
	if err != nil {
		return 0, err
	}
	if employeeID <= 0 {
		return 0, ErrEquipmentEmployeeNotFound
	}
	return employeeID, nil
}

func equipmentUserCanManage(user *authPkg.AuthUser) bool {
	if authPkg.LegacyAuthorizationMode() {
		// Legacy HasAppPermission intentionally degrades to app membership.
		// Do not inherit that broad fallback for this state-changing action.
		return authPkg.IsBusinessAdmin(user) || authPkg.HasAnyRole(user, "manager", "hr_admin", "admin", "owner", "console_admin")
	}
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
	res, err := cacher.Of[listEquipmentResult]().GetOrComputeSingle(ctx, key,
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

	if !s.equipmentApproverMatchesLevel(ctx, businessID, actorEmployeeID, req.CurrentApprovalLevel, chain.Levels) {
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

func (s *Service) equipmentApproverMatchesLevel(ctx context.Context, businessID, employeeID int64, currentLevel int32, levels []domain.ApprovalLevelInfo) bool {
	if len(levels) == 0 {
		return false
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
		return false
	}

	level := levels[levelIndex]
	if level.ApproverEmployeeID > 0 {
		return level.ApproverEmployeeID == employeeID
	}
	if level.ApproverRoleID <= 0 || s.db == nil {
		return false
	}
	var exists int
	err := s.db.QueryRowContext(ctx, `
		SELECT 1
		FROM hr_employee_roles
		WHERE business_id = ? AND employee_id = ? AND role_id = ?
		LIMIT 1`, businessID, employeeID, level.ApproverRoleID).Scan(&exists)
	return err == nil
}

func (s *Service) updateEquipmentRequestStatus(
	ctx context.Context,
	businessID, requestID int64,
	expectedLevel, newLevel int32,
	status string,
	approvedBy int64,
	reason string,
) error {
	if s.db == nil {
		return errors.New("equipment repository database is unavailable")
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE hr_equipment_requests
		SET status = ?, current_approval_level = ?, approved_by = ?, rejection_reason = ?
		WHERE business_id = ? AND id = ? AND status = 'pending' AND current_approval_level = ?`,
		status, newLevel, sql.NullInt64{Int64: approvedBy, Valid: approvedBy > 0}, sql.NullString{String: reason, Valid: reason != ""}, businessID, requestID, expectedLevel)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrEquipmentRequestNotPending
	}
	return nil
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
	res, err := cacher.Of[listAssetsResult]().GetOrComputeSingle(ctx, key,
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
