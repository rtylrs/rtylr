//! Converted from sqlc by `tools/sqlc2sqlx`.
//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.
//! SQL text is byte-identical to the Go sources; migrations
//! stay under `internal/modules/*/sql` (they must not move).

use sqlx::mysql::MySql;

/// Params for `CreateCannedResponse` (Go `CreateCannedResponseParams`).
pub struct CreateCannedResponseParams {
    pub business_id: i64,
    pub name: String,
    pub body: String,
    pub category: Option<String>,
    pub shortcut: Option<String>,
}

/// `CreateCannedResponse`: `execresult`.
pub async fn create_canned_response<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &CreateCannedResponseParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("INSERT INTO support_canned_responses (business_id, name, body, category, shortcut)\nVALUES (?, ?, ?, ?, ?)")
        .bind(params.business_id)
        .bind(&params.name)
        .bind(&params.body)
        .bind(&params.category)
        .bind(&params.shortcut)
        .execute(exec)
        .await
}

/// Params for `DeleteCannedResponse` (Go `DeleteCannedResponseParams`).
pub struct DeleteCannedResponseParams {
    pub business_id: i64,
    pub id: i64,
}

/// `DeleteCannedResponse`: `exec`.
pub async fn delete_canned_response<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &DeleteCannedResponseParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("DELETE FROM support_canned_responses WHERE business_id = ? AND id = ?")
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}

/// Params for `ListCannedResponses` (Go `ListCannedResponsesParams`).
pub struct ListCannedResponsesParams {
    pub business_id: i64,
}

/// `ListCannedResponses`: `many`.
pub async fn list_canned_responses<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &ListCannedResponsesParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("SELECT id, business_id, name, body, category, shortcut, created_at\nFROM support_canned_responses WHERE business_id = ?\nORDER BY name ASC\nLIMIT 200")
        .bind(params.business_id)
        .execute(exec)
        .await
}

/// Params for `UpdateCannedResponse` (Go `UpdateCannedResponseParams`).
pub struct UpdateCannedResponseParams {
    pub name: String,
    pub body: String,
    pub category: Option<String>,
    pub shortcut: Option<String>,
    pub business_id: i64,
    pub id: i64,
}

/// `UpdateCannedResponse`: `exec`.
pub async fn update_canned_response<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &UpdateCannedResponseParams,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {
    sqlx::query("UPDATE support_canned_responses SET name = ?, body = ?, category = ?, shortcut = ?\nWHERE business_id = ? AND id = ?")
        .bind(&params.name)
        .bind(&params.body)
        .bind(&params.category)
        .bind(&params.shortcut)
        .bind(params.business_id)
        .bind(params.id)
        .execute(exec)
        .await
}
