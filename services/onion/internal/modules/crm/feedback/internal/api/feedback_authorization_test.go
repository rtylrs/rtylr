package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/feedback/internal/application"
	"github.com/voxire/onion/internal/modules/crm/feedback/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	crmfeedbackv1 "github.com/voxire/voxie/gen/go/crm/feedback/v1"
)

type feedbackAuthorizationProbeRepository struct {
	listCalls                   int
	createCalls                 int
	customerOwnershipCheckCalls int
}

func (r *feedbackAuthorizationProbeRepository) List(context.Context, int64, domain.ListFilters, int, int) ([]domain.Feedback, int, error) {
	r.listCalls++
	return []domain.Feedback{}, 0, nil
}

func (r *feedbackAuthorizationProbeRepository) Create(context.Context, domain.FeedbackInput) (int64, error) {
	r.createCalls++
	return 1, nil
}

func (r *feedbackAuthorizationProbeRepository) CustomerExistsInBusiness(context.Context, int64, int64) bool {
	r.customerOwnershipCheckCalls++
	return true
}

func TestFeedbackAPIRequiresCustomerPermissions(t *testing.T) {
	repo := &feedbackAuthorizationProbeRepository{}
	service := application.NewService(context.Background(), repo)
	api := NewFeedbackAPI(service)
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		UserID:     12,
		BusinessID: 34,
		EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
			AppID:     authorization.AppCRM,
			Licensed:  true,
			HasAccess: true,
		}}},
	})

	if _, err := api.ListFeedback(ctx, &crmfeedbackv1.ListFeedbackRequest{}); err == nil {
		t.Fatal("ListFeedback succeeded without crm:customers:view")
	}
	if _, err := api.CreateFeedback(ctx, &crmfeedbackv1.CreateFeedbackRequest{
		CustomerId: 1,
		Type:       "review",
		Score:      5,
	}); err == nil {
		t.Fatal("CreateFeedback succeeded without crm:customers:edit")
	}
	if repo.listCalls != 0 || repo.createCalls != 0 || repo.customerOwnershipCheckCalls != 0 {
		t.Fatalf("repository calls = list:%d create:%d ownership:%d, want none", repo.listCalls, repo.createCalls, repo.customerOwnershipCheckCalls)
	}
}
