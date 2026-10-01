package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/careers/internal/application"
	"github.com/voxire/onion/internal/modules/careers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/testutil"
)

// The business ids and user ids below are deliberately different values:
// businesses.id and users.id are independent sequences, and several defects
// only show when owner_id != business_id.
const (
	publicCacheTestBusinessID int64 = 993201
	publicCacheTestOwnerID    int64 = 993202

	createCareerTestBusinessID int64 = 993301
	createCareerTestUserID     int64 = 993302

	exportTestBusinessID int64 = 993401
	exportTestUserID     int64 = 993402

	sectionEchoTestBusinessID     int64 = 993501
	sectionEchoTestUserID         int64 = 993502
	sectionPositionTestBusinessID int64 = 993511
	sectionPositionTestUserID     int64 = 993512
	formFieldEchoTestBusinessID   int64 = 993601
	formFieldEchoTestUserID       int64 = 993602

	missingWriteTestBusinessID int64 = 993701
	missingWriteTestUserID     int64 = 993702
	missingWriteTestRowID      int64 = 993799
)

// recruitTestAuthContext builds the server-resolved auth user the recruit
// handlers expect: tenant scope plus the given recruit permissions.
func recruitTestAuthContext(businessID, userID int64, permissions ...string) context.Context {
	user := &authPkg.AuthUser{UserID: userID, BusinessID: businessID}
	user.EffectiveAccess = &authorization.EffectiveAccess{
		BusinessID:           businessID,
		IsOwner:              true,
		IsAdmin:              true,
		AuthorizationVersion: 1,
		Apps: []authorization.EffectiveAppAccess{{
			AppID:       authorization.AppRecruit,
			Licensed:    true,
			HasAccess:   true,
			Permissions: permissions,
		}},
	}
	return authPkg.WithAuthUser(context.Background(), user)
}

// A scoped UPDATE that matches no row is not a successful write. These three
// endpoints used to discard the sql.Result and return 200 for a stale, foreign,
// or unknown id. Keep the regression at the handler boundary so it also proves
// that the repository sentinel is translated through the service to HTTP 404.
func TestRecruitWritesReturnNotFoundWhenNoRowMatches(t *testing.T) {
	db := testutil.OpenDB(t)
	testutil.GrantBusinessAccess(t, db, missingWriteTestBusinessID, missingWriteTestUserID)
	ctx := recruitTestAuthContext(
		missingWriteTestBusinessID,
		missingWriteTestUserID,
		"recruit:jobs:manage",
		"recruit:candidates:edit",
	)
	service := application.NewService(ctx, db)

	cases := []struct {
		name string
		call func() error
	}{
		{
			name: "update career",
			call: func() error {
				_, err := (&UpdateRecruitCareerHandler{CareersService: service}).Serve(ctx, &UpdateRecruitCareerRequest{
					ID:    missingWriteTestRowID,
					Title: "Missing Career",
				})
				return err
			},
		},
		{
			name: "delete career",
			call: func() error {
				_, err := (&DeleteRecruitCareerHandler{CareersService: service}).Serve(ctx, &DeleteRecruitCareerRequest{ID: missingWriteTestRowID})
				return err
			},
		},
		{
			name: "update submission status",
			call: func() error {
				_, err := (&UpdateSubmissionStatusHandler{CareersService: service}).Serve(ctx, &UpdateSubmissionStatusRequest{
					ID:     missingWriteTestRowID,
					Status: "reviewed",
				})
				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if status := recruitStatus(tc.call()); status != http.StatusNotFound {
				t.Fatalf("status = %d, want %d", status, http.StatusNotFound)
			}
		})
	}
}

