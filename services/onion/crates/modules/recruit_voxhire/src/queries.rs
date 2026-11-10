//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `AnalyticsApplicationDailyCounts` (Go `AnalyticsApplicationDailyCountsParams`).
pub struct AnalyticsApplicationDailyCountsParams {
    pub owner_id: i64,
}

/// `AnalyticsApplicationDailyCounts`: `many`.
pub async fn analytics_application_daily_counts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AnalyticsApplicationDailyCountsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT DATE_FORMAT(a.created_at, '%Y-%m-%d') AS day, COUNT(*) AS count\nFROM voxire_career_applications a\nJOIN careers c ON c.id = a.job_id AND c.owner_id = ?\nWHERE a.created_at >= DATE_SUB(UTC_DATE(), INTERVAL 13 DAY)\nGROUP BY day")
        .bind(params.owner_id)
        .execute(exec)
        .await
}

/// Params for `AnalyticsApplicationTotals` (Go `AnalyticsApplicationTotalsParams`).
pub struct AnalyticsApplicationTotalsParams {
    pub owner_id: i64,
}

/// `AnalyticsApplicationTotals`: `one`.
pub async fn analytics_application_totals<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AnalyticsApplicationTotalsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS applications_total,\n       COALESCE(SUM(CASE WHEN a.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 7 DAY) THEN 1 ELSE 0 END), 0) AS applications_7d,\n       COALESCE(SUM(CASE WHEN a.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 30 DAY) THEN 1 ELSE 0 END), 0) AS applications_30d,\n       COALESCE(SUM(CASE WHEN LOWER(COALESCE(a.status, 'new')) = 'new' THEN 1 ELSE 0 END), 0) AS new_applications,\n       COALESCE(SUM(CASE WHEN a.cv_key IS NOT NULL AND a.cv_key <> '' THEN 1 ELSE 0 END), 0) AS cv_count,\n       AVG(NULLIF(a.salary_expectation, 0)) AS avg_salary_expectation\nFROM voxire_career_applications a\nJOIN careers c ON c.id = a.job_id AND c.owner_id = ?")
        .bind(params.owner_id)
        .fetch_optional(exec)
        .await
}

/// Params for `AnalyticsApplicationsByStatus` (Go `AnalyticsApplicationsByStatusParams`).
pub struct AnalyticsApplicationsByStatusParams {
    pub owner_id: i64,
}

/// `AnalyticsApplicationsByStatus`: `many`.
pub async fn analytics_applications_by_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AnalyticsApplicationsByStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT LOWER(COALESCE(NULLIF(a.status, ''), 'new')) AS bucket, COUNT(*) AS count\nFROM voxire_career_applications a\nJOIN careers c ON c.id = a.job_id AND c.owner_id = ?\nWHERE a.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)\nGROUP BY bucket\nORDER BY count DESC, bucket ASC")
        .bind(params.owner_id)
        .execute(exec)
        .await
}

/// Params for `AnalyticsCVProcessingTotals` (Go `AnalyticsCVProcessingTotalsParams`).
pub struct AnalyticsCVProcessingTotalsParams {
    pub owner_id: i64,
}

/// `AnalyticsCVProcessingTotals`: `one`.
pub async fn analytics_cv_processing_totals<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AnalyticsCVProcessingTotalsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(CASE WHEN status = 'processed' THEN 1 ELSE 0 END), 0) AS cv_processed_total,\n       COALESCE(SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END), 0) AS cv_error_total,\n       COALESCE(SUM(CASE WHEN status NOT IN ('processed', 'error') THEN 1 ELSE 0 END), 0) AS cv_pending_total\nFROM (\n    SELECT COALESCE(a.cv_processing_status, 'pending') AS status\n      FROM voxire_career_applications a\n      JOIN careers c ON c.id = a.job_id AND c.owner_id = ?\n     WHERE a.cv_key IS NOT NULL AND a.cv_key <> ''\n    UNION ALL\n    SELECT COALESCE(cv_processing_status, 'pending') AS status\n      FROM voxhire_talent_applications\n     WHERE cv_key IS NOT NULL AND cv_key <> ''\n) cv_statuses")
        .bind(params.owner_id)
        .fetch_optional(exec)
        .await
}

/// Params for `AnalyticsCVStatusMix` (Go `AnalyticsCVStatusMixParams`).
pub struct AnalyticsCVStatusMixParams {
    pub owner_id: i64,
}

