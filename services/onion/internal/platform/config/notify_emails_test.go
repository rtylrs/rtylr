package config

import (
	"reflect"
	"testing"
)

func TestParseEmailList(t *testing.T) {
	fallback := []string{"info@voxire.com"}
	cases := []struct {
		name string
		raw  string
		want []string
	}{
		{"unset uses fallback", "", fallback},
		{"whitespace only uses fallback", " \t ", fallback},
		{"single", "a@voxire.com", []string{"a@voxire.com"}},
		{"trims and lowercases", " A@Voxire.com , b@voxire.com ", []string{"a@voxire.com", "b@voxire.com"}},
		{"drops empties and duplicates", "a@voxire.com,,a@voxire.com,", []string{"a@voxire.com"}},
		{"keeps order of first appearance", "b@voxire.com,a@voxire.com,b@voxire.com", []string{"b@voxire.com", "a@voxire.com"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseEmailList(tc.raw, fallback)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("ParseEmailList(%q) = %v, want %v", tc.raw, got, tc.want)
			}
		})
	}
}

func TestParseEmailListReturnsCopyOfFallback(t *testing.T) {
	fallback := []string{"info@voxire.com"}
	got := ParseEmailList("", fallback)
	got[0] = "changed"
	if fallback[0] != "info@voxire.com" {
		t.Fatal("fallback slice was mutated through the returned list")
	}
}

func TestDefaultNotifyLists(t *testing.T) {
	wantLead := []string{
		"info@voxire.com",
		"sales@voxire.com",
		"mohammad.saab@voxire.com",
		"mohamad.homsi@voxire.com",
		"amounehabed@gmail.com",
	}
	if !reflect.DeepEqual(DefaultLeadNotifyEmails, wantLead) {
		t.Fatalf("DefaultLeadNotifyEmails = %v", DefaultLeadNotifyEmails)
	}
	if !reflect.DeepEqual(DefaultNewsletterNotifyEmails, []string{"info@voxire.com"}) {
		t.Fatalf("DefaultNewsletterNotifyEmails = %v", DefaultNewsletterNotifyEmails)
	}
}

func TestLoadAndValidateReadsNotifyLists(t *testing.T) {
	setLicenseConfigEnv(t)
	t.Setenv("LEAD_NOTIFY_EMAILS", "x@voxire.com, y@voxire.com")
	t.Setenv("NEWSLETTER_NOTIFY_EMAILS", "")
	cfg, err := LoadAndValidate()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg.LeadNotifyEmails, []string{"x@voxire.com", "y@voxire.com"}) {
		t.Fatalf("LeadNotifyEmails = %v", cfg.LeadNotifyEmails)
	}
	if !reflect.DeepEqual(cfg.NewsletterNotifyEmails, DefaultNewsletterNotifyEmails) {
		t.Fatalf("NewsletterNotifyEmails = %v", cfg.NewsletterNotifyEmails)
	}
}
