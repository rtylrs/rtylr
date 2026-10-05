package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/hr/benefits/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	paginationv1 "github.com/voxire/voxie/gen/go/common/pagination/v1"
	hrbenefitv1 "github.com/voxire/voxie/gen/go/hr/benefit/v1"
)

const (
	defaultEnrollmentLimit int32 = 50
	maxEnrollmentLimit     int32 = 500
)

// EnrollInBenefit enrolls the calling employee in a plan. The employee is
// resolved from the auth context and is never read from the request.
func (a *BenefitAPI) EnrollInBenefit(ctx context.Context, req *hrbenefitv1.EnrollInBenefitRequest) (*hrbenefitv1.EnrollInBenefitResponse, error) {
	planID := req.GetPlanId()
	if planID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "plan_id is required")
	}

	businessID, employeeID, err := a.callerEmployee(ctx)
	if err != nil {
		return nil, err
	}

	enrollment, enrollErr := a.service.Enroll(ctx, businessID, employeeID, planID)
	if enrollErr != nil {
		switch {
		case errors.Is(enrollErr, domain.ErrAlreadyEnrolled):
			return nil, onekiterrors.New(http.StatusConflict, "Already enrolled in this benefit plan")
		case errors.Is(enrollErr, domain.ErrBenefitPlanNotFound):
			return nil, onekiterrors.New(http.StatusNotFound, "Benefit plan not found")
		default:
			logger.Ctx(ctx).Error("failed to enroll in benefit plan",
				logger.Any("business_id", businessID),
				logger.Any("employee_id", employeeID),
				logger.Any("plan_id", planID),
				logger.Error(enrollErr))
			return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to enroll in benefit plan")
		}
	}

	return &hrbenefitv1.EnrollInBenefitResponse{
		Enrollment: newBenefitEnrollment(enrollment),
	}, nil
}

// ListMyBenefitEnrollments returns the calling employee's own enrollments.
func (a *BenefitAPI) ListMyBenefitEnrollments(ctx context.Context, req *hrbenefitv1.ListMyBenefitEnrollmentsRequest) (*hrbenefitv1.ListMyBenefitEnrollmentsResponse, error) {
	businessID, employeeID, err := a.callerEmployee(ctx)
	if err != nil {
		return nil, err
	}

	limit := req.GetLimit()
	if limit <= 0 {
		limit = defaultEnrollmentLimit
	}
	if limit > maxEnrollmentLimit {
		limit = maxEnrollmentLimit
	}
	page := req.GetPage()
	if page < 1 {
		page = 1
	}

	items, total, listErr := a.service.ListMyEnrollments(ctx, businessID, employeeID, page, limit)
	if listErr != nil {
		logger.Ctx(ctx).Error("failed to list benefit enrollments",
			logger.Any("business_id", businessID),
			logger.Any("employee_id", employeeID),
			logger.Error(listErr))
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list benefit enrollments")
	}

	enrollments := make([]*hrbenefitv1.BenefitEnrollment, 0, len(items))
	for _, item := range items {
		enrollments = append(enrollments, newBenefitEnrollment(item))
	}

	return &hrbenefitv1.ListMyBenefitEnrollmentsResponse{
		Enrollments: enrollments,
		Pagination:  newPaginationResponse(page, limit, total),
	}, nil
}

// callerEmployee resolves the caller's business scope and employee record,
// failing closed if either is unavailable.
func (a *BenefitAPI) callerEmployee(ctx context.Context) (businessID, employeeID int64, err error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return 0, 0, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return 0, 0, onekiterrors.PermissionDenied()
	}

	employeeID, empErr := a.service.ResolveEmployeeID(ctx, businessID, device.User.UserID)
	if empErr != nil {
		logger.Ctx(ctx).Warn("failed to resolve employee for benefit self-service",
			logger.Any("business_id", businessID),
			logger.Any("user_id", device.User.UserID),
			logger.Error(empErr))
		return 0, 0, onekiterrors.PermissionDenied()
	}

	return businessID, employeeID, nil
}

func newBenefitEnrollment(enrollment domain.BenefitEnrollment) *hrbenefitv1.BenefitEnrollment {
	payload := &hrbenefitv1.BenefitEnrollment{
		Id:         enrollment.ID,
		OwnerId:    enrollment.OwnerID,
		EmployeeId: enrollment.EmployeeID,
		PlanId:     enrollment.PlanID,
		Status:     enrollment.Status,
		EnrolledAt: enrollment.EnrolledAt.Format(time.RFC3339),
	}
	if enrollment.CancelledAt != nil {
		cancelledAt := enrollment.CancelledAt.Format(time.RFC3339)
		payload.CancelledAt = &cancelledAt
	}
	return payload
}

func newPaginationResponse(page, limit, total int32) *paginationv1.PaginationResponse {
	totalPages := int32(0)
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}
	return &paginationv1.PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}
