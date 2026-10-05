package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

func recruitStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func TestRecruitCareerOperationsRequireExactPermissionBeforeServiceAccess(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:     authorization.AppRecruit,
				Licensed:  true,
				HasAccess: true,
			}},
		},
	})
	publishCtx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppRecruit,
				Licensed:    true,
				HasAccess:   true,
				Permissions: []string{"recruit:jobs:manage"},
			}},
		},
	})
	cases := []struct {
		name       string
		permission string
		call       func() error
	}{
		{"jobs list", "recruit:jobs:view", func() error {
			_, err := (&ListRecruitCareersHandler{}).Serve(ctx, &ListRecruitCareersRequest{})
			return err
		}},
		{"jobs create", "recruit:jobs:manage", func() error {
			_, err := (&CreateRecruitCareerHandler{}).Serve(ctx, &CreateRecruitCareerRequest{})
			return err
		}},
		{"jobs update", "recruit:jobs:manage", func() error {
			_, err := (&UpdateRecruitCareerHandler{}).Serve(ctx, &UpdateRecruitCareerRequest{})
			return err
		}},
		{"jobs publish", "recruit:jobs:publish", func() error {
			_, err := (&UpdateRecruitCareerHandler{}).Serve(publishCtx, &UpdateRecruitCareerRequest{IsPublished: true})
			return err
		}},
		{"jobs delete", "recruit:jobs:manage", func() error {
			_, err := (&DeleteRecruitCareerHandler{}).Serve(ctx, &DeleteRecruitCareerRequest{})
			return err
		}},
		{"candidate list", "recruit:candidates:view", func() error {
			_, err := (&ListSubmissionsHandler{}).Serve(ctx, &ListSubmissionsRequest{})
			return err
		}},
		{"candidate status", "recruit:candidates:edit", func() error {
			_, err := (&UpdateSubmissionStatusHandler{}).Serve(ctx, &UpdateSubmissionStatusRequest{})
			return err
		}},
		{"resume download", "recruit:candidates:sensitive_view", func() error {
			_, err := (&ResumeDownloadURLHandler{}).Serve(ctx, &ResumeDownloadURLRequest{})
			return err
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); recruitStatus(err) != http.StatusForbidden {
				t.Fatalf("permission %s returned %v, want 403 before service access", tc.permission, err)
			}
		})
	}
}

func TestRecruitPermissionHelperRejectsWrongAppAndMembersWithoutAccess(t *testing.T) {
	wrongApp := &authPkg.AuthUser{BusinessID: 9, UserID: 7, EffectiveAccess: &authorization.EffectiveAccess{
		Apps: []authorization.EffectiveAppAccess{{AppID: authorization.AppHR, HasAccess: true, Permissions: []string{"recruit:jobs:view"}}},
	}}
	if _, err := requireRecruitPermission(authPkg.WithAuthUser(context.Background(), wrongApp), "recruit:jobs:view"); recruitStatus(err) != http.StatusForbidden {
		t.Fatalf("wrong app returned %v, want 403", err)
	}
	noAccess := &authPkg.AuthUser{BusinessID: 9, UserID: 7, EffectiveAccess: &authorization.EffectiveAccess{
		Apps: []authorization.EffectiveAppAccess{{AppID: authorization.AppRecruit, HasAccess: false, Permissions: []string{"recruit:jobs:view"}}},
	}}
	if _, err := requireRecruitPermission(authPkg.WithAuthUser(context.Background(), noAccess), "recruit:jobs:view"); recruitStatus(err) != http.StatusForbidden {
		t.Fatalf("app without access returned %v, want 403", err)
	}
	// A permission assignment cannot replace the app's active license.
	unlicensed := &authPkg.AuthUser{BusinessID: 9, UserID: 7, EffectiveAccess: &authorization.EffectiveAccess{
		Apps: []authorization.EffectiveAppAccess{{AppID: authorization.AppRecruit, Licensed: false, HasAccess: true, Permissions: []string{"recruit:jobs:view"}}},
	}}
	if _, err := requireRecruitPermission(authPkg.WithAuthUser(context.Background(), unlicensed), "recruit:jobs:view"); recruitStatus(err) != http.StatusForbidden {
		t.Fatalf("an unlicensed app returned %v, want 403", err)
	}
}
