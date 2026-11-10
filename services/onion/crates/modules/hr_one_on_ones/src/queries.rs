//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateOneOnOne` (Go `CreateOneOnOneParams`).
pub struct CreateOneOnOneParams {
    pub owner_id: i64,
    pub business_id: i64,
    pub manager_id: i64,
    pub staff_id: i64,
    pub scheduled_at: chrono::NaiveDate,
    pub notes: Option<String>,
    pub action_items: String,
    pub status: String,
}

/// `CreateOneOnOne`: `execresult`.
pub async fn create_one_on_one<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateOneOnOneParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO hr_one_on_ones (owner_id, business_id, manager_id, staff_id, scheduled_at, notes, action_items, status)\nVALUES (?, ?, ?, ?, ?, ?, ?, ?)")
        .bind(params.owner_id)
        .bind(params.business_id)
        .bind(params.manager_id)
        .bind(params.staff_id)
        .bind(params.scheduled_at)
        .bind(&params.notes)
        .bind(&params.action_items)
        .bind(&params.status)
        .execute(exec)
        .await
}

/// Params for `ListOneOnOnes` (Go `ListOneOnOnesParams`).
pub struct ListOneOnOnesParams {
    pub business_id: i64,
}

/// `ListOneOnOnes`: `many`.
pub async fn list_one_on_ones<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListOneOnOnesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, owner_id, business_id, manager_id, staff_id, scheduled_at, notes, action_items, status, created_at\nFROM hr_one_on_ones\nWHERE business_id = ?\nORDER BY scheduled_at DESC\nLIMIT 200")
        .bind(params.business_id)
        .execute(exec)
        .await
}