// The public careers page used to be cached under the customization's
// owner_id while every write handler invalidated the business id. With
// owner_id != business_id the invalidation missed the entry, so the page kept
// serving the pre-write careers list for the whole TTL.
func TestListPublicCareersCacheInvalidatedByBusinessID(t *testing.T) {
	db := testutil.OpenDB(t)
	if err := cacher.InitializeLocalCache(context.Background()); err != nil {
		t.Fatalf("InitializeLocalCache: %v", err)
	}
	testutil.GrantBusinessAccess(t, db, publicCacheTestBusinessID, publicCacheTestOwnerID)

	ctx := context.Background()
	service := application.NewService(ctx, db)
	slug := fmt.Sprintf("public-cache-regression-%d", publicCacheTestBusinessID)
	if err := service.UpsertCustomization(ctx, domain.CustomizationInput{
		OwnerID:    publicCacheTestOwnerID,
		BusinessID: publicCacheTestBusinessID,
		Slug:       slug,
		BrandName:  "Cache Regression Co",
	}); err != nil {
		t.Fatalf("UpsertCustomization: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM careers WHERE business_id = ?`, publicCacheTestBusinessID)
		_, _ = db.ExecContext(ctx, `DELETE FROM careers_customizations WHERE owner_id = ?`, publicCacheTestOwnerID)
	})

	handler := &ListPublicCareersHandler{CareersService: service}

	// First read caches the (empty) page for this business.
	first, err := handler.Serve(ctx, &ListPublicCareersRequest{Slug: slug})
	if err != nil {
		t.Fatalf("first Serve: %v", err)
	}
	if len(first.Careers) != 0 {
		t.Fatalf("fresh page has %d careers, want 0", len(first.Careers))
	}

	// Publish a role; the cached page must still win until invalidation.
	if _, err := service.CreateCareer(ctx, domain.RecruitCareerInput{
		OwnerID:     publicCacheTestOwnerID,
		BusinessID:  publicCacheTestBusinessID,
		Slug:        slug + "-role",
		Title:       "Cache Regression Role",
		Details:     "Exercises public cache invalidation",
		IsPublished: true,
	}); err != nil {
		t.Fatalf("CreateCareer: %v", err)
	}

	second, err := handler.Serve(ctx, &ListPublicCareersRequest{Slug: slug})
	if err != nil {
		t.Fatalf("second Serve: %v", err)
	}
	if len(second.Careers) != 0 {
		t.Fatalf("cached page returned %d careers, want the cached 0", len(second.Careers))
	}

	// This is what every careers/customization/section write handler calls.
	invalidatePublicCareersCache(ctx, publicCacheTestBusinessID)

	third, err := handler.Serve(ctx, &ListPublicCareersRequest{Slug: slug})
	if err != nil {
		t.Fatalf("third Serve: %v", err)
	}
	if len(third.Careers) != 1 {
		t.Fatalf("after invalidation page has %d careers, want 1", len(third.Careers))
	}
}

