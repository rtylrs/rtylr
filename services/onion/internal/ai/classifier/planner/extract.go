package planner

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Entity is one "Label: value" pair extracted from a source document.
type Entity struct {
	Label string
	Value string
}

// ExtractEntityLimit caps how many candidates travel into scoring.
const ExtractEntityLimit = 128

const (
	extractLabelBytes = 80
	extractValueBytes = 200
	fieldLabelBytes   = 80
	fieldValueBytes   = 200
	fieldRoleBytes    = 32
	fieldIDBytes      = 64
	formNameBytes     = 200
)

// ExtractEntities parses "Label: value" lines, strips bullet prefixes,
// deduplicates pairs, and caps the result at limit (or ExtractEntityLimit
// when limit <= 0).
func ExtractEntities(document string, limit int) []Entity {
	if limit <= 0 {
		limit = ExtractEntityLimit
	}
	seen := make(map[string]struct{}, 32)
	out := make([]Entity, 0, 32)
	for _, rawLine := range strings.Split(document, "\n") {
		line := strings.TrimSpace(strings.TrimSuffix(rawLine, "\r"))
		line = strings.TrimSpace(strings.TrimLeft(line, "-*•"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		label, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		label = trimToBytes(strings.TrimSpace(label), extractLabelBytes)
		value = trimToBytes(strings.TrimSpace(value), extractValueBytes)
		if IsURLNoise(label, value) {
			continue
		}
		if label == "" || value == "" || !hasLetter(label) {
			continue
		}
		key := strings.ToLower(label) + "\x00" + strings.ToLower(value)
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, Entity{Label: label, Value: value})
		if len(out) >= limit {
			break
		}
	}
	return out
}

// NormalizeEntity trims an entity to the extraction byte caps, applies the
// same URL-noise filter as document extraction, and reports whether it
// remains usable.
func NormalizeEntity(e Entity) (Entity, bool) {
	e.Label = trimToBytes(strings.TrimSpace(e.Label), extractLabelBytes)
	e.Value = trimToBytes(strings.TrimSpace(e.Value), extractValueBytes)
	if IsURLNoise(e.Label, e.Value) {
		return Entity{}, false
	}
	if e.Label == "" || e.Value == "" || !hasLetter(e.Label) {
		return Entity{}, false
	}
	return e, true
}

// NormalizeForm trims a form name to the context length bound.
func NormalizeForm(name string) string {
	return trimToBytes(strings.TrimSpace(name), formNameBytes)
}

// NormalizeField trims a form field to bounded lengths and reports whether
// it remains usable.
func NormalizeField(f Field) (Field, bool) {
	f.ID = trimToBytes(strings.TrimSpace(f.ID), fieldIDBytes)
	f.Label = trimToBytes(strings.TrimSpace(f.Label), fieldLabelBytes)
	f.Role = trimToBytes(strings.TrimSpace(f.Role), fieldRoleBytes)
	f.Value = trimToBytes(strings.TrimSpace(f.Value), fieldValueBytes)
	if f.Label == "" {
		return Field{}, false
	}
	return f, true
}

// IsURLNoise reports whether a label/value pair is a link fragment or
// prose that embeds a URL, so neither extracted text nor pre-extracted
// entities smuggle links into form fields. A clean URL value (no
// surrounding prose) is allowed, since website-style fields are legitimate
// targets.
func IsURLNoise(label, value string) bool {
	lower := strings.ToLower(label)
	if strings.HasPrefix(value, "//") || strings.HasSuffix(lower, "http") || strings.HasSuffix(lower, "https") {
		return true
	}
	return strings.Contains(value, "://") && strings.ContainsAny(value, " \t")
}

func trimToBytes(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	cut := limit
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return strings.TrimSpace(s[:cut])
}

func hasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}
