package domain

import "testing"

func TestIsAllowedNewsletterSource(t *testing.T) {
	for _, ok := range []string{"intel_drop", "report_banner", "article_pdf_mid", "article_pdf_end"} {
		if !IsAllowedNewsletterSource(ok) {
			t.Errorf("%q should be allowed", ok)
		}
	}
	for _, bad := range []string{"", "web3forms", "INTEL_DROP", "article_pdf"} {
		if IsAllowedNewsletterSource(bad) {
			t.Errorf("%q should be rejected", bad)
		}
	}
}

func TestIsAllowedLanguage(t *testing.T) {
	if !IsAllowedLanguage("en") || !IsAllowedLanguage("ar") {
		t.Fatal("en and ar must be allowed")
	}
	if IsAllowedLanguage("") || IsAllowedLanguage("fr") || IsAllowedLanguage("EN") {
		t.Fatal("only lowercase en and ar are allowed")
	}
}
