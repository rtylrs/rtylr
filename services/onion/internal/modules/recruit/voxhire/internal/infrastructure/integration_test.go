package infrastructure_test

import (
	"context"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestVoxhire_CreateCompanyLead_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateCompanyLead(context.Background(), domain.CompanyLeadInput{
		CompanyName: "Acme Corp",
		ContactName: "John Doe",
		WorkEmail:   "john@acme.com",
		Role:        "Software Engineer",
		Service:     "staffing",
	})
	if err != nil {
		t.Fatalf("CreateCompanyLead: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestVoxhire_CreateCompanyLead_WithAllFields(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateCompanyLead(context.Background(), domain.CompanyLeadInput{
		CompanyName:  "Beta Inc",
		ContactName:  "Jane Smith",
		ContactTitle: "CTO",
		WorkEmail:    "jane@beta.com",
		Phone:        "+1-555-0100",
		Role:         "Backend Developer",
		Headcount:    "5-10",
		Timeline:     "1-3 months",
		Service:      "agency",
		Message:      "We need help ASAP",
		AgencyOptIn:  true,
	})
	if err != nil {
		t.Fatalf("CreateCompanyLead all fields: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestVoxhire_CreateCompanyLead_Multiple(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	for i := 0; i < 3; i++ {
		id, err := repo.CreateCompanyLead(context.Background(), domain.CompanyLeadInput{
			CompanyName: "Company",
			ContactName: "Contact",
			WorkEmail:   "contact@company.com",
			Role:        "Dev",
		})
		if err != nil {
			t.Fatalf("CreateCompanyLead #%d: %v", i, err)
		}
		if id <= 0 {
			t.Errorf("iteration %d: expected positive ID", i)
		}
	}
}

func TestVoxhire_CreateTalentApplication_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateTalentApplication(context.Background(), domain.TalentApplicationInput{
		FullName: "Alice Johnson",
		Email:    "alice@example.com",
		Track:    "engineering",
		Consent:  true,
	})
	if err != nil {
		t.Fatalf("CreateTalentApplication: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestVoxhire_CreateTalentApplication_WithAllFields(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateTalentApplication(context.Background(), domain.TalentApplicationInput{
		FullName:           "Bob Wilson",
		Email:              "bob@example.com",
		Phone:              "+1-555-0200",
		CountryOfResidence: "USA",
		Nationality:        "American",
		Track:              "design",
		YearsExperience:    "5-10",
		NoticePeriod:       "1 month",
		OpenTo:             "remote,hybrid",
		LinkedinURL:        "https://linkedin.com/in/bobwilson",
		PortfolioURL:       "https://bobwilson.design",
		Pitch:              "I am a passionate designer...",
		Consent:            true,
		CVKey:              "cvs/bob-wilson-2026.pdf",
	})
	if err != nil {
		t.Fatalf("CreateTalentApplication all fields: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestVoxhire_CreateTalentApplication_NoConsent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateTalentApplication(context.Background(), domain.TalentApplicationInput{
		FullName: "Carol Test",
		Email:    "carol@example.com",
		Track:    "sales",
		Consent:  false,
	})
	if err != nil {
		t.Fatalf("CreateTalentApplication without consent: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestVoxhire_CreateTalentApplication_Multiple(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	tracks := []string{"engineering", "design", "sales"}
	for _, track := range tracks {
		id, err := repo.CreateTalentApplication(context.Background(), domain.TalentApplicationInput{
			FullName: "Test Applicant",
			Email:    "applicant@example.com",
			Track:    track,
			Consent:  true,
		})
		if err != nil {
			t.Fatalf("CreateTalentApplication track=%s: %v", track, err)
		}
		if id <= 0 {
			t.Errorf("track %s: expected positive ID", track)
		}
	}
}

func TestVoxhire_AutoMigrate_TablesExist(t *testing.T) {
	db := testutil.OpenDB(t)
	// The test database applies Recruit-owned migrations before running integration tests.
	repo := infrastructure.NewRepository(db)
	_ = repo

	// Verify tables are queryable
	var count int
	if err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM voxhire_company_leads").Scan(&count); err != nil {
		t.Fatalf("voxhire_company_leads not accessible: %v", err)
	}
	if err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM voxhire_talent_applications").Scan(&count); err != nil {
		t.Fatalf("voxhire_talent_applications not accessible: %v", err)
	}
}

