package application

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/voxire/onion/internal/platform/storage"
)

const (
	cvProcessingBatchSize = 25
	maxCVExtractionBytes  = 10 << 20
	maxCVStoredTextBytes  = 128 << 10
	// maxCVExtractedTextBytes bounds the decompressed text materialized from a
	// CV before it is normalized and truncated to maxCVStoredTextBytes, so a
	// compression bomb cannot exhaust the worker's memory.
	maxCVExtractedTextBytes = 1 << 20
	// maxCVExtractedXMLBytes bounds the decompressed DOCX XML read while looking
	// for the text above. XML carries markup on top of the text, so it is
	// allowed more room than the text itself.
	maxCVExtractedXMLBytes = 4 << 20
	// maxCVExtractedPDFPages bounds how many PDF pages are inspected.
	maxCVExtractedPDFPages = 500
	maxCVSummaryRunes      = 700
)

var (
	whitespaceRE      = regexp.MustCompile(`\s+`)
	yearsExperienceRE = regexp.MustCompile(`(?i)(\d{1,2})\+?\s*(?:years?|yrs?|yoe)`)
	salarySignalRE    = regexp.MustCompile(`(?i)(?:usd|\$)\s?[\d,]+|[\d,]+\s?(?:usd|dollars?)`)
)

type cvKeyword struct {
	Label string
	Terms []string
}

var cvSkillKeywords = []cvKeyword{
	{"Excel", []string{"excel", "spreadsheet"}},
	{"Sales", []string{"sales", "business development", "account executive"}},
	{"Customer support", []string{"customer support", "customer success", "call center", "client support"}},
	{"Operations", []string{"operations", "process improvement", "workflow", "scheduling"}},
	{"HR", []string{"human resources", "hr", "recruiting", "recruitment", "talent acquisition"}},
	{"Finance", []string{"finance", "accounting", "bookkeeping", "payroll", "budgeting"}},
	{"Marketing", []string{"marketing", "campaign", "seo", "content", "social media"}},
	{"Project management", []string{"project management", "scrum", "agile", "pmp"}},
	{"Data analysis", []string{"data analysis", "analytics", "reporting", "dashboard", "sql"}},
	{"SQL", []string{"sql", "postgresql", "mysql"}},
	{"Python", []string{"python"}},
	{"JavaScript", []string{"javascript", "typescript", "react", "node.js", "nodejs"}},
	{"Design", []string{"figma", "ux", "ui design", "graphic design"}},
	{"Supply chain", []string{"supply chain", "logistics", "procurement", "inventory"}},
	{"Healthcare operations", []string{"patient", "clinic", "medical", "healthcare", "insurance coordination"}},
}

var cvFunctionKeywords = []cvKeyword{
	{"Operations", []string{"operations", "scheduling", "workflow", "process"}},
	{"Sales", []string{"sales", "pipeline", "account management", "business development"}},
	{"Customer Support", []string{"customer support", "customer success", "client support"}},
	{"Finance", []string{"finance", "accounting", "payroll", "budget"}},
	{"HR", []string{"human resources", "recruiting", "talent acquisition"}},
	{"Marketing", []string{"marketing", "campaign", "content", "seo"}},
	{"Product", []string{"product manager", "product owner", "roadmap"}},
	{"Design", []string{"designer", "figma", "ux", "ui"}},
	{"Engineering", []string{"engineer", "developer", "software", "backend", "frontend"}},
	{"Data", []string{"data", "analytics", "business intelligence", "sql"}},
	{"Healthcare", []string{"clinic", "patient", "medical", "healthcare"}},
	{"Retail", []string{"retail", "store", "merchandising"}},
	{"Logistics", []string{"logistics", "supply chain", "warehouse", "inventory"}},
}

var cvIndustryKeywords = []cvKeyword{
	{"Healthcare", []string{"healthcare", "clinic", "hospital", "patient", "medical"}},
	{"Retail", []string{"retail", "store", "merchandising", "pos"}},
	{"SaaS", []string{"saas", "software as a service"}},
	{"Hospitality", []string{"hotel", "restaurant", "hospitality", "f&b"}},
	{"Finance", []string{"bank", "fintech", "finance", "accounting"}},
	{"Education", []string{"school", "university", "education", "training"}},
	{"Logistics", []string{"logistics", "supply chain", "freight", "warehouse"}},
	{"E-commerce", []string{"e-commerce", "ecommerce", "shopify", "online store"}},
	{"Real estate", []string{"real estate", "property", "brokerage"}},
}

