package domain_test

import (
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/business/internal/domain"
)

func TestReservedBusinessSlug_AllListEntries(t *testing.T) {
	expectedEntries := domain.ReservedBusinessSlugsList()
	if len(expectedEntries) != 45 {
		t.Fatalf("expected 45 reserved slugs, got %d", len(expectedEntries))
	}

	for _, slug := range expectedEntries {
		t.Run(slug, func(t *testing.T) {
			if !domain.ReservedBusinessSlug(slug) {
				t.Errorf("ReservedBusinessSlug(%q) = false, want true", slug)
			}
		})
	}
}

func TestReservedBusinessSlug_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		slug     string
		reserved bool
	}{
		// Normal slugs accepted
		{name: "normal slug acme", slug: "acme", reserved: false},
		{name: "normal slug joes-pizza", slug: "joes-pizza", reserved: false},
		{name: "normal slug bistro123", slug: "bistro123", reserved: false},

		// 'menu' vs 'menus' vs 'my-menu'
		{name: "exact menu", slug: "menu", reserved: true},
		{name: "exact menus", slug: "menus", reserved: true},
		{name: "my-menu is not reserved", slug: "my-menu", reserved: false},
		{name: "menu-bar is not reserved", slug: "menu-bar", reserved: false},

		// Case variants
		{name: "uppercase MENU", slug: "MENU", reserved: true},
		{name: "mixed case MeNu", slug: "MeNu", reserved: true},
		{name: "uppercase PORTAL", slug: "PORTAL", reserved: true},
		{name: "mixed case OrDeRs", slug: "OrDeRs", reserved: true},
		{name: "uppercase SETTINGS", slug: "SETTINGS", reserved: true},

		// Whitespace variants
		{name: "leading whitespace", slug: "  menu", reserved: true},
		{name: "trailing whitespace", slug: "menu  ", reserved: true},
		{name: "surrounding whitespace and tabs", slug: "\t portal \n", reserved: true},
		{name: "whitespace around normal slug", slug: "  acme-coffee  ", reserved: false},

		// Empty and blank
		{name: "empty string", slug: "", reserved: false},
		{name: "only whitespace", slug: "   ", reserved: false},

		// Specific list entries with case/whitespace variations
		{name: "automations uppercase with spaces", slug: "  AUTOMATIONS  ", reserved: true},
		{name: "webhooks mixed case", slug: "WebHooks", reserved: true},
		{name: "create-item uppercase", slug: "CREATE-ITEM", reserved: true},
		{name: "create-banner trailing newline", slug: "create-banner\n", reserved: true},
		{name: "qr-codes uppercase", slug: "QR-CODES", reserved: true},
		{name: "b single letter uppercase", slug: "B", reserved: true},
		{name: "api uppercase", slug: "API", reserved: true},
		{name: "auth uppercase", slug: "AUTH", reserved: true},
		{name: "login uppercase", slug: "LOGIN", reserved: true},
		{name: "signup uppercase", slug: "SIGNUP", reserved: true},
		{name: "www uppercase", slug: "WWW", reserved: true},
		{name: "app uppercase", slug: "APP", reserved: true},
		{name: "portal uppercase", slug: "PORTAL", reserved: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := domain.ReservedBusinessSlug(tc.slug)
			if got != tc.reserved {
				t.Errorf("ReservedBusinessSlug(%q) = %v, want %v", tc.slug, got, tc.reserved)
			}
		})
	}
}

func TestReservedBusinessSlug_EveryListEntryWithVariants(t *testing.T) {
	for _, entry := range domain.ReservedBusinessSlugsList() {
		// Uppercase
		upper := strings.ToUpper(entry)
		if !domain.ReservedBusinessSlug(upper) {
			t.Errorf("ReservedBusinessSlug(%q) = false, want true for uppercase", upper)
		}

		// With whitespace
		withSpaces := "  " + entry + "  "
		if !domain.ReservedBusinessSlug(withSpaces) {
			t.Errorf("ReservedBusinessSlug(%q) = false, want true for whitespace", withSpaces)
		}
	}
}
