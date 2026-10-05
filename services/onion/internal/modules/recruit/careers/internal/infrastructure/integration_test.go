package infrastructure_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
	"github.com/voxire/onion/internal/modules/recruit/careers/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetActiveCareers_ReturnsSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	careers, err := repo.GetActiveCareers(ctx)
	if err != nil {
		t.Fatalf("GetActiveCareers: %v", err)
	}
	// Result may be empty if no postings exist; the call itself must succeed.
	_ = careers
}

func TestGetActiveCareers_NeverReturnsNilOnEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	careers, err := repo.GetActiveCareers(ctx)
	if err != nil {
		t.Fatalf("GetActiveCareers: %v", err)
	}
	if careers == nil {
		t.Error("expected non-nil slice, got nil")
	}
}

func TestGetActiveCareers_ActivePostingsHaveTitle(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	careers, err := repo.GetActiveCareers(ctx)
	if err != nil {
		t.Fatalf("GetActiveCareers: %v", err)
	}
	for _, c := range careers {
		if c.Title == "" {
			t.Errorf("career id=%d has empty Title", c.Id)
		}
		if !c.IsActive {
			t.Errorf("career id=%d returned but IsActive=false", c.Id)
		}
	}
}

func TestGetActiveCareers_IdsArePositive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	careers, err := repo.GetActiveCareers(ctx)
	if err != nil {
		t.Fatalf("GetActiveCareers: %v", err)
	}
	for _, c := range careers {
		if c.Id <= 0 {
			t.Errorf("expected positive ID, got %d", c.Id)
		}
	}
}

func TestGetActiveCareers_StableAcrossMultipleCalls(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.GetActiveCareers(ctx)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	second, err := repo.GetActiveCareers(ctx)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if len(first) != len(second) {
		t.Errorf("expected stable result: first=%d second=%d", len(first), len(second))
	}
}

func TestCareers_CreateCareer_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	slug := fmt.Sprintf("integration-test-role-%d", testutil.OwnerID)
	id, err := repo.CreateCareer(ctx, domain.RecruitCareerInput{
		OwnerID:     testutil.OwnerID,
		BusinessID:  testutil.BusinessID,
		Slug:        slug,
		Title:       "Integration Test Role",
		IsPublished: false,
	})
	if err != nil {
		t.Fatalf("CreateCareer: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM careers WHERE owner_id = ? AND slug = ?", testutil.OwnerID, slug)
	})
}

func TestCareers_CreateCareer_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCareer(ctx, domain.RecruitCareerInput{
		OwnerID: testutil.OwnerID, Slug: "appear-in-list-role", Title: "List Test Role",
		BusinessID: testutil.BusinessID,
	})
	if err != nil {
		t.Fatalf("CreateCareer: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM careers WHERE owner_id = ? AND slug = ?", testutil.OwnerID, "appear-in-list-role")
	})

	careers, err := repo.ListCareers(ctx, testutil.OwnerID, false)
	if err != nil {
		t.Fatalf("ListCareers: %v", err)
	}
	found := false
	for _, c := range careers {
		if c.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("created career id=%d not found in list", id)
	}
}

func TestCareers_ListCareers_OwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	careers, err := repo.ListCareers(ctx, testutil.OwnerID, false)
	if err != nil {
		t.Fatalf("ListCareers: %v", err)
	}
	for _, c := range careers {
		if c.OwnerID != testutil.OwnerID {
			t.Errorf("expected owner_id=%d, got %d", testutil.OwnerID, c.OwnerID)
		}
	}
}