/// `AnalyticsCVStatusMix`: `many`.
pub async fn analytics_cv_status_mix<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AnalyticsCVStatusMixParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT status AS bucket, COUNT(*) AS count\nFROM (\n    SELECT LOWER(COALESCE(a.cv_processing_status, 'pending')) AS status\n      FROM voxire_career_applications a\n      JOIN careers c ON c.id = a.job_id AND c.owner_id = ?\n     WHERE a.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)\n       AND a.cv_key IS NOT NULL AND a.cv_key <> ''\n    UNION ALL\n    SELECT LOWER(COALESCE(cv_processing_status, 'pending')) AS status\n      FROM voxhire_talent_applications\n     WHERE created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)\n       AND cv_key IS NOT NULL AND cv_key <> ''\n) statuses\nGROUP BY status\nORDER BY count DESC, status ASC")
        .bind(params.owner_id)
        .execute(exec)
        .await
}

/// Params for `AnalyticsCareerJobTotals` (Go `AnalyticsCareerJobTotalsParams`).
pub struct AnalyticsCareerJobTotalsParams {
    pub owner_id: i64,
}

/// `AnalyticsCareerJobTotals`: `one`.
pub async fn analytics_career_job_totals<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AnalyticsCareerJobTotalsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(SUM(CASE WHEN is_active = 1 THEN 1 ELSE 0 END), 0) AS open_jobs,\n       COALESCE(SUM(CASE WHEN is_active = 0 THEN 1 ELSE 0 END), 0) AS closed_jobs\nFROM careers\nWHERE owner_id = ?")
        .bind(params.owner_id)
        .fetch_optional(exec)
        .await
}

/// `AnalyticsLeadDailyCounts`: `many`.
pub async fn analytics_lead_daily_counts<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT DATE_FORMAT(created_at, '%Y-%m-%d') AS day, COUNT(*) AS count\nFROM voxhire_company_leads\nWHERE created_at >= DATE_SUB(UTC_DATE(), INTERVAL 13 DAY)\nGROUP BY day")
        .execute(exec)
        .await
}

/// `AnalyticsLeadServiceMix`: `many`.
pub async fn analytics_lead_service_mix<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT LOWER(COALESCE(NULLIF(service, ''), 'unsure')) AS bucket, COUNT(*) AS count\nFROM voxhire_company_leads\nWHERE created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)\nGROUP BY bucket\nORDER BY count DESC, bucket ASC")
        .execute(exec)
        .await
}

/// `AnalyticsLeadTimelineMix`: `many`.
pub async fn analytics_lead_timeline_mix<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT LOWER(COALESCE(NULLIF(timeline, ''), 'unknown')) AS bucket, COUNT(*) AS count\nFROM voxhire_company_leads\nWHERE created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)\nGROUP BY bucket\nORDER BY count DESC, bucket ASC")
        .execute(exec)
        .await
}

/// `AnalyticsLeadTotals`: `one`.
pub async fn analytics_lead_totals<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS leads_total,\n       COALESCE(SUM(CASE WHEN created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 7 DAY) THEN 1 ELSE 0 END), 0) AS leads_7d,\n       COALESCE(SUM(CASE WHEN created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 30 DAY) THEN 1 ELSE 0 END), 0) AS leads_30d\nFROM voxhire_company_leads")
        .fetch_optional(exec)
        .await
}

/// Params for `AnalyticsRecentActivity` (Go `AnalyticsRecentActivityParams`).
pub struct AnalyticsRecentActivityParams {
    pub owner_id: i64,
}

/// `AnalyticsRecentActivity`: `many`.
pub async fn analytics_recent_activity<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AnalyticsRecentActivityParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT kind, id, primary_text, secondary_text, meta, created_at\nFROM (\n    SELECT 'application' AS kind,\n           a.id AS id,\n           c.title AS primary_text,\n           a.name AS secondary_text,\n           COALESCE(a.status, 'new') AS meta,\n           a.created_at AS created_at\n      FROM voxire_career_applications a\n      JOIN careers c ON c.id = a.job_id AND c.owner_id = ?\n    UNION ALL\n    SELECT 'lead' AS kind,\n           id,\n           company_name AS primary_text,\n           COALESCE(role, '') AS secondary_text,\n           COALESCE(timeline, '') AS meta,\n           created_at\n      FROM voxhire_company_leads\n) recent\nORDER BY created_at DESC, id DESC\nLIMIT 10")
        .bind(params.owner_id)
        .execute(exec)
        .await
}

