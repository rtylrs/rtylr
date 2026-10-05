package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateContact_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Contact Lead Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	contact, err := repo.CreateContact(ctx, testutil.BusinessID, testutil.OwnerID, domain.CreateContactRequest{
		LeadID: lead.ID,
		Name:   "Alice Smith",
		Email:  "alice@corp.test",
	})
	if err != nil {
		t.Fatalf("CreateContact: %v", err)
	}
	if contact.ID <= 0 {
		t.Fatalf("expected positive contact ID, got %d", contact.ID)
	}
}

func TestCreateContact_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Name Contact Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	contact, err := repo.CreateContact(ctx, testutil.BusinessID, testutil.OwnerID, domain.CreateContactRequest{
		LeadID: lead.ID,
		Name:   "Bob Jones",
	})
	if err != nil {
		t.Fatalf("CreateContact: %v", err)
	}
	if contact.Name != "Bob Jones" {
		t.Errorf("Name = %q, want %q", contact.Name, "Bob Jones")
	}
}

func TestCreateContact_EmailIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Email Contact Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	contact, err := repo.CreateContact(ctx, testutil.BusinessID, testutil.OwnerID, domain.CreateContactRequest{
		LeadID: lead.ID,
		Name:   "Email Person",
		Email:  "email@corp.test",
	})
	if err != nil {
		t.Fatalf("CreateContact: %v", err)
	}
	if contact.Email != "email@corp.test" {
		t.Errorf("Email = %q, want %q", contact.Email, "email@corp.test")
	}
}

func TestGetContact_NotFound_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	_, err := repo.GetContact(ctx, 999999999, testutil.OwnerID)
	if err == nil {
		t.Fatal("expected error for non-existent contact, got nil")
	}
}

func TestGetContact_WrongOwner_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "WrongOwner Contact Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	contact, err := repo.CreateContact(ctx, testutil.BusinessID, testutil.OwnerID, domain.CreateContactRequest{
		LeadID: lead.ID,
		Name:   "WrongOwner Person",
	})
	if err != nil {
		t.Fatalf("CreateContact: %v", err)
	}

	_, err = repo.GetContact(ctx, contact.ID, 999999)
	if err == nil {
		t.Fatal("expected error with wrong ownerID, got nil")
	}
}

func TestUpdateContact_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Update Contact Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	contact, err := repo.CreateContact(ctx, testutil.BusinessID, testutil.OwnerID, domain.CreateContactRequest{
		LeadID: lead.ID,
		Name:   "Before Name",
	})
	if err != nil {
		t.Fatalf("CreateContact: %v", err)
	}

	updated, err := repo.UpdateContact(ctx, contact.ID, testutil.OwnerID, domain.UpdateContactRequest{
		Name: "After Name",
	})
	if err != nil {
		t.Fatalf("UpdateContact: %v", err)
	}
	if updated.Name != "After Name" {
		t.Errorf("Name = %q, want %q", updated.Name, "After Name")
	}
}

func TestDeleteContact_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Delete Contact Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	contact, err := repo.CreateContact(ctx, testutil.BusinessID, testutil.OwnerID, domain.CreateContactRequest{
		LeadID: lead.ID,
		Name:   "Delete Me Contact",
	})
	if err != nil {
		t.Fatalf("CreateContact: %v", err)
	}

	if err := repo.DeleteContact(ctx, contact.ID, testutil.OwnerID); err != nil {
		t.Fatalf("DeleteContact: %v", err)
	}

	_, err = repo.GetContact(ctx, contact.ID, testutil.OwnerID)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestListContacts_FilterByLeadID_ReturnsOnlyThatLead(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := configuredRepository(db)
	ctx := context.Background()

	lead, err := repo.CreateLead(ctx, testutil.OwnerID, domain.CreateLeadRequest{
		CompanyName: "Filter Lead Corp",
		Stage:       domain.LeadStageCold,
	})
	if err != nil {
		t.Fatalf("CreateLead: %v", err)
	}

	_, err = repo.CreateContact(ctx, testutil.BusinessID, testutil.OwnerID, domain.CreateContactRequest{
		LeadID: lead.ID,
		Name:   "Filtered Contact",
	})
	if err != nil {
		t.Fatalf("CreateContact: %v", err)
	}

	contacts, err := repo.ListContacts(ctx, testutil.OwnerID, &lead.ID, 100, 0)
	if err != nil {
		t.Fatalf("ListContacts filtered: %v", err)
	}

	for _, c := range contacts {
		if c.LeadID != lead.ID {
			t.Errorf("contact %d has LeadID=%d, want %d", c.ID, c.LeadID, lead.ID)
		}
	}
}
