package application

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
)

var matchTokenRE = regexp.MustCompile(`[a-z0-9]+`)

func enrichApplicationMatches(apps []domain.Application, opening *domain.Opening) []domain.Application {
	for i := range apps {
		apps[i].MatchSignal = scoreApplication(opening, apps[i])
	}
	return apps
}

func enrichCandidateMatches(candidates []domain.Candidate) []domain.Candidate {
	for i := range candidates {
		candidates[i].MatchSignal = scoreCandidate(candidates[i])
	}
	return candidates
}

func enrichShortlistMatches(shortlists []domain.Shortlist) []domain.Shortlist {
	for i := range shortlists {
		for j := range shortlists[i].Items {
			shortlists[i].Items[j].MatchSignal = scoreShortlistItem(shortlists[i], shortlists[i].Items[j])
		}
	}
	return shortlists
}

func scoreApplication(opening *domain.Opening, app domain.Application) domain.MatchSignal {
	score := 30
	strengths := make([]string, 0, 6)
	risks := make([]string, 0, 4)
	matched := make([]string, 0, 8)

	roleText := ""
	if opening != nil {
		roleText = strings.Join([]string{
			opening.Title,
			opening.Department,
			opening.Location,
			opening.WorkMode,
			opening.Seniority,
			opening.Details,
		}, " ")
	}
	candidateText := strings.Join([]string{
		app.Name,
		app.Location,
		app.ExperienceSummary,
		app.WorkHistory,
		app.Education,
		app.CoverLetter,
		app.CVSummary,
		strings.Join(app.CVProfile.Skills, " "),
		strings.Join(app.CVProfile.Functions, " "),
		strings.Join(app.CVProfile.Industries, " "),
		strings.Join(app.CVProfile.RoleTitles, " "),
		app.CVProfile.Seniority,
		app.CVProfile.YearsExperience,
	}, " ")

	roleTokens := importantTokens(roleText)
	candidateTokens := tokenSet(candidateText)
	for _, token := range roleTokens {
		if _, ok := candidateTokens[token]; ok {
			matched = append(matched, token)
		}
	}

	switch {
	case len(matched) >= 8:
		score += 28
		strengths = append(strengths, "Strong overlap with the role brief")
	case len(matched) >= 4:
		score += 20
		strengths = append(strengths, "Useful overlap with the role brief")
	case len(matched) >= 2:
		score += 12
		strengths = append(strengths, "Some role keywords match")
	default:
		risks = append(risks, "Limited obvious role overlap")
	}

	if len(app.CVProfile.Functions) > 0 {
		score += 10
		strengths = append(strengths, "CV function extracted")
	}
	if len(app.CVProfile.Skills) >= 4 {
		score += 10
		strengths = append(strengths, "Several skills identified")
	} else if len(app.CVProfile.Skills) > 0 {
		score += 5
	}
	if app.CVProfile.Seniority != "" {
		score += 6
		strengths = append(strengths, "Seniority signal present")
	}
	if app.CVProfile.YearsExperience != "" {
		score += 6
		strengths = append(strengths, app.CVProfile.YearsExperience)
	}
	if opening != nil && app.Location != "" && opening.Location != "" && locationMatches(app.Location, opening.Location) {
		score += 8
		strengths = append(strengths, "Location aligns")
	}
	if opening != nil && opening.SalaryMax > 0 && app.SalaryExpectation > 0 {
		if app.SalaryExpectation <= opening.SalaryMax {
			score += 8
			strengths = append(strengths, "Salary expectation is in range")
		} else {
			score -= 8
			risks = append(risks, "Salary expectation is above range")
		}
	}
	switch app.CVProcessingStatus {
	case "processed":
		score += 6
	case "error":
		score -= 12
		risks = append(risks, "CV extraction failed")
	default:
		score -= 5
		risks = append(risks, "CV extraction still pending")
	}
	if app.ReviewRating >= 4 {
		score += 10
		strengths = append(strengths, "Recruiter rated strong")
	} else if app.ReviewRating > 0 && app.ReviewRating <= 2 {
		score -= 10
		risks = append(risks, "Recruiter rating is low")
	}

	return buildSignal(score, strengths, risks, matched)
}