// POST /recruit/careers with an empty slug must echo the row that was stored:
// server-generated slug, stored publish state, real timestamps. The handler
// used to look the row up with the empty input slug and fall back to a
// zero-value career.
func TestCreateRecruitCareerEchoesGeneratedSlug(t *testing.T) {
	db := testutil.OpenDB(t)
	testutil.GrantBusinessAccess(t, db, createCareerTestBusinessID, createCareerTestUserID)
	ctx := recruitTestAuthContext(createCareerTestBusinessID, createCareerTestUserID, "recruit:jobs:manage")
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM careers WHERE business_id = ?`, createCareerTestBusinessID)
	})

	title := fmt.Sprintf("Backend Engineer %d", createCareerTestUserID)
	handler := &CreateRecruitCareerHandler{CareersService: application.NewService(ctx, db)}
	response, err := handler.Serve(ctx, &CreateRecruitCareerRequest{
		Title:       title,
		Details:     "Own the recruiting API",
		IsPublished: true,
	})
	if err != nil {
		t.Fatalf("Serve: %v", err)
	}
	if response.Career == nil {
		t.Fatal("response career is nil")
	}

	career := response.Career
	if career.ID <= 0 {
		t.Errorf("career id = %d, want a stored id", career.ID)
	}
	if want := fmt.Sprintf("backend-engineer-%d", createCareerTestUserID); career.Slug != want {
		t.Errorf("career slug = %q, want server-generated %q", career.Slug, want)
	}
	if !career.IsPublished {
		t.Error("career is_published = false, want the stored true")
	}
	if !career.IsActive {
		t.Error("career is_active = false, want true")
	}
	if strings.HasPrefix(career.CreatedAt, "0001-01-01") || career.CreatedAt == "" {
		t.Errorf("career created_at = %q, want a real timestamp", career.CreatedAt)
	}
	if strings.HasPrefix(career.UpdatedAt, "0001-01-01") || career.UpdatedAt == "" {
		t.Errorf("career updated_at = %q, want a real timestamp", career.UpdatedAt)
	}
}

// GET /recruit/submissions/export used to reuse the UI list query, which is
// capped at 100 rows, so businesses with more candidates silently lost rows.
func TestExportSubmissionsIncludesRowsBeyondFirstPage(t *testing.T) {
	db := testutil.OpenDB(t)
	testutil.GrantBusinessAccess(t, db, exportTestBusinessID, exportTestUserID)
	ctx := recruitTestAuthContext(exportTestBusinessID, exportTestUserID, "recruit:candidates:export")

	total := int(submissionsExportPageSize) + 5
	placeholders := make([]string, 0, total)
	args := make([]any, 0, total*6)
	for i := 0; i < total; i++ {
		placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?)")
		args = append(args,
			exportTestBusinessID, // career_id: the export's LEFT JOIN tolerates no match
			exportTestUserID,
			exportTestBusinessID,
			fmt.Sprintf("Candidate %d", i),
			fmt.Sprintf("candidate%d@example.test", i),
			"new",
		)
	}
	if _, err := db.ExecContext(ctx,
		"INSERT INTO careers_submissions (career_id, owner_id, business_id, candidate_name, email, status) VALUES "+
			strings.Join(placeholders, ", "), args...); err != nil {
		t.Fatalf("seed submissions: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM careers_submissions WHERE business_id = ?`, exportTestBusinessID)
	})

	handler := &ExportSubmissionsHandler{CareersService: application.NewService(ctx, db)}
	response, err := handler.Serve(ctx, &ExportSubmissionsRequest{})
	if err != nil {
		t.Fatalf("Serve: %v", err)
	}

	// One header line plus `total` data lines.
	if dataRows := strings.Count(strings.TrimSpace(response.CSV), "\n"); dataRows != total {
		t.Fatalf("export has %d data rows, want %d; the export must walk every page", dataRows, total)
	}
	if oldest := fmt.Sprintf("Candidate %d", total-1); !strings.Contains(response.CSV, oldest) {
		t.Errorf("export is missing %q, the oldest seeded submission", oldest)
	}
}

// POST /recruit/sections used to discard the id CreateSection returned and
// re-read the ordered list, answering with its last row. With a client-supplied
// position the created section is not necessarily last, so the response named a
// different section than the one just created.
func TestCreateSectionEchoesCreatedRow(t *testing.T) {
	db := testutil.OpenDB(t)
	testutil.GrantBusinessAccess(t, db, sectionEchoTestBusinessID, sectionEchoTestUserID)
	ctx := recruitTestAuthContext(sectionEchoTestBusinessID, sectionEchoTestUserID, "recruit:jobs:manage")
	service := application.NewService(ctx, db)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM careers_sections WHERE business_id = ?`, sectionEchoTestBusinessID)
	})

	for _, seed := range []struct {
		title    string
		position int32
	}{
		{"First", 0},
		{"Last", 5},
	} {
		if _, err := service.CreateSection(ctx, domain.SectionInput{
			OwnerID:    sectionEchoTestUserID,
			BusinessID: sectionEchoTestBusinessID,
			Kind:       "text",
			Title:      seed.title,
			Position:   seed.position,
			Enabled:    true,
		}); err != nil {
			t.Fatalf("seed section %q: %v", seed.title, err)
		}
	}

	handler := &CreateSectionHandler{CareersService: service}
	response, err := handler.Serve(ctx, &CreateSectionRequest{Kind: "text", Title: "Middle", Position: 1, Enabled: true})
	if err != nil {
		t.Fatalf("Serve: %v", err)
	}
	if response.Section == nil {
		t.Fatal("response section is nil")
	}
	if response.Section.Title != "Middle" {
		t.Errorf("response section title = %q, want the created %q", response.Section.Title, "Middle")
	}
	if response.Section.Position != 1 {
		t.Errorf("response section position = %d, want 1", response.Section.Position)
	}
}

// A section created without a position used to get position len(existing). With
// gaps left by deletes that lands the new row mid-list, so it must append after
// the highest existing position instead.
func TestCreateSectionWithoutPositionAppendsAfterHighest(t *testing.T) {
	db := testutil.OpenDB(t)
	testutil.GrantBusinessAccess(t, db, sectionPositionTestBusinessID, sectionPositionTestUserID)
	ctx := recruitTestAuthContext(sectionPositionTestBusinessID, sectionPositionTestUserID, "recruit:jobs:manage")
	service := application.NewService(ctx, db)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM careers_sections WHERE business_id = ?`, sectionPositionTestBusinessID)
	})

	firstID, err := service.CreateSection(ctx, domain.SectionInput{
		OwnerID:    sectionPositionTestUserID,
		BusinessID: sectionPositionTestBusinessID,
		Kind:       "text",
		Title:      "First",
		Enabled:    true,
	})
	if err != nil {
		t.Fatalf("CreateSection first: %v", err)
	}
	if _, err := service.CreateSection(ctx, domain.SectionInput{
		OwnerID:    sectionPositionTestUserID,
		BusinessID: sectionPositionTestBusinessID,
		Kind:       "text",
		Title:      "Kept",
		Position:   5,
		Enabled:    true,
	}); err != nil {
		t.Fatalf("CreateSection kept: %v", err)
	}
	// Delete the low row so len(existing)=1 but the highest position is 5.
	if err := service.DeleteSection(ctx, sectionPositionTestBusinessID, firstID); err != nil {
		t.Fatalf("DeleteSection: %v", err)
	}

	handler := &CreateSectionHandler{CareersService: service}
	response, err := handler.Serve(ctx, &CreateSectionRequest{Kind: "text", Title: "Appended", Enabled: true})
	if err != nil {
		t.Fatalf("Serve: %v", err)
	}
	if response.Section == nil {
		t.Fatal("response section is nil")
	}
	if response.Section.Title != "Appended" {
		t.Errorf("response section title = %q, want the created %q", response.Section.Title, "Appended")
	}
	if response.Section.Position != 6 {
		t.Errorf("appended section position = %d, want the highest existing position + 1 = 6", response.Section.Position)
	}
}

