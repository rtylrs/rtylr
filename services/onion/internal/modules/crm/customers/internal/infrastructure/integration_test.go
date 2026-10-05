package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/business"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListCustomers_ReturnsJamieAndAlex(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)

	customers, total, err := repo.ListCustomers(context.Background(), domain.CustomerListFilter{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Page:       1,
		PageSize:   50,
	})
	if err != nil {
		t.Fatalf("ListCustomers error: %v", err)
	}
	if total == 0 {
		t.Fatal("expected non-zero total, got 0")
	}
	if len(customers) == 0 {
		t.Fatal("expected at least one customer row, got none")
	}

	names := make(map[string]bool)
	for _, c := range customers {
		names[c.FirstName] = true
	}

	if !names["Jamie"] {
		t.Errorf("expected to find customer 'Jamie', got names: %v", names)
	}
	if !names["Alex"] {
		t.Errorf("expected to find customer 'Alex', got names: %v", names)
	}
}

func TestListCustomers_AlexHasGoldTier(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)

	customers, _, err := repo.ListCustomers(context.Background(), domain.CustomerListFilter{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Page:       1,
		PageSize:   50,
	})
	if err != nil {
		t.Fatalf("ListCustomers error: %v", err)
	}

	for _, c := range customers {
		if c.FirstName == "Alex" {
			if c.LoyaltyTier != "gold" {
				t.Errorf("expected Alex to have loyalty_tier='gold', got %q", c.LoyaltyTier)
			}
			return
		}
	}
	t.Fatal("customer 'Alex' not found in results")
}

func TestListCustomers_SearchGold_FindsAlex(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	customers, total, err := repo.ListCustomers(ctx, domain.CustomerListFilter{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Search:     "Gold",
		Page:       1,
		PageSize:   20,
	})
	if err != nil {
		t.Fatalf("ListCustomers search 'Gold': unexpected error: %v", err)
	}
	if total == 0 || len(customers) == 0 {
		t.Fatal("ListCustomers search 'Gold': expected at least one result")
	}

	found := false
	for _, c := range customers {
		if c.ID == 5302 {
			found = true
		}
	}
	if !found {
		t.Error("ListCustomers search 'Gold': expected customer id=5302 (Alex Gold) in results")
	}
}

func TestGetCustomerByID_AlexHasGoldTierAndHighSpend(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	c, err := repo.GetCustomerByID(ctx, testutil.OwnerID, 5302)
	if err != nil {
		t.Fatalf("GetCustomerByID(5302): unexpected error: %v", err)
	}

	if c.LoyaltyTier != "gold" {
		t.Errorf("GetCustomerByID(5302): expected loyalty_tier='gold', got %q", c.LoyaltyTier)
	}

	if c.TotalSpent < 2500 {
		t.Errorf("GetCustomerByID(5302): expected total_spent >= 2500, got %v", c.TotalSpent)
	}
}

func TestGetCustomerByID_JamieBronze(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	c, err := repo.GetCustomerByID(ctx, testutil.OwnerID, 5301)
	if err != nil {
		t.Fatalf("GetCustomerByID(5301): unexpected error: %v", err)
	}

	if c.LoyaltyTier != "bronze" {
		t.Errorf("GetCustomerByID(5301): expected loyalty_tier='bronze', got %q", c.LoyaltyTier)
	}

	if c.Email != "jamie@test.local" {
		t.Errorf("GetCustomerByID(5301): expected email='jamie@test.local', got %q", c.Email)
	}
}

func TestListCustomers_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	customers, total, err := repo.ListCustomers(ctx, domain.CustomerListFilter{
		OwnerID:  0,
		Page:     1,
		PageSize: 50,
	})
	if err != nil {
		t.Fatalf("ListCustomers(ownerID=0): unexpected error: %v", err)
	}
	if total != 0 || len(customers) != 0 {
		t.Fatalf("expected 0 customers for ownerID=0, got total=%d len=%d", total, len(customers))
	}
}

func TestListCustomers_SearchJamie_FindsCustomer(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	customers, _, err := repo.ListCustomers(ctx, domain.CustomerListFilter{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Search:     "Jamie",
		Page:       1,
		PageSize:   20,
	})
	if err != nil {
		t.Fatalf("ListCustomers search 'Jamie': %v", err)
	}
	found := false
	for _, c := range customers {
		if c.ID == 5301 {
			found = true
		}
	}
	if !found {
		t.Error("expected customer id=5301 (Jamie) in search results")
	}
}

func TestListCustomers_SearchNonExistent_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	customers, total, err := repo.ListCustomers(ctx, domain.CustomerListFilter{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Search:     "xyzzy_no_such_customer_9999",
		Page:       1,
		PageSize:   20,
	})
	if err != nil {
		t.Fatalf("ListCustomers search non-existent: %v", err)
	}
	if total != 0 || len(customers) != 0 {
		t.Fatalf("expected 0 results for non-existent search, got total=%d len=%d", total, len(customers))
	}
}

