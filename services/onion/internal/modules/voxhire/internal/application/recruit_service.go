package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/storage"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

const (
	adminJWTIssuer = "voxhire-admin"
	adminTokenTTL  = 7 * 24 * time.Hour
	cvURLTTL       = 10 * time.Minute

	minAdminPasswordLen = 6
	careersCatalogKey   = "cache:careers:active"
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
	claims := &domain.AdminClaims{AdminID: admin.ID, Email: admin.Email}
	return helper.CreateJWT(claims)
}

// ValidateAdminToken parses and validates an admin bearer token.
func (s *Service) ValidateAdminToken(ctx context.Context, token string) (*domain.AdminClaims, error) {
	helper, err := adminJWTHelper()
	if err != nil {
		return nil, err
	}
	var claims domain.AdminClaims
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
	cacher.GetInMemoryCache(ctx).DeleteByPattern(ctx, careersCatalogKey)
	cacher.GetRedisCache().DeleteByPattern(ctx, careersCatalogKey)
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
	return s.repo.ListApplications(ctx, f)
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
		return 0, err
	}
	go s.notifyCareerApplication(context.WithoutCancel(ctx), id, in)
	return id, nil
}
