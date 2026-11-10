//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateCareer` (Go `CreateCareerParams`).
pub struct CreateCareerParams {
    pub title: String,
    pub department: Option<String>,
    pub location: Option<String>,
    pub employment_type: Option<String>,
    pub work_mode: Option<String>,
    pub seniority: Option<String>,
    pub details: String,
    pub is_active: bool,
}

/// `CreateCareer`: `execresult`.
pub async fn create_career<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateCareerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO careers (\n    title,\n    department,\n    location,\n    employment_type,\n    work_mode,\n    seniority,\n    details,\n    is_active\n) VALUES (\n    ?, ?, ?, ?, ?, ?, ?, ?\n)")
        .bind(&params.title)
        .bind(&params.department)
        .bind(&params.location)
        .bind(&params.employment_type)
        .bind(&params.work_mode)
        .bind(&params.seniority)
        .bind(&params.details)
        .bind(params.is_active)
        .execute(exec)
        .await
}

/// Params for `CreateCareerSubmission` (Go `CreateCareerSubmissionParams`).
pub struct CreateCareerSubmissionParams {
    pub career_id: i64,
    pub owner_id: i64,
    pub business_id: i64,
    pub candidate_name: String,
    pub email: String,
    pub phone: Option<String>,
    pub resume_url: Option<String>,
    pub cover_letter: Option<String>,
    pub answers_json: String,
}

/// `CreateCareerSubmission`: `execresult`.
pub async fn create_career_submission<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateCareerSubmissionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO careers_submissions (\n    career_id,\n    owner_id,\n    business_id,\n    candidate_name,\n    email,\n    phone,\n    resume_url,\n    cover_letter,\n    answers_json\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.career_id)
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.candidate_name)
        .bind(&params.email)
        .bind(&params.phone)
        .bind(&params.resume_url)
        .bind(&params.cover_letter)
        .bind(&params.answers_json)
        .execute(exec)
        .await
}

/// Params for `CreateFormField` (Go `CreateFormFieldParams`).
pub struct CreateFormFieldParams {
    pub career_id: i64,
    pub owner_id: i64,
    pub business_id: i64,
    pub kind: String,
    pub label: String,
    pub placeholder: Option<String>,
    pub required: bool,
    pub options_json: String,
    pub position: i32,
}

/// `CreateFormField`: `execresult`.
pub async fn create_form_field<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateFormFieldParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO careers_form_fields (\n    career_id,\n    owner_id,\n    business_id,\n    kind,\n    label,\n    placeholder,\n    required,\n    options_json,\n    position\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.career_id)
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.kind)
        .bind(&params.label)
        .bind(&params.placeholder)
        .bind(params.required)
        .bind(&params.options_json)
        .bind(params.position)
        .execute(exec)
        .await
}

/// Params for `CreateRecruitCareer` (Go `CreateRecruitCareerParams`).
pub struct CreateRecruitCareerParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub slug: String,
    pub title: String,
    pub department: Option<String>,
    pub location: Option<String>,
    pub employment_type: Option<String>,
    pub work_mode: Option<String>,
    pub seniority: Option<String>,
    pub details: String,
    pub requirements: Option<String>,
    pub salary_min: Option<String>,
    pub salary_max: Option<String>,
    pub salary_currency: Option<String>,
    pub is_published: bool,
}

/// `CreateRecruitCareer`: `execresult`.
pub async fn create_recruit_career<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateRecruitCareerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO careers (\n    owner_id,\n    business_id,\n    slug,\n    title,\n    department,\n    location,\n    employment_type,\n    work_mode,\n    seniority,\n    details,\n    requirements,\n    salary_min,\n    salary_max,\n    salary_currency,\n    is_published,\n    is_active\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, true)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.slug)
        .bind(&params.title)
        .bind(&params.department)
        .bind(&params.location)
        .bind(&params.employment_type)
        .bind(&params.work_mode)
        .bind(&params.seniority)
        .bind(&params.details)
        .bind(&params.requirements)
        .bind(&params.salary_min)
        .bind(&params.salary_max)
        .bind(&params.salary_currency)
        .bind(params.is_published)
        .execute(exec)
        .await
}

