package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// Repository persists VoxHire hiring leads, applications and admin workflow state.
// CV files live in R2; this package stores their object keys.
type Repository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db, queries: sqlc.New(db)}
}

// CreateCompanyLead inserts a hiring inquiry and returns its ID.
func (r *Repository) CreateCompanyLead(ctx context.Context, input domain.CompanyLeadInput) (int64, error) {
	result, err := r.queries.CreateVoxhireCompanyLead(ctx, sqlc.CreateVoxhireCompanyLeadParams{
		CompanyName:  input.CompanyName,
		ContactName:  input.ContactName,
		ContactTitle: sql.NullString{String: input.ContactTitle, Valid: input.ContactTitle != ""},
		WorkEmail:    input.WorkEmail,
		Phone:        sql.NullString{String: input.Phone, Valid: input.Phone != ""},
		Role:         input.Role,
		Headcount:    sql.NullString{String: input.Headcount, Valid: input.Headcount != ""},
		Timeline:     sql.NullString{String: input.Timeline, Valid: input.Timeline != ""},
		Service:      sql.NullString{String: input.Service, Valid: input.Service != ""},
		Message:      sql.NullString{String: input.Message, Valid: input.Message != ""},
		AgencyOptIn:  input.AgencyOptIn,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert voxhire company lead", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

// CreateTalentApplication inserts a talent application and returns its ID.
func (r *Repository) CreateTalentApplication(ctx context.Context, input domain.TalentApplicationInput) (int64, error) {
	result, err := r.queries.CreateVoxhireTalentApplication(ctx, sqlc.CreateVoxhireTalentApplicationParams{
		FullName:           input.FullName,
		Email:              input.Email,
		Phone:              sql.NullString{String: input.Phone, Valid: input.Phone != ""},
		CountryOfResidence: sql.NullString{String: input.CountryOfResidence, Valid: input.CountryOfResidence != ""},
		Nationality:        sql.NullString{String: input.Nationality, Valid: input.Nationality != ""},
		Track:              sql.NullString{String: input.Track, Valid: input.Track != ""},
		YearsExperience:    sql.NullString{String: input.YearsExperience, Valid: input.YearsExperience != ""},
		NoticePeriod:       sql.NullString{String: input.NoticePeriod, Valid: input.NoticePeriod != ""},
		OpenTo:             sql.NullString{String: input.OpenTo, Valid: input.OpenTo != ""},
		CurrentRole:        sql.NullString{String: input.CurrentRole, Valid: input.CurrentRole != ""},
		DesiredRoles:       sql.NullString{String: input.DesiredRoles, Valid: input.DesiredRoles != ""},
		SalaryExpectation:  sql.NullString{String: input.SalaryExpectation, Valid: input.SalaryExpectation != ""},
		WorkAuthorization:  sql.NullString{String: input.WorkAuthorization, Valid: input.WorkAuthorization != ""},
		WorkHistory:        sql.NullString{String: input.WorkHistory, Valid: input.WorkHistory != ""},
		Education:          sql.NullString{String: input.Education, Valid: input.Education != ""},
		LinkedinUrl:        sql.NullString{String: input.LinkedinURL, Valid: input.LinkedinURL != ""},
		PortfolioUrl:       sql.NullString{String: input.PortfolioURL, Valid: input.PortfolioURL != ""},
		Pitch:              sql.NullString{String: input.Pitch, Valid: input.Pitch != ""},
		Consent:            input.Consent,
		CvKey:              sql.NullString{String: input.CVKey, Valid: input.CVKey != ""},
		CvProcessingStatus: domain.SubmissionCVStatus(input.CVUploadFailed),
		CvProcessingError:  sql.NullString{String: input.CVUploadError, Valid: input.CVUploadFailed && input.CVUploadError != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert voxhire talent application", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

// FindRecentTalentApplicationID returns the newest talent application from
// email in the last 10 minutes, or 0 when there is none.
func (r *Repository) FindRecentTalentApplicationID(ctx context.Context, email string) (int64, error) {
	id, err := r.queries.FindRecentTalentApplicationID(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return id, err
}
