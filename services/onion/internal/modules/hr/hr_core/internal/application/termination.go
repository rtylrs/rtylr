package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// TerminationRevoker is the port hr_core calls when an employee's employment
// ends, so hub-side app permissions drop without this module importing the
// entitlement owner (#834). team implements it; the default no-op keeps
// unwired deployments (and tests) working while the daily reconciliation job
// in team remains the durability guarantee.
type TerminationRevoker interface {
	RevokeAccessForTerminated(ctx context.Context, businessID, userID int64) error
}

// SetTerminationRevoker wires the entitlement owner's implementation into the
// termination path. Called once from cmd/api/main.go.
func (s *Service) SetTerminationRevoker(r TerminationRevoker) {
	if r == nil {
		return
	}
	s.terminationRevoker = r
}

// revokeOnTermination fires after a successful employee update whose new
// status is terminated. It resolves user_id from storage when the caller
// updated status alone (the delete-shaped handler passes only ID and status),
// logs rather than fails on revocation errors: the status change is already
// committed and the reconciliation job will catch any miss within a day.
func (s *Service) revokeOnTermination(ctx context.Context, ownerID int64, emp domain.EmployeeRecord) {
	if emp.Status != "terminated" {
		return
	}
	if s.terminationRevoker == nil {
		logger.Ctx(ctx).Warn("employee terminated but no revoker wired; relying on reconciliation job",
			logger.Any("employee_id", emp.ID))
		return
	}
	userID := emp.UserID
	if userID <= 0 {
		rec, err := s.repo.GetEmployee(ctx, ownerID, emp.ID)
		if err != nil || rec == nil || rec.UserID <= 0 {
			logger.Ctx(ctx).Warn("terminated employee has no resolvable user; skipping immediate revoke",
				logger.Any("employee_id", emp.ID), logger.Error(err))
			return
		}
		userID = rec.UserID
	}
	businessID := emp.BusinessID
	if businessID <= 0 {
		businessID = ownerID
	}
	if err := s.terminationRevoker.RevokeAccessForTerminated(ctx, businessID, userID); err != nil {
		logger.Ctx(ctx).Error("immediate post-termination revocation failed; reconciliation will retry",
			logger.Any("business_id", businessID), logger.Any("user_id", userID), logger.Error(err))
	}
}
