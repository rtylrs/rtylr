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
		{"branding read", "recruit:jobs:view", func() error {
			_, err := (&GetCustomizationHandler{}).Serve(ctx, &GetCustomizationRequest{})
			return err
		}},
		{"branding write", "recruit:jobs:manage", func() error {
			_, err := (&PutCustomizationHandler{}).Serve(ctx, &PutCustomizationRequest{})
			return err
		}},
		{"sections list", "recruit:jobs:view", func() error {
			_, err := (&ListSectionsHandler{}).Serve(ctx, &ListSectionsRequest{})
			return err
		}},
		{"sections create", "recruit:jobs:manage", func() error {
			_, err := (&CreateSectionHandler{}).Serve(ctx, &CreateSectionRequest{})
			return err
		}},
		{"sections update", "recruit:jobs:manage", func() error {
			_, err := (&UpdateSectionHandler{}).Serve(ctx, &UpdateSectionRequest{})
			return err
		}},
		{"sections delete", "recruit:jobs:manage", func() error {
			_, err := (&DeleteSectionHandler{}).Serve(ctx, &DeleteSectionRequest{})
			return err
		}},
		{"sections reorder", "recruit:jobs:manage", func() error {
			_, err := (&ReorderSectionsHandler{}).Serve(ctx, &ReorderSectionsRequest{})
			return err
		}},
		{"form fields list", "recruit:jobs:view", func() error {
			_, err := (&ListFormFieldsHandler{}).Serve(ctx, &ListFormFieldsRequest{})
			return err
		}},
		{"form fields create", "recruit:jobs:manage", func() error {
			_, err := (&CreateFormFieldHandler{}).Serve(ctx, &CreateFormFieldRequest{})
			return err
		}},
		{"form fields update", "recruit:jobs:manage", func() error {
			_, err := (&UpdateFormFieldHandler{}).Serve(ctx, &UpdateFormFieldRequest{})
			return err
		}},
		{"form fields delete", "recruit:jobs:manage", func() error {
			_, err := (&DeleteFormFieldHandler{}).Serve(ctx, &DeleteFormFieldRequest{})
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