// POST /recruit/form-fields used to discard the list error and answer with the
// last row of the (position-ordered) list, which need not be the created field.
func TestCreateFormFieldEchoesCreatedRow(t *testing.T) {
	db := testutil.OpenDB(t)
	testutil.GrantBusinessAccess(t, db, formFieldEchoTestBusinessID, formFieldEchoTestUserID)
	ctx := recruitTestAuthContext(formFieldEchoTestBusinessID, formFieldEchoTestUserID, "recruit:jobs:manage")
	service := application.NewService(ctx, db)

	career, err := service.CreateCareer(ctx, domain.RecruitCareerInput{
		OwnerID:     formFieldEchoTestUserID,
		BusinessID:  formFieldEchoTestBusinessID,
		Slug:        fmt.Sprintf("form-field-echo-%d", formFieldEchoTestBusinessID),
		Title:       "Form Field Role",
		IsPublished: true,
	})
	if err != nil {
		t.Fatalf("CreateCareer: %v", err)
	}
	careerID := career.ID
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM careers_form_fields WHERE business_id = ?`, formFieldEchoTestBusinessID)
		_, _ = db.ExecContext(context.Background(), `DELETE FROM careers WHERE business_id = ?`, formFieldEchoTestBusinessID)
	})

	for _, seed := range []struct {
		label    string
		position int32
	}{
		{"First", 0},
		{"Last", 9},
	} {
		if _, err := service.CreateFormField(ctx, domain.FormFieldInput{
			CareerID:   careerID,
			OwnerID:    formFieldEchoTestUserID,
			BusinessID: formFieldEchoTestBusinessID,
			Kind:       "text",
			Label:      seed.label,
			Position:   seed.position,
		}); err != nil {
			t.Fatalf("seed field %q: %v", seed.label, err)
		}
	}

	handler := &CreateFormFieldHandler{CareersService: service}
	response, err := handler.Serve(ctx, &CreateFormFieldRequest{CareerID: careerID, Kind: "text", Label: "Middle", Position: 1})
	if err != nil {
		t.Fatalf("Serve: %v", err)
	}
	if response.Field == nil {
		t.Fatal("response field is nil")
	}
	if response.Field.Label != "Middle" {
		t.Errorf("response field label = %q, want the created %q", response.Field.Label, "Middle")
	}
	if response.Field.ID <= 0 {
		t.Errorf("response field id = %d, want the stored id", response.Field.ID)
	}
}
