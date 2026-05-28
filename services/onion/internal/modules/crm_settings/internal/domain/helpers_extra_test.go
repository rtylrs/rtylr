package domain

import "testing"

func TestIsFullyConfigured_WhenAllFieldsSet(t *testing.T) {
	s := Settings{FromEmail: "a@b.com", FromName: "Shop", Timezone: "UTC"}
	if !s.IsFullyConfigured() {
		t.Error("expected IsFullyConfigured=true when all fields are set")
	}
}

func TestIsFullyConfigured_WhenEmailMissing(t *testing.T) {
	s := Settings{FromEmail: "", FromName: "Shop", Timezone: "UTC"}
	if s.IsFullyConfigured() {
		t.Error("expected IsFullyConfigured=false when email is missing")
	}
}

func TestIsFullyConfigured_WhenFromNameMissing(t *testing.T) {
	s := Settings{FromEmail: "a@b.com", FromName: "", Timezone: "UTC"}
	if s.IsFullyConfigured() {
		t.Error("expected IsFullyConfigured=false when from name is missing")
	}
}

func TestIsFullyConfigured_WhenTimezoneMissing(t *testing.T) {
	s := Settings{FromEmail: "a@b.com", FromName: "Shop", Timezone: ""}
	if s.IsFullyConfigured() {
		t.Error("expected IsFullyConfigured=false when timezone is missing")
	}
}

func TestHasSmsCapability_WhenSet(t *testing.T) {
	s := Settings{SmsSenderID: "BRAND"}
	if !s.HasSmsCapability() {
		t.Error("expected HasSmsCapability=true when SMS sender ID is set")
	}
}

func TestHasSmsCapability_WhenEmpty(t *testing.T) {
	s := Settings{SmsSenderID: ""}
	if s.HasSmsCapability() {
		t.Error("expected HasSmsCapability=false when SMS sender ID is empty")
	}
}

func TestConsentRequired_AndIsFullyConfigured_AreIndependent(t *testing.T) {
	// Consent and full configuration are independent concerns
	s := Settings{
		FromEmail:      "a@b.com",
		FromName:       "Shop",
		Timezone:       "UTC",
		DefaultConsent: true,
	}
	if !s.IsFullyConfigured() {
		t.Error("expected IsFullyConfigured=true")
	}
	if s.ConsentRequired() {
		t.Error("expected ConsentRequired=false when DefaultConsent=true")
	}
}