func TestVoxhire_ListApplications_SearchesExtractedCVText(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := domain.WithOwnerID(context.Background(), testutil.OwnerID)
	opening, err := repo.CreateCareer(ctx, domain.CareerInput{
		Title:   "CV Search Role",
		Details: "Integration test opening for CV search",
	})
	if err != nil {
		t.Fatalf("CreateCareer: %v", err)
	}
	jobID := opening.ID

	id, err := repo.CreateCareerApplication(ctx, domain.CareerApplicationInput{
		JobID:             jobID,
		Name:              "CV Search Candidate",
		Email:             "cv-search@example.com",
		Location:          "Beirut",
		ExperienceSummary: "General operations profile",
		CVKey:             "voxhire/applications/cv-search/resume.pdf",
	})
	if err != nil {
		t.Fatalf("CreateCareerApplication: %v", err)
	}

	_, err = db.ExecContext(ctx, `
		UPDATE voxire_career_applications
		   SET cv_processing_status = 'processed',
		       cv_summary = 'Clinic operations and staffing background',
		       cv_extracted_text = 'Led oncology clinic scheduling, patient intake, and insurance coordination.'
		 WHERE id = ?`, id)
	if err != nil {
		t.Fatalf("seed extracted cv text: %v", err)
	}

	apps, total, err := repo.ListApplications(domain.WithOwnerID(ctx, testutil.OwnerID), domain.ApplicationFilter{
		JobID: jobID,
		Query: "oncology insurance",
		Limit: 20,
	})
	if err != nil {
		t.Fatalf("ListApplications: %v", err)
	}
	if total != 1 || len(apps) != 1 {
		t.Fatalf("expected one CV-text match, got total=%d len=%d", total, len(apps))
	}
	if apps[0].ID != id {
		t.Fatalf("expected application id %d, got %d", id, apps[0].ID)
	}
	excerpt := strings.ToLower(apps[0].CVSearchExcerpt)
	if !strings.Contains(excerpt, "oncology") || !strings.Contains(excerpt, "insurance") {
		t.Fatalf("expected CV excerpt to include search phrase, got %q", apps[0].CVSearchExcerpt)
	}
}

func TestVoxhire_CreateCompanyLead_UniqueIDs(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.CreateCompanyLead(ctx, domain.CompanyLeadInput{
		CompanyName: "Unique Corp A", ContactName: "A", WorkEmail: "a@example.com", Role: "Dev",
	})
	if err != nil {
		t.Fatalf("CreateCompanyLead A: %v", err)
	}
	id2, err := repo.CreateCompanyLead(ctx, domain.CompanyLeadInput{
		CompanyName: "Unique Corp B", ContactName: "B", WorkEmail: "b@example.com", Role: "Dev",
	})
	if err != nil {
		t.Fatalf("CreateCompanyLead B: %v", err)
	}
	if id1 == id2 {
		t.Errorf("expected distinct IDs, both got %d", id1)
	}
}

func TestVoxhire_CreateCompanyLead_HeadcountVariants(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	for _, hc := range []string{"1-5", "5-10", "10-50", "50-200", "200+"} {
		id, err := repo.CreateCompanyLead(ctx, domain.CompanyLeadInput{
			CompanyName: "HC Corp", WorkEmail: "hc@example.com", Role: "PM", Headcount: hc,
		})
		if err != nil {
			t.Fatalf("headcount=%s: %v", hc, err)
		}
		if id <= 0 {
			t.Errorf("headcount=%s: expected positive ID", hc)
		}
	}
}