/// Params for `CreateSection` (Go `CreateSectionParams`).
pub struct CreateSectionParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub kind: String,
    pub title: Option<String>,
    pub body: Option<String>,
    pub media_url: Option<String>,
    pub items_json: String,
    pub position: i32,
    pub enabled: bool,
}

/// `CreateSection`: `execresult`.
pub async fn create_section<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateSectionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO careers_sections (\n    owner_id,\n    business_id,\n    kind,\n    title,\n    body,\n    media_url,\n    items_json,\n    position,\n    enabled\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.kind)
        .bind(&params.title)
        .bind(&params.body)
        .bind(&params.media_url)
        .bind(&params.items_json)
        .bind(params.position)
        .bind(params.enabled)
        .execute(exec)
        .await
}

/// Params for `DeleteFormField` (Go `DeleteFormFieldParams`).
pub struct DeleteFormFieldParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteFormField`: `execresult`.
pub async fn delete_form_field<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteFormFieldParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM careers_form_fields\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `DeleteSection` (Go `DeleteSectionParams`).
pub struct DeleteSectionParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteSection`: `execresult`.
pub async fn delete_section<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteSectionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM careers_sections\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// `GetAllCareers`: `many`.
pub async fn get_all_careers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    id,\n    title,\n    department,\n    location,\n    employment_type,\n    work_mode,\n    seniority,\n    details,\n    is_active\n FROM careers\nWHERE is_active = true\nORDER BY created_at DESC\nLIMIT 100")
        .execute(exec)
        .await
}

/// Params for `GetCareerSubmissionResumeURL` (Go `GetCareerSubmissionResumeURLParams`).
pub struct GetCareerSubmissionResumeURLParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetCareerSubmissionResumeURL`: `one`.
pub async fn get_career_submission_resume_url<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetCareerSubmissionResumeURLParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT COALESCE(resume_url, '') AS resume_url\nFROM careers_submissions\nWHERE id = ? AND business_id = ?\nLIMIT 1")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetCustomizationByBusiness` (Go `GetCustomizationByBusinessParams`).
pub struct GetCustomizationByBusinessParams {
    pub business_id: i64,
}

