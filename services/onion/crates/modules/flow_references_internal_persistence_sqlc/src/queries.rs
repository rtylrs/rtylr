//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `ProjectExists` (Go `ProjectExistsParams`).
pub struct ProjectExistsParams {
    pub business_id: i64,
    pub id: i64,
}

/// `ProjectExists`: `one`.
pub async fn project_exists<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ProjectExistsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT EXISTS(SELECT 1 FROM flow_projects WHERE business_id = ? AND id = ?) AS present",
    )
    .bind(params.business_id)
    .bind(params.id)
    .fetch_optional(exec)
    .await
}
