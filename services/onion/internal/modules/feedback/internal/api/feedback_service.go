package api

import (
	"context"
	"net/http"

	crmfeedbackv1 "github.com/voxire/voxie/gen/go/crm/feedback/v1"

	"github.com/voxire/onion/internal/modules/feedback/internal/application"
	"github.com/voxire/onion/internal/modules/feedback/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type FeedbackAPI struct {
	service *application.Service
}

func NewFeedbackAPI(service *application.Service) *FeedbackAPI {
	return &FeedbackAPI{service: service}
}

func (a *FeedbackAPI) ListFeedback(ctx context.Context, req *crmfeedbackv1.ListFeedbackRequest) (*crmfeedbackv1.ListFeedbackResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	page := req.GetPage()
	if page < 1 {
		page = 1
	}

	pageSize := req.GetPageSize()
	if pageSize < 1 {
		pageSize = 20
	}

	filters := domain.ListFilters{
		Type:   req.GetType(),
		Source: req.GetSource(),
	}

	if req.MinScore != nil {
		filters.MinScore = int(req.GetMinScore())
		filters.MinScoreSet = true
	}

	if req.MaxScore != nil {
		filters.MaxScore = int(req.GetMaxScore())
		filters.MaxScoreSet = true
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	items, total, err := a.service.List(ctx, businessID, filters, int(page), int(pageSize))
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list feedback")
	}

	feedback := make([]*crmfeedbackv1.Feedback, len(items))
	for i, item := range items {
		feedback[i] = &crmfeedbackv1.Feedback{
			Id:           item.ID,
			OwnerId:      item.OwnerID,
			CustomerId:   item.CustomerID,
			CustomerName: item.CustomerName,
			Type:         item.Type,
			Score:        int32(item.Score),
			Comment:      item.Comment,
			Source:       item.Source,
			CreatedAt:    item.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return &crmfeedbackv1.ListFeedbackResponse{
		Feedback: feedback,
		Total:    int32(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (a *FeedbackAPI) CreateFeedback(ctx context.Context, req *crmfeedbackv1.CreateFeedbackRequest) (*crmfeedbackv1.CreateFeedbackResponse, error) {
	if req.GetCustomerId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Customer ID is required")
	}

	if req.GetType() == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Feedback type is required")
	}

	validTypes := map[string]bool{"nps": true, "review": true}
	if !validTypes[req.GetType()] {
		return nil, onekiterrors.New(http.StatusBadRequest, "Feedback type must be 'nps' or 'review'")
	}

	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	id, err := a.service.Create(ctx, domain.FeedbackInput{
		OwnerID:    device.User.UserID,
		BusinessID: businessID,
		CustomerID: req.GetCustomerId(),
		Type:       req.GetType(),
		Score:      int(req.GetScore()),
		Comment:    req.GetComment(),
		Source:     req.GetSource(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create feedback")
	}

	return &crmfeedbackv1.CreateFeedbackResponse{
		FeedbackId: id,
		Message:    "Feedback created successfully",
	}, nil
}
