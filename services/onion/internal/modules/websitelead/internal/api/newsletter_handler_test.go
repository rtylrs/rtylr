package api

import (
	"context"
	"strings"
	"testing"
)

func validNewsletter() *NewsletterRequest {
	return &NewsletterRequest{Email: "rania@example.com", Source: "intel_drop"}
}

func TestNewsletterHandler_Validate_HappyPathEverySource(t *testing.T) {
	h := &NewsletterHandler{}
	for _, src := range []string{"intel_drop", "report_banner", "article_pdf_mid", "article_pdf_end"} {
		req := validNewsletter()
		req.Source = src
		if err := h.Validate(context.Background(), req); err != nil {
			t.Errorf("source %q: unexpected error %v", src, err)
		}
	}
}

func TestNewsletterHandler_Validate_Rejects(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*NewsletterRequest)
		want   string
	}{
		{"missing email", func(r *NewsletterRequest) { r.Email = "" }, "email is required"},
		{"malformed email", func(r *NewsletterRequest) { r.Email = "not-an-email" }, "email is invalid"},
		{"email without tld", func(r *NewsletterRequest) { r.Email = "a@b" }, "email is invalid"},
		{"email too long", func(r *NewsletterRequest) { r.Email = strings.Repeat("a", 250) + "@x.com" }, "email is invalid"},
		{"missing source", func(r *NewsletterRequest) { r.Source = "" }, "source is required"},
		{"unknown source", func(r *NewsletterRequest) { r.Source = "web3forms" }, "source is invalid"},
		{"bad language", func(r *NewsletterRequest) { r.Language = "fr" }, "language is invalid"},
		{"slug too long", func(r *NewsletterRequest) { r.PostSlug = strings.Repeat("s", 256) }, "post_slug is too long"},
		{"title too long", func(r *NewsletterRequest) { r.PostTitle = strings.Repeat("t", 256) }, "post_title is too long"},
		{"category too long", func(r *NewsletterRequest) { r.PostCategory = strings.Repeat("c", 101) }, "post_category is too long"},
	}
	h := &NewsletterHandler{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := validNewsletter()
			tc.mutate(req)
			err := h.Validate(context.Background(), req)
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.want)
			}
		})
	}
}

func TestNewsletterHandler_Validate_DefaultsLanguageAndTrims(t *testing.T) {
	h := &NewsletterHandler{}
	req := &NewsletterRequest{Email: "  Rania@Example.com ", Source: "intel_drop"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if req.Email != "rania@example.com" {
		t.Fatalf("email not normalised: %q", req.Email)
	}
	if req.Language != "en" {
		t.Fatalf("language not defaulted: %q", req.Language)
	}
}
