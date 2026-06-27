package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// Repository persists VoxHire submissions in MySQL. Two tables only:
// company hiring inquiries and talent applications. CV files live in
// R2; we store the object key.
type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	autoMigrate(context.Background(), db)
	return &Repository{db: db}
}

func autoMigrate(ctx context.Context, db *sql.DB) {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS voxhire_company_leads (
			id            BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			company_name  VARCHAR(255) NOT NULL,
			contact_name  VARCHAR(255) NOT NULL,
			contact_title VARCHAR(255),
			work_email    VARCHAR(255) NOT NULL,
			phone         VARCHAR(64),
			role          VARCHAR(255) NOT NULL,
			headcount     VARCHAR(32),
			timeline      VARCHAR(32),
			service       VARCHAR(32),
			message       TEXT,
			agency_opt_in TINYINT(1) NOT NULL DEFAULT 0,
			status        VARCHAR(32) NOT NULL DEFAULT 'new',
			created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_voxhire_company_leads_email   (work_email),
			INDEX idx_voxhire_company_leads_status  (status),
			INDEX idx_voxhire_company_leads_created (created_at)
		)
	`); err != nil {
		logger.Ctx(ctx).Error("Failed to auto-migrate voxhire_company_leads", logger.Error(err))
	}

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS voxhire_talent_applications (
			id                   BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			full_name            VARCHAR(255) NOT NULL,
			email                VARCHAR(255) NOT NULL,
			phone                VARCHAR(64),
			country_of_residence VARCHAR(128),
			nationality          VARCHAR(128),
			track                VARCHAR(64),
			years_experience     VARCHAR(32),
			notice_period        VARCHAR(32),
			open_to              VARCHAR(255),
			current_role         VARCHAR(255),
			desired_roles        TEXT,
			salary_expectation   VARCHAR(128),
			work_authorization   VARCHAR(255),
			work_history         TEXT,
			education            TEXT,
			linkedin_url         VARCHAR(512),
			portfolio_url        VARCHAR(512),
			pitch                TEXT,
			consent              TINYINT(1) NOT NULL DEFAULT 0,
			cv_key               VARCHAR(512),
			cv_processing_status VARCHAR(32) NOT NULL DEFAULT 'pending',
			cv_processing_error  TEXT,
			cv_extracted_text    MEDIUMTEXT,
			cv_summary           TEXT,
			cv_processed_at      TIMESTAMP NULL,
			status               VARCHAR(32) NOT NULL DEFAULT 'new',
			created_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_voxhire_talent_applications_email   (email),
			INDEX idx_voxhire_talent_applications_track   (track),
			INDEX idx_voxhire_talent_applications_status  (status),
			INDEX idx_voxhire_talent_applications_created (created_at)
		)
	`); err != nil {
		logger.Ctx(ctx).Error("Failed to auto-migrate voxhire_talent_applications", logger.Error(err))
	}
	ensureColumn(ctx, db, "voxhire_talent_applications", "current_role", "current_role VARCHAR(255)")
	ensureColumn(ctx, db, "voxhire_talent_applications", "desired_roles", "desired_roles TEXT")
	ensureColumn(ctx, db, "voxhire_talent_applications", "salary_expectation", "salary_expectation VARCHAR(128)")
	ensureColumn(ctx, db, "voxhire_talent_applications", "work_authorization", "work_authorization VARCHAR(255)")
	ensureColumn(ctx, db, "voxhire_talent_applications", "work_history", "work_history TEXT")
	ensureColumn(ctx, db, "voxhire_talent_applications", "education", "education TEXT")
	ensureColumn(ctx, db, "voxhire_talent_applications", "cv_processing_status", "cv_processing_status VARCHAR(32) NOT NULL DEFAULT 'pending'")
	ensureColumn(ctx, db, "voxhire_talent_applications", "cv_processing_error", "cv_processing_error TEXT")
	ensureColumn(ctx, db, "voxhire_talent_applications", "cv_extracted_text", "cv_extracted_text MEDIUMTEXT")
	ensureColumn(ctx, db, "voxhire_talent_applications", "cv_summary", "cv_summary TEXT")
	ensureColumn(ctx, db, "voxhire_talent_applications", "cv_processed_at", "cv_processed_at TIMESTAMP NULL")
	ensureColumn(ctx, db, "voxire_career_applications", "cv_processing_status", "cv_processing_status VARCHAR(32) NOT NULL DEFAULT 'pending'")
	ensureColumn(ctx, db, "voxire_career_applications", "cv_processing_error", "cv_processing_error TEXT")
	ensureColumn(ctx, db, "voxire_career_applications", "cv_extracted_text", "cv_extracted_text MEDIUMTEXT")
	ensureColumn(ctx, db, "voxire_career_applications", "cv_summary", "cv_summary TEXT")
	ensureColumn(ctx, db, "voxire_career_applications", "cv_processed_at", "cv_processed_at TIMESTAMP NULL")
	ensureColumn(ctx, db, "voxire_career_applications", "review_rating", "review_rating TINYINT NOT NULL DEFAULT 0")
	ensureColumn(ctx, db, "voxire_career_applications", "review_notes", "review_notes TEXT")
	ensureColumn(ctx, db, "voxire_career_applications", "next_step", "next_step VARCHAR(255)")
	ensureColumn(ctx, db, "voxire_career_applications", "reviewed_at", "reviewed_at TIMESTAMP NULL")

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS voxhire_admins (
			id            BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			email         VARCHAR(255) NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			name          VARCHAR(255),
			created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_voxhire_admins_email (email)
		)
	`); err != nil {
		logger.Ctx(ctx).Error("Failed to auto-migrate voxhire_admins", logger.Error(err))
	}
}

func ensureColumn(ctx context.Context, db *sql.DB, tableName, columnName, ddl string) {
	var count int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		  FROM INFORMATION_SCHEMA.COLUMNS
		 WHERE TABLE_SCHEMA = DATABASE()
		   AND TABLE_NAME = ?
		   AND COLUMN_NAME = ?
	`, tableName, columnName).Scan(&count)
	if err != nil {
		logger.Ctx(ctx).Warn("Failed to inspect column for voxhire auto-migrate",
			logger.String("table", tableName),
			logger.String("column", columnName),
			logger.Error(err),
		)
		return
	}
	if count > 0 {
		return
	}
	if _, err := db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, ddl)); err != nil {
		logger.Ctx(ctx).Warn("Failed to add column for voxhire auto-migrate",
			logger.String("table", tableName),
			logger.String("column", columnName),
			logger.Error(err),
		)
	}
}