/// `GetCustomizationByBusiness`: `one`.
pub async fn get_customization_by_business<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetCustomizationByBusinessParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    id,\n    owner_id,\n    business_id,\n    slug,\n    brand_name,\n    COALESCE(tagline, '') AS tagline,\n    COALESCE(about_blurb, '') AS about_blurb,\n    COALESCE(logo_url, '') AS logo_url,\n    COALESCE(primary_color, '') AS primary_color,\n    COALESCE(accent_color, '') AS accent_color,\n    COALESCE(bg_color, '') AS bg_color,\n    COALESCE(text_color, '') AS text_color,\n    COALESCE(muted_text_color, '') AS muted_text_color,\n    COALESCE(font_pairing, '') AS font_pairing,\n    COALESCE(hero_image_url, '') AS hero_image_url,\n    COALESCE(favicon_url, '') AS favicon_url,\n    template,\n    COALESCE(header_html, '') AS header_html,\n    COALESCE(meta_title, '') AS meta_title,\n    COALESCE(meta_description, '') AS meta_description,\n    COALESCE(privacy_url, '') AS privacy_url,\n    COALESCE(terms_url, '') AS terms_url,\n    COALESCE(ga_measurement_id, '') AS ga_measurement_id,\n    COALESCE(meta_pixel_id, '') AS meta_pixel_id,\n    require_consent,\n    COALESCE(notification_email, '') AS notification_email,\n    email_notifications,\n    created_at,\n    updated_at\nFROM careers_customizations\nWHERE business_id = ?\nORDER BY business_id ASC\nLIMIT 1")
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetCustomizationBySlug` (Go `GetCustomizationBySlugParams`).
pub struct GetCustomizationBySlugParams {
    pub slug: String,
}

/// `GetCustomizationBySlug`: `one`.
pub async fn get_customization_by_slug<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetCustomizationBySlugParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    id,\n    owner_id,\n    business_id,\n    slug,\n    brand_name,\n    COALESCE(tagline, '') AS tagline,\n    COALESCE(about_blurb, '') AS about_blurb,\n    COALESCE(logo_url, '') AS logo_url,\n    COALESCE(primary_color, '') AS primary_color,\n    COALESCE(accent_color, '') AS accent_color,\n    COALESCE(bg_color, '') AS bg_color,\n    COALESCE(text_color, '') AS text_color,\n    COALESCE(muted_text_color, '') AS muted_text_color,\n    COALESCE(font_pairing, '') AS font_pairing,\n    COALESCE(hero_image_url, '') AS hero_image_url,\n    COALESCE(favicon_url, '') AS favicon_url,\n    template,\n    COALESCE(header_html, '') AS header_html,\n    COALESCE(meta_title, '') AS meta_title,\n    COALESCE(meta_description, '') AS meta_description,\n    COALESCE(privacy_url, '') AS privacy_url,\n    COALESCE(terms_url, '') AS terms_url,\n    COALESCE(ga_measurement_id, '') AS ga_measurement_id,\n    COALESCE(meta_pixel_id, '') AS meta_pixel_id,\n    require_consent,\n    COALESCE(notification_email, '') AS notification_email,\n    email_notifications,\n    created_at,\n    updated_at\nFROM careers_customizations\nWHERE slug = ?\nORDER BY slug ASC\nLIMIT 1")
        .bind(&params.slug)
        .fetch_optional(exec)
        .await
}

/// Params for `GetFormField` (Go `GetFormFieldParams`).
pub struct GetFormFieldParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetFormField`: `one`.
pub async fn get_form_field<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetFormFieldParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    id,\n    career_id,\n    owner_id,\n    business_id,\n    kind,\n    label,\n    COALESCE(placeholder, '') AS placeholder,\n    required,\n    COALESCE(CAST(options_json AS CHAR), '') AS options_json,\n    position,\n    created_at,\n    updated_at\nFROM careers_form_fields\nWHERE id = ? AND business_id = ?\nORDER BY id ASC\nLIMIT 1")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetRecruitCareerByID` (Go `GetRecruitCareerByIDParams`).
pub struct GetRecruitCareerByIDParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetRecruitCareerByID`: `one`.
pub async fn get_recruit_career_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetRecruitCareerByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    id,\n    owner_id,\n    business_id,\n    COALESCE(slug, '') AS slug,\n    title,\n    COALESCE(department, '') AS department,\n    COALESCE(location, '') AS location,\n    COALESCE(employment_type, '') AS employment_type,\n    COALESCE(work_mode, '') AS work_mode,\n    COALESCE(seniority, '') AS seniority,\n    details,\n    COALESCE(requirements, '') AS requirements,\n    COALESCE(salary_min, 0) AS salary_min,\n    COALESCE(salary_max, 0) AS salary_max,\n    COALESCE(salary_currency, '') AS salary_currency,\n    COALESCE(is_published, false) AS is_published,\n    is_active,\n    created_at,\n    updated_at\nFROM careers\nWHERE id = ? AND business_id = ?\nLIMIT 1")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `GetRecruitCareerBySlug` (Go `GetRecruitCareerBySlugParams`).
pub struct GetRecruitCareerBySlugParams {
    pub business_id: i64,
    pub slug: String,
}

/// `GetRecruitCareerBySlug`: `one`.
pub async fn get_recruit_career_by_slug<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetRecruitCareerBySlugParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    id,\n    owner_id,\n    business_id,\n    COALESCE(slug, '') AS slug,\n    title,\n    COALESCE(department, '') AS department,\n    COALESCE(location, '') AS location,\n    COALESCE(employment_type, '') AS employment_type,\n    COALESCE(work_mode, '') AS work_mode,\n    COALESCE(seniority, '') AS seniority,\n    details,\n    COALESCE(requirements, '') AS requirements,\n    COALESCE(salary_min, 0) AS salary_min,\n    COALESCE(salary_max, 0) AS salary_max,\n    COALESCE(salary_currency, '') AS salary_currency,\n    COALESCE(is_published, false) AS is_published,\n    is_active,\n    created_at,\n    updated_at\nFROM careers\nWHERE business_id = ? AND slug = ?\nORDER BY slug ASC\nLIMIT 1")
        .bind(params.business_id)
        .bind(&params.slug)
        .fetch_optional(exec)
        .await
}

/// Params for `GetSection` (Go `GetSectionParams`).
pub struct GetSectionParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetSection`: `one`.
pub async fn get_section<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetSectionParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT\n    id,\n    owner_id,\n    business_id,\n    kind,\n    COALESCE(title, '') AS title,\n    COALESCE(body, '') AS body,\n    COALESCE(media_url, '') AS media_url,\n    COALESCE(CAST(items_json AS CHAR), '') AS items_json,\n    position,\n    enabled,\n    created_at,\n    updated_at\nFROM careers_sections\nWHERE id = ? AND business_id = ?\nORDER BY id ASC\nLIMIT 1")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListCareerSubmissions` (Go `ListCareerSubmissionsParams`).
pub struct ListCareerSubmissionsParams {
    pub business_id: i64,
}

