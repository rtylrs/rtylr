package domain

import "testing"

func TestIsCancelled_WhenCancelledStatus(t *testing.T) {
	c := Campaign{Status: StatusCancelled}
	if !c.IsCancelled() {
		t.Error("expected IsCancelled=true for cancelled campaign")
	}
}

func TestIsCancelled_WhenSent(t *testing.T) {
	c := Campaign{Status: StatusSent}
	if c.IsCancelled() {
		t.Error("expected IsCancelled=false for sent campaign")
	}
}

func TestHasBody_WhenSet(t *testing.T) {
	c := Campaign{Body: "Hello subscribers!"}
	if !c.HasBody() {
		t.Error("expected HasBody=true when body is set")
	}
}

func TestHasBody_WhenEmpty(t *testing.T) {
	c := Campaign{Body: ""}
	if c.HasBody() {
		t.Error("expected HasBody=false when body is empty")
	}
}

func TestHasSubject_WhenSet(t *testing.T) {
	c := Campaign{Subject: "Summer Sale is Here!"}
	if !c.HasSubject() {
		t.Error("expected HasSubject=true when subject is set")
	}
}

func TestHasSubject_WhenEmpty(t *testing.T) {
	c := Campaign{Subject: ""}
	if c.HasSubject() {
		t.Error("expected HasSubject=false when subject is empty")
	}
}

func TestValidateCampaignUpdateInput_Valid(t *testing.T) {
	err := ValidateCampaignUpdateInput(CampaignUpdateInput{ID: 1, OwnerID: 2, Name: "Summer"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateCampaignUpdateInput_ZeroID(t *testing.T) {
	err := ValidateCampaignUpdateInput(CampaignUpdateInput{ID: 0, OwnerID: 2, Name: "Summer"})
	if err == nil {
		t.Error("expected error for zero id")
	}
}

func TestValidateCampaignUpdateInput_ZeroOwner(t *testing.T) {
	err := ValidateCampaignUpdateInput(CampaignUpdateInput{ID: 1, OwnerID: 0, Name: "Summer"})
	if err == nil {
		t.Error("expected error for zero owner_id")
	}
}

func TestValidateCampaignUpdateInput_EmptyName(t *testing.T) {
	err := ValidateCampaignUpdateInput(CampaignUpdateInput{ID: 1, OwnerID: 2, Name: ""})
	if err == nil {
		t.Error("expected error for empty name")
	}
}
