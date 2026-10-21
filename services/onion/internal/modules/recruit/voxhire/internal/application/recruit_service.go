package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/recruit/careers"
	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/storage"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

const (
	adminJWTIssuer = "voxhire-admin"
	adminTokenTTL  = 7 * 24 * time.Hour
	cvURLTTL       = 10 * time.Minute

	// careerApplicationNotifyTimeout bounds the background notification that
	// follows a public application submission, so a stalled email provider
	// cannot pin the goroutine and its connection.
	careerApplicationNotifyTimeout = 10 * time.Second

	minAdminPasswordLen = 6
)

var errJWTSecretMissing = errors.New("jwt secret not configured")

func adminJWTHelper() (*jwtutils.JWTHelper, error) {
	secret := ""
	if config.GlobalConfig != nil {
		secret = config.GlobalConfig.JWTSecret
	}
	if secret == "" {
		return nil, errJWTSecretMissing
	}
	return jwtutils.NewJWTHelperWithDuration(secret, adminJWTIssuer, adminTokenTTL), nil
}

// ── Authentication ───────────────────────────────────────────────────

// AdminLogin verifies credentials and returns a signed admin JWT.
func (s *Service) AdminLogin(ctx context.Context, email, password string) (string, *domain.Admin, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	admin, err := s.repo.GetAdminByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrAdminNotFound) {
			return "", nil, domain.ErrInvalidCredentials
		}
		return "", nil, err
	}

	if !passwordutils.CheckPasswordHash(password, admin.PasswordHash) {
		return "", nil, domain.ErrInvalidCredentials
	}

	token, err := s.issueAdminToken(admin)
	if err != nil {
		return "", nil, err
	}
	return token, admin, nil
}

func (s *Service) issueAdminToken(admin *domain.Admin) (string, error) {
	helper, err := adminJWTHelper()
	if err != nil {
		return "", err
	}
	claims := &AdminClaims{AdminID: admin.ID, Email: admin.Email}
	return helper.CreateJWT(claims)
}

// ValidateAdminToken parses and validates an admin bearer token.
func (s *Service) ValidateAdminToken(ctx context.Context, token string) (*AdminClaims, error) {
	helper, err := adminJWTHelper()
	if err != nil {
		return nil, err
	}
	var claims AdminClaims
	if err := helper.ValidateToken(ctx, token, &claims); err != nil {
		return nil, err
	}
	if claims.AdminID == 0 {
		return nil, domain.ErrInvalidCredentials
	}
	return &claims, nil
}

// ── Admin account management ─────────────────────────────────────────

// CreateAdmin hashes the password and creates a new recruiter-admin.
func (s *Service) CreateAdmin(ctx context.Context, email, password, name string) (*domain.Admin, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, domain.ErrInvalidInput
	}
	if len(password) < minAdminPasswordLen {
		return nil, domain.ErrInvalidInput
	}

	hash, err := passwordutils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	id, err := s.repo.CreateAdmin(ctx, email, hash, strings.TrimSpace(name))
	if err != nil {
		return nil, err
	}
	return &domain.Admin{ID: id, Email: email, Name: strings.TrimSpace(name)}, nil
}

// ListAdmins returns all recruiter-admin accounts.
func (s *Service) ListAdmins(ctx context.Context) ([]domain.Admin, error) {
	return s.repo.ListAdmins(ctx)
}

// GetAdmin returns a single admin by id.
func (s *Service) GetAdmin(ctx context.Context, id int64) (*domain.Admin, error) {
	return s.repo.GetAdminByID(ctx, id)
}

// ── Openings & applications (recruiter panel) ────────────────────────

// ListOpenings returns active job openings with application counts.
func (s *Service) ListOpenings(ctx context.Context) ([]domain.Opening, error) {
	return s.repo.ListOpenings(ctx)
}

func (s *Service) ListOpeningsPage(ctx context.Context, limit, offset int32) (domain.OpeningPage, error) {
	const maxOpeningsPageSize int32 = 100
	if limit <= 0 || limit > maxOpeningsPageSize {
		limit = maxOpeningsPageSize
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListOpeningsPage(ctx, limit, offset)
}

// GetAnalytics returns aggregate funnel metrics for the admin overview.
func (s *Service) GetAnalytics(ctx context.Context) (*domain.Analytics, error) {
	return s.repo.GetAnalytics(ctx)
}

// CreateCareer creates a public VoxHire job posting.
func (s *Service) CreateCareer(ctx context.Context, in domain.CareerInput) (*domain.Opening, error) {
	in = normalizeCareerInput(in)
	if in.Title == "" {
		return nil, domain.ErrInvalidInput
	}
	career, err := s.repo.CreateCareer(ctx, in)
	if err != nil {
		return nil, err
	}
	invalidateCareersCatalog(ctx)
	return career, nil
}

// UpdateCareer updates an existing VoxHire job posting.
func (s *Service) UpdateCareer(ctx context.Context, in domain.CareerInput) (*domain.Opening, error) {
	in = normalizeCareerInput(in)
	if in.ID <= 0 || in.Title == "" {
		return nil, domain.ErrInvalidInput
	}
	career, err := s.repo.UpdateCareer(ctx, in)
	if err != nil {
		return nil, err
	}
	invalidateCareersCatalog(ctx)
	return career, nil
}

// SetCareerActive closes or reopens a VoxHire job posting.
func (s *Service) SetCareerActive(ctx context.Context, id int64, active bool) (*domain.Opening, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}
	career, err := s.repo.SetCareerActive(ctx, id, active)
	if err != nil {
		return nil, err
	}
	invalidateCareersCatalog(ctx)
	return career, nil
}

