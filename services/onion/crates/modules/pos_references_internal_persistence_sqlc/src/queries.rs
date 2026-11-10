//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `LocationExists` (Go `LocationExistsParams`).
pub struct LocationExistsParams {
    pub business_id: i64,
    pub id: i64,
}

/// `LocationExists`: `one`.
pub async fn location_exists<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &LocationExistsParams,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {
    sqlx::query(
        "SELECT EXISTS(SELECT 1 FROM pos_locations WHERE business_id = ? AND id = ?) AS present",
    )
    .bind(params.business_id)
    .bind(params.id)
    .fetch_optional(exec)
    .await
}
