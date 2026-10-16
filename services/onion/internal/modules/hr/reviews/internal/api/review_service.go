package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr/reviews/internal/application"
	"github.com/voxire/onion/internal/modules/hr/reviews/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrreviewv1 "github.com/voxire/voxie/gen/go/hr/review/v1"
)

type ReviewAPI struct {
	service *application.Service
}

func NewReviewAPI(service *application.Service) *ReviewAPI {
	return &ReviewAPI{service: service}
}

func (a *ReviewAPI) ListReviews(ctx context.Context, _ *hrreviewv1.ListReviewsRequest) (*hrreviewv1.ListReviewsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	items, err := a.service.List(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list reviews")
	}

	reviews := make([]*hrreviewv1.Review, 0, len(items))
	for _, item := range items {
		reviews = append(reviews, newReview(item))
	}

	return &hrreviewv1.ListReviewsResponse{Reviews: reviews}, nil
}

func (a *ReviewAPI) CreateReview(ctx context.Context, req *hrreviewv1.CreateReviewRequest) (*hrreviewv1.CreateReviewResponse, error) {
	if req.GetStaffId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "staffId is required")
	}
	if req.GetPeriod() == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "period is required")
	}

	status, statusErr := domain.NormalizeStatus(req.GetStatus())
	if statusErr != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "status must be draft, submitted or acknowledged")
	}
	if ratingErr := domain.ValidateRatings(req.GetRating(), req.GetGoalsRating(), req.GetValuesRating(), req.GetSkillsRating()); ratingErr != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "ratings must be between 0 and 5")
	}

	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	id, err := a.service.Create(ctx, businessID, domain.ReviewInput{
		StaffID:      req.GetStaffId(),
		ReviewerID:   req.GetReviewerId(),
		Period:       req.GetPeriod(),
		Rating:       req.GetRating(),
		GoalsRating:  req.GetGoalsRating(),
		ValuesRating: req.GetValuesRating(),
		SkillsRating: req.GetSkillsRating(),
		Notes:        req.GetNotes(),
		Status:       status,
	})
	if errors.Is(err, domain.ErrStaffNotFound) {
		return nil, onekiterrors.New(http.StatusNotFound, "Staff member not found")
	}
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create review")
	}

	return &hrreviewv1.CreateReviewResponse{Id: id, Message: "Review created successfully"}, nil
}

func newReview(review domain.Review) *hrreviewv1.Review {
	return &hrreviewv1.Review{
		Id:           review.ID,
		OwnerId:      review.OwnerID,
		StaffId:      review.StaffID,
		ReviewerId:   review.ReviewerID,
		Period:       review.Period,
		Rating:       review.Rating,
		GoalsRating:  review.GoalsRating,
		ValuesRating: review.ValuesRating,
		SkillsRating: review.SkillsRating,
		Notes:        review.Notes,
		Status:       review.Status,
		CreatedAt:    review.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    review.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
