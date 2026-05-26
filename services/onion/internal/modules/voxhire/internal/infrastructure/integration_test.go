package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	"github.com/voxire/onion/internal/modules/voxhire/internal/infrastructure"
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
	// NewRepository runs autoMigrate internally
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
