package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
)

type mockRepo struct {
	companyID   int64
	companyErr  error
	talentID    int64
	talentErr   error
	career      *domain.Opening
	careerErr   error
	cvTargets   []domain.CVProcessingTarget
	cvResults   []domain.CVProcessingResult
	cvUpdateErr error
	reviewID    int64
	reviewInput domain.ApplicationReviewInput
	reviewErr   error
}

func (m *mockRepo) CreateCompanyLead(_ context.Context, _ domain.CompanyLeadInput) (int64, error) {
	return m.companyID, m.companyErr
}

func (m *mockRepo) CreateTalentApplication(_ context.Context, _ domain.TalentApplicationInput) (int64, error) {
	return m.talentID, m.talentErr
}

func (m *mockRepo) CreateAdmin(_ context.Context, _, _, _ string) (int64, error) { return 0, nil }

func (m *mockRepo) GetAdminByEmail(_ context.Context, _ string) (*domain.Admin, error) {
	return nil, domain.ErrAdminNotFound
}

func (m *mockRepo) GetAdminByID(_ context.Context, _ int64) (*domain.Admin, error) {
	return nil, domain.ErrAdminNotFound
}

func (m *mockRepo) ListAdmins(_ context.Context) ([]domain.Admin, error) { return nil, nil }

func (m *mockRepo) CountAdmins(_ context.Context) (int64, error) { return 0, nil }

func (m *mockRepo) ListCompanyLeads(_ context.Context) ([]domain.Lead, error) { return nil, nil }

func (m *mockRepo) ListOpenings(_ context.Context) ([]domain.Opening, error) { return nil, nil }

func (m *mockRepo) GetAnalytics(_ context.Context) (*domain.Analytics, error) { return nil, nil }

func (m *mockRepo) CreateCareer(_ context.Context, _ domain.CareerInput) (*domain.Opening, error) {
	return m.career, m.careerErr
}

func (m *mockRepo) UpdateCareer(_ context.Context, _ domain.CareerInput) (*domain.Opening, error) {
	return m.career, m.careerErr
}

func (m *mockRepo) SetCareerActive(_ context.Context, _ int64, _ bool) (*domain.Opening, error) {
	return m.career, m.careerErr
}

func (m *mockRepo) ListApplications(_ context.Context, _ domain.ApplicationFilter) ([]domain.Application, int64, error) {
	return nil, 0, nil
}

func (m *mockRepo) UpdateApplicationStatus(_ context.Context, _ int64, _ string) error { return nil }

func (m *mockRepo) UpdateApplicationReview(_ context.Context, id int64, in domain.ApplicationReviewInput) error {
	m.reviewID = id
	m.reviewInput = in
	return m.reviewErr
}

func (m *mockRepo) GetApplicationCVKey(_ context.Context, _ int64) (string, error) { return "", nil }

func (m *mockRepo) CreateCareerApplication(_ context.Context, _ domain.CareerApplicationInput) (int64, error) {
	return 0, nil
}

func (m *mockRepo) ListPendingCVProcessingTargets(_ context.Context, _ int) ([]domain.CVProcessingTarget, error) {
	return m.cvTargets, nil
}

func (m *mockRepo) UpdateCVProcessingResult(_ context.Context, result domain.CVProcessingResult) error {
	m.cvResults = append(m.cvResults, result)
	return m.cvUpdateErr
}

