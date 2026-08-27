package application

import (
	"strings"
	"testing"
)

func TestValidateScheduledReport_OK(t *testing.T) {
	p := ScheduledReportPayload{
		TemplateID: "weekly_sales",
		Cadence:    "weekly",
		Channel:    "email",
		Recipient:  "owner@example.com",
	}
	if err := validateScheduledReport(p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateScheduledReport_RejectsMissingFields(t *testing.T) {
	cases := []struct {
		name  string
		p     ScheduledReportPayload
		wants string
	}{
		{
			name:  "missing template_id",
			p:     ScheduledReportPayload{Cadence: "daily", Channel: "email", Recipient: "a@b"},
			wants: "template_id",
		},
		{
			name:  "invalid cadence",
			p:     ScheduledReportPayload{TemplateID: "t", Cadence: "yearly", Channel: "email", Recipient: "a@b"},
			wants: "cadence",
		},
		{
			name:  "invalid channel",
			p:     ScheduledReportPayload{TemplateID: "t", Cadence: "daily", Channel: "sms", Recipient: "a@b"},
			wants: "channel",
		},
		{
			name:  "missing recipient",
			p:     ScheduledReportPayload{TemplateID: "t", Cadence: "daily", Channel: "email"},
			wants: "recipient",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateScheduledReport(tc.p)
			if err == nil {
				t.Fatalf("expected error mentioning %q", tc.wants)
			}
			if !strings.Contains(err.Error(), tc.wants) {
				t.Fatalf("expected error to mention %q, got: %v", tc.wants, err)
			}
		})
	}
}

func TestValidCadencesAndChannelsHaveExpectedKeys(t *testing.T) {
	for _, c := range []string{"daily", "weekly", "monthly"} {
		if !validCadences[c] {
			t.Errorf("expected cadence %q to be valid", c)
		}
	}
	for _, c := range []string{"email"} {
		if !validChannels[c] {
			t.Errorf("expected channel %q to be valid", c)
		}
	}
}

func TestValidateScheduledReport_AllCadences(t *testing.T) {
	for cadence := range validCadences {
		p := ScheduledReportPayload{
			TemplateID: "tmpl",
			Cadence:    cadence,
			Channel:    "email",
			Recipient:  "user@test.com",
		}
		if err := validateScheduledReport(p); err != nil {
			t.Errorf("cadence %q should be valid, got %v", cadence, err)
		}
	}
}

func TestValidateScheduledReport_AllChannels(t *testing.T) {
	for channel := range validChannels {
		p := ScheduledReportPayload{
			TemplateID: "tmpl",
			Cadence:    "weekly",
			Channel:    channel,
			Recipient:  "user@test.com",
		}
		if err := validateScheduledReport(p); err != nil {
			t.Errorf("channel %q should be valid, got %v", channel, err)
		}
	}
}

func TestValidateScheduledReport_EmptyCadence(t *testing.T) {
	p := ScheduledReportPayload{TemplateID: "t", Cadence: "", Channel: "email", Recipient: "a@b"}
	err := validateScheduledReport(p)
	if err == nil || !strings.Contains(err.Error(), "cadence") {
		t.Fatalf("empty cadence should fail with cadence mention, got %v", err)
	}
}

func TestValidateScheduledReport_EmptyChannel(t *testing.T) {
	p := ScheduledReportPayload{TemplateID: "t", Cadence: "daily", Channel: "", Recipient: "a@b"}
	err := validateScheduledReport(p)
	if err == nil || !strings.Contains(err.Error(), "channel") {
		t.Fatalf("empty channel should fail with channel mention, got %v", err)
	}
}

func TestValidateScheduledReport_AllFieldsPresent(t *testing.T) {
	p := ScheduledReportPayload{
		TemplateID: "quarterly_revenue",
		Cadence:    "monthly",
		Channel:    "email",
		Recipient:  "alerts@company.com",
		IsActive:   true,
	}
	if err := validateScheduledReport(p); err != nil {
		t.Fatalf("all valid fields should pass, got %v", err)
	}
}

func TestValidateScheduledReport_CaseSensitiveCadence(t *testing.T) {
	p := ScheduledReportPayload{TemplateID: "t", Cadence: "Daily", Channel: "email", Recipient: "a@b"}
	err := validateScheduledReport(p)
	if err == nil {
		t.Fatal("cadence matching should be case-sensitive")
	}
}

func TestValidateScheduledReport_CaseSensitiveChannel(t *testing.T) {
	p := ScheduledReportPayload{TemplateID: "t", Cadence: "daily", Channel: "Email", Recipient: "a@b"}
	err := validateScheduledReport(p)
	if err == nil {
		t.Fatal("channel matching should be case-sensitive")
	}
}

func TestValidCadences_DoesNotContainYearly(t *testing.T) {
	if validCadences["yearly"] {
		t.Error("yearly should not be a valid cadence")
	}
}

func TestValidChannels_DoesNotContainUnsupportedDeliveryChannels(t *testing.T) {
	for _, channel := range []string{"sms", "slack"} {
		if validChannels[channel] {
			t.Errorf("%s should not be a valid channel without a delivery worker", channel)
		}
	}
}

func TestValidateScheduledReport_MissingRecipientAndTemplate(t *testing.T) {
	p := ScheduledReportPayload{Cadence: "daily", Channel: "email"}
	err := validateScheduledReport(p)
	if err == nil {
		t.Fatal("missing template_id should fail")
	}
	if !strings.Contains(err.Error(), "template_id") {
		t.Fatalf("expected error mentioning template_id, got %v", err)
	}
}
