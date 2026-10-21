package api

import (
	"context"
	"net/http"
	"strings"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

func recruitUserWith(permissions ...string) context.Context {
	return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppRecruit,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
			}},
		},
	})
}

func TestCreateRecruitCareerPublishedRequiresPublishPermission(t *testing.T) {
	ctx := recruitUserWith("recruit:jobs:manage")
	_, err := (&CreateRecruitCareerHandler{}).Serve(ctx, &CreateRecruitCareerRequest{Title: "Eng", Details: "x", IsPublished: true})
	if recruitStatus(err) != http.StatusForbidden {
		t.Fatalf("publishing on create without recruit:jobs:publish returned %v, want 403", err)
	}
}

func TestRecruitSalaryAmountsRejectsInvalidRanges(t *testing.T) {
	if _, _, err := recruitSalaryAmounts(-1, 10); recruitStatus(err) != http.StatusBadRequest {
		t.Errorf("negative salary returned %v, want 400", err)
	}
	if _, _, err := recruitSalaryAmounts(200, 100); recruitStatus(err) != http.StatusBadRequest {
		t.Errorf("min above max returned %v, want 400", err)
	}
	if _, _, err := recruitSalaryAmounts(100, 0); err != nil {
		t.Errorf("open-ended range returned %v, want nil", err)
	}
	if _, _, err := recruitSalaryAmounts(100, 200); err != nil {
		t.Errorf("valid range returned %v, want nil", err)
	}
}

func TestUpdateSubmissionStatusValidateRejectsUnknownStatus(t *testing.T) {
	h := &UpdateSubmissionStatusHandler{}
	if err := h.Validate(context.Background(), &UpdateSubmissionStatusRequest{ID: 1, Status: "pwned"}); recruitStatus(err) != http.StatusBadRequest {
		t.Fatalf("unknown status returned %v, want 400", err)
	}
	for _, status := range []string{"new", "reviewed", "interviewing", "offered", "hired", "rejected"} {
		if err := h.Validate(context.Background(), &UpdateSubmissionStatusRequest{ID: 1, Status: status}); err != nil {
			t.Errorf("status %q rejected: %v", status, err)
		}
	}
}

func TestSubmitApplicationValidateBoundsInput(t *testing.T) {
	h := &SubmitApplicationHandler{}
	valid := func() *SubmitApplicationRequest {
		return &SubmitApplicationRequest{
			Slug: "acme", CareerSlug: "eng", CandidateName: "Jane Doe",
			Email: "jane@example.com", Phone: "+961 1 234 567",
			AnswersJSON: `{"12":"yes"}`,
		}
	}
	if err := h.Validate(context.Background(), valid()); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}
	cases := map[string]func(*SubmitApplicationRequest){
		"name too long":      func(r *SubmitApplicationRequest) { r.CandidateName = strings.Repeat("a", maxCandidateNameLen+1) },
		"bad email":          func(r *SubmitApplicationRequest) { r.Email = "not an email" },
		"display-name email": func(r *SubmitApplicationRequest) { r.Email = "Jane <jane@example.com>" },
		"phone too long":     func(r *SubmitApplicationRequest) { r.Phone = strings.Repeat("1", maxPhoneLen+1) },
		"resume too long": func(r *SubmitApplicationRequest) {
			r.ResumeURL = "https://x.test/" + strings.Repeat("a", maxResumeURLLen)
		},
		"cover too long":     func(r *SubmitApplicationRequest) { r.CoverLetter = strings.Repeat("a", maxCoverLetterLen+1) },
		"answers not json":   func(r *SubmitApplicationRequest) { r.AnswersJSON = "{oops" },
		"answers not object": func(r *SubmitApplicationRequest) { r.AnswersJSON = `["a"]` },
		"answers too large": func(r *SubmitApplicationRequest) {
			r.AnswersJSON = `{"a":"` + strings.Repeat("a", maxAnswersJSONBytes) + `"}`
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			req := valid()
			mutate(req)
			if err := h.Validate(context.Background(), req); recruitStatus(err) != http.StatusBadRequest {
				t.Fatalf("returned %v, want 400", err)
			}
		})
	}
}