/// `AnalyticsTalentPoolTotals`: `one`.
pub async fn analytics_talent_pool_totals<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS talent_pool_total,\n       COALESCE(SUM(CASE WHEN created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 7 DAY) THEN 1 ELSE 0 END), 0) AS talent_pool_7d,\n       COALESCE(SUM(CASE WHEN created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 30 DAY) THEN 1 ELSE 0 END), 0) AS talent_pool_30d\nFROM voxhire_talent_applications")
        .fetch_optional(exec)
        .await
}

/// Params for `AnalyticsTopOpenings` (Go `AnalyticsTopOpeningsParams`).
pub struct AnalyticsTopOpeningsParams {
    pub owner_id: i64,
}

/// `AnalyticsTopOpenings`: `many`.
pub async fn analytics_top_openings<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &AnalyticsTopOpeningsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT c.id,\n       c.title,\n       c.is_active,\n       COALESCE(COUNT(a.id), 0) AS total_count,\n       COALESCE(SUM(CASE WHEN LOWER(COALESCE(a.status, 'new')) = 'new' THEN 1 ELSE 0 END), 0) AS new_count,\n       COALESCE(SUM(CASE WHEN a.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 30 DAY) THEN 1 ELSE 0 END), 0) AS recent_count,\n       MAX(a.created_at) AS last_application_at\nFROM careers c\nLEFT JOIN voxire_career_applications a ON a.job_id = c.id\nWHERE c.owner_id = ?\nGROUP BY c.id, c.title, c.is_active, c.created_at\nORDER BY recent_count DESC, total_count DESC, c.created_at DESC\nLIMIT 6")
        .bind(params.owner_id)
        .execute(exec)
        .await
}

/// `CountAdmins`: `one`.
pub async fn count_admins<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COUNT(*) AS count FROM voxhire_admins")
        .fetch_optional(exec)
        .await
}

/// Params for `CreateAdmin` (Go `CreateAdminParams`).
pub struct CreateAdminParams {
    pub email: String,
    pub password_hash: String,
    pub name: Option<String>,
}

/// `CreateAdmin`: `execresult`.
pub async fn create_admin<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateAdminParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO voxhire_admins (email, password_hash, name) VALUES (?, ?, ?)")
        .bind(&params.email)
        .bind(&params.password_hash)
        .bind(&params.name)
        .execute(exec)
        .await
}

/// Params for `CreateCareer` (Go `CreateCareerParams`).
pub struct CreateCareerParams {
    pub owner_id: i64,
    pub slug: String,
    pub title: String,
    pub department: Option<String>,
    pub location: Option<String>,
    pub employment_type: Option<String>,
    pub work_mode: Option<String>,
    pub seniority: Option<String>,
    pub details: String,
    pub salary_min: Option<String>,
    pub salary_max: Option<String>,
    pub salary_currency: Option<String>,
}

/// `CreateCareer`: `execresult`.
pub async fn create_career<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateCareerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO careers (\n    owner_id, slug, title, department, location, employment_type,\n    work_mode, seniority, details, salary_min, salary_max, salary_currency,\n    is_active, is_published\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, TRUE, TRUE)")
        .bind(params.owner_id)
        .bind(&params.slug)
        .bind(&params.title)
        .bind(&params.department)
        .bind(&params.location)
        .bind(&params.employment_type)
        .bind(&params.work_mode)
        .bind(&params.seniority)
        .bind(&params.details)
        .bind(&params.salary_min)
        .bind(&params.salary_max)
        .bind(&params.salary_currency)
        .execute(exec)
        .await
}

/// Params for `CreateCareerApplication` (Go `CreateCareerApplicationParams`).
pub struct CreateCareerApplicationParams {
    pub job_id: i64,
    pub name: String,
    pub email: String,
    pub phone: Option<String>,
    pub location: Option<String>,
    pub salary_expectation: Option<String>,
    pub work_history: Option<String>,
    pub education: Option<String>,
    pub experience_summary: Option<String>,
    pub cover_letter: Option<String>,
    pub cv_key: Option<String>,
    pub cv_processing_status: String,
    pub cv_processing_error: Option<String>,
}

/// `CreateCareerApplication`: `execresult`.
pub async fn create_career_application<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateCareerApplicationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO voxire_career_applications (\n    job_id, name, email, phone, location, salary_expectation,\n    work_history, education, experience_summary, cover_letter, cv_key, status,\n    cv_processing_status, cv_processing_error\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'new', ?, ?)")
        .bind(params.job_id)
        .bind(&params.name)
        .bind(&params.email)
        .bind(&params.phone)
        .bind(&params.location)
        .bind(&params.salary_expectation)
        .bind(&params.work_history)
        .bind(&params.education)
        .bind(&params.experience_summary)
        .bind(&params.cover_letter)
        .bind(&params.cv_key)
        .bind(&params.cv_processing_status)
        .bind(&params.cv_processing_error)
        .execute(exec)
        .await
}

