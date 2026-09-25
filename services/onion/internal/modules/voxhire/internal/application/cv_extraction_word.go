package application

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
)

func extractWordXMLText(r io.Reader) (string, error) {
	decoder := xml.NewDecoder(r)
	var out strings.Builder
	for {
		tok, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return out.String(), err
		}
		if t, ok := tok.(xml.CharData); ok {
			remaining := maxCVExtractedTextBytes - out.Len()
			if remaining <= 0 {
				break
			}
			if len(t) > remaining {
				out.Write([]byte(t[:remaining]))
				break
			}
			out.Write([]byte(t))
			out.WriteByte(' ')
		}
	}
	return out.String(), nil
}

func normalizeCVText(text string) string {
	return strings.TrimSpace(whitespaceRE.ReplaceAllString(text, " "))
}

func summarizeCVText(text string) string {
	runes := []rune(text)
	if len(runes) > maxCVSummaryRunes {
		runes = runes[:maxCVSummaryRunes]
	}
	return strings.TrimSpace(string(runes))
}

func buildCVProfile(text string) domain.CVProfile {
	lower := strings.ToLower(text)
	profile := domain.CVProfile{
		Skills:          matchCVKeywords(lower, cvSkillKeywords),
		Functions:       matchCVKeywords(lower, cvFunctionKeywords),
		Industries:      matchCVKeywords(lower, cvIndustryKeywords),
		RoleTitles:      matchCVKeywords(lower, cvRoleKeywords),
		Education:       matchCVKeywords(lower, cvEducationKeywords),
		Languages:       matchCVKeywords(lower, cvLanguageKeywords),
		Certifications:  matchCVKeywords(lower, cvCertificationKeywords),
		Locations:       matchCVKeywords(lower, cvLocationKeywords),
		Availability:    detectAvailability(lower),
		SalarySignals:   detectSalarySignals(text),
		YearsExperience: detectYearsExperience(text),
		Seniority:       detectSeniority(lower),
	}
	return profile
}

func matchCVKeywords(lowerText string, keywords []cvKeyword) []string {
	out := make([]string, 0)
	for _, keyword := range keywords {
		for _, term := range keyword.Terms {
			pattern := cvKeywordPatterns[strings.ToLower(strings.TrimSpace(term))]
			if pattern != nil && pattern.MatchString(lowerText) {
				out = append(out, keyword.Label)
				break
			}
		}
	}
	return uniqueSorted(out)
}

func detectSeniority(lowerText string) string {
	switch {
	case strings.Contains(lowerText, "chief ") || strings.Contains(lowerText, "director") || strings.Contains(lowerText, "head of"):
		return "Executive"
	case strings.Contains(lowerText, "lead ") || strings.Contains(lowerText, "manager") || strings.Contains(lowerText, "senior"):
		return "Senior"
	case strings.Contains(lowerText, "mid-level") || strings.Contains(lowerText, "mid level"):
		return "Mid-level"
	case strings.Contains(lowerText, "junior") || strings.Contains(lowerText, "intern") || strings.Contains(lowerText, "entry level"):
		return "Junior"
	default:
		return ""
	}
}

func detectYearsExperience(text string) string {
	matches := yearsExperienceRE.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return ""
	}
	maxYears := 0
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		var years int
		if _, err := fmt.Sscanf(match[1], "%d", &years); err == nil && years > maxYears {
			maxYears = years
		}
	}
	if maxYears == 0 {
		return ""
	}
	return fmt.Sprintf("%d+ years", maxYears)
}

func detectAvailability(lowerText string) string {
	switch {
	case strings.Contains(lowerText, "immediately available") || strings.Contains(lowerText, "available immediately"):
		return "Immediate"
	case strings.Contains(lowerText, "2 weeks") || strings.Contains(lowerText, "two weeks"):
		return "2 weeks"
	case strings.Contains(lowerText, "1 month") || strings.Contains(lowerText, "one month"):
		return "1 month"
	case strings.Contains(lowerText, "notice period"):
		return "Notice period mentioned"
	default:
		return ""
	}
}

func detectSalarySignals(text string) []string {
	matches := salarySignalRE.FindAllString(text, 8)
	return uniqueSorted(matches)
}

func uniqueSorted(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func truncateBytes(text string, maxBytes int) string {
	if len(text) <= maxBytes {
		return text
	}
	cut := 0
	for i := range text {
		if i > maxBytes {
			break
		}
		cut = i
	}
	if cut == 0 {
		return ""
	}
	return strings.TrimSpace(text[:cut])
}
