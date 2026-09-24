package insights

import (
	"strings"
	"testing"
)

func TestParseScheduleReportToolInputNormalizesAliases(t *testing.T) {
	got, err := parseScheduleReportToolInput(map[string]any{
		"report":    "daily revenue",
		"frequency": "every day",
		"email":     "ops@example.com",
	})
	if err != nil {
		t.Fatalf("parseScheduleReportToolInput returned error: %v", err)
	}
	if got.TemplateID != "revenue-daily" {
		t.Fatalf("TemplateID = %q, want revenue-daily", got.TemplateID)
	}
	if got.Cadence != "daily" {
		t.Fatalf("Cadence = %q, want daily", got.Cadence)
	}
	if got.Channel != "email" {
		t.Fatalf("Channel = %q, want email", got.Channel)
	}
	if got.Recipient != "ops@example.com" {
		t.Fatalf("Recipient = %q, want ops@example.com", got.Recipient)
	}
	if got.IsActive != nil {
		t.Fatalf("IsActive = %v, want nil so persistence applies the active default", *got.IsActive)
	}
}

func TestParseScheduleReportToolInputSupportsSlackRecipientArray(t *testing.T) {
	got, err := parseScheduleReportToolInput(map[string]any{
		"template":   "top customers",
		"cadence":    "weekly",
		"recipients": []any{"#sales"},
		"is_active":  false,
	})
	if err != nil {
		t.Fatalf("parseScheduleReportToolInput returned error: %v", err)
	}
	if got.TemplateID != "top-customers" {
		t.Fatalf("TemplateID = %q, want top-customers", got.TemplateID)
	}
	if got.Cadence != "weekly" {
		t.Fatalf("Cadence = %q, want weekly", got.Cadence)
	}
	if got.Channel != "slack" {
		t.Fatalf("Channel = %q, want slack", got.Channel)
	}
	if got.Recipient != "#sales" {
		t.Fatalf("Recipient = %q, want #sales", got.Recipient)
	}
	if got.IsActive == nil || *got.IsActive {
		t.Fatalf("IsActive = %v, want explicit false", got.IsActive)
	}
}

func TestParseScheduleReportToolInputRejectsAmbiguousRecipient(t *testing.T) {
	_, err := parseScheduleReportToolInput(map[string]any{
		"template_id": "revenue-monthly",
		"cadence":     "monthly",
		"recipient":   "me",
	})
	if err == nil {
		t.Fatal("expected ambiguous recipient error")
	}
	if !strings.Contains(err.Error(), "explicit email address or Slack channel") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseScheduleReportToolInputRejectsBadEmailRecipient(t *testing.T) {
	_, err := parseScheduleReportToolInput(map[string]any{
		"template_id": "revenue-monthly",
		"cadence":     "monthly",
		"channel":     "email",
		"recipient":   "not-an-email",
	})
	if err == nil {
		t.Fatal("expected invalid email error")
	}
	if !strings.Contains(err.Error(), "valid email") {
		t.Fatalf("unexpected error: %v", err)
	}
}
