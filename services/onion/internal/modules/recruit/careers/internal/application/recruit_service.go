package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
)

// Recruit extends the careers module with the tenant-scoped recruit
// product. Keeping it on the existing Service struct means handlers
// can use a single dependency; the methods below are grouped by the
// three entities they touch.

// ── careers ──────────────────────────────────────────────────────────

func (s *Service) ListRecruitCareers(ctx context.Context, businessID int64, publishedOnly bool) ([]domain.RecruitCareer, error) {
	return s.repo.ListCareers(ctx, businessID, publishedOnly)
}

func (s *Service) GetCareerBySlug(ctx context.Context, businessID int64, slug string) (domain.RecruitCareer, error) {
	c, err := s.repo.GetCareerBySlug(ctx, businessID, slug)
	if err != nil {
		if errors.Is(err, domain.ErrCareerNotFound) {
			return domain.RecruitCareer{}, ErrNotFound
		}
		return domain.RecruitCareer{}, err
	}
	return c, nil
}

// CreateCareer inserts a career and returns the stored row. The slug may be
// server-generated from the title, so callers must echo the returned row
// instead of re-deriving it from the input (which still has the empty slug).
func (s *Service) CreateCareer(ctx context.Context, in domain.RecruitCareerInput) (domain.RecruitCareer, error) {
	if in.Slug == "" {
		in.Slug = slugify(in.Title)
	}
	id, err := s.repo.CreateCareer(ctx, in)
	if err != nil {
		return domain.RecruitCareer{}, err
	}

	created, err := s.repo.GetCareerByID(ctx, in.BusinessID, id)
	if err != nil {
		// The row is committed. Return the input's own values (with the
		// resolved slug and publish state) rather than failing a write that
		// already succeeded.
		return domain.RecruitCareer{
			ID:             id,
			OwnerID:        in.OwnerID,
			BusinessID:     in.BusinessID,
			Slug:           in.Slug,
			Title:          in.Title,
			Department:     in.Department,
			Location:       in.Location,
			EmploymentType: in.EmploymentType,
			WorkMode:       in.WorkMode,
			Seniority:      in.Seniority,
			Details:        in.Details,
			Requirements:   in.Requirements,
			SalaryMin:      in.SalaryMin,
			SalaryMax:      in.SalaryMax,
			SalaryCurrency: in.SalaryCurrency,
			IsPublished:    in.IsPublished,
			IsActive:       true,
		}, nil
	}
	return created, nil
}

func (s *Service) UpdateCareer(ctx context.Context, in domain.RecruitCareerInput) error {
	if in.Slug == "" {
		in.Slug = slugify(in.Title)
	}
	if err := s.repo.UpdateCareer(ctx, in); err != nil {
		if errors.Is(err, domain.ErrCareerNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *Service) DeleteCareer(ctx context.Context, businessID, careerID int64) error {
	if err := s.repo.DeleteCareer(ctx, businessID, careerID); err != nil {
		if errors.Is(err, domain.ErrCareerNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// ── customization ────────────────────────────────────────────────────

func (s *Service) GetMyCustomization(ctx context.Context, businessID int64) (domain.Customization, error) {
	c, err := s.repo.GetCustomizationByOwner(ctx, businessID)
	if err != nil {
		if errors.Is(err, domain.ErrCustomizationNotFound) {
			return domain.Customization{}, ErrNotFound
		}
		return domain.Customization{}, err
	}
	return c, nil
}

func (s *Service) GetCustomizationBySlug(ctx context.Context, slug string) (domain.Customization, error) {
	c, err := s.repo.GetCustomizationBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, domain.ErrCustomizationNotFound) {
			return domain.Customization{}, ErrNotFound
		}
		return domain.Customization{}, err
	}
	return c, nil
}

func (s *Service) UpsertCustomization(ctx context.Context, in domain.CustomizationInput) error {
	in.Slug = slugify(in.Slug)
	if in.Template == "" {
		in.Template = "minimal"
	}
	return s.repo.UpsertCustomization(ctx, in)
}

// ── submissions ──────────────────────────────────────────────────────

func (s *Service) CreateSubmission(ctx context.Context, in domain.SubmissionInput) (int64, error) {
	return s.repo.CreateSubmission(ctx, in)
}

func (s *Service) ListSubmissions(ctx context.Context, businessID int64, status string) ([]domain.Submission, error) {
	return s.repo.ListSubmissions(ctx, businessID, status)
}

// ListSubmissionsPage returns one page of submissions, newest first. Exports
// walk it until a short page comes back; ListSubmissions keeps the UI's
// fixed first-page shape.
func (s *Service) ListSubmissionsPage(ctx context.Context, businessID int64, status string, limit, offset int32) ([]domain.Submission, error) {
	return s.repo.ListSubmissionsPage(ctx, businessID, status, limit, offset)
}

func (s *Service) GetSubmissionResumeURL(ctx context.Context, businessID, submissionID int64) (string, error) {
	resumeURL, err := s.repo.GetSubmissionResumeURL(ctx, businessID, submissionID)
	if errors.Is(err, domain.ErrSubmissionNotFound) {
		return "", ErrNotFound
	}
	return resumeURL, err
}

func (s *Service) UpdateSubmissionStatus(ctx context.Context, businessID, submissionID int64, status string) error {
	if err := s.repo.UpdateSubmissionStatus(ctx, businessID, submissionID, status); err != nil {
		if errors.Is(err, domain.ErrSubmissionNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// ── helpers ──────────────────────────────────────────────────────────

// ErrNotFound is returned from service methods when a row the handler
// expected to exist isn't in the database. The handler translates this
// into a 404.
var ErrNotFound = errors.New("not found")

var (
	slugInvalid  = regexp.MustCompile(`[^a-z0-9-]+`)
	slugCollapse = regexp.MustCompile(`-+`)
)

// slugify lowercases, strips non-alphanumerics, collapses repeated
// dashes. Accepts partially-slugged input ("Backend Engineer" →
// "backend-engineer") so admins can paste a job title and have the
// URL come out clean. Inputs with no ASCII slug characters get a stable
// hash fallback so they can never produce an unreachable empty URL.
func slugify(s string) string {
	source := s
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = slugInvalid.ReplaceAllString(s, "")
	s = slugCollapse.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		sum := sha256.Sum256([]byte(source))
		return "slug-" + hex.EncodeToString(sum[:8])
	}
	return s
}
