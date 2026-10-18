package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/testutil"
)

func TestContract_StoresTermsAndLinksTheLeadsCustomer(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.BusinessID, domain.CreateLeadRequest{
		CompanyName: "Generic Terms Hotel",
		Stage:       domain.LeadStageProposalSent,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}
	contact, err := repo.CreateContact(ctx, testutil.BusinessID, testutil.OwnerID, domain.CreateContactRequest{
		LeadID: lead.ID, Name: "Karim Nassar", Email: "karim@terms.test", IsPrimary: true,
	})
	if err != nil {
		t.Fatalf("CreateContact: %v", err)
	}
	// The customers module links a captured contact back to its customer.
	const customerID int64 = 424242
	if err := repo.LinkCustomerToContact(ctx, testutil.BusinessID, contact.ID, customerID); err != nil {
		t.Fatalf("LinkCustomerToContact: %v", err)
	}

	start, end := "2026-11-01", "2027-10-31"
	created, err := repo.CreateContract(ctx, testutil.BusinessID, domain.CreateContractRequest{
		LeadID:    lead.ID,
		Plan:      "Annual maintenance for 3 branches",
		Value:     money.Amount(1200000),
		Billing:   domain.BillingYearly,
		StartDate: &start,
		EndDate:   &end,
	})
	if err != nil {
		t.Fatalf("CreateContract: %v", err)
	}
	if created.Plan != "Annual maintenance for 3 branches" || created.Value != 1200000 ||
		created.Billing != domain.BillingYearly {
		t.Fatalf("terms = %q %d %q, want the sold item, 12000.00, yearly", created.Plan, created.Value, created.Billing)
	}
	if created.StartDate == nil || *created.StartDate != start || created.EndDate == nil || *created.EndDate != end {
		t.Fatalf("term = %v..%v, want %s..%s", created.StartDate, created.EndDate, start, end)
	}
	if created.CustomerID == nil || *created.CustomerID != customerID {
		t.Fatalf("CustomerID = %v, want the lead contact's customer %d", created.CustomerID, customerID)
	}

	// An update that names no customer keeps the link.
	updated, err := repo.UpdateContract(ctx, created.ID, testutil.BusinessID, domain.UpdateContractRequest{
		Plan:    "Monthly support",
		Value:   money.Amount(50000),
		Billing: domain.BillingMonthly,
	})
	if err != nil {
		t.Fatalf("UpdateContract: %v", err)
	}
	if updated.Billing != domain.BillingMonthly || updated.Value != 50000 {
		t.Fatalf("updated terms = %d %q, want 500.00 monthly", updated.Value, updated.Billing)
	}
	if updated.CustomerID == nil || *updated.CustomerID != customerID {
		t.Fatalf("CustomerID after update = %v, want %d", updated.CustomerID, customerID)
	}
}
