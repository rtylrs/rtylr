package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/websitelead/internal/domain"
	"github.com/voxire/onion/internal/modules/websitelead/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestWebsiteLead_Create_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.LeadInput{
		Name:    "Alice Brown",
		Email:   "alice@example.com",
		Company: "Startup Inc",
		Service: "pos",
		Project: "New POS system for retail",
		Budget:  "$5000-$10000",
		Source:  "google",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestWebsiteLead_Create_MinimalFields(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.LeadInput{
		Name:  "Minimal Lead",
		Email: "minimal@example.com",
	})
	if err != nil {
		t.Fatalf("Create minimal: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestWebsiteLead_Create_WithPhone(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.LeadInput{
		Name:  "Phone Lead",
		Email: "phone@example.com",
		Phone: "+1-555-9999",
	})
	if err != nil {
		t.Fatalf("Create with phone: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestWebsiteLead_Create_Multiple(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	for i := 0; i < 5; i++ {
		id, err := repo.Create(context.Background(), domain.LeadInput{
			Name:  "Lead",
			Email: "lead@example.com",
		})
		if err != nil {
			t.Fatalf("Create #%d: %v", i, err)
		}
		if id <= 0 {
			t.Errorf("iteration %d: expected positive ID", i)
		}
	}
}

func TestWebsiteLead_Create_DifferentServices(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	services := []string{"pos", "erp", "crm", "hrm"}
	for _, service := range services {
		id, err := repo.Create(context.Background(), domain.LeadInput{
			Name:    "Service Lead",
			Email:   "service@example.com",
			Service: service,
		})
		if err != nil {
			t.Fatalf("Create service=%s: %v", service, err)
		}
		if id <= 0 {
			t.Errorf("service %s: expected positive ID", service)
		}
	}
}

func TestWebsiteLead_Create_AutoMigrated(t *testing.T) {
	db := testutil.OpenDB(t)
	// NewRepository runs autoMigrate
	repo := infrastructure.NewRepository(db)

	var count int
	if err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM voxire_website_leads").Scan(&count); err != nil {
		t.Fatalf("voxire_website_leads table not accessible: %v", err)
	}
	_ = repo
}

func TestWebsiteLead_Create_WithLongProject(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	longProject := "We are looking for a comprehensive solution that handles all aspects of our business operations including inventory management, customer relationship management, point of sale, and more."
	id, err := repo.Create(context.Background(), domain.LeadInput{
		Name:    "Long Project Lead",
		Email:   "longproject@example.com",
		Project: longProject,
	})
	if err != nil {
		t.Fatalf("Create with long project: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestWebsiteLead_Create_DifferentSources(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	sources := []string{"google", "linkedin", "referral", "direct"}
	for _, source := range sources {
		id, err := repo.Create(context.Background(), domain.LeadInput{
			Name:   "Source Lead",
			Email:  "sourcelead@example.com",
			Source: source,
		})
		if err != nil {
			t.Fatalf("Create source=%s: %v", source, err)
		}
		if id <= 0 {
			t.Errorf("source %s: expected positive ID", source)
		}
	}
}
