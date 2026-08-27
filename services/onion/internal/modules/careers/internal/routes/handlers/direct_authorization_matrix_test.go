package handlers

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/careers/internal/application"

	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "recruit jobs view", App: authorization.AppRecruit, Permission: "recruit:jobs:view", Call: func(ctx context.Context) (any, error) {
			response, err := (&ListRecruitCareersHandler{CareersService: application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, &ListRecruitCareersRequest{PublishedOnly: "true"})
			return response, err
		}},
		{Name: "recruit jobs manage", App: authorization.AppRecruit, Permission: "recruit:jobs:manage", Call: func(ctx context.Context) (any, error) {
			response, err := (&CreateRecruitCareerHandler{CareersService: application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, &CreateRecruitCareerRequest{Slug: "direct-authorization-create-2026", Title: "Authorization Probe", Details: "Authorization probe details"})
			return response, err
		}},
		{Name: "recruit jobs publish", App: authorization.AppRecruit, Permission: "recruit:jobs:publish", Extra: []string{"recruit:jobs:manage"}, Call: func(ctx context.Context) (any, error) {
			response, err := (&UpdateRecruitCareerHandler{CareersService: application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, &UpdateRecruitCareerRequest{ID: 7001, IsPublished: true})
			return response, err
		}},
		{Name: "recruit candidates view", App: authorization.AppRecruit, Permission: "recruit:candidates:view", Call: func(ctx context.Context) (any, error) {
			response, err := (&ListSubmissionsHandler{CareersService: application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, &ListSubmissionsRequest{})
			return response, err
		}},
		{Name: "recruit candidates edit", App: authorization.AppRecruit, Permission: "recruit:candidates:edit", Call: func(ctx context.Context) (any, error) {
			response, err := (&UpdateSubmissionStatusHandler{CareersService: application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, &UpdateSubmissionStatusRequest{ID: 7002, Status: "reviewed"})
			return response, err
		}},
		{Name: "recruit candidates sensitive", App: authorization.AppRecruit, Permission: "recruit:candidates:sensitive_view", Call: func(ctx context.Context) (any, error) {
			response, err := (&ResumeDownloadURLHandler{CareersService: application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, &ResumeDownloadURLRequest{SubmissionID: 7002})
			return response, err
		}},
		{Name: "recruit candidates export", App: authorization.AppRecruit, Permission: "recruit:candidates:export", Call: func(ctx context.Context) (any, error) {
			response, err := (&ExportSubmissionsHandler{CareersService: application.NewService(ctx, testutil.DirectAuthorizationDB(ctx))}).Serve(ctx, nil)
			return response, err
		}},
	})
}
