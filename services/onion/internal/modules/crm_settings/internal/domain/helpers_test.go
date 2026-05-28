package domain

import "testing"

func TestHasFromEmail_WhenSet(t *testing.T) {
	s := Settings{FromEmail: "no-reply@example.com"}
	if !s.HasFromEmail() {
		t.Error("expected HasFromEmail=true when from_email is set")
	}
}

func TestHasFromEmail_WhenEmpty(t *testing.T) {
	s := Settings{FromEmail: ""}
	if s.HasFromEmail() {
		t.Error("expected HasFromEmail=false when from_email is empty")
	}
}

func TestHasFromName_WhenSet(t *testing.T) {
	s := Settings{FromName: "Acme Support"}
	if !s.HasFromName() {
		t.Error("expected HasFromName=true when from_name is set")
	}
}

func TestHasFromName_WhenEmpty(t *testing.T) {
	s := Settings{FromName: ""}
	if s.HasFromName() {
		t.Error("expected HasFromName=false when from_name is empty")
	}
}

func TestHasSmsSenderID_WhenSet(t *testing.T) {
	s := Settings{SmsSenderID: "ACME"}
	if !s.HasSmsSenderID() {
		t.Error("expected HasSmsSenderID=true when sms_sender_id is set")
	}
}

func TestHasSmsSenderID_WhenEmpty(t *testing.T) {
	s := Settings{SmsSenderID: ""}
	if s.HasSmsSenderID() {
		t.Error("expected HasSmsSenderID=false when sms_sender_id is empty")
	}
}

func TestHasTimezone_WhenSet(t *testing.T) {
	s := Settings{Timezone: "America/New_York"}
	if !s.HasTimezone() {
		t.Error("expected HasTimezone=true when timezone is set")
	}
}

func TestHasTimezone_WhenEmpty(t *testing.T) {
	s := Settings{Timezone: ""}
	if s.HasTimezone() {
		t.Error("expected HasTimezone=false when timezone is empty")
	}
}

func TestConsentRequired_WhenDefaultConsentFalse(t *testing.T) {
	s := Settings{DefaultConsent: false}
	if !s.ConsentRequired() {
		t.Error("expected ConsentRequired=true when default_consent is false")
	}
}

func TestConsentRequired_WhenDefaultConsentTrue(t *testing.T) {
	s := Settings{DefaultConsent: true}
	if s.ConsentRequired() {
		t.Error("expected ConsentRequired=false when default_consent is true")
	}
}

func TestValidateSettingsInput_Valid(t *testing.T) {
	msg := ValidateSettingsInput(SettingsInput{OwnerID: 1})
	if msg != "" {
		t.Errorf("expected no error, got %q", msg)
	}
}

func TestValidateSettingsInput_ZeroOwner(t *testing.T) {
	msg := ValidateSettingsInput(SettingsInput{OwnerID: 0})
	if msg == "" {
		t.Error("expected error for zero owner_id")
	}
}