var cvRoleKeywords = []cvKeyword{
	{"Operations Manager", []string{"operations manager", "operation manager"}},
	{"Sales Manager", []string{"sales manager", "account manager"}},
	{"Customer Support Specialist", []string{"customer support specialist", "support specialist", "customer service representative"}},
	{"HR Specialist", []string{"hr specialist", "recruiter", "talent acquisition"}},
	{"Finance Manager", []string{"finance manager", "accountant", "controller"}},
	{"Project Manager", []string{"project manager", "program manager"}},
	{"Product Manager", []string{"product manager", "product owner"}},
	{"Designer", []string{"designer", "ux designer", "ui designer"}},
	{"Engineer", []string{"software engineer", "developer", "backend engineer", "frontend engineer"}},
	{"Data Analyst", []string{"data analyst", "business analyst", "bi analyst"}},
}

var cvEducationKeywords = []cvKeyword{
	{"Bachelor's degree", []string{"bachelor", "b.a.", "b.s.", "bsc", "ba "}},
	{"Master's degree", []string{"master", "m.a.", "m.s.", "msc", "mba"}},
	{"MBA", []string{"mba"}},
	{"Computer science", []string{"computer science", "software engineering"}},
	{"Business", []string{"business administration", "management", "commerce"}},
	{"Finance", []string{"finance", "accounting", "economics"}},
}

var cvLanguageKeywords = []cvKeyword{
	{"English", []string{"english"}},
	{"Arabic", []string{"arabic"}},
	{"French", []string{"french"}},
	{"Spanish", []string{"spanish"}},
	{"German", []string{"german"}},
}

var cvCertificationKeywords = []cvKeyword{
	{"PMP", []string{"pmp", "project management professional"}},
	{"CPA", []string{"cpa"}},
	{"CFA", []string{"cfa"}},
	{"SHRM", []string{"shrm"}},
	{"AWS", []string{"aws certified", "amazon web services"}},
	{"Scrum", []string{"scrum master", "csm"}},
	{"Google Analytics", []string{"google analytics"}},
}

var cvLocationKeywords = []cvKeyword{
	{"Beirut", []string{"beirut"}},
	{"Lebanon", []string{"lebanon", "lebanese"}},
	{"Dubai", []string{"dubai"}},
	{"UAE", []string{"uae", "united arab emirates"}},
	{"Riyadh", []string{"riyadh"}},
	{"Saudi Arabia", []string{"saudi", "ksa"}},
	{"Qatar", []string{"qatar", "doha"}},
	{"GCC", []string{"gcc", "gulf"}},
	{"MENA", []string{"mena", "middle east"}},
	{"Remote", []string{"remote"}},
	{"Europe", []string{"europe", "eu"}},
}

// cvKeywordPatterns holds one precompiled word-boundary pattern per keyword
// term. Matching terms as raw substrings tags CVs with unrelated skills: "hr"
// matches "three", "ui" matches "built", "commerce" matches "e-commerce".
var cvKeywordPatterns = buildCVKeywordPatterns(
	cvSkillKeywords,
	cvFunctionKeywords,
	cvIndustryKeywords,
	cvRoleKeywords,
	cvEducationKeywords,
	cvLanguageKeywords,
	cvCertificationKeywords,
	cvLocationKeywords,
)

func buildCVKeywordPatterns(groups ...[]cvKeyword) map[string]*regexp.Regexp {
	patterns := make(map[string]*regexp.Regexp)
	for _, keywords := range groups {
		for _, keyword := range keywords {
			for _, term := range keyword.Terms {
				term = strings.ToLower(strings.TrimSpace(term))
				if term == "" {
					continue
				}
				if _, ok := patterns[term]; ok {
					continue
				}
				// The pattern consumes the character before and after the term;
				// RE2 has no lookaround. A term only matches when it is neither
				// preceded nor followed by a letter, digit, or hyphen — hyphens
				// join compounds ("e-commerce"), so "commerce" must not match
				// inside "e-commerce" while "b.a." and "e-commerce" still match
				// themselves.
				patterns[term] = regexp.MustCompile(`(?:^|[^a-z0-9-])` + regexp.QuoteMeta(term) + `(?:[^a-z0-9-]|$)`)
			}
		}
	}
	return patterns
}

type cvStorage interface {
	Download(ctx context.Context, key string) (io.ReadCloser, int64, error)
}

// ProcessPendingCVs downloads queued CV files, extracts searchable text,
// and writes the result back onto the application row.
func (s *Service) ProcessPendingCVs(ctx context.Context) error {
	targets, err := s.repo.ListPendingCVProcessingTargets(ctx, cvProcessingBatchSize)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return nil
	}

	client, err := storage.Default()
	if err != nil {
		storageErr := fmt.Errorf("cv storage unavailable: %w", err)
		if updateErr := s.markCVTargetsErrored(ctx, targets, storageErr.Error()); updateErr != nil {
			return errors.Join(storageErr, updateErr)
		}
		return storageErr
	}

	for _, target := range targets {
		result := s.processCVTarget(ctx, client, target)
		if err := s.repo.UpdateCVProcessingResult(ctx, result); err != nil {
			return err
		}
	}
	return nil
}
