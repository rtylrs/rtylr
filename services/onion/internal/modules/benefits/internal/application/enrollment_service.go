package application

import (
	"context"
	"fmt"

	"github.com/voxire/onion/internal/modules/benefits/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	// Enrollments get their own cache scope so enrolling invalidates only
	// enrollment reads, leaving the plans list cached.
	benefitEnrollmentsScope = "benefit_enrollments"

	defaultEnrollmentPageSize int32 = 50
	maxEnrollmentPageSize     int32 = 500
)

// enrollmentPage bundles a page and its unpaginated total so both survive a
// single cache round-trip.
type enrollmentPage struct {
	Items []domain.BenefitEnrollment `json:"items"`
	Total int32                      `json:"total"`
}

// ResolveEmployeeID maps the calling platform user to their employee record.
// Callers must use this rather than trusting any client-supplied employee id.
func (s *Service) ResolveEmployeeID(ctx context.Context, businessID, userID int64) (int64, error) {
	if s.employees == nil {
		return 0, domain.ErrEmployeeDirectoryUnavailable
	}
	return s.employees.GetEmployeeIDByUserID(ctx, businessID, userID)
}

func invalidateBenefitEnrollmentsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(benefitEnrollmentsScope, businessID))
}

// Enroll enrolls one employee in one plan. employeeID is resolved from the auth
// context by the caller and is never taken from the request body.
func (s *Service) Enroll(ctx context.Context, businessID, employeeID, planID int64) (domain.BenefitEnrollment, error) {
	enrollment, err := s.repo.CreateEnrollment(ctx, businessID, employeeID, planID)
	if err != nil {
		return domain.BenefitEnrollment{}, err
	}

	invalidateBenefitEnrollmentsScope(ctx, businessID)
	return enrollment, nil
}

// ListMyEnrollments returns one employee's own enrollments plus the
// unpaginated total.
func (s *Service) ListMyEnrollments(ctx context.Context, businessID, employeeID int64, page, limit int32) ([]domain.BenefitEnrollment, int32, error) {
	limit = normalizeEnrollmentLimit(limit)
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	key := cacheKeys.Scoped(benefitEnrollmentsScope, businessID,
		fmt.Sprintf("employee:%d:page:%d:limit:%d", employeeID, page, limit))

	cached, err := cacher.Of[enrollmentPage]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (enrollmentPage, error) {
			items, total, listErr := s.repo.ListEnrollmentsByEmployee(ctx, businessID, employeeID, limit, offset)
			if listErr != nil {
				return enrollmentPage{}, listErr
			}
			return enrollmentPage{Items: items, Total: total}, nil
		},
		cacher.WithExpiry(ctx, benefitsTTL),
	)
	if err != nil {
		return nil, 0, err
	}

	return cached.Items, cached.Total, nil
}

func normalizeEnrollmentLimit(limit int32) int32 {
	if limit <= 0 {
		return defaultEnrollmentPageSize
	}
	if limit > maxEnrollmentPageSize {
		return maxEnrollmentPageSize
	}
	return limit
}
