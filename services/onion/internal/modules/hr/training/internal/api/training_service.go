package api

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/hr/training/internal/application"
	"github.com/voxire/onion/internal/modules/hr/training/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrtrainingv1 "github.com/voxire/voxie/gen/go/hr/training/v1"
)

type TrainingAPI struct {
	service *application.Service
}

func NewTrainingAPI(service *application.Service) *TrainingAPI {
	return &TrainingAPI{service: service}
}

func (a *TrainingAPI) ListTrainingCourses(ctx context.Context, _ *hrtrainingv1.ListTrainingCoursesRequest) (*hrtrainingv1.ListTrainingCoursesResponse, error) {
	businessID, errModel := resolveBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	items, err := a.service.List(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list training courses")
	}

	return &hrtrainingv1.ListTrainingCoursesResponse{Courses: newTrainingCourses(items)}, nil
}

func (a *TrainingAPI) CreateTrainingCourse(ctx context.Context, req *hrtrainingv1.CreateTrainingCourseRequest) (*hrtrainingv1.CreateTrainingCourseResponse, error) {
	if req.GetName() == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "name is required")
	}

	businessID, errModel := resolveBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	id, err := a.service.Create(ctx, businessID, domain.TrainingCourseInput{
		Name:          req.GetName(),
		Description:   req.GetDescription(),
		DurationHours: req.GetDurationHours(),
		IsRequired:    req.GetIsRequired(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create training course")
	}

	return &hrtrainingv1.CreateTrainingCourseResponse{Id: id, Message: "Training course created successfully"}, nil
}

func resolveBusinessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

func newTrainingCourses(items []domain.TrainingCourse) []*hrtrainingv1.TrainingCourse {
	out := make([]*hrtrainingv1.TrainingCourse, 0, len(items))
	for _, item := range items {
		out = append(out, &hrtrainingv1.TrainingCourse{
			Id:            item.ID,
			OwnerId:       item.OwnerID,
			Name:          item.Name,
			Description:   item.Description,
			DurationHours: item.DurationHours,
			IsRequired:    item.IsRequired,
			CreatedAt:     formatTime(item.CreatedAt),
		})
	}
	return out
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