func invalidateCareersCatalog(ctx context.Context) {
	careers.InvalidateActiveCareers(ctx)
}

func normalizeCareerInput(in domain.CareerInput) domain.CareerInput {
	in.Title = strings.TrimSpace(in.Title)
	in.Department = strings.TrimSpace(in.Department)
	in.Location = strings.TrimSpace(in.Location)
	in.EmploymentType = strings.TrimSpace(in.EmploymentType)
	in.WorkMode = strings.TrimSpace(in.WorkMode)
	in.Seniority = strings.TrimSpace(in.Seniority)
	in.Details = strings.TrimSpace(in.Details)
	in.SalaryCurrency = strings.ToUpper(strings.TrimSpace(in.SalaryCurrency))
	if in.SalaryCurrency == "" {
		in.SalaryCurrency = "USD"
	}
	return in
}

// ListCompanyLeads returns company hiring inquiries for the admin panel.
func (s *Service) ListCompanyLeads(ctx context.Context) ([]domain.Lead, error) {
	return s.repo.ListCompanyLeads(ctx)
}

// ListApplications returns the applications for a single opening,
// filtered/sorted, plus the total matching count.
func (s *Service) ListApplications(ctx context.Context, f domain.ApplicationFilter) ([]domain.Application, int64, error) {
	apps, total, err := s.repo.ListApplications(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	opening, err := s.repo.GetOpening(ctx, f.JobID)
	if err != nil {
		// The applications are valid even when the opening lookup fails, but
		// scoring them against a nil opening would silently report every
		// applicant as a poor match. Leave them unscored and log the cause.
		if errors.Is(err, domain.ErrCareerNotFound) {
			logger.Ctx(ctx).Info("voxhire application match enrichment skipped for unknown opening",
				logger.Any("job_id", f.JobID),
			)
		} else {
			logger.Ctx(ctx).Error("voxhire application match enrichment skipped: get opening failed",
				logger.Any("job_id", f.JobID),
				logger.Error(err),
			)
		}
		return apps, total, nil
	}
	return enrichApplicationMatches(apps, opening), total, nil
}

func (s *Service) ListCandidates(ctx context.Context, f domain.CandidateFilter) ([]domain.Candidate, int64, error) {
	f.Query = strings.TrimSpace(f.Query)
	f.Source = strings.ToLower(strings.TrimSpace(f.Source))
	f.CVStatus = strings.ToLower(strings.TrimSpace(f.CVStatus))
	candidates, total, err := s.repo.ListCandidates(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	return enrichCandidateMatches(candidates), total, nil
}

func (s *Service) ListCVJobs(ctx context.Context, f domain.CandidateFilter) ([]domain.CVJob, error) {
	f.Query = strings.TrimSpace(f.Query)
	f.Source = strings.ToLower(strings.TrimSpace(f.Source))
	f.CVStatus = strings.ToLower(strings.TrimSpace(f.CVStatus))
	return s.repo.ListCVJobs(ctx, f)
}

func (s *Service) RetryCVProcessing(ctx context.Context, source string, id int64) error {
	source = strings.ToLower(strings.TrimSpace(source))
	if id <= 0 || (source != "career" && source != "talent") {
		return domain.ErrInvalidInput
	}
	return s.repo.RetryCVProcessing(ctx, source, id)
}

// UpdateApplicationStatus changes an application's status (e.g. marking
// it "viewed" when HR opens it).
func (s *Service) UpdateApplicationStatus(ctx context.Context, id int64, status string) error {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		return domain.ErrInvalidInput
	}
	return s.repo.UpdateApplicationStatus(ctx, id, status)
}

// UpdateApplicationReview stores HR review state such as rating, notes,
// next step, and optionally the current pipeline status.
func (s *Service) UpdateApplicationReview(ctx context.Context, id int64, in domain.ApplicationReviewInput) error {
	if id <= 0 {
		return domain.ErrInvalidInput
	}
	in.Status = strings.ToLower(strings.TrimSpace(in.Status))
	in.ReviewNotes = strings.TrimSpace(in.ReviewNotes)
	in.NextStep = strings.TrimSpace(in.NextStep)
	if in.ReviewRating < 0 || in.ReviewRating > 5 {
		return domain.ErrInvalidInput
	}
	if err := s.repo.UpdateApplicationReview(ctx, id, in); err != nil {
		return err
	}
	if in.NextStep != "" {
		_, err := s.repo.UpsertTask(ctx, domain.TaskInput{
			ApplicationID: id,
			Title:         in.NextStep,
			Status:        "open",
		})
		return err
	}
	return nil
}

func (s *Service) ListApplicationTimeline(ctx context.Context, id int64) ([]domain.TimelineEvent, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}
	return s.repo.ListApplicationTimeline(ctx, id)
}

