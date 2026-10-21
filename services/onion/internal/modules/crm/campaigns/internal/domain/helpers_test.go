package domain

import (
	"errors"
	"testing"
)

func TestIsSent_WhenSent(t *testing.T) {
	c := Campaign{Status: StatusSent}
	if !c.IsSent() {
		t.Error("expected IsSent=true for status='sent'")
	}
}

func TestIsSent_WhenDraft(t *testing.T) {
	c := Campaign{Status: StatusDraft}
	if c.IsSent() {
		t.Error("expected IsSent=false for status='draft'")
	}
}

func TestIsDraft_WhenDraft(t *testing.T) {
	c := Campaign{Status: StatusDraft}
	if !c.IsDraft() {
		t.Error("expected IsDraft=true for status='draft'")
	}
}

func TestIsScheduled_WhenScheduled(t *testing.T) {
	c := Campaign{Status: StatusScheduled}
	if !c.IsScheduled() {
		t.Error("expected IsScheduled=true for status='scheduled'")
	}
}

func TestOpenRate_WhenNoSent(t *testing.T) {
	c := Campaign{SentCount: 0, OpenCount: 0}
	if r := c.OpenRate(); r != 0 {
		t.Errorf("expected OpenRate=0 when SentCount=0, got %f", r)
	}
}

func TestOpenRate_CalculatesCorrectly(t *testing.T) {
	c := Campaign{SentCount: 100, OpenCount: 25}
	if r := c.OpenRate(); r != 0.25 {
		t.Errorf("expected OpenRate=0.25, got %f", r)
	}
}

func TestOpenRate_FullOpen(t *testing.T) {
	c := Campaign{SentCount: 50, OpenCount: 50}
	if r := c.OpenRate(); r != 1.0 {
		t.Errorf("expected OpenRate=1.0, got %f", r)
	}
}

func TestHasAudience_WithSegment(t *testing.T) {
	id := int64(42)
	c := Campaign{AudienceSegmentID: &id}
	if !c.HasAudience() {
		t.Error("expected HasAudience=true when segment is set")
	}
}

func TestHasAudience_WithoutSegment(t *testing.T) {
	c := Campaign{AudienceSegmentID: nil}
	if c.HasAudience() {
		t.Error("expected HasAudience=false when no segment")
	}
}

func TestValidateCampaignInput_EmptyName(t *testing.T) {
	err := ValidateCampaignInput(CampaignInput{Name: "", Type: "email"})
	if err == nil {
		t.Error("expected validation error for empty name")
	}
}

func TestValidateCampaignInput_EmptyType(t *testing.T) {
	err := ValidateCampaignInput(CampaignInput{Name: "Summer", Type: ""})
	if err == nil {
		t.Error("expected validation error for empty type")
	}
}

func TestValidateCampaignInput_Valid(t *testing.T) {
	err := ValidateCampaignInput(CampaignInput{Name: "Summer Sale", Type: "email"})
	if err != nil {
		t.Errorf("expected no validation error, got %v", err)
	}
}

func TestValidateCampaignStatus(t *testing.T) {
	for _, status := range []string{StatusDraft, StatusScheduled, StatusSent, StatusActive, StatusPaused, StatusCompleted, StatusCancelled} {
		if err := ValidateCampaignStatus(status); err != nil {
			t.Errorf("ValidateCampaignStatus(%q) = %v", status, err)
		}
	}
	if err := ValidateCampaignStatus("bogus"); !errors.Is(err, ErrInvalidCampaignStatus) {
		t.Fatalf("expected invalid status error, got %v", err)
	}
}
