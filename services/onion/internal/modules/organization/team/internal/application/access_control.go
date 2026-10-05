package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/platform/logger"
)

// RevokeAccessForTerminated deactivates one user's app permissions in one
// business because hr_core has marked that user's employee record terminated
// (#834). Idempotent: already-inactive rows stay untouched. Called from the
// termination write path so access drops immediately, and from the daily
// reconciliation job as the catch-up guarantee.
func (s *Service) RevokeAccessForTerminated(ctx context.Context, businessID, userID int64) error {
	if businessID <= 0 || userID <= 0 {
		return nil
	}
	rows, err := s.repo.RevokeBusinessUserPermissions(ctx, businessID, userID)
	if err != nil {
		return err
	}
	if rows > 0 {
		invalidateUserScope(ctx, userID)
		invalidateBusinessTeamScope(ctx, businessID)
		logger.Ctx(ctx).Info("revoked terminated employee's app access",
			logger.Any("business_id", businessID), logger.Any("user_id", userID),
			logger.Any("rows", rows))
	}
	return nil
}

// RunTerminatedAccessReconciliation sweeps every business and deactivates any
// still-active permission row whose user is a terminated employee of that
// same business. This is the durability half of #834: a crash between the
// status write and the synchronous revoke, or an unwired revoker, is
// corrected within a day.
func (s *Service) RunTerminatedAccessReconciliation(ctx context.Context) error {
	if s.terminatedEmployees == nil {
		return errors.New("terminated employee source is not configured")
	}
	employees, err := s.terminatedEmployees.ListTerminatedEmployees(ctx)
	if err != nil {
		return err
	}
	var rows int64
	for _, employee := range employees {
		if employee.BusinessID <= 0 || employee.UserID <= 0 {
			continue
		}
		changed, revokeErr := s.repo.RevokeBusinessUserPermissions(ctx, employee.BusinessID, employee.UserID)
		if revokeErr != nil {
			return revokeErr
		}
		rows += changed
	}
	logger.Ctx(ctx).Info("terminated-employee access reconciliation complete",
		logger.Any("rows", rows))
	return nil
}
