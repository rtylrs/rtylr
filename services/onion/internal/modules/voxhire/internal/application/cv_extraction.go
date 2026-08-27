package application

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/ledongthuc/pdf"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/storage"
)

const (
	cvProcessingBatchSize = 25
	maxCVExtractionBytes  = 10 << 20
	maxCVStoredTextBytes  = 128 << 10
	maxCVSummaryRunes     = 700
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

func (s *Service) markCVTargetsErrored(ctx context.Context, targets []domain.CVProcessingTarget, message string) error {
	var updateErr error
	for _, target := range targets {
		err := s.repo.UpdateCVProcessingResult(ctx, domain.CVProcessingResult{
			Source: target.Source,
			ID:     target.ID,
			Status: "error",
			Error:  message,
		})
		if err != nil {
			updateErr = errors.Join(updateErr, err)
		}
	}
	return updateErr
}

func (s *Service) processCVTarget(ctx context.Context, store cvStorage, target domain.CVProcessingTarget) domain.CVProcessingResult {
	result := domain.CVProcessingResult{
		Source: target.Source,
		ID:     target.ID,
		Status: "processed",
	}

	body, size, err := store.Download(ctx, target.CVKey)
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		return result
	}
	defer body.Close()

	if size > maxCVExtractionBytes {
		result.Status = "error"
		result.Error = "CV is too large to extract"
		return result
	}

	raw, err := io.ReadAll(io.LimitReader(body, maxCVExtractionBytes+1))
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		return result
	}
	if len(raw) > maxCVExtractionBytes {
		result.Status = "error"
		result.Error = "CV is too large to extract"
		return result
	}

	text, err := extractCVText(target.CVKey, raw)
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		logger.Ctx(ctx).Warn("voxhire cv extraction failed",
			logger.String("source", target.Source),
			logger.Any("id", target.ID),
			logger.String("key", target.CVKey),
			logger.Error(err),
		)
		return result
	}

	text = normalizeCVText(text)
	result.ExtractedText = truncateBytes(text, maxCVStoredTextBytes)
	result.Summary = summarizeCVText(result.ExtractedText)
	result.Profile = buildCVProfile(result.ExtractedText)
	return result
}

func extractCVText(key string, raw []byte) (string, error) {
	lower := strings.ToLower(key)
	switch {
	case strings.HasSuffix(lower, ".pdf"):
		return extractPDFText(raw)
	case strings.HasSuffix(lower, ".docx"):
		return extractDOCXText(raw)
	case strings.HasSuffix(lower, ".doc"):
		return "", fmt.Errorf("legacy .doc CV parsing is not supported; ask candidate for PDF or DOCX")
	default:
		if bytes.HasPrefix(raw, []byte("%PDF")) {
			return extractPDFText(raw)
		}
		if len(raw) >= 4 && raw[0] == 'P' && raw[1] == 'K' {
			return extractDOCXText(raw)
		}
		return "", fmt.Errorf("unsupported CV format")
	}
}

func extractPDFText(raw []byte) (string, error) {
	tmp, err := os.CreateTemp("", "voxhire-cv-*.pdf")
	if err != nil {
		return "", err
	}
	path := tmp.Name()
	defer os.Remove(path)

	if _, err := tmp.Write(raw); err != nil {
		if closeErr := tmp.Close(); closeErr != nil {
			return "", fmt.Errorf("write temporary CV: %w; close file: %v", err, closeErr)
		}
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}

	file, reader, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	plain, err := reader.GetPlainText()
	if err != nil {
		return "", err
	}
	out, err := io.ReadAll(io.LimitReader(plain, maxCVStoredTextBytes+1))
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func extractDOCXText(raw []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return "", err
	}
	for _, f := range zr.File {
		if f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()
		return extractWordXMLText(rc)
	}
	return "", fmt.Errorf("DOCX missing word/document.xml")
}

func extractWordXMLText(r io.Reader) (string, error) {
	decoder := xml.NewDecoder(r)
	var out strings.Builder
	for {
		tok, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.CharData:
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
			if strings.Contains(lowerText, strings.ToLower(term)) {
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
