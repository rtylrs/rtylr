package handlers

import (
	"regexp"
	"strings"
)

const answerQualityHoldBytes = 96

var answerQualityWordReplacements = []struct {
	pattern *regexp.Regexp
	value   string
}{
	{regexp.MustCompile(`(?i)\byour tenant\b`), "your business"},
	{regexp.MustCompile(`(?i)\btenant\b`), "business"},
	{regexp.MustCompile(`(?i)\brows returned\b`), "results"},
	{regexp.MustCompile(`(?i)\brow count\b`), "result count"},
	{regexp.MustCompile(`(?i)\bserver-validated\b`), "checked"},
	{regexp.MustCompile(`(?i)\bread-only analytics query\b`), "business data"},
	{regexp.MustCompile(`(?i)\banalytics query\b`), "business data"},
	{regexp.MustCompile(`(?i)\bgrounding\b`), "context"},
}

var answerQualityPhraseReplacements = []struct {
	old string
	new string
}{
	{"Based on the provided analytics query, here are the results:", ""},
	{"Based on the provided analytics query, here are", "Here are"},
	{"Based on the provided analytics query,", ""},
	{"Based on the analytics query,", ""},
	{"Based on the provided query,", ""},
	{"Based on the data provided,", ""},
	{"Based on the query results,", ""},
	{"To answer your question, I ran a server-validated, read-only analytics query scoped to your tenant.", "I checked your business data."},
	{"To answer your question, I ran a read-only analytics query scoped to your tenant.", "I checked your business data."},
	{"To answer your question, I ran a checked, business data scoped to your business.", "I checked your business data."},
	{"To answer your question, I ran a business data scoped to your business.", "I checked your business data."},
	{"To answer your question, I ran", "I checked"},
	{"The query results are:", ""},
	{"Invalid tool call marker ignored.", ""},
	{"the revenue value is null", "I do not see revenue yet"},
	{"the value is null", "I do not see a value yet"},
}

func polishCaretAnswerText(text string) string {
	if text == "" {
		return ""
	}

	out := text
	for _, replacement := range answerQualityPhraseReplacements {
		out = strings.ReplaceAll(out, replacement.old, replacement.new)
		out = strings.ReplaceAll(out, strings.ToLower(replacement.old), replacement.new)
	}
	for _, replacement := range answerQualityWordReplacements {
		out = replacement.pattern.ReplaceAllString(out, replacement.value)
	}

	out = strings.ReplaceAll(out, "  ", " ")
	out = strings.ReplaceAll(out, "\n \n", "\n\n")
	return out
}
