//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateTrainingCourse` (Go `CreateTrainingCourseParams`).
pub struct CreateTrainingCourseParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub description: Option<String>,
    pub duration_hours: Option<String>,
    pub is_required: Option<bool>,
}

/// `CreateTrainingCourse`: `execresult`.
pub async fn create_training_course<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateTrainingCourseParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_training_courses (owner_id, business_id, name, description, duration_hours, is_required)\nVALUES (?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.description)
        .bind(&params.duration_hours)
        .bind(params.is_required)
        .execute(exec)
        .await
}

/// Params for `ListTrainingCourses` (Go `ListTrainingCoursesParams`).
pub struct ListTrainingCoursesParams {
    pub business_id: i64,
}

/// `ListTrainingCourses`: `many`.
pub async fn list_training_courses<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListTrainingCoursesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, description, duration_hours, is_required, created_at\nFROM hr_training_courses\nWHERE business_id = ?\nORDER BY name ASC")
        .bind(params.business_id)
        .execute(exec)
        .await
}