func TestCareers_UpdateCareer_ChangesTitle(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCareer(ctx, domain.RecruitCareerInput{
		OwnerID: testutil.OwnerID, Slug: "update-career-slug", Title: "Old Title",
		BusinessID: testutil.BusinessID,
	})
	if err != nil {
		t.Fatalf("CreateCareer: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM careers WHERE owner_id = ? AND slug = ?", testutil.OwnerID, "update-career-slug")
	})

	if err := repo.UpdateCareer(ctx, domain.RecruitCareerInput{
		ID: id, OwnerID: testutil.OwnerID, Slug: "update-career-slug", Title: "New Title",
		BusinessID: testutil.BusinessID,
	}); err != nil {
		t.Fatalf("UpdateCareer: %v", err)
	}

	careers, err := repo.ListCareers(ctx, testutil.OwnerID, false)
	if err != nil {
		t.Fatalf("ListCareers: %v", err)
	}
	for _, c := range careers {
		if c.ID == id && c.Title != "New Title" {
			t.Errorf("expected 'New Title', got %q", c.Title)
		}
	}
}

func TestCareers_DeleteCareer_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateCareer(ctx, domain.RecruitCareerInput{
		OwnerID: testutil.OwnerID, Slug: "delete-career-slug", Title: "Delete Me",
		BusinessID: testutil.BusinessID,
	})
	if err != nil {
		t.Fatalf("CreateCareer: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM careers WHERE owner_id = ? AND slug = ?", testutil.OwnerID, "delete-career-slug")
	})

	if err := repo.DeleteCareer(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("DeleteCareer: %v", err)
	}

	careers, _ := repo.ListCareers(ctx, testutil.OwnerID, false)
	for _, c := range careers {
		if c.ID == id {
			t.Errorf("deleted career id=%d still in list", id)
		}
	}
}

func TestCareers_GetCareerBySlug_ReturnsCareer(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const slug = "get-by-slug-career"
	id, err := repo.CreateCareer(ctx, domain.RecruitCareerInput{
		OwnerID: testutil.OwnerID, Slug: slug, Title: "Slug Career",
		BusinessID: testutil.BusinessID,
	})
	if err != nil {
		t.Fatalf("CreateCareer: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM careers WHERE owner_id = ? AND slug = ?", testutil.OwnerID, slug)
	})

	career, err := repo.GetCareerBySlug(ctx, testutil.OwnerID, slug)
	if err != nil {
		t.Fatalf("GetCareerBySlug: %v", err)
	}
	if career.ID != id {
		t.Errorf("expected id=%d, got %d", id, career.ID)
	}
}

func TestCareers_SaveApplication_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// First create a career to apply to.
	careerID, err := repo.CreateCareer(ctx, domain.RecruitCareerInput{
		OwnerID: testutil.OwnerID, Slug: "apply-career-slug", Title: "Apply Target",
		BusinessID: testutil.BusinessID,
	})
	if err != nil {
		t.Fatalf("CreateCareer: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM careers WHERE owner_id = ? AND slug = ?", testutil.OwnerID, "apply-career-slug")
	})

	appID, err := repo.SaveApplication(ctx, domain.ApplicationInput{
		JobID: careerID,
		Name:  "Test Applicant",
		Email: "applicant@example.com",
	})
	if err != nil {
		t.Fatalf("SaveApplication: %v", err)
	}
	if appID <= 0 {
		t.Errorf("expected positive ID, got %d", appID)
	}
}

func TestCareers_GetSubmissionResumeURL_IsBusinessScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	careerID, err := repo.CreateCareer(ctx, domain.RecruitCareerInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Slug: "submission-resume-url", Title: "Resume URL Test",
	})
	if err != nil {
		t.Fatalf("CreateCareer: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM careers_submissions WHERE career_id = ?", careerID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM careers WHERE id = ?", careerID)
	})

	const resumeURL = "resumes/1001/1/test-resume.pdf"
	submissionID, err := repo.CreateSubmission(ctx, domain.SubmissionInput{
		CareerID: careerID, OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		CandidateName: "Ada Lovelace", Email: "ada@example.com", ResumeURL: resumeURL,
	})
	if err != nil {
		t.Fatalf("CreateSubmission: %v", err)
	}

	got, err := repo.GetSubmissionResumeURL(ctx, testutil.BusinessID, submissionID)
	if err != nil {
		t.Fatalf("GetSubmissionResumeURL: %v", err)
	}
	if got != resumeURL {
		t.Fatalf("resume URL = %q, want %q", got, resumeURL)
	}

	_, err = repo.GetSubmissionResumeURL(ctx, testutil.BusinessID+1, submissionID)
	if !errors.Is(err, domain.ErrSubmissionNotFound) {
		t.Fatalf("GetSubmissionResumeURL with another business error = %v, want ErrSubmissionNotFound", err)
	}
}