func scoreCandidate(candidate domain.Candidate) domain.MatchSignal {
	score := 25
	strengths := make([]string, 0, 5)
	risks := make([]string, 0, 3)
	matched := make([]string, 0, 6)

	if candidate.Source == "career" && candidate.OpeningTitle != "" {
		roleTokens := importantTokens(candidate.OpeningTitle)
		candidateTokens := tokenSet(strings.Join([]string{
			candidate.CurrentRole,
			candidate.DesiredRoles,
			candidate.Track,
			candidate.CVSummary,
			strings.Join(candidate.CVProfile.Skills, " "),
			strings.Join(candidate.CVProfile.Functions, " "),
			strings.Join(candidate.CVProfile.RoleTitles, " "),
		}, " "))
		for _, token := range roleTokens {
			if _, ok := candidateTokens[token]; ok {
				matched = append(matched, token)
			}
		}
		score += len(matched) * 6
	}

	if len(candidate.CVProfile.Functions) > 0 {
		score += 14
		strengths = append(strengths, "Function identified")
	}
	if len(candidate.CVProfile.Skills) >= 5 {
		score += 18
		strengths = append(strengths, "Deep skill profile")
	} else if len(candidate.CVProfile.Skills) > 0 {
		score += 9
		strengths = append(strengths, "Skills identified")
	}
	if candidate.CVProfile.Seniority != "" {
		score += 8
		strengths = append(strengths, candidate.CVProfile.Seniority)
	}
	switch candidate.CVProcessingStatus {
	case "processed":
		score += 10
	case "error":
		score -= 15
		risks = append(risks, "CV extraction failed")
	default:
		score -= 6
		risks = append(risks, "CV extraction pending")
	}
	if candidate.Status == "rejected" {
		score -= 20
		risks = append(risks, "Previously rejected")
	}

	return buildSignal(score, strengths, risks, matched)
}

func scoreShortlistItem(shortlist domain.Shortlist, item domain.ShortlistItem) domain.MatchSignal {
	score := 35
	strengths := make([]string, 0, 4)
	risks := make([]string, 0, 2)
	matched := make([]string, 0, 5)

	candidateTokens := tokenSet(strings.Join([]string{
		strings.Join(item.CVProfile.Skills, " "),
		strings.Join(item.CVProfile.Functions, " "),
		strings.Join(item.CVProfile.RoleTitles, " "),
		item.CVProfile.Seniority,
	}, " "))
	for _, token := range importantTokens(shortlist.JobTitle) {
		if _, ok := candidateTokens[token]; ok {
			matched = append(matched, token)
		}
	}
	score += len(matched) * 7
	if len(item.CVProfile.Skills) > 0 {
		score += 10
		strengths = append(strengths, "Profile ready")
	}
	if item.Rank > 0 && item.Rank <= 3 {
		score += 10
		strengths = append(strengths, fmt.Sprintf("Ranked #%d", item.Rank))
	}
	if item.Status == "rejected" {
		score -= 20
		risks = append(risks, "Rejected status")
	}
	return buildSignal(score, strengths, risks, matched)
}

func buildSignal(score int, strengths, risks, matched []string) domain.MatchSignal {
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	label := "Needs review"
	switch {
	case score >= 78:
		label = "Strong match"
	case score >= 58:
		label = "Possible match"
	case score < 40:
		label = "Weak signal"
	}
	matched = uniqueStrings(matched)
	if len(matched) > 8 {
		matched = matched[:8]
	}
	strengths = uniqueStrings(strengths)
	risks = uniqueStrings(risks)
	summary := label
	if len(strengths) > 0 {
		summary = fmt.Sprintf("%s: %s", label, strengths[0])
	}
	return domain.MatchSignal{
		Score:        score,
		Label:        label,
		Summary:      summary,
		Strengths:    clampStrings(strengths, 4),
		Risks:        clampStrings(risks, 3),
		MatchedTerms: matched,
	}
}

func importantTokens(text string) []string {
	tokens := matchTokenRE.FindAllString(strings.ToLower(text), -1)
	seen := make(map[string]struct{}, len(tokens))
	out := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if len(token) < 3 || matchStopWords[token] {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		out = append(out, token)
	}
	sort.Strings(out)
	if len(out) > 24 {
		out = out[:24]
	}
	return out
}

func tokenSet(text string) map[string]struct{} {
	tokens := matchTokenRE.FindAllString(strings.ToLower(text), -1)
	set := make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		if len(token) < 3 || matchStopWords[token] {
			continue
		}
		set[token] = struct{}{}
	}
	return set
}

func locationMatches(candidate, role string) bool {
	c := strings.ToLower(candidate)
	r := strings.ToLower(role)
	if strings.Contains(c, r) || strings.Contains(r, c) {
		return true
	}
	if strings.Contains(c, "leban") && strings.Contains(r, "leban") {
		return true
	}
	if strings.Contains(c, "remote") && strings.Contains(r, "remote") {
		return true
	}
	if strings.Contains(c, "gcc") && (strings.Contains(r, "gcc") || strings.Contains(r, "gulf")) {
		return true
	}
	return false
}

func uniqueStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
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
	return out
}

func clampStrings(values []string, limit int) []string {
	if len(values) <= limit {
		return values
	}
	return values[:limit]
}

var matchStopWords = map[string]bool{
	"and": true, "are": true, "for": true, "from": true, "has": true, "have": true,
	"into": true, "our": true, "the": true, "this": true, "that": true, "with": true,
	"your": true, "you": true, "role": true, "team": true, "work": true, "will": true,
	"candidate": true, "candidates": true, "experience": true, "years": true,
}
