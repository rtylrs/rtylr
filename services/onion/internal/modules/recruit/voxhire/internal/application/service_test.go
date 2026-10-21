package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
)

type mockRepo struct {
	companyID   int64
	companyErr  error
	talentID    int64
	talentErr   error
	career      *domain.Opening
	careerErr   error
	apps        []domain.Application
	appsErr     error
	cvTargets   []domain.CVProcessingTarget
	cvResults   []domain.CVProcessingResult
	cvUpdateErr error
	reviewID    int64
	reviewInput domain.ApplicationReviewInput
	reviewErr   error

	careerAppID  int64
	careerAppErr error
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

func (m *mockRepo) ListOpeningsPage(_ context.Context, limit, offset int32) (domain.OpeningPage, error) {
	return domain.OpeningPage{Limit: limit, Offset: offset}, nil
}

func (m *mockRepo) GetOpening(_ context.Context, _ int64) (*domain.Opening, error) {
	return m.career, m.careerErr
}

func (m *mockRepo) GetAnalytics(_ context.Context) (*domain.Analytics, error) { return nil, nil }

func TestListOpeningsPageClampsBounds(t *testing.T) {
	svc := &Service{repo: &mockRepo{}}
	page, err := svc.ListOpeningsPage(context.Background(), 500, -10)
	if err != nil {
		t.Fatalf("ListOpeningsPage: %v", err)
	}
	if page.Limit != 100 || page.Offset != 0 {
		t.Fatalf("page bounds = limit %d offset %d, want 100 and 0", page.Limit, page.Offset)
	}
}

// A failed opening lookup must not silently score every applicant against a
// nil opening; the applications come back without a match signal and the error
// is logged instead of discarded.
func TestListApplicationsLeavesMatchesUnscoredWhenOpeningLookupFails(t *testing.T) {
	repo := &mockRepo{
		apps:      []domain.Application{{ID: 1, Name: "Ada Lovelace"}},
		careerErr: domain.ErrCareerNotFound,
	}
	svc := &Service{repo: repo}

	apps, total, err := svc.ListApplications(context.Background(), domain.ApplicationFilter{JobID: 404})
	if err != nil {
		t.Fatalf("ListApplications: %v", err)
	}
	if total != 1 || len(apps) != 1 {
		t.Fatalf("apps = %d, total = %d, want 1 and 1", len(apps), total)
	}
	if apps[0].MatchSignal.Score != 0 || apps[0].MatchSignal.Label != "" {
		t.Fatalf("MatchSignal = %#v, want zero value when the opening lookup fails", apps[0].MatchSignal)
	}
}

func TestListApplicationsEnrichesMatchesWhenOpeningExists(t *testing.T) {
	repo := &mockRepo{
		apps:   []domain.Application{{ID: 1, Name: "Ada Lovelace", CVProcessingStatus: "processed"}},
		career: &domain.Opening{ID: 5, Title: "Operations Manager", Location: "Beirut"},
	}
	svc := &Service{repo: repo}

	apps, _, err := svc.ListApplications(context.Background(), domain.ApplicationFilter{JobID: 5})
	if err != nil {
		t.Fatalf("ListApplications: %v", err)
	}
	if apps[0].MatchSignal.Score == 0 || apps[0].MatchSignal.Label == "" {
		t.Fatalf("MatchSignal = %#v, want an enriched signal when the opening exists", apps[0].MatchSignal)
	}
}

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
	return m.apps, int64(len(m.apps)), m.appsErr
}

func (m *mockRepo) ListCandidates(_ context.Context, _ domain.CandidateFilter) ([]domain.Candidate, int64, error) {
	return nil, 0, nil
}

func (m *mockRepo) ListCVJobs(_ context.Context, _ domain.CandidateFilter) ([]domain.CVJob, error) {
	return nil, nil
}

func (m *mockRepo) RetryCVProcessing(_ context.Context, _ string, _ int64) error { return nil }

func (m *mockRepo) UpdateApplicationStatus(_ context.Context, _ int64, _ string) error { return nil }

func (m *mockRepo) UpdateApplicationReview(_ context.Context, id int64, in domain.ApplicationReviewInput) error {
	m.reviewID = id
	m.reviewInput = in
	return m.reviewErr
}

func (m *mockRepo) GetApplicationCVKey(_ context.Context, _ int64) (string, error) { return "", nil }

func (m *mockRepo) ListApplicationTimeline(_ context.Context, _ int64) ([]domain.TimelineEvent, error) {
	return nil, nil
}

func (m *mockRepo) ListTasks(_ context.Context, _ string) ([]domain.Task, error) { return nil, nil }

func (m *mockRepo) UpsertTask(_ context.Context, in domain.TaskInput) (*domain.Task, error) {
	return &domain.Task{ApplicationID: in.ApplicationID, Title: in.Title, Status: in.Status}, nil
}

func (m *mockRepo) UpdateTask(_ context.Context, in domain.TaskInput) (*domain.Task, error) {
	return &domain.Task{ID: in.ID, ApplicationID: in.ApplicationID, Title: in.Title, Status: in.Status}, nil
}

func (m *mockRepo) ListShortlists(_ context.Context, _ int64) ([]domain.Shortlist, error) {
	return nil, nil
}

func (m *mockRepo) CreateShortlist(_ context.Context, in domain.ShortlistInput) (*domain.Shortlist, error) {
	return &domain.Shortlist{ID: 1, JobID: in.JobID, Name: in.Name, Status: in.Status}, nil
}

func (m *mockRepo) AddShortlistItem(_ context.Context, in domain.ShortlistItemInput) (*domain.ShortlistItem, error) {
	return &domain.ShortlistItem{ID: 1, ShortlistID: in.ShortlistID, ApplicationID: in.ApplicationID, Rank: in.Rank, Notes: in.Notes}, nil
}

func (m *mockRepo) CreateCareerApplication(_ context.Context, _ domain.CareerApplicationInput) (int64, error) {
	return m.careerAppID, m.careerAppErr
}

func (m *mockRepo) FindRecentTalentApplicationID(_ context.Context, _ string) (int64, error) {
	return 0, nil
}

func (m *mockRepo) FindRecentCareerApplicationID(_ context.Context, _ int64, _ string) (int64, error) {
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