/// Params for `CreateShortlist` (Go `CreateShortlistParams`).
pub struct CreateShortlistParams {
    pub job_id: i64,
    pub name: String,
    pub status: String,
    pub notes: Option<String>,
}

/// `CreateShortlist`: `execresult`.
pub async fn create_shortlist<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateShortlistParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO voxhire_shortlists (job_id, name, status, notes) VALUES (?, ?, ?, ?)")
        .bind(params.job_id)
        .bind(&params.name)
        .bind(&params.status)
        .bind(&params.notes)
        .execute(exec)
        .await
}

/// Params for `CreateVoxhireCompanyLead` (Go `CreateVoxhireCompanyLeadParams`).
pub struct CreateVoxhireCompanyLeadParams {
    pub company_name: String,
    pub contact_name: String,
    pub contact_title: Option<String>,
    pub work_email: String,
    pub phone: Option<String>,
    pub role: String,
    pub headcount: Option<String>,
    pub timeline: Option<String>,
    pub service: Option<String>,
    pub message: Option<String>,
    pub agency_opt_in: bool,
}

/// `CreateVoxhireCompanyLead`: `execresult`.
pub async fn create_voxhire_company_lead<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateVoxhireCompanyLeadParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO voxhire_company_leads (\n    company_name,\n    contact_name,\n    contact_title,\n    work_email,\n    phone,\n    role,\n    headcount,\n    timeline,\n    service,\n    message,\n    agency_opt_in\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(&params.company_name)
        .bind(&params.contact_name)
        .bind(&params.contact_title)
        .bind(&params.work_email)
        .bind(&params.phone)
        .bind(&params.role)
        .bind(&params.headcount)
        .bind(&params.timeline)
        .bind(&params.service)
        .bind(&params.message)
        .bind(params.agency_opt_in)
        .execute(exec)
        .await
}

/// Params for `CreateVoxhireTalentApplication` (Go `CreateVoxhireTalentApplicationParams`).
pub struct CreateVoxhireTalentApplicationParams {
    pub full_name: String,
    pub email: String,
    pub phone: Option<String>,
    pub country_of_residence: Option<String>,
    pub nationality: Option<String>,
    pub track: Option<String>,
    pub years_experience: Option<String>,
    pub notice_period: Option<String>,
    pub open_to: Option<String>,
    pub current_role: Option<String>,
    pub desired_roles: Option<String>,
    pub salary_expectation: Option<String>,
    pub work_authorization: Option<String>,
    pub work_history: Option<String>,
    pub education: Option<String>,
    pub linkedin_url: Option<String>,
    pub portfolio_url: Option<String>,
    pub pitch: Option<String>,
    pub consent: bool,
    pub cv_key: Option<String>,
    pub cv_processing_status: String,
    pub cv_processing_error: Option<String>,
}

/// `CreateVoxhireTalentApplication`: `execresult`.
pub async fn create_voxhire_talent_application<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateVoxhireTalentApplicationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO voxhire_talent_applications (\n    full_name,\n    email,\n    phone,\n    country_of_residence,\n    nationality,\n    track,\n    years_experience,\n    notice_period,\n    open_to,\n    `current_role`,\n    desired_roles,\n    salary_expectation,\n    work_authorization,\n    work_history,\n    education,\n    linkedin_url,\n    portfolio_url,\n    pitch,\n    consent,\n    cv_key,\n    cv_processing_status,\n    cv_processing_error\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(&params.full_name)
        .bind(&params.email)
        .bind(&params.phone)
        .bind(&params.country_of_residence)
        .bind(&params.nationality)
        .bind(&params.track)
        .bind(&params.years_experience)
        .bind(&params.notice_period)
        .bind(&params.open_to)
        .bind(&params.current_role)
        .bind(&params.desired_roles)
        .bind(&params.salary_expectation)
        .bind(&params.work_authorization)
        .bind(&params.work_history)
        .bind(&params.education)
        .bind(&params.linkedin_url)
        .bind(&params.portfolio_url)
        .bind(&params.pitch)
        .bind(params.consent)
        .bind(&params.cv_key)
        .bind(&params.cv_processing_status)
        .bind(&params.cv_processing_error)
        .execute(exec)
        .await
}

