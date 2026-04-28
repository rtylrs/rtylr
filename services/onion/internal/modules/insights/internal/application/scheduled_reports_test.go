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
	for _, c := range []string{"email", "slack"} {
		if !validChannels[c] {
			t.Errorf("expected channel %q to be valid", c)
		}
	}
}