/// `ListCareerSubmissions`: `many`.
pub async fn list_career_submissions<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListCareerSubmissionsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    s.id,\n    s.career_id,\n    s.owner_id,\n    s.business_id,\n    s.candidate_name,\n    s.email,\n    COALESCE(s.phone, '') AS phone,\n    COALESCE(s.resume_url, '') AS resume_url,\n    COALESCE(s.cover_letter, '') AS cover_letter,\n    s.status,\n    COALESCE(CAST(s.answers_json AS CHAR), '') AS answers_json,\n    s.created_at,\n    s.updated_at,\n    COALESCE(c.title, '') AS career_title\nFROM careers_submissions s\nLEFT JOIN careers c ON c.id = s.career_id\nWHERE s.business_id = ?\nORDER BY s.created_at DESC\nLIMIT 100")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListCareerSubmissionsByStatus` (Go `ListCareerSubmissionsByStatusParams`).
pub struct ListCareerSubmissionsByStatusParams {
    pub business_id: i64,
    pub status: String,
}

/// `ListCareerSubmissionsByStatus`: `many`.
pub async fn list_career_submissions_by_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListCareerSubmissionsByStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    s.id,\n    s.career_id,\n    s.owner_id,\n    s.business_id,\n    s.candidate_name,\n    s.email,\n    COALESCE(s.phone, '') AS phone,\n    COALESCE(s.resume_url, '') AS resume_url,\n    COALESCE(s.cover_letter, '') AS cover_letter,\n    s.status,\n    COALESCE(CAST(s.answers_json AS CHAR), '') AS answers_json,\n    s.created_at,\n    s.updated_at,\n    COALESCE(c.title, '') AS career_title\nFROM careers_submissions s\nLEFT JOIN careers c ON c.id = s.career_id\nWHERE s.business_id = ? AND s.status = ?\nORDER BY s.created_at DESC\nLIMIT 100")
        .bind(params.business_id)
        .bind(&params.status)
        .execute(exec)
        .await
}

/// Params for `ListCareerSubmissionsByStatusPage` (Go `ListCareerSubmissionsByStatusPageParams`).
pub struct ListCareerSubmissionsByStatusPageParams {
    pub business_id: i64,
    pub status: String,
    pub limit: i32,
    pub offset: i32,
}

