package handlers

import (
	"regexp"
	"strings"
)

type caretAnswerQuality struct {
	Score     int
	Cacheable bool
	Findings  []string
}

var unsupportedSamplePattern = regexp.MustCompile(`(?i)\b(?:john smith|jane doe|bob brown|customer id\s+c\d{3})\b`)

// assessCaretAnswer is a final server-side quality gate. It cannot prove that
// every sentence is correct, but it can prevent obvious protocol leakage and
// synthetic sample data from becoming a cached answer or a durable memory.
func assessCaretAnswer(response string) caretAnswerQuality {
	response = strings.TrimSpace(response)
	quality := caretAnswerQuality{Score: 100, Cacheable: true}
	if response == "" {
		return caretAnswerQuality{Score: 0, Cacheable: false, Findings: []string{"empty response"}}
	}
	for _, marker := range []string{"<tool>", "</tool>", "<thinking>", "</thinking>"} {
		if strings.Contains(response, marker) {
			quality.Score -= 60
			quality.Cacheable = false
			quality.Findings = append(quality.Findings, "structured marker leaked into answer")
			break
		}
	}
	if unsupportedSamplePattern.MatchString(response) {
		quality.Score -= 50
		quality.Cacheable = false
		quality.Findings = append(quality.Findings, "known synthetic sample data detected")
	}
	for _, phrase := range []string{"internal business snapshot evidence", "server request plan", "query spec", "row count", "BEGIN REFERENCE CONTEXT"} {
		if strings.Contains(strings.ToLower(response), strings.ToLower(phrase)) {
			quality.Score -= 20
			quality.Cacheable = false
			quality.Findings = append(quality.Findings, "internal process text leaked into answer")
		}
	}
	if quality.Score < 0 {
		quality.Score = 0
	}
	return quality
}