func TestVoxhire_CreateCompanyLead_TimelineVariants(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	for _, tl := range []string{"immediately", "1-3 months", "3-6 months", "6+ months"} {
		id, err := repo.CreateCompanyLead(ctx, domain.CompanyLeadInput{
			CompanyName: "TL Corp", WorkEmail: "tl@example.com", Role: "Eng", Timeline: tl,
		})
		if err != nil {
			t.Fatalf("timeline=%s: %v", tl, err)
		}
		if id <= 0 {
			t.Errorf("timeline=%s: expected positive ID", tl)
		}
	}
}

func TestVoxhire_CreateCompanyLead_EmptyMessage(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateCompanyLead(context.Background(), domain.CompanyLeadInput{
		CompanyName: "No Message Corp", WorkEmail: "nm@example.com", Role: "Dev", Message: "",
	})
	if err != nil {
		t.Fatalf("CreateCompanyLead empty message: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestVoxhire_CreateTalentApplication_PhoneOptional(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateTalentApplication(context.Background(), domain.TalentApplicationInput{
		FullName: "No Phone Person", Email: "nophone@example.com", Track: "engineering", Consent: true,
	})
	if err != nil {
		t.Fatalf("CreateTalentApplication no phone: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestVoxhire_CreateTalentApplication_CVKeyOptional(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateTalentApplication(context.Background(), domain.TalentApplicationInput{
		FullName: "No CV Person", Email: "nocv@example.com", Track: "design", Consent: true, CVKey: "",
	})
	if err != nil {
		t.Fatalf("CreateTalentApplication no cv: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestVoxhire_CreateTalentApplication_UniqueIDs(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.CreateTalentApplication(ctx, domain.TalentApplicationInput{
		FullName: "Person One", Email: "one@example.com", Track: "engineering",
	})
	if err != nil {
		t.Fatalf("application 1: %v", err)
	}
	id2, err := repo.CreateTalentApplication(ctx, domain.TalentApplicationInput{
		FullName: "Person Two", Email: "two@example.com", Track: "design",
	})
	if err != nil {
		t.Fatalf("application 2: %v", err)
	}
	if id1 == id2 {
		t.Errorf("expected distinct IDs, both got %d", id1)
	}
}

func TestVoxhire_CreateTalentApplication_WithLinkedIn(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateTalentApplication(context.Background(), domain.TalentApplicationInput{
		FullName:    "LinkedIn User",
		Email:       "linkedin@example.com",
		Track:       "sales",
		LinkedinURL: "https://linkedin.com/in/testuser",
		Consent:     true,
	})
	if err != nil {
		t.Fatalf("CreateTalentApplication with linkedin: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestVoxhire_CreateCompanyLead_LongMessage(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	longMsg := "We are a fast-growing startup in the fintech space and we need a full-stack engineering team that can integrate seamlessly with our existing codebase and deliver results in a tight timeline."
	id, err := repo.CreateCompanyLead(context.Background(), domain.CompanyLeadInput{
		CompanyName: "Long Msg Corp", WorkEmail: "longmsg@example.com", Role: "Dev", Message: longMsg,
	})
	if err != nil {
		t.Fatalf("CreateCompanyLead long message: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestVoxhire_CreateTalentApplication_ExperienceVariants(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	for _, exp := range []string{"0-2", "2-5", "5-10", "10+"} {
		id, err := repo.CreateTalentApplication(ctx, domain.TalentApplicationInput{
			FullName: "Exp User", Email: "exp@example.com", Track: "engineering",
			YearsExperience: exp, Consent: true,
		})
		if err != nil {
			t.Fatalf("experience=%s: %v", exp, err)
		}
		if id <= 0 {
			t.Errorf("experience=%s: expected positive ID", exp)
		}
	}
}