func TestListCustomers_Pagination_LimitWorks(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	customers, _, err := repo.ListCustomers(ctx, domain.CustomerListFilter{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Page:       1,
		PageSize:   1,
	})
	if err != nil {
		t.Fatalf("ListCustomers pageSize=1: %v", err)
	}
	if len(customers) > 1 {
		t.Fatalf("expected at most 1 customer with pageSize=1, got %d", len(customers))
	}
}

func TestGetCustomerByID_5301_IsJamie(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	c, err := repo.GetCustomerByID(ctx, testutil.OwnerID, 5301)
	if err != nil {
		t.Fatalf("GetCustomerByID(5301): %v", err)
	}
	if c.FirstName != "Jamie" {
		t.Errorf("expected FirstName='Jamie', got %q", c.FirstName)
	}
}

func TestGetCustomerByID_5302_IsAlex(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	c, err := repo.GetCustomerByID(ctx, testutil.OwnerID, 5302)
	if err != nil {
		t.Fatalf("GetCustomerByID(5302): %v", err)
	}
	if c.FirstName != "Alex" {
		t.Errorf("expected FirstName='Alex', got %q", c.FirstName)
	}
}

func TestGetCustomerByID_NotFound_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	_, err := repo.GetCustomerByID(ctx, testutil.OwnerID, 999999999)
	if err == nil {
		t.Fatal("expected error for non-existent customer, got nil")
	}
}

func TestGetCustomerByID_WrongOwner_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	_, err := repo.GetCustomerByID(ctx, 999999, 5301)
	if err == nil {
		t.Fatal("expected error when fetching customer with wrong ownerID, got nil")
	}
}

func TestCreateCustomer_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	id, err := repo.CreateCustomer(ctx, domain.CustomerInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FirstName:  "Test",
		LastName:   "Create",
		Email:      "test.create@test.local",
	})
	if err != nil {
		t.Fatalf("CreateCustomer: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreateCustomer_EmailIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	email := "email.stored@test.local"
	id, err := repo.CreateCustomer(ctx, domain.CustomerInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FirstName:  "Email",
		LastName:   "Test",
		Email:      email,
	})
	if err != nil {
		t.Fatalf("CreateCustomer: %v", err)
	}

	c, err := repo.GetCustomerByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetCustomerByID after create: %v", err)
	}
	if c.Email != email {
		t.Errorf("Email = %q, want %q", c.Email, email)
	}
}

func TestUpdateCustomerLoyalty_IncreasesPoints(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	if err := repo.UpdateCustomerLoyalty(ctx, testutil.OwnerID, 5301, 999, "bronze"); err != nil {
		t.Fatalf("UpdateCustomerLoyalty: %v", err)
	}

	c, err := repo.GetCustomerByID(ctx, testutil.OwnerID, 5301)
	if err != nil {
		t.Fatalf("GetCustomerByID after loyalty update: %v", err)
	}
	if c.LoyaltyPoints != 999 {
		t.Errorf("expected LoyaltyPoints=999, got %d", c.LoyaltyPoints)
	}

	// Restore
	if err := repo.UpdateCustomerLoyalty(ctx, testutil.OwnerID, 5301, 50, "bronze"); err != nil {
		t.Errorf("restore UpdateCustomerLoyalty: %v", err)
	}
}

func TestUpdateCustomerLoyalty_ChangesToGold(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	if err := repo.UpdateCustomerLoyalty(ctx, testutil.OwnerID, 5301, 2000, "gold"); err != nil {
		t.Fatalf("UpdateCustomerLoyalty: %v", err)
	}

	c, err := repo.GetCustomerByID(ctx, testutil.OwnerID, 5301)
	if err != nil {
		t.Fatalf("GetCustomerByID after tier update: %v", err)
	}
	if c.LoyaltyTier != "gold" {
		t.Errorf("expected LoyaltyTier='gold', got %q", c.LoyaltyTier)
	}

	// Restore
	if err := repo.UpdateCustomerLoyalty(ctx, testutil.OwnerID, 5301, 50, "bronze"); err != nil {
		t.Errorf("restore UpdateCustomerLoyalty: %v", err)
	}
}

func TestRecordVisit_IncreasesVisitCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	before, err := repo.GetCustomerByID(ctx, testutil.OwnerID, 5301)
	if err != nil {
		t.Fatalf("GetCustomerByID before visit: %v", err)
	}

	if err := repo.RecordVisit(ctx, testutil.OwnerID, 5301, 25.00); err != nil {
		t.Fatalf("RecordVisit: %v", err)
	}

	after, err := repo.GetCustomerByID(ctx, testutil.OwnerID, 5301)
	if err != nil {
		t.Fatalf("GetCustomerByID after visit: %v", err)
	}
	if after.VisitCount <= before.VisitCount {
		t.Errorf("expected VisitCount to increase: before=%d after=%d", before.VisitCount, after.VisitCount)
	}
}

func TestDeleteCustomer_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	id, err := repo.CreateCustomer(ctx, domain.CustomerInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		FirstName:  "Delete",
		LastName:   "Me",
		Email:      "delete.me@test.local",
	})
	if err != nil {
		t.Fatalf("CreateCustomer: %v", err)
	}

	if err := repo.DeleteCustomer(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("DeleteCustomer: %v", err)
	}

	_, err = repo.GetCustomerByID(ctx, testutil.OwnerID, id)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}