func TestCareers_ListSections_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	sections, err := repo.ListSections(ctx, testutil.OwnerID, false)
	if err != nil {
		t.Fatalf("ListSections: %v", err)
	}
	_ = sections
}

func TestCareers_CreateSection_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateSection(ctx, domain.SectionInput{
		OwnerID: testutil.OwnerID, Kind: "perks", Title: "Our Perks", Enabled: true,
		BusinessID: testutil.BusinessID,
	})
	if err != nil {
		t.Fatalf("CreateSection: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.DeleteSection(ctx, testutil.OwnerID, id) })
}

func TestCareers_CreateSection_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateSection(ctx, domain.SectionInput{
		OwnerID: testutil.OwnerID, Kind: "culture", Title: "Our Culture", Enabled: true,
		BusinessID: testutil.BusinessID,
	})
	if err != nil {
		t.Fatalf("CreateSection: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteSection(ctx, testutil.OwnerID, id) })

	sections, err := repo.ListSections(ctx, testutil.OwnerID, false)
	if err != nil {
		t.Fatalf("ListSections: %v", err)
	}
	found := false
	for _, s := range sections {
		if s.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("created section id=%d not in list", id)
	}
}

func TestCareers_DeleteSection_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateSection(ctx, domain.SectionInput{
		OwnerID: testutil.OwnerID, Kind: "values", Title: "Our Values", Enabled: true,
		BusinessID: testutil.BusinessID,
	})
	if err != nil {
		t.Fatalf("CreateSection: %v", err)
	}

	if err := repo.DeleteSection(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("DeleteSection: %v", err)
	}

	sections, _ := repo.ListSections(ctx, testutil.OwnerID, false)
	for _, s := range sections {
		if s.ID == id {
			t.Errorf("deleted section id=%d still in list", id)
		}
	}
}

func TestCareers_ListCareers_PublishedOnlyFilter(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Unpublished career should not appear in publishedOnly=true result.
	id, err := repo.CreateCareer(ctx, domain.RecruitCareerInput{
		OwnerID: testutil.OwnerID, Slug: "unpublished-career", Title: "Unpublished", IsPublished: false,
		BusinessID: testutil.BusinessID,
	})
	if err != nil {
		t.Fatalf("CreateCareer: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM careers WHERE owner_id = ? AND slug = ?", testutil.OwnerID, "unpublished-career")
	})

	published, err := repo.ListCareers(ctx, testutil.OwnerID, true)
	if err != nil {
		t.Fatalf("ListCareers publishedOnly: %v", err)
	}
	for _, c := range published {
		if c.ID == id {
			t.Error("unpublished career appeared in publishedOnly=true list")
		}
	}
}

func TestCareers_CareerSalary_RoundTripsExactCents(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const slug = "salary-exact-cents"
	id, err := repo.CreateCareer(ctx, domain.RecruitCareerInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Slug:       slug,
		Title:      "Exact Salary Role",
		SalaryMin:  money.Amount(123456),
		SalaryMax:  money.Amount(234567),
	})
	if err != nil {
		t.Fatalf("CreateCareer: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM careers WHERE id = ?", id)
	})

	careers, err := repo.ListCareers(ctx, testutil.BusinessID, false)
	if err != nil {
		t.Fatalf("ListCareers: %v", err)
	}
	for _, career := range careers {
		if career.ID == id {
			if career.SalaryMin != money.Amount(123456) || career.SalaryMax != money.Amount(234567) {
				t.Fatalf("salary = (%s, %s), want (1234.56, 2345.67)", career.SalaryMin.String(), career.SalaryMax.String())
			}
			return
		}
	}
	t.Fatalf("career id=%d not found", id)
}