/// Params for `FindRecentCareerApplicationID` (Go `FindRecentCareerApplicationIDParams`).
pub struct FindRecentCareerApplicationIDParams {
    pub job_id: i64,
    pub email: String,
}

/// `FindRecentCareerApplicationID`: `one`.
pub async fn find_recent_career_application_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FindRecentCareerApplicationIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM voxire_career_applications\nWHERE job_id = ? AND email = ? AND created_at >= NOW() - INTERVAL 10 MINUTE\nORDER BY id DESC\nLIMIT 1")
        .bind(params.job_id)
        .bind(&params.email)
        .fetch_optional(exec)
        .await
}

/// Params for `FindRecentTalentApplicationID` (Go `FindRecentTalentApplicationIDParams`).
pub struct FindRecentTalentApplicationIDParams {
    pub email: String,
}

/// `FindRecentTalentApplicationID`: `one`.
pub async fn find_recent_talent_application_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &FindRecentTalentApplicationIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id FROM voxhire_talent_applications\nWHERE email = ? AND created_at >= NOW() - INTERVAL 10 MINUTE\nORDER BY id DESC\nLIMIT 1")
        .bind(&params.email)
        .fetch_optional(exec)
        .await
}

/// Params for `GetAdminByEmail` (Go `GetAdminByEmailParams`).
pub struct GetAdminByEmailParams {
    pub email: String,
}

/// `GetAdminByEmail`: `one`.
pub async fn get_admin_by_email<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetAdminByEmailParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, email, COALESCE(name,'') AS name, password_hash, created_at\nFROM voxhire_admins WHERE email = ?")
        .bind(&params.email)
        .fetch_optional(exec)
        .await
}

/// Params for `GetAdminByID` (Go `GetAdminByIDParams`).
pub struct GetAdminByIDParams {
    pub id: i64,
}

/// `GetAdminByID`: `one`.
pub async fn get_admin_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetAdminByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, email, COALESCE(name,'') AS name, password_hash, created_at\nFROM voxhire_admins WHERE id = ?")
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetApplicationCVKey` (Go `GetApplicationCVKeyParams`).
pub struct GetApplicationCVKeyParams {
    pub owner_id: i64,
    pub id: i64,
}

/// `GetApplicationCVKey`: `one`.
pub async fn get_application_cv_key<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetApplicationCVKeyParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(a.cv_key, '') AS cv_key\nFROM voxire_career_applications a\nJOIN careers c ON c.id = a.job_id AND c.owner_id = ?\nWHERE a.id = ?")
        .bind(params.owner_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetCareerOpeningByID` (Go `GetCareerOpeningByIDParams`).
pub struct GetCareerOpeningByIDParams {
    pub owner_id: i64,
    pub id: i64,
}

