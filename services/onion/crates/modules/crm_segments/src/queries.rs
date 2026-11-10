//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateSegment` (Go `CreateSegmentParams`).
pub struct CreateSegmentParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub name: String,
    pub description: Option<String>,
    pub rules: String,
}

/// `CreateSegment`: `execresult`.
pub async fn create_segment<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateSegmentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO crm_segments (owner_id, business_id, name, description, rules) VALUES (?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.description)
        .bind(&params.rules)
        .execute(exec)
        .await
}

/// Params for `DeleteSegment` (Go `DeleteSegmentParams`).
pub struct DeleteSegmentParams {
    pub id: i64,
    pub business_id: i64,
}

/// `DeleteSegment`: `execresult`.
pub async fn delete_segment<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteSegmentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM crm_segments WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `GetSegmentByID` (Go `GetSegmentByIDParams`).
pub struct GetSegmentByIDParams {
    pub id: i64,
    pub business_id: i64,
}

/// `GetSegmentByID`: `one`.
pub async fn get_segment_by_id<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &GetSegmentByIDParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, description, rules, customer_count, last_calculated_at, created_at, updated_at\nFROM crm_segments WHERE id = ? AND business_id = ?")
        .bind(params.id)
        .bind(params.business_id)
        .fetch_optional(exec)
        .await
}

/// Params for `ListSegments` (Go `ListSegmentsParams`).
pub struct ListSegmentsParams {
    pub business_id: i64,
}

/// `ListSegments`: `many`.
pub async fn list_segments<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListSegmentsParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, name, description, rules, customer_count, last_calculated_at, created_at, updated_at\nFROM crm_segments WHERE business_id = ? ORDER BY created_at DESC LIMIT 100")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateSegment` (Go `UpdateSegmentParams`).
pub struct UpdateSegmentParams {
    pub name: String,
    pub description: Option<String>,
    pub rules: String,
    pub id: i64,
    pub business_id: i64,
}

/// `UpdateSegment`: `execresult`.
pub async fn update_segment<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateSegmentParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE crm_segments SET name = ?, description = ?, rules = ? WHERE id = ? AND business_id = ?")
        .bind(&params.name)
        .bind(&params.description)
        .bind(&params.rules)
        .bind(params.id)
        .bind(params.business_id)
        .execute(exec)
        .await
}
