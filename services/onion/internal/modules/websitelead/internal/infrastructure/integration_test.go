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

func TestWebsiteLead_Create_UniqueIDsPerInsert(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.Create(ctx, domain.LeadInput{Name: "Lead X", Email: "x@example.com"})
	if err != nil {
		t.Fatalf("Create X: %v", err)
	}
	id2, err := repo.Create(ctx, domain.LeadInput{Name: "Lead Y", Email: "y@example.com"})
	if err != nil {
		t.Fatalf("Create Y: %v", err)
	}
	if id1 == id2 {
		t.Errorf("expected distinct IDs, both %d", id1)
	}
}

func TestWebsiteLead_Create_WithBudget(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.LeadInput{
		Name: "Budget Lead", Email: "budget@example.com", Budget: "$10000-$50000",
	})
	if err != nil {
		t.Fatalf("Create with budget: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestWebsiteLead_Create_WithCompanyName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.LeadInput{
		Name: "Corp Lead", Email: "corp@example.com", Company: "MegaCorp Ltd",
	})
	if err != nil {
		t.Fatalf("Create with company: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestWebsiteLead_Create_AllBudgetRanges(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	for _, budget := range []string{"<$1000", "$1000-$5000", "$5000-$10000", "$10000+"} {
		id, err := repo.Create(ctx, domain.LeadInput{
			Name: "Budget Range Lead", Email: "brl@example.com", Budget: budget,
		})
		if err != nil {
			t.Fatalf("budget=%s: %v", budget, err)
		}
		if id <= 0 {
			t.Errorf("budget=%s: expected positive ID", budget)
		}
	}
}

func TestWebsiteLead_Create_ERP_Service(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.LeadInput{
		Name: "ERP Lead", Email: "erp@example.com", Service: "erp",
	})
	if err != nil {
		t.Fatalf("Create ERP: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestWebsiteLead_Create_NoBudgetNoPanic(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.LeadInput{
		Name: "No Budget Lead", Email: "nobudget@example.com",
	})
	if err != nil {
		t.Fatalf("Create no budget: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestWebsiteLead_Create_Idempotent_SameEmailAllowed(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Same email is allowed — each create should return a unique ID.
	id1, err := repo.Create(ctx, domain.LeadInput{Name: "Dup A", Email: "dup@example.com"})
	if err != nil {
		t.Fatalf("Create dup A: %v", err)
	}
	id2, err := repo.Create(ctx, domain.LeadInput{Name: "Dup B", Email: "dup@example.com"})
	if err != nil {
		t.Fatalf("Create dup B: %v", err)
	}
	if id1 == id2 {
		t.Errorf("duplicate email should produce distinct IDs, got same %d", id1)
	}
}

func TestWebsiteLead_Create_TableRowCountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	var before int
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM voxire_website_leads").Scan(&before)

	_, err := repo.Create(ctx, domain.LeadInput{Name: "Counter Lead", Email: "counter@example.com"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	var after int
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM voxire_website_leads").Scan(&after)

	if after != before+1 {
		t.Errorf("expected row count to increase by 1: before=%d after=%d", before, after)
	}
}

func TestWebsiteLead_Create_HRMService(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.LeadInput{
		Name: "HRM Lead", Email: "hrm@example.com", Service: "hrm",
	})
	if err != nil {
		t.Fatalf("Create HRM: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestWebsiteLead_Create_WithAllFields(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.LeadInput{
		Name:    "Full Lead",
		Email:   "full@example.com",
		Phone:   "+1-555-1234",
		Company: "Full Corp",
		Service: "pos",
		Project: "Full featured project",
		Budget:  "$5000-$10000",
		Source:  "google",
	})
	if err != nil {
		t.Fatalf("Create all fields: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}