/// `GetCareerOpeningByID`: `one`.
pub async fn get_career_opening_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetCareerOpeningByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT c.id,\n       COALESCE(c.slug, '') AS slug,\n       c.title,\n       COALESCE(c.department, '') AS department,\n       COALESCE(c.location, '') AS location,\n       COALESCE(c.employment_type, '') AS employment_type,\n       COALESCE(c.work_mode, '') AS work_mode,\n       COALESCE(c.seniority, '') AS seniority,\n       COALESCE(c.details, '') AS details,\n       COALESCE(c.salary_min, 0) AS salary_min,\n       COALESCE(c.salary_max, 0) AS salary_max,\n       COALESCE(c.salary_currency, 'USD') AS salary_currency,\n       c.is_active,\n       COALESCE(SUM(CASE WHEN a.status = 'new' THEN 1 ELSE 0 END), 0) AS new_count,\n       COALESCE(COUNT(a.id), 0) AS total_count\n  FROM careers c\n  LEFT JOIN voxire_career_applications a ON a.job_id = c.id\n WHERE c.owner_id = ? AND c.id = ?\n GROUP BY c.id, c.slug, c.title, c.department, c.location, c.employment_type,\n          c.work_mode, c.seniority, c.details, c.salary_min, c.salary_max,\n          c.salary_currency, c.is_active")
        .bind(params.owner_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetTaskByID` (Go `GetTaskByIDParams`).
pub struct GetTaskByIDParams {
    pub owner_id: i64,
    pub id: i64,
}

/// `GetTaskByID`: `one`.
pub async fn get_task_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetTaskByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT t.id, t.application_id, a.job_id, COALESCE(c.title, '') AS opening_title,\n       a.name AS candidate_name, a.email AS candidate_email,\n       t.title AS title, t.status AS status, t.due_at, t.completed_at, t.created_at AS created_at\nFROM voxhire_follow_up_tasks t\nJOIN voxire_career_applications a ON a.id = t.application_id\nJOIN careers c ON c.id = a.job_id AND c.owner_id = ?\nWHERE t.id = ?")
        .bind(params.owner_id)
        .bind(params.id)
        .fetch_optional(exec)
        .await
}

/// Params for `InsertApplicationEvent` (Go `InsertApplicationEventParams`).
pub struct InsertApplicationEventParams {
    pub application_id: i64,
    pub kind: String,
    pub title: String,
    pub detail: Option<String>,
}

/// `InsertApplicationEvent`: `exec`.
pub async fn insert_application_event<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &InsertApplicationEventParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO voxhire_application_events (application_id, kind, title, detail)\nVALUES (?, ?, ?, ?)")
        .bind(params.application_id)
        .bind(&params.kind)
        .bind(&params.title)
        .bind(&params.detail)
        .execute(exec)
        .await
}

/// `ListAdmins`: `many`.
pub async fn list_admins<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, email, COALESCE(name,'') AS name, created_at\nFROM voxhire_admins ORDER BY created_at DESC LIMIT 100")
        .execute(exec)
        .await
}

/// `ListCompanyLeads`: `many`.
pub async fn list_company_leads<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, company_name, contact_name, COALESCE(contact_title,'') AS contact_title,\n       work_email, COALESCE(phone,'') AS phone, role, COALESCE(headcount,'') AS headcount,\n       COALESCE(timeline,'') AS timeline, COALESCE(service,'') AS service,\n       COALESCE(message,'') AS message, agency_opt_in, status, created_at\nFROM voxhire_company_leads\nORDER BY created_at DESC\nLIMIT 100")
        .execute(exec)
        .await
}

/// Params for `ListOpenings` (Go `ListOpeningsParams`).
pub struct ListOpeningsParams {
    pub owner_id: i64,
}

/// `ListOpenings`: `many`.
pub async fn list_openings<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListOpeningsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT c.id,\n       COALESCE(c.slug, '') AS slug,\n       c.title,\n       COALESCE(c.department, '') AS department,\n       COALESCE(c.location, '') AS location,\n       COALESCE(c.employment_type, '') AS employment_type,\n       COALESCE(c.work_mode, '') AS work_mode,\n       COALESCE(c.seniority, '') AS seniority,\n       COALESCE(c.details, '') AS details,\n       COALESCE(c.salary_min, 0) AS salary_min,\n       COALESCE(c.salary_max, 0) AS salary_max,\n       COALESCE(c.salary_currency, 'USD') AS salary_currency,\n       c.is_active,\n       COALESCE(SUM(CASE WHEN a.status = 'new' THEN 1 ELSE 0 END), 0) AS new_count,\n       COALESCE(COUNT(a.id), 0) AS total_count\n  FROM careers c\n  LEFT JOIN voxire_career_applications a ON a.job_id = c.id\n WHERE c.owner_id = ?\n GROUP BY c.id, c.slug, c.title, c.department, c.location, c.employment_type,\n          c.work_mode, c.seniority, c.details, c.salary_min, c.salary_max,\n          c.salary_currency, c.is_active\n ORDER BY c.is_active DESC, c.created_at DESC")
        .bind(params.owner_id)
        .execute(exec)
        .await
}

/// Params for `ListOpeningsPage` (Go `ListOpeningsPageParams`).
pub struct ListOpeningsPageParams {
    pub owner_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListOpeningsPage`: `many`.
pub async fn list_openings_page<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListOpeningsPageParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("WITH opening_page AS (\n    SELECT c.id,\n           COALESCE(c.slug, '') AS slug,\n           c.title,\n           COALESCE(c.department, '') AS department,\n           COALESCE(c.location, '') AS location,\n           COALESCE(c.employment_type, '') AS employment_type,\n           COALESCE(c.work_mode, '') AS work_mode,\n           COALESCE(c.seniority, '') AS seniority,\n           COALESCE(c.details, '') AS details,\n           COALESCE(c.salary_min, 0) AS salary_min,\n           COALESCE(c.salary_max, 0) AS salary_max,\n           COALESCE(c.salary_currency, 'USD') AS salary_currency,\n           c.is_active,\n           c.created_at\n    FROM careers c\n    WHERE c.owner_id = ?\n    ORDER BY c.is_active DESC, c.created_at DESC, c.id DESC\n    LIMIT ? OFFSET ?\n), application_counts AS (\n    SELECT a.job_id,\n           SUM(a.status = 'new') AS new_count,\n           COUNT(*) AS total_count\n    FROM voxire_career_applications a\n    JOIN opening_page p ON p.id = a.job_id\n    GROUP BY a.job_id\n)\nSELECT p.id, p.slug, p.title, p.department, p.location, p.employment_type,\n       p.work_mode, p.seniority, p.details, p.salary_min, p.salary_max,\n       p.salary_currency, p.is_active,\n       COALESCE(ac.new_count, 0) AS new_count,\n       COALESCE(ac.total_count, 0) AS total_count\nFROM opening_page p\nLEFT JOIN application_counts ac ON ac.job_id = p.id\nORDER BY p.is_active DESC, p.created_at DESC, p.id DESC")
        .bind(params.owner_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListShortlists` (Go `ListShortlistsParams`).
pub struct ListShortlistsParams {
    pub owner_id: i64,
    pub job_id: Option<i64>,
}

/// `ListShortlists`: `many`.
pub async fn list_shortlists<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListShortlistsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT s.id, s.job_id, COALESCE(c.title, '') AS job_title, s.name, s.status,\n       COALESCE(s.notes, '') AS notes, s.created_at, s.updated_at\nFROM voxhire_shortlists s\nJOIN careers c ON c.id = s.job_id AND c.owner_id = ?\nWHERE (? IS NULL OR s.job_id = ?)\nORDER BY s.updated_at DESC\nLIMIT 100")
        .bind(params.owner_id)
        .bind(params.job_id)
        .bind(params.job_id)
        .execute(exec)
        .await
}

/// Params for `ListTaskTimelineEvents` (Go `ListTaskTimelineEventsParams`).
pub struct ListTaskTimelineEventsParams {
    pub owner_id: i64,
    pub application_id: i64,
}

/// `ListTaskTimelineEvents`: `many`.
pub async fn list_task_timeline_events<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListTaskTimelineEventsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT t.title, t.status, t.created_at, t.completed_at\nFROM voxhire_follow_up_tasks t\nJOIN voxire_career_applications a ON a.id = t.application_id\nJOIN careers c ON c.id = a.job_id AND c.owner_id = ?\nWHERE t.application_id = ?")
        .bind(params.owner_id)
        .bind(params.application_id)
        .execute(exec)
        .await
}

/// Params for `ListTasks` (Go `ListTasksParams`).
pub struct ListTasksParams {
    pub owner_id: i64,
    pub filter_status: Option<String>,
}

/// `ListTasks`: `many`.
pub async fn list_tasks<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListTasksParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT t.id, t.application_id, a.job_id, COALESCE(c.title, '') AS opening_title,\n       a.name AS candidate_name, a.email AS candidate_email,\n       t.title AS title, t.status AS status, t.due_at, t.completed_at, t.created_at AS created_at\nFROM voxhire_follow_up_tasks t\nJOIN voxire_career_applications a ON a.id = t.application_id\nJOIN careers c ON c.id = a.job_id AND c.owner_id = ?\nWHERE (? IS NULL OR t.status = ?)\nORDER BY\n  CASE t.status WHEN 'open' THEN 0 WHEN 'done' THEN 2 ELSE 1 END,\n  t.due_at IS NULL,\n  t.due_at ASC,\n  t.created_at DESC\nLIMIT 100")
        .bind(params.owner_id)
        .bind(&params.filter_status)
        .bind(&params.filter_status)
        .execute(exec)
        .await
}

/// Params for `SetCareerActive` (Go `SetCareerActiveParams`).
pub struct SetCareerActiveParams {
    pub is_active: bool,
    pub is_published: bool,
    pub id: i64,
    pub owner_id: i64,
}

/// `SetCareerActive`: `exec`.
pub async fn set_career_active<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SetCareerActiveParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE careers\n   SET is_active = ?,\n       is_published = ?\n WHERE id = ? AND owner_id = ?")
        .bind(params.is_active)
        .bind(params.is_published)
        .bind(params.id)
        .bind(params.owner_id)
        .execute(exec)
        .await
}

/// Params for `UpdateApplicationReview` (Go `UpdateApplicationReviewParams`).
pub struct UpdateApplicationReviewParams {
    pub column1: String,
    pub status: String,
    pub review_rating: i8,
    pub review_notes: Option<String>,
    pub next_step: Option<String>,
    pub id: i64,
    pub owner_id: i64,
}

/// `UpdateApplicationReview`: `exec`.
pub async fn update_application_review<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateApplicationReviewParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE voxire_career_applications\nJOIN careers c ON c.id = voxire_career_applications.job_id\n   SET voxire_career_applications.status = CASE WHEN ? <> '' THEN ? ELSE voxire_career_applications.status END,\n       voxire_career_applications.review_rating = ?,\n       voxire_career_applications.review_notes = ?,\n       voxire_career_applications.next_step = ?,\n       reviewed_at = CURRENT_TIMESTAMP\n WHERE voxire_career_applications.id = ? AND c.owner_id = ?")
        .bind(&params.column1)
        .bind(&params.status)
        .bind(params.review_rating)
        .bind(&params.review_notes)
        .bind(&params.next_step)
        .bind(params.id)
        .bind(params.owner_id)
        .execute(exec)
        .await
}

/// Params for `UpdateApplicationStatus` (Go `UpdateApplicationStatusParams`).
pub struct UpdateApplicationStatusParams {
    pub status: String,
    pub id: i64,
    pub owner_id: i64,
}

/// `UpdateApplicationStatus`: `exec`.
pub async fn update_application_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateApplicationStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE voxire_career_applications a\nJOIN careers c ON c.id = a.job_id\n   SET a.status = ?\n WHERE a.id = ? AND c.owner_id = ?")
        .bind(&params.status)
        .bind(params.id)
        .bind(params.owner_id)
        .execute(exec)
        .await
}

/// Params for `UpdateCareer` (Go `UpdateCareerParams`).
pub struct UpdateCareerParams {
    pub slug: String,
    pub title: String,
    pub department: Option<String>,
    pub location: Option<String>,
    pub employment_type: Option<String>,
    pub work_mode: Option<String>,
    pub seniority: Option<String>,
    pub details: String,
    pub salary_min: Option<String>,
    pub salary_max: Option<String>,
    pub salary_currency: Option<String>,
    pub id: i64,
    pub owner_id: i64,
}

/// `UpdateCareer`: `exec`.
pub async fn update_career<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateCareerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE careers\n   SET slug = ?,\n       title = ?,\n       department = ?,\n       location = ?,\n       employment_type = ?,\n       work_mode = ?,\n       seniority = ?,\n       details = ?,\n       salary_min = ?,\n       salary_max = ?,\n       salary_currency = ?\n WHERE id = ? AND owner_id = ?")
        .bind(&params.slug)
        .bind(&params.title)
        .bind(&params.department)
        .bind(&params.location)
        .bind(&params.employment_type)
        .bind(&params.work_mode)
        .bind(&params.seniority)
        .bind(&params.details)
        .bind(&params.salary_min)
        .bind(&params.salary_max)
        .bind(&params.salary_currency)
        .bind(params.id)
        .bind(params.owner_id)
        .execute(exec)
        .await
}

/// Params for `UpsertFollowUpTask` (Go `UpsertFollowUpTaskParams`).
pub struct UpsertFollowUpTaskParams {
    pub application_id: i64,
    pub title: String,
    pub status: String,
    pub due_at: Option<chrono::DateTime<chrono::Utc>>,
}

/// `UpsertFollowUpTask`: `execresult`.
pub async fn upsert_follow_up_task<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertFollowUpTaskParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO voxhire_follow_up_tasks (application_id, title, status, due_at)\nVALUES (?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n    status = VALUES(status),\n    due_at = VALUES(due_at),\n    completed_at = CASE\n        WHEN VALUES(status) = 'done' THEN COALESCE(completed_at, CURRENT_TIMESTAMP)\n        ELSE NULL\n    END,\n    id = LAST_INSERT_ID(id)")
        .bind(params.application_id)
        .bind(&params.title)
        .bind(&params.status)
        .bind(params.due_at)
        .execute(exec)
        .await
}

/// Params for `UpsertShortlistItem` (Go `UpsertShortlistItemParams`).
pub struct UpsertShortlistItemParams {
    pub shortlist_id: i64,
    pub application_id: i64,
    pub sort_order: i32,
    pub notes: Option<String>,
}

/// `UpsertShortlistItem`: `execresult`.
pub async fn upsert_shortlist_item<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertShortlistItemParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO voxhire_shortlist_items (shortlist_id, application_id, sort_order, notes)\nVALUES (?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n    sort_order = VALUES(sort_order),\n    notes = VALUES(notes),\n    id = LAST_INSERT_ID(id)")
        .bind(params.shortlist_id)
        .bind(params.application_id)
        .bind(params.sort_order)
        .bind(&params.notes)
        .execute(exec)
        .await
}
