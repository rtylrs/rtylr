package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	"github.com/voxire/onion/internal/modules/voxhire/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// Repository persists VoxHire submissions in MySQL. Two tables only:
// company hiring inquiries and talent applications. CV files live in
// R2; we store the object key.
type Repository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	autoMigrate(context.Background(), db)
	return &Repository{db: db, queries: sqlc.New(db)}
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
			cv_profile_json      TEXT,
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
	ensureColumn(ctx, db, "voxhire_talent_applications", "cv_profile_json", "cv_profile_json TEXT")
	ensureColumn(ctx, db, "voxhire_talent_applications", "cv_processed_at", "cv_processed_at TIMESTAMP NULL")
	ensureColumn(ctx, db, "voxire_career_applications", "cv_processing_status", "cv_processing_status VARCHAR(32) NOT NULL DEFAULT 'pending'")
	ensureColumn(ctx, db, "voxire_career_applications", "cv_processing_error", "cv_processing_error TEXT")
	ensureColumn(ctx, db, "voxire_career_applications", "cv_extracted_text", "cv_extracted_text MEDIUMTEXT")
	ensureColumn(ctx, db, "voxire_career_applications", "cv_summary", "cv_summary TEXT")
	ensureColumn(ctx, db, "voxire_career_applications", "cv_profile_json", "cv_profile_json TEXT")
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

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS voxhire_application_events (
			id             BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			application_id BIGINT NOT NULL,
			admin_id       BIGINT,
			kind           VARCHAR(64) NOT NULL,
			title          VARCHAR(255) NOT NULL,
			detail         TEXT,
			created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_voxhire_events_application (application_id),
			INDEX idx_voxhire_events_created (created_at)
		)
	`); err != nil {
		logger.Ctx(ctx).Error("Failed to auto-migrate voxhire_application_events", logger.Error(err))
	}

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS voxhire_follow_up_tasks (
			id             BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			application_id BIGINT NOT NULL,
			title          VARCHAR(255) NOT NULL,
			status         VARCHAR(32) NOT NULL DEFAULT 'open',
			due_at         TIMESTAMP NULL,
			completed_at   TIMESTAMP NULL,
			created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_voxhire_tasks_application (application_id),
			INDEX idx_voxhire_tasks_status_due (status, due_at),
			UNIQUE KEY uk_voxhire_task_application_title (application_id, title)
		)
	`); err != nil {
		logger.Ctx(ctx).Error("Failed to auto-migrate voxhire_follow_up_tasks", logger.Error(err))
	}

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS voxhire_shortlists (
			id         BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			job_id     BIGINT NOT NULL,
			name       VARCHAR(255) NOT NULL,
			status     VARCHAR(32) NOT NULL DEFAULT 'draft',
			notes      TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_voxhire_shortlists_job (job_id)
		)
	`); err != nil {
		logger.Ctx(ctx).Error("Failed to auto-migrate voxhire_shortlists", logger.Error(err))
	}

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS voxhire_shortlist_items (
			id             BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			shortlist_id   BIGINT NOT NULL,
			application_id BIGINT NOT NULL,
			sort_order     INT NOT NULL DEFAULT 0,
			notes          TEXT,
			created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_voxhire_shortlist_items_shortlist (shortlist_id),
			UNIQUE KEY uk_voxhire_shortlist_application (shortlist_id, application_id)
		)
	`); err != nil {
		logger.Ctx(ctx).Error("Failed to auto-migrate voxhire_shortlist_items", logger.Error(err))
	}
}

var allowedVoxhireSchemaColumns = map[[2]string]string{
	{"voxhire_talent_applications", "current_role"}:         "current_role VARCHAR(255)",
	{"voxhire_talent_applications", "desired_roles"}:        "desired_roles TEXT",
	{"voxhire_talent_applications", "salary_expectation"}:   "salary_expectation VARCHAR(128)",
	{"voxhire_talent_applications", "work_authorization"}:   "work_authorization VARCHAR(255)",
	{"voxhire_talent_applications", "work_history"}:         "work_history TEXT",
	{"voxhire_talent_applications", "education"}:            "education TEXT",
	{"voxhire_talent_applications", "cv_processing_status"}: "cv_processing_status VARCHAR(32) NOT NULL DEFAULT 'pending'",
	{"voxhire_talent_applications", "cv_processing_error"}:  "cv_processing_error TEXT",
	{"voxhire_talent_applications", "cv_extracted_text"}:    "cv_extracted_text MEDIUMTEXT",
	{"voxhire_talent_applications", "cv_summary"}:           "cv_summary TEXT",
	{"voxhire_talent_applications", "cv_profile_json"}:      "cv_profile_json TEXT",
	{"voxhire_talent_applications", "cv_processed_at"}:      "cv_processed_at TIMESTAMP NULL",
	{"voxire_career_applications", "cv_processing_status"}:  "cv_processing_status VARCHAR(32) NOT NULL DEFAULT 'pending'",
	{"voxire_career_applications", "cv_processing_error"}:   "cv_processing_error TEXT",
	{"voxire_career_applications", "cv_extracted_text"}:     "cv_extracted_text MEDIUMTEXT",
	{"voxire_career_applications", "cv_summary"}:            "cv_summary TEXT",
	{"voxire_career_applications", "cv_profile_json"}:       "cv_profile_json TEXT",
	{"voxire_career_applications", "cv_processed_at"}:       "cv_processed_at TIMESTAMP NULL",
	{"voxire_career_applications", "review_rating"}:         "review_rating TINYINT NOT NULL DEFAULT 0",
	{"voxire_career_applications", "review_notes"}:          "review_notes TEXT",
	{"voxire_career_applications", "next_step"}:             "next_step VARCHAR(255)",
	{"voxire_career_applications", "reviewed_at"}:           "reviewed_at TIMESTAMP NULL",
}

func ensureColumn(ctx context.Context, db *sql.DB, tableName, columnName, ddl string) {
	allowedDDL, ok := allowedVoxhireSchemaColumns[[2]string{tableName, columnName}]
	if !ok || allowedDDL != ddl {
		logger.Ctx(ctx).Error("Rejected unknown voxhire schema change",
			logger.String("table", tableName),
			logger.String("column", columnName),
		)
		return
	}

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
	// nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query
	if _, err := db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, allowedDDL)); err != nil {
		logger.Ctx(ctx).Warn("Failed to add column for voxhire auto-migrate",
			logger.String("table", tableName),
			logger.String("column", columnName),
			logger.Error(err),
		)
	}
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
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert voxhire talent application", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}
