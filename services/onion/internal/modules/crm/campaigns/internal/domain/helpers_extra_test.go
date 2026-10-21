package domain

import "testing"

func TestIsCancelled_WhenCancelled(t *testing.T) {
	c := Campaign{Status: StatusCancelled}
	if c.Status != StatusCancelled {
		t.Errorf("expected status %q, got %q", StatusCancelled, c.Status)
	}
}

func TestTypeConstants_AreDistinct(t *testing.T) {
	types := []string{TypeEmail, TypeSMS, TypePush}
	seen := make(map[string]bool)
	for _, tp := range types {
		if seen[tp] {
			t.Errorf("duplicate type constant: %q", tp)
		}
		seen[tp] = true
	}
}

func TestStatusConstants_AreDistinct(t *testing.T) {
	statuses := []string{StatusDraft, StatusScheduled, StatusSent, StatusCancelled}
	seen := make(map[string]bool)
	for _, s := range statuses {
		if seen[s] {
			t.Errorf("duplicate status constant: %q", s)
		}
		seen[s] = true
	}
}

func TestIsSent_WhenScheduled(t *testing.T) {
	c := Campaign{Status: StatusScheduled}
	if c.IsSent() {
		t.Error("expected IsSent=false for scheduled campaign")
	}
}

func TestIsDraft_WhenSent(t *testing.T) {
	c := Campaign{Status: StatusSent}
	if c.IsDraft() {
		t.Error("expected IsDraft=false for sent campaign")
	}
}

func TestIsScheduled_WhenDraft(t *testing.T) {
	c := Campaign{Status: StatusDraft}
	if c.IsScheduled() {
		t.Error("expected IsScheduled=false for draft campaign")
	}
}

func TestOpenRate_OverTotalSent(t *testing.T) {
	c := Campaign{SentCount: 200, OpenCount: 80}
	want := 0.4
	if got := c.OpenRate(); got != want {
		t.Errorf("OpenRate = %f, want %f", got, want)
	}
}

func TestValidateCampaignInput_SMSType_Valid(t *testing.T) {
	err := ValidateCampaignInput(CampaignInput{Name: "SMS blast", Type: TypeSMS})
	if err != nil {
		t.Errorf("expected no error for SMS type, got %v", err)
	}
}

func TestValidateCampaignInput_PushType_Valid(t *testing.T) {
	err := ValidateCampaignInput(CampaignInput{Name: "Push alert", Type: TypePush})
	if err != nil {
		t.Errorf("expected no error for push type, got %v", err)
	}
}