/// `ListCareerSubmissionsByStatusPage`: `many`.
pub async fn list_career_submissions_by_status_page<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListCareerSubmissionsByStatusPageParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    s.id,\n    s.career_id,\n    s.owner_id,\n    s.business_id,\n    s.candidate_name,\n    s.email,\n    COALESCE(s.phone, '') AS phone,\n    COALESCE(s.resume_url, '') AS resume_url,\n    COALESCE(s.cover_letter, '') AS cover_letter,\n    s.status,\n    COALESCE(CAST(s.answers_json AS CHAR), '') AS answers_json,\n    s.created_at,\n    s.updated_at,\n    COALESCE(c.title, '') AS career_title\nFROM careers_submissions s\nLEFT JOIN careers c ON c.id = s.career_id\nWHERE s.business_id = ? AND s.status = ?\nORDER BY s.created_at DESC, s.id DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(&params.status)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListCareerSubmissionsPage` (Go `ListCareerSubmissionsPageParams`).
pub struct ListCareerSubmissionsPageParams {
    pub business_id: i64,
    pub limit: i32,
    pub offset: i32,
}

/// `ListCareerSubmissionsPage`: `many`.
pub async fn list_career_submissions_page<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListCareerSubmissionsPageParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    s.id,\n    s.career_id,\n    s.owner_id,\n    s.business_id,\n    s.candidate_name,\n    s.email,\n    COALESCE(s.phone, '') AS phone,\n    COALESCE(s.resume_url, '') AS resume_url,\n    COALESCE(s.cover_letter, '') AS cover_letter,\n    s.status,\n    COALESCE(CAST(s.answers_json AS CHAR), '') AS answers_json,\n    s.created_at,\n    s.updated_at,\n    COALESCE(c.title, '') AS career_title\nFROM careers_submissions s\nLEFT JOIN careers c ON c.id = s.career_id\nWHERE s.business_id = ?\nORDER BY s.created_at DESC, s.id DESC\nLIMIT ? OFFSET ?")
        .bind(params.business_id)
        .bind(params.limit)
        .bind(params.offset)
        .execute(exec)
        .await
}

/// Params for `ListEnabledSections` (Go `ListEnabledSectionsParams`).
pub struct ListEnabledSectionsParams {
    pub business_id: i64,
}

/// `ListEnabledSections`: `many`.
pub async fn list_enabled_sections<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListEnabledSectionsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    id,\n    owner_id,\n    business_id,\n    kind,\n    COALESCE(title, '') AS title,\n    COALESCE(body, '') AS body,\n    COALESCE(media_url, '') AS media_url,\n    COALESCE(CAST(items_json AS CHAR), '') AS items_json,\n    position,\n    enabled,\n    created_at,\n    updated_at\nFROM careers_sections\nWHERE business_id = ? AND enabled = true\nORDER BY position ASC, id ASC\nLIMIT 100")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListFormFields` (Go `ListFormFieldsParams`).
pub struct ListFormFieldsParams {
    pub business_id: i64,
    pub career_id: i64,
}

/// `ListFormFields`: `many`.
pub async fn list_form_fields<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListFormFieldsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    id,\n    career_id,\n    owner_id,\n    business_id,\n    kind,\n    label,\n    COALESCE(placeholder, '') AS placeholder,\n    required,\n    COALESCE(CAST(options_json AS CHAR), '') AS options_json,\n    position,\n    created_at,\n    updated_at\nFROM careers_form_fields\nWHERE business_id = ? AND career_id = ?\nORDER BY position ASC, id ASC\nLIMIT 100")
        .bind(params.business_id)
        .bind(params.career_id)
        .execute(exec)
        .await
}

/// Params for `ListFormFieldsByCareer` (Go `ListFormFieldsByCareerParams`).
pub struct ListFormFieldsByCareerParams {
    pub career_id: i64,
}