// CreateCompanyLead inserts a hiring inquiry and returns its ID.
func (r *Repository) CreateCompanyLead(ctx context.Context, input domain.CompanyLeadInput) (int64, error) {
	query, args, err := sq.
		Insert("voxhire_company_leads").
		Columns("company_name", "contact_name", "contact_title", "work_email", "phone", "role", "headcount", "timeline", "service", "message", "agency_opt_in").
		Values(
			input.CompanyName,
			input.ContactName,
			input.ContactTitle,
			input.WorkEmail,
			input.Phone,
			input.Role,
			input.Headcount,
			input.Timeline,
			input.Service,
			input.Message,
			input.AgencyOptIn,
		).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build insert voxhire company lead query", logger.Error(err))
		return 0, err
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert voxhire company lead", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

// CreateTalentApplication inserts a talent application and returns its ID.
func (r *Repository) CreateTalentApplication(ctx context.Context, input domain.TalentApplicationInput) (int64, error) {
	query, args, err := sq.
		Insert("voxhire_talent_applications").
		Columns("full_name", "email", "phone", "country_of_residence", "nationality", "track", "years_experience", "notice_period", "open_to", "current_role", "desired_roles", "salary_expectation", "work_authorization", "work_history", "education", "linkedin_url", "portfolio_url", "pitch", "consent", "cv_key", "cv_processing_status").
		Values(
			input.FullName,
			input.Email,
			input.Phone,
			input.CountryOfResidence,
			input.Nationality,
			input.Track,
			input.YearsExperience,
			input.NoticePeriod,
			input.OpenTo,
			input.CurrentRole,
			input.DesiredRoles,
			input.SalaryExpectation,
			input.WorkAuthorization,
			input.WorkHistory,
			input.Education,
			input.LinkedinURL,
			input.PortfolioURL,
			input.Pitch,
			input.Consent,
			input.CVKey,
			"pending",
		).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build insert voxhire talent application query", logger.Error(err))
		return 0, err
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert voxhire talent application", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}
