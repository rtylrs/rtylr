package domain

import (
	"errors"
	"strings"
)

var (
	// ErrSlugReserved indicates that a business slug matches a reserved system or routing word.
	ErrSlugReserved = errors.New("this slug is reserved")
)

var reservedBusinessSlugs = map[string]struct{}{
	"automations":   {},
	"webhooks":      {},
	"menus":         {},
	"menu":          {},
	"create-item":   {},
	"create-banner": {},
	"orders":        {},
	"qr-codes":      {},
	"design":        {},
	"chat":          {},
	"caret":         {},
	"team":          {},
	"roles":         {},
	"audit":         {},
	"notifications": {},
	"account":       {},
	"billing":       {},
	"theme":         {},
	"help":          {},
	"extensions":    {},
	"store":         {},
	"integrations":  {},
	"insights":      {},
	"flow":          {},
	"crm":           {},
	"erp":           {},
	"finance":       {},
	"hr":            {},
	"recruit":       {},
	"settings":      {},
	"security":      {},
	"api-keys":      {},
	"pos":           {},
	"b":             {},
	"api":           {},
	"auth":          {},
	"login":         {},
	"signup":        {},
	"assets":        {},
	"static":        {},
	"public":        {},
	"admin":         {},
	"www":           {},
	"app":           {},
	"portal":        {},
}

// ReservedBusinessSlug reports whether the given slug is a reserved business slug.
// Comparison is case-insensitive after trimming surrounding whitespace.
func ReservedBusinessSlug(slug string) bool {
	normalized := strings.ToLower(strings.TrimSpace(slug))
	if normalized == "" {
		return false
	}
	_, ok := reservedBusinessSlugs[normalized]
	return ok
}

// ReservedBusinessSlugsList returns a slice of all reserved business slug entries.
func ReservedBusinessSlugsList() []string {
	return []string{
		"automations", "webhooks", "menus", "menu", "create-item", "create-banner",
		"orders", "qr-codes", "design", "chat", "caret", "team", "roles", "audit",
		"notifications", "account", "billing", "theme", "help", "extensions",
		"store", "integrations", "insights", "flow", "crm", "erp", "finance",
		"hr", "recruit", "settings", "security", "api-keys", "pos", "b", "api",
		"auth", "login", "signup", "assets", "static", "public", "admin", "www",
		"app", "portal",
	}
}