/// `ListFormFieldsByCareer`: `many`.
pub async fn list_form_fields_by_career<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListFormFieldsByCareerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    id,\n    career_id,\n    owner_id,\n    business_id,\n    kind,\n    label,\n    COALESCE(placeholder, '') AS placeholder,\n    required,\n    COALESCE(CAST(options_json AS CHAR), '') AS options_json,\n    position,\n    created_at,\n    updated_at\nFROM careers_form_fields\nWHERE career_id = ?\nORDER BY position ASC, id ASC\nLIMIT 100")
        .bind(params.career_id)
        .execute(exec)
        .await
}

/// Params for `ListRecruitCareers` (Go `ListRecruitCareersParams`).
pub struct ListRecruitCareersParams {
    pub business_id: i64,
}

/// `ListRecruitCareers`: `many`.
pub async fn list_recruit_careers<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListRecruitCareersParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    id,\n    owner_id,\n    business_id,\n    COALESCE(slug, '') AS slug,\n    title,\n    COALESCE(department, '') AS department,\n    COALESCE(location, '') AS location,\n    COALESCE(employment_type, '') AS employment_type,\n    COALESCE(work_mode, '') AS work_mode,\n    COALESCE(seniority, '') AS seniority,\n    details,\n    COALESCE(requirements, '') AS requirements,\n    COALESCE(salary_min, 0) AS salary_min,\n    COALESCE(salary_max, 0) AS salary_max,\n    COALESCE(salary_currency, '') AS salary_currency,\n    COALESCE(is_published, false) AS is_published,\n    is_active,\n    created_at,\n    updated_at\nFROM careers\nWHERE business_id = ? AND is_active = true\nORDER BY created_at DESC")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListRecruitCareersPublished` (Go `ListRecruitCareersPublishedParams`).
pub struct ListRecruitCareersPublishedParams {
    pub business_id: i64,
}

/// `ListRecruitCareersPublished`: `many`.
pub async fn list_recruit_careers_published<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListRecruitCareersPublishedParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    id,\n    owner_id,\n    business_id,\n    COALESCE(slug, '') AS slug,\n    title,\n    COALESCE(department, '') AS department,\n    COALESCE(location, '') AS location,\n    COALESCE(employment_type, '') AS employment_type,\n    COALESCE(work_mode, '') AS work_mode,\n    COALESCE(seniority, '') AS seniority,\n    details,\n    COALESCE(requirements, '') AS requirements,\n    COALESCE(salary_min, 0) AS salary_min,\n    COALESCE(salary_max, 0) AS salary_max,\n    COALESCE(salary_currency, '') AS salary_currency,\n    COALESCE(is_published, false) AS is_published,\n    is_active,\n    created_at,\n    updated_at\nFROM careers\nWHERE business_id = ? AND is_active = true AND is_published = true\nORDER BY created_at DESC\nLIMIT 100")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `ListSections` (Go `ListSectionsParams`).
pub struct ListSectionsParams {
    pub business_id: i64,
}

/// `ListSections`: `many`.
pub async fn list_sections<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListSectionsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT\n    id,\n    owner_id,\n    business_id,\n    kind,\n    COALESCE(title, '') AS title,\n    COALESCE(body, '') AS body,\n    COALESCE(media_url, '') AS media_url,\n    COALESCE(CAST(items_json AS CHAR), '') AS items_json,\n    position,\n    enabled,\n    created_at,\n    updated_at\nFROM careers_sections\nWHERE business_id = ?\nORDER BY position ASC, id ASC\nLIMIT 100")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `SaveCareerApplication` (Go `SaveCareerApplicationParams`).
pub struct SaveCareerApplicationParams {
    pub job_id: i64,
    pub name: String,
    pub email: String,
    pub phone: Option<String>,
    pub work_history: Option<String>,
    pub education: Option<String>,
    pub experience_summary: Option<String>,
    pub cover_letter: Option<String>,
    pub cv_key: Option<String>,
}

/// `SaveCareerApplication`: `execresult`.
pub async fn save_career_application<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SaveCareerApplicationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO voxire_career_applications\n    (job_id, name, email, phone, work_history, education, experience_summary, cover_letter, cv_key)\nVALUES\n    (?, ?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.job_id)
        .bind(&params.name)
        .bind(&params.email)
        .bind(&params.phone)
        .bind(&params.work_history)
        .bind(&params.education)
        .bind(&params.experience_summary)
        .bind(&params.cover_letter)
        .bind(&params.cv_key)
        .execute(exec)
        .await
}