func (s *Service) ListTasks(ctx context.Context, status string) ([]domain.Task, error) {
	return s.repo.ListTasks(ctx, strings.ToLower(strings.TrimSpace(status)))
}

func (s *Service) UpsertTask(ctx context.Context, in domain.TaskInput) (*domain.Task, error) {
	in.Title = strings.TrimSpace(in.Title)
	in.Status = strings.ToLower(strings.TrimSpace(in.Status))
	if in.Status == "" {
		in.Status = "open"
	}
	if in.ApplicationID <= 0 || in.Title == "" {
		return nil, domain.ErrInvalidInput
	}
	return s.repo.UpsertTask(ctx, in)
}

func (s *Service) UpdateTask(ctx context.Context, in domain.TaskInput) (*domain.Task, error) {
	in.Title = strings.TrimSpace(in.Title)
	in.Status = strings.ToLower(strings.TrimSpace(in.Status))
	if in.ID <= 0 || (in.Status == "" && in.Title == "" && !in.DueAtSet) {
		return nil, domain.ErrInvalidInput
	}
	return s.repo.UpdateTask(ctx, in)
}

func (s *Service) ListShortlists(ctx context.Context, jobID int64) ([]domain.Shortlist, error) {
	shortlists, err := s.repo.ListShortlists(ctx, jobID)
	if err != nil {
		return nil, err
	}
	return enrichShortlistMatches(shortlists), nil
}

func (s *Service) CreateShortlist(ctx context.Context, in domain.ShortlistInput) (*domain.Shortlist, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Status = strings.ToLower(strings.TrimSpace(in.Status))
	in.Notes = strings.TrimSpace(in.Notes)
	if in.Status == "" {
		in.Status = "draft"
	}
	if in.JobID <= 0 || in.Name == "" {
		return nil, domain.ErrInvalidInput
	}
	return s.repo.CreateShortlist(ctx, in)
}

func (s *Service) AddShortlistItem(ctx context.Context, in domain.ShortlistItemInput) (*domain.ShortlistItem, error) {
	in.Notes = strings.TrimSpace(in.Notes)
	if in.ShortlistID <= 0 || in.ApplicationID <= 0 {
		return nil, domain.ErrInvalidInput
	}
	return s.repo.AddShortlistItem(ctx, in)
}

// GetApplicationCVURL returns a short-lived presigned URL to download an
// application's CV.
func (s *Service) GetApplicationCVURL(ctx context.Context, id int64) (string, error) {
	key, err := s.repo.GetApplicationCVKey(ctx, id)
	if err != nil {
		return "", err
	}
	if key == "" {
		return "", domain.ErrApplicationNotFound
	}

	client, err := storage.Default()
	if err != nil {
		return "", err
	}
	return client.PresignDownload(ctx, key, cvURLTTL)
}

// SubmitCareerApplication persists a public application from the
// hire.voxire.com careers form and notifies the team.
func (s *Service) SubmitCareerApplication(ctx context.Context, in domain.CareerApplicationInput) (int64, error) {
	id, err := s.repo.CreateCareerApplication(ctx, in)
	if err != nil {
		// The email is the only record left: send it flagged as not saved.
		s.notifyCareerApplicationAsync(ctx, 0, in, noticeNotSaved)
		return 0, err
	}

	s.notifyCareerApplicationAsync(ctx, id, in, noticeForSaved(in.CVUploadFailed))

	return id, nil
}

// RecentCareerApplication returns the ID of an application the same email
// sent to jobID in the last few minutes, or 0. See RecentTalentApplication.
func (s *Service) RecentCareerApplication(ctx context.Context, jobID int64, email string) (int64, error) {
	return s.repo.FindRecentCareerApplicationID(ctx, jobID, email)
}

func (s *Service) notifyCareerApplicationAsync(ctx context.Context, id int64, in domain.CareerApplicationInput, notice submissionNotice) {
	// Detach from the request context — the goroutine outlives the handler — but
	// bound it, so a stalled email provider cannot hold the goroutine forever.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Global().Error("voxhire career application notification panicked", logger.Any("panic", r))
			}
		}()

		notifyCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), careerApplicationNotifyTimeout)
		defer cancel()
		s.notifyCareerApplication(notifyCtx, id, in, notice)
	}()
}