func TestSubmitCompanyLead_ReturnsID(t *testing.T) {
	s := &Service{repo: &mockRepo{companyID: 42}}
	id, err := s.SubmitCompanyLead(context.Background(), domain.CompanyLeadInput{
		CompanyName: "Acme",
		ContactName: "Alice",
		WorkEmail:   "alice@acme.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Errorf("expected id=42, got %d", id)
	}
}

func TestSubmitCompanyLead_PropagatesError(t *testing.T) {
	want := errors.New("db fail")
	s := &Service{repo: &mockRepo{companyErr: want}}
	if _, err := s.SubmitCompanyLead(context.Background(), domain.CompanyLeadInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestSubmitCompanyLead_ZeroIDOnError(t *testing.T) {
	s := &Service{repo: &mockRepo{companyErr: errors.New("oops")}}
	id, _ := s.SubmitCompanyLead(context.Background(), domain.CompanyLeadInput{})
	if id != 0 {
		t.Errorf("expected 0 on error, got %d", id)
	}
}

func TestSubmitTalentApplication_ReturnsID(t *testing.T) {
	s := &Service{repo: &mockRepo{talentID: 99}}
	id, err := s.SubmitTalentApplication(context.Background(), domain.TalentApplicationInput{
		FullName: "Bob",
		Email:    "bob@example.com",
		Track:    "engineering",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 99 {
		t.Errorf("expected id=99, got %d", id)
	}
}

func TestSubmitTalentApplication_PropagatesError(t *testing.T) {
	want := errors.New("repo down")
	s := &Service{repo: &mockRepo{talentErr: want}}
	if _, err := s.SubmitTalentApplication(context.Background(), domain.TalentApplicationInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestSubmitTalentApplication_ZeroIDOnError(t *testing.T) {
	s := &Service{repo: &mockRepo{talentErr: errors.New("bad")}}
	id, _ := s.SubmitTalentApplication(context.Background(), domain.TalentApplicationInput{})
	if id != 0 {
		t.Errorf("expected 0 on error, got %d", id)
	}
}

func TestCreateCareerRequiresTitle(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if _, err := s.CreateCareer(context.Background(), domain.CareerInput{}); !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected invalid input, got %v", err)
	}
}

func TestOrDash_EmptyReturnsHyphen(t *testing.T) {
	if got := orDash(""); got != "-" {
		t.Errorf("orDash(\"\") = %q, want \"-\"", got)
	}
}

func TestOrDash_NonEmptyPassThrough(t *testing.T) {
	if got := orDash("hello"); got != "hello" {
		t.Errorf("orDash(\"hello\") = %q, want \"hello\"", got)
	}
}

func TestFormatTitle_EmptyReturnsEmpty(t *testing.T) {
	if got := formatTitle(""); got != "" {
		t.Errorf("formatTitle(\"\") = %q, want empty", got)
	}
}

func TestFormatTitle_NonEmptyWrapsInParens(t *testing.T) {
	got := formatTitle("CTO")
	if got != " (CTO)" {
		t.Errorf("formatTitle(\"CTO\") = %q, want \" (CTO)\"", got)
	}
}

func TestBoolLabel_True(t *testing.T) {
	got := boolLabel(true, "yes", "no")
	if got != "yes" {
		t.Errorf("boolLabel(true) = %q", got)
	}
}

func TestBoolLabel_False(t *testing.T) {
	got := boolLabel(false, "yes", "no")
	if got != "no" {
		t.Errorf("boolLabel(false) = %q", got)
	}
}

func TestLinkOrDash_EmptyReturnsDash(t *testing.T) {
	if got := linkOrDash(""); got != "-" {
		t.Errorf("linkOrDash(\"\") = %q, want \"-\"", got)
	}
}

func TestLinkOrDash_NonEmptyReturnsAnchor(t *testing.T) {
	got := linkOrDash("https://example.com")
	if len(got) == 0 || got == "-" {
		t.Errorf("linkOrDash(url) should return anchor, got %q", got)
	}
}

func TestSubmitCompanyLead_MinimalInput(t *testing.T) {
	s := &Service{repo: &mockRepo{companyID: 1}}
	id, err := s.SubmitCompanyLead(context.Background(), domain.CompanyLeadInput{})
	if err != nil || id != 1 {
		t.Errorf("id=%d err=%v", id, err)
	}
}

func TestSubmitTalentApplication_MinimalInput(t *testing.T) {
	s := &Service{repo: &mockRepo{talentID: 1}}
	id, err := s.SubmitTalentApplication(context.Background(), domain.TalentApplicationInput{})
	if err != nil || id != 1 {
		t.Errorf("id=%d err=%v", id, err)
	}
}

func TestUpdateApplicationReview_NormalizesInput(t *testing.T) {
	repo := &mockRepo{}
	s := &Service{repo: repo}

	err := s.UpdateApplicationReview(context.Background(), 77, domain.ApplicationReviewInput{
		Status:       " Shortlisted ",
		ReviewRating: 4,
		ReviewNotes:  "  Strong operations profile.  ",
		NextStep:     "  Send to client  ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.reviewID != 77 {
		t.Fatalf("reviewID = %d, want 77", repo.reviewID)
	}
	want := domain.ApplicationReviewInput{
		Status:       "shortlisted",
		ReviewRating: 4,
		ReviewNotes:  "Strong operations profile.",
		NextStep:     "Send to client",
	}
	if repo.reviewInput != want {
		t.Fatalf("reviewInput = %#v, want %#v", repo.reviewInput, want)
	}
}

func TestUpdateApplicationReview_RejectsInvalidRating(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.UpdateApplicationReview(context.Background(), 77, domain.ApplicationReviewInput{ReviewRating: 6})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}