/// Params for `SoftDeleteRecruitCareer` (Go `SoftDeleteRecruitCareerParams`).
pub struct SoftDeleteRecruitCareerParams {
    pub id: i64,
    pub business_id: i64,
}

/// `SoftDeleteRecruitCareer`: `execresult`.
pub async fn soft_delete_recruit_career<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &SoftDeleteRecruitCareerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE careers\nSET is_active = false, is_published = false\nWHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateCareerSubmissionStatus` (Go `UpdateCareerSubmissionStatusParams`).
pub struct UpdateCareerSubmissionStatusParams {
    pub status: String,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateCareerSubmissionStatus`: `execresult`.
pub async fn update_career_submission_status<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateCareerSubmissionStatusParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE careers_submissions\nSET status = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.status)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateFormField` (Go `UpdateFormFieldParams`).
pub struct UpdateFormFieldParams {
    pub kind: String,
    pub label: String,
    pub placeholder: Option<String>,
    pub required: bool,
    pub options_json: String,
    pub position: i32,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateFormField`: `exec`.
pub async fn update_form_field<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateFormFieldParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE careers_form_fields\nSET\n    kind = ?,\n    label = ?,\n    placeholder = ?,\n    required = ?,\n    options_json = ?,\n    position = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.kind)
        .bind(&params.label)
        .bind(&params.placeholder)
        .bind(params.required)
        .bind(&params.options_json)
        .bind(params.position)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateRecruitCareer` (Go `UpdateRecruitCareerParams`).
pub struct UpdateRecruitCareerParams {
    pub slug: String,
    pub title: String,
    pub department: Option<String>,
    pub location: Option<String>,
    pub employment_type: Option<String>,
    pub work_mode: Option<String>,
    pub seniority: Option<String>,
    pub details: String,
    pub requirements: Option<String>,
    pub salary_min: Option<String>,
    pub salary_max: Option<String>,
    pub salary_currency: Option<String>,
    pub is_published: bool,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateRecruitCareer`: `execresult`.
pub async fn update_recruit_career<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateRecruitCareerParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE careers\nSET\n    slug = ?,\n    title = ?,\n    department = ?,\n    location = ?,\n    employment_type = ?,\n    work_mode = ?,\n    seniority = ?,\n    details = ?,\n    requirements = ?,\n    salary_min = ?,\n    salary_max = ?,\n    salary_currency = ?,\n    is_published = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.slug)
        .bind(&params.title)
        .bind(&params.department)
        .bind(&params.location)
        .bind(&params.employment_type)
        .bind(&params.work_mode)
        .bind(&params.seniority)
        .bind(&params.details)
        .bind(&params.requirements)
        .bind(&params.salary_min)
        .bind(&params.salary_max)
        .bind(&params.salary_currency)
        .bind(params.is_published)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateSection` (Go `UpdateSectionParams`).
pub struct UpdateSectionParams {
    pub kind: String,
    pub title: Option<String>,
    pub body: Option<String>,
    pub media_url: Option<String>,
    pub items_json: String,
    pub position: i32,
    pub enabled: bool,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateSection`: `exec`.
pub async fn update_section<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateSectionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE careers_sections\nSET\n    kind = ?,\n    title = ?,\n    body = ?,\n    media_url = ?,\n    items_json = ?,\n    position = ?,\n    enabled = ?\nWHERE id = ? AND business_id = ?")
        .bind(&params.kind)
        .bind(&params.title)
        .bind(&params.body)
        .bind(&params.media_url)
        .bind(&params.items_json)
        .bind(params.position)
        .bind(params.enabled)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateSectionPosition` (Go `UpdateSectionPositionParams`).
pub struct UpdateSectionPositionParams {
    pub position: i32,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateSectionPosition`: `exec`.
pub async fn update_section_position<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateSectionPositionParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE careers_sections\nSET position = ?\nWHERE id = ? AND business_id = ?")
        .bind(params.position)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpsertCustomization` (Go `UpsertCustomizationParams`).
pub struct UpsertCustomizationParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub slug: String,
    pub brand_name: String,
    pub tagline: Option<String>,
    pub about_blurb: Option<String>,
    pub logo_url: Option<String>,
    pub primary_color: Option<String>,
    pub accent_color: Option<String>,
    pub bg_color: Option<String>,
    pub text_color: Option<String>,
    pub muted_text_color: Option<String>,
    pub font_pairing: Option<String>,
    pub hero_image_url: Option<String>,
    pub favicon_url: Option<String>,
    pub template: String,
    pub header_html: Option<String>,
    pub meta_title: Option<String>,
    pub meta_description: Option<String>,
    pub privacy_url: Option<String>,
    pub terms_url: Option<String>,
    pub ga_measurement_id: Option<String>,
    pub meta_pixel_id: Option<String>,
    pub require_consent: bool,
    pub notification_email: Option<String>,
    pub email_notifications: bool,
}

/// `UpsertCustomization`: `exec`.
pub async fn upsert_customization<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpsertCustomizationParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO careers_customizations (\n    owner_id,\n    business_id,\n    slug,\n    brand_name,\n    tagline,\n    about_blurb,\n    logo_url,\n    primary_color,\n    accent_color,\n    bg_color,\n    text_color,\n    muted_text_color,\n    font_pairing,\n    hero_image_url,\n    favicon_url,\n    template,\n    header_html,\n    meta_title,\n    meta_description,\n    privacy_url,\n    terms_url,\n    ga_measurement_id,\n    meta_pixel_id,\n    require_consent,\n    notification_email,\n    email_notifications\n) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)\nON DUPLICATE KEY UPDATE\n    slug = VALUES(slug),\n    brand_name = VALUES(brand_name),\n    tagline = VALUES(tagline),\n    about_blurb = VALUES(about_blurb),\n    logo_url = VALUES(logo_url),\n    primary_color = VALUES(primary_color),\n    accent_color = VALUES(accent_color),\n    bg_color = VALUES(bg_color),\n    text_color = VALUES(text_color),\n    muted_text_color = VALUES(muted_text_color),\n    font_pairing = VALUES(font_pairing),\n    hero_image_url = VALUES(hero_image_url),\n    favicon_url = VALUES(favicon_url),\n    template = VALUES(template),\n    header_html = VALUES(header_html),\n    meta_title = VALUES(meta_title),\n    meta_description = VALUES(meta_description),\n    privacy_url = VALUES(privacy_url),\n    terms_url = VALUES(terms_url),\n    ga_measurement_id = VALUES(ga_measurement_id),\n    meta_pixel_id = VALUES(meta_pixel_id),\n    require_consent = VALUES(require_consent),\n    notification_email = VALUES(notification_email),\n    email_notifications = VALUES(email_notifications)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.slug)
        .bind(&params.brand_name)
        .bind(&params.tagline)
        .bind(&params.about_blurb)
        .bind(&params.logo_url)
        .bind(&params.primary_color)
        .bind(&params.accent_color)
        .bind(&params.bg_color)
        .bind(&params.text_color)
        .bind(&params.muted_text_color)
        .bind(&params.font_pairing)
        .bind(&params.hero_image_url)
        .bind(&params.favicon_url)
        .bind(&params.template)
        .bind(&params.header_html)
        .bind(&params.meta_title)
        .bind(&params.meta_description)
        .bind(&params.privacy_url)
        .bind(&params.terms_url)
        .bind(&params.ga_measurement_id)
        .bind(&params.meta_pixel_id)
        .bind(params.require_consent)
        .bind(&params.notification_email)
        .bind(params.email_notifications)
        .execute(exec)
        .await
}
